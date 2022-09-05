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

	"github.com/TykTechnologies/tyk-operator/api/model"
	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/pkg/environmet"
	"github.com/TykTechnologies/tyk-operator/pkg/keys"
	"github.com/go-logr/logr"
	netV1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

// IngressReconciler watches and reconciles Ingress objects
type IngressReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Env      environmet.Env
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch

// Reconcile perform reconciliation logic for Ingress resource that is managed
// by the operator.
func (r *IngressReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	nsl := r.Log.WithValues("name", req.NamespacedName)

	desired := &netV1.Ingress{}
	if err := r.Get(ctx, req.NamespacedName, desired); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// set context for all api calls inside this reconciliation loop
	env, ctx, err := httpContext(ctx, r.Client, r.Env, desired, nsl)
	if err != nil {
		return ctrl.Result{}, err
	}

	key, ok := desired.Annotations[keys.IngressTemplateAnnotation]
	template := r.keyless()

	if ok {
		template = &v1alpha1.ApiDefinition{}

		err := r.Get(ctx, types.NamespacedName{Name: key, Namespace: req.Namespace}, template)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	nsl.Info("Sync ingress")

	op, err := util.CreateOrUpdate(ctx, r.Client, desired, func() error {
		if !desired.DeletionTimestamp.IsZero() {
			if util.ContainsFinalizer(desired, keys.IngressFinalizerName) {
				util.RemoveFinalizer(desired, keys.IngressFinalizerName)
			}
			return nil
		}

		if !util.ContainsFinalizer(desired, keys.IngressFinalizerName) {
			util.AddFinalizer(desired, keys.IngressFinalizerName)
			return nil
		}

		return nil
	})
	if err != nil {
		nsl.Error(err, "failed to update ingress object", "Op", op)
		return ctrl.Result{}, err
	}

	if !desired.DeletionTimestamp.IsZero() {
		nsl.Info("Deleted ingress resource")
		return ctrl.Result{}, nil
	}

	err = r.createAPI(ctx, nsl, template, req.Namespace, desired, env)
	if err != nil {
		nsl.Error(err, "failed to create api's")
		return ctrl.Result{}, err
	}

	nsl.Info("Sync ingress OK")

	return ctrl.Result{}, nil
}

func (r *IngressReconciler) keyless() *v1alpha1.ApiDefinition {
	return &v1alpha1.ApiDefinition{
		Spec: v1alpha1.APIDefinitionSpec{
			APIDefinitionSpec: model.APIDefinitionSpec{
				Name:             "default-keyless",
				Protocol:         "http",
				UseKeylessAccess: true,
				Active:           true,
				Proxy: model.Proxy{
					TargetURL: "http://example.com",
				},
				VersionData: model.VersionData{
					NotVersioned: true,
				},
			},
		},
	}
}

func (r *IngressReconciler) createAPI(
	ctx context.Context, lg logr.Logger,
	template *v1alpha1.ApiDefinition,
	ns string,
	desired *netV1.Ingress,
	env environmet.Env,
) error {
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
					keys.IngressLabel: desired.Name,
					keys.APIDefLabel:  hash,
				})
				api.Spec = *template.Spec.DeepCopy()
				api.Spec.Name = name
				api.Spec.Proxy.ListenPath = p.Path
				svc := p.Backend.Service
				api.Spec.Proxy.TargetURL = fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", svc.Name,
					ns, svc.Port.Number)
				if rule.Host != "" {
					api.Spec.Domain = r.translateHost(rule.Host)
				}
				if env.Ingress.HTTPPort != 0 {
					api.Spec.ListenPort = env.Ingress.HTTPPort
				}
				if !strings.Contains(p.Path, ".well-known/acme-challenge") && !strings.Contains(svc.Name, "cm-acme-http-solver") {
					for _, tls := range desired.Spec.TLS {
						for _, host := range tls.Hosts {
							if rule.Host == host {
								api.Spec.Protocol = "https"
								api.Spec.CertificateSecretNames = []string{
									tls.SecretName,
								}
								api.Spec.ListenPort = env.Ingress.HTTPSPort
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

			lg.Info("successful sync api definition", "name", name, "op", op)
		}
	}

	lg.Info("deleting orphan api's")

	return r.deleteOrphanAPI(ctx, lg, ns, desired)
}

func (r *IngressReconciler) translateHost(host string) string {
	return strings.Replace(host, "*", "{?:[^.]+}", 1)
}

func (r *IngressReconciler) deleteOrphanAPI(ctx context.Context, lg logr.Logger, ns string, desired *netV1.Ingress) error {
	var ids []string

	for _, rule := range desired.Spec.Rules {
		for _, p := range rule.HTTP.Paths {
			hash := shortHash(rule.Host + p.Path)
			ids = append(ids, hash)
		}
	}

	s := labels.NewSelector()

	exists, err := labels.NewRequirement(keys.APIDefLabel, selection.Exists, []string{})
	if err != nil {
		return err
	}

	s = s.Add(*exists)

	notIn, err := labels.NewRequirement(keys.APIDefLabel, selection.NotIn, ids)
	if err != nil {
		return err
	}

	name, err := labels.NewRequirement(keys.IngressLabel, selection.DoubleEquals, []string{desired.Name})
	if err != nil {
		return err
	}

	s = s.Add(*name)
	s = s.Add(*notIn)

	lg.Info("deleting orphan api definitions", "selector", s, "count", len(ids))

	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		return r.DeleteAllOf(ctx, &v1alpha1.ApiDefinition{}, &client.DeleteAllOfOptions{
			ListOptions: client.ListOptions{
				LabelSelector: s,
				Namespace:     ns,
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
	watch := keys.DefaultIngressClassAnnotationValue

	if override := r.Env.IngressClass; override != "" {
		watch = override
	}

	isOurIngress := func(o runtime.Object) bool {
		switch e := o.(type) {
		case *netV1.Ingress:
			return e.GetAnnotations()[keys.IngressClassAnnotation] == watch
		default:
			return false
		}
	}

	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return isOurIngress(e.Object)
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return isOurIngress(e.ObjectNew)
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return isOurIngress(e.Object)
		},
	}
}

// SetupWithManager initializes ingress controller manager
func (r *IngressReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&netV1.Ingress{}).
		Owns(&v1alpha1.ApiDefinition{}).
		WithEventFilter(r.ingressClassEventFilter()).
		Complete(r)
}
