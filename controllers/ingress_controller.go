/*


Licensed under the Mozilla Public License (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.mozilla.org/en-US/MPL/2.0/

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/pkg/universal_client"
	"github.com/go-logr/logr"
	"k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	util "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	ingressLabelKey                    = "tyk.io/ingress"
	apidefLabelKey                     = "tyk.io/apidefinition"
	ingressFinalizerName               = "finalizers.tyk.io/ingress"
	ingressClassAnnotationKey          = "kubernetes.io/ingress.class"
	ingressTemplateAnnotationKey       = "tyk.io/template"
	defaultIngressClassAnnotationValue = "tyk"
)

// CertificateReconciler reconciles a CertificateSecret object
type IngressReconciler struct {
	client.Client
	Log             logr.Logger
	Scheme          *runtime.Scheme
	UniversalClient universal_client.UniversalClient
	Recorder        record.EventRecorder
}

// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch

func (r *IngressReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	desired := &v1beta1.Ingress{}
	if err := r.Get(ctx, req.NamespacedName, desired); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	nsl := r.Log.WithValues("name", req.NamespacedName)
	nsl.Info("updating  ingress object")
	op, err := util.CreateOrUpdate(ctx, r.Client, desired, func() error {
		if !util.ContainsFinalizer(desired, ingressFinalizerName) {
			nsl.Info("adding ingress finalizer")
			util.RemoveFinalizer(desired, ingressFinalizerName)
			return nil
		}
		if !desired.ObjectMeta.DeletionTimestamp.IsZero() {
			nsl.Info("deleting ingress resource")
			if util.ContainsFinalizer(desired, ingressFinalizerName) {
				nsl.Info("deleting api's")
				err := r.deleteAPIAll(ctx, nsl, req.Namespace, desired)
				if err != nil {
					nsl.Info("failed to delete api's")
					return err
				}
				util.RemoveFinalizer(desired, ingressFinalizerName)
				nsl.Info("successful deleted api's")
			}
			nsl.Info("successful deleted  ingress resource")
			return nil
		}
		return nil
	})
	if err != nil {
		nsl.Error(err, "Failed to update ingress object", "Op", op)
		return ctrl.Result{}, err
	}
	nsl.Info("creating api defintiions")
	err = r.createAPI(ctx, nsl, req.Namespace, desired)
	if err != nil {
		nsl.Error(err, "Failed to create api's")
		return ctrl.Result{}, err
	}
	nsl.Info("successful created api's")
	return ctrl.Result{}, nil
}

func (r *IngressReconciler) createAPI(ctx context.Context, lg logr.Logger, ns string, desired *v1beta1.Ingress) error {
	key, ok := desired.Annotations[ingressTemplateAnnotationKey]
	if !ok {
		return fmt.Errorf("expecting template annotation %s", ingressTemplateAnnotationKey)
	}
	template := &v1alpha1.ApiDefinition{}
	err := r.Get(ctx, types.NamespacedName{Name: key, Namespace: ns}, template)
	if err != nil {
		return err
	}
	for _, rule := range desired.Spec.Rules {
		for _, p := range rule.HTTP.Paths {
			hash := shortHash(rule.Host + p.Path)
			name := r.buildAPIName(ns, desired.Name, hash)
			api := &v1alpha1.ApiDefinition{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: ns,
				},
			}
			lg.Info("sync api definition", "name", name)
			op, err := util.CreateOrUpdate(ctx, r.Client, api, func() error {
				api.SetLabels(map[string]string{
					ingressLabelKey: desired.Name,
					apidefLabelKey:  hash,
				})
				api.Spec = *template.Spec.DeepCopy()
				api.Spec.Name = name
				api.Spec.Proxy.ListenPath = p.Path
				api.Spec.Proxy.TargetURL = fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", p.Backend.ServiceName,
					ns, p.Backend.ServicePort.IntValue())
				api.Spec.Domain = rule.Host
				if !strings.Contains(p.Path, ".well-known/acme-challenge") && !strings.Contains(p.Backend.ServiceName, "cm-acme-http-solver") {
					for _, tls := range desired.Spec.TLS {
						for _, host := range tls.Hosts {
							if rule.Host == host {
								api.Spec.Protocol = "https"
								api.Spec.CertificateSecretNames = []string{
									tls.SecretName,
								}
								api.Spec.ListenPort = 443
							}
						}
					}
				} else {
					// for the acme challenge
					api.Spec.Proxy.StripListenPath = false
					api.Spec.Proxy.PreserveHostHeader = true
				}
				return util.SetControllerReference(desired, api, r.Scheme)
			})
			if err != nil {
				lg.Error(err, "failed to sync api definition", "name", name, "op", op)
				return nil
			}
			lg.Info("successful sync api defintion", "name", name, "op", op)
		}
	}
	lg.Info("deleting orphan api's")
	return r.deleteOrphanAPI(ctx, lg, ns, desired)
}

func (r *IngressReconciler) deleteAPIAll(ctx context.Context, lg logr.Logger, ns string, desired *v1beta1.Ingress) error {
	var keys []string
	for _, rule := range desired.Spec.Rules {
		for _, p := range rule.HTTP.Paths {
			hash := shortHash(rule.Host + p.Path)
			keys = append(keys, hash)
		}
	}
	s := labels.NewSelector()
	exists, err := labels.NewRequirement(apidefLabelKey, selection.Exists, keys)
	if err != nil {
		return err
	}
	s = s.Add(*exists)
	fs := fields.OneTermEqualSelector(".metadata.ownerReferences[*].name", desired.GetName())
	lg.Info("deleting all api's", "fieldSelector", fs, "count", len(keys))
	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		return r.DeleteAllOf(ctx, &v1alpha1.ApiDefinition{}, &client.DeleteAllOfOptions{
			ListOptions: client.ListOptions{
				LabelSelector: s,
				Namespace:     ns,
				FieldSelector: fs,
			},
			DeleteOptions: client.DeleteOptions{},
		})
	})
}

func (r *IngressReconciler) deleteOrphanAPI(ctx context.Context, lg logr.Logger, ns string, desired *v1beta1.Ingress) error {
	var keys []string
	for _, rule := range desired.Spec.Rules {
		for _, p := range rule.HTTP.Paths {
			hash := shortHash(rule.Host + p.Path)
			keys = append(keys, hash)
		}
	}
	s := labels.NewSelector()
	exists, err := labels.NewRequirement(apidefLabelKey, selection.Exists, []string{})
	if err != nil {
		return err
	}
	s = s.Add(*exists)
	notIn, err := labels.NewRequirement(apidefLabelKey, selection.NotIn, keys)
	if err != nil {
		return err
	}
	s = s.Add(*notIn)
	lg.Info("deleting orphan api definitions", "selector", s, "count", len(keys))
	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		return r.DeleteAllOf(ctx, &v1alpha1.ApiDefinition{}, &client.DeleteAllOfOptions{
			ListOptions: client.ListOptions{
				LabelSelector: s,
				Namespace:     ns,
				FieldSelector: fields.OneTermEqualSelector(".metadata.ownerReferences[*].name", desired.GetName()),
			},
			DeleteOptions: client.DeleteOptions{},
		})
	})
}

func (r *IngressReconciler) buildAPIName(nameSpace, name, hash string) string {
	return fmt.Sprintf("%s-%s-%s", nameSpace, name, hash)
}

func shortHash(txt string) string {
	h := sha256.New()
	h.Write([]byte(txt))
	return fmt.Sprintf("%x", h.Sum(nil))[:9]
}

func (r *IngressReconciler) ingressClassEventFilter() predicate.Predicate {
	isOurIngress := func(annotations map[string]string) bool {
		if ingressClass, ok := annotations[ingressClassAnnotationKey]; !ok {
			return false
		} else if ingressClass == defaultIngressClassAnnotationValue {
			// if the ingress class is `tyk` it's for us
			return true
		}
		// TODO: env var?

		return false
	}

	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return isOurIngress(e.Meta.GetAnnotations())
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return isOurIngress(e.MetaNew.GetAnnotations())
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return isOurIngress(e.Meta.GetAnnotations())
		},
	}
}

func (r *IngressReconciler) SetupWithManager(mgr ctrl.Manager) error {
	//err := mgr.GetFieldIndexer().
	//	IndexField(context.TODO(), &v1alpha1.ApiDefinition{}, apiOwnerKey, func(rawObj runtime.Object) []string {
	//		// grab the apiDef object, extract the owner...
	//		apiDefinition := rawObj.(*v1alpha1.ApiDefinition)
	//		owner := metav1.GetControllerOf(apiDefinition)
	//		if owner == nil {
	//			return nil
	//		}
	//
	//		if owner.APIVersion != ingressGVString || owner.Kind != "Ingress" {
	//			return nil
	//		}
	//
	//		return []string{owner.Name}
	//	})

	//if err != nil {
	//	return err
	//}

	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.Ingress{}).
		Owns(&v1alpha1.ApiDefinition{}).
		WithEventFilter(r.ingressClassEventFilter()).
		Complete(r)
}
