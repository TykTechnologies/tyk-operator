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

	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/pkg/universal_client"
	"github.com/go-logr/logr"
	"k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
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
	namespacedName := req.NamespacedName

	log := r.Log.WithValues("Ingress", namespacedName.String())

	desired := &v1beta1.Ingress{}
	if err := r.Get(ctx, req.NamespacedName, desired); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// ALL MANAGED API DEFINITIONS =====================================================================================
	lbls := map[string]string{
		"ingress": req.Name,
	}

	oldAPIs := v1alpha1.ApiDefinitionList{}
	opts := []client.ListOption{
		client.InNamespace(req.Namespace),
		client.MatchingLabels(lbls),
	}
	if err := r.List(ctx, &oldAPIs, opts...); err != nil {
		log.Error(err, "unable to list apis")
		return ctrl.Result{}, err
	}

	// /MANAGED API DEFINITIONS ========================================================================================

	// FINALIZER LOGIC =================================================================================================

	// If object is being deleted
	if !desired.ObjectMeta.DeletionTimestamp.IsZero() {
		// If our finalizer is present, need to delete from Tyk still
		if containsString(desired.ObjectMeta.Finalizers, ingressFinalizerName) {
			log.Info("resource is being deleted - removing associated api definitions")

			for _, a := range oldAPIs.Items {
				_ = r.Delete(ctx, &a)
			}

			log.Info("removing finalizer from ingress")
			desired.ObjectMeta.Finalizers = removeString(desired.ObjectMeta.Finalizers, ingressFinalizerName)
			if err := r.Update(ctx, desired); err != nil {
				return ctrl.Result{}, err
			}
		}

		return ctrl.Result{}, nil
	}

	// If finalizer not present, add it; This is a new object
	if !containsString(desired.ObjectMeta.Finalizers, ingressFinalizerName) {
		log.Info("adding finalizer", "name", ingressFinalizerName)
		desired.ObjectMeta.Finalizers = append(desired.ObjectMeta.Finalizers, ingressFinalizerName)
		err := r.Update(ctx, desired)
		// Return either way because the update will
		// issue a requeue anyway
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// /FINALIZER LOGIC ================================================================================================

	// ensure there is a template api definition
	templateAnnotationValue, ok := desired.Annotations[ingressTemplateAnnotationKey]
	if !ok {
		return ctrl.Result{}, fmt.Errorf("expecting template annotation %s", ingressTemplateAnnotationKey)
	}

	// we have parameters - as such, we should ensure that there is an api definition resource
	template := &v1alpha1.ApiDefinition{}
	err := r.Get(ctx, types.NamespacedName{Name: templateAnnotationValue, Namespace: req.Namespace}, template)
	if err != nil {
		log.Error(err, "error getting api definition to use as a template", "name", templateAnnotationValue)
		return ctrl.Result{}, err
	}

	apisToCreateOrUpdate := v1alpha1.ApiDefinitionList{}
	apisToUpdate := v1alpha1.ApiDefinitionList{}
	apisToCreate := v1alpha1.ApiDefinitionList{}
	apisToDelete := v1alpha1.ApiDefinitionList{}

	for _, rule := range desired.Spec.Rules {
		hostName := rule.Host
		for _, p := range rule.HTTP.Paths {
			apiName := r.buildAPIName(req.Namespace, req.Name, hostName, p.Path)
			api := v1alpha1.ApiDefinition{
				ObjectMeta: metav1.ObjectMeta{
					Name:      apiName,
					Namespace: req.Namespace,
					Labels:    lbls,
				},
				Spec: *template.Spec.DeepCopy(),
			}

			api.Spec.Name = apiName
			api.Spec.Proxy.ListenPath = p.Path
			api.Spec.Proxy.TargetURL = fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", p.Backend.ServiceName, namespacedName.Namespace, p.Backend.ServicePort.IntValue())

			// TODO: Translate to Tyk custom domain
			api.Spec.Domain = hostName

			apisToCreateOrUpdate.Items = append(apisToCreateOrUpdate.Items, api)
		}
	}

	// All the apis we should delete or update
	for _, x := range oldAPIs.Items {
		found := false
		for _, y := range apisToCreateOrUpdate.Items {
			if y.Name == x.Name {
				apisToUpdate.Items = append(apisToUpdate.Items, y)
				found = true
				break
			}
		}

		if !found {
			apisToDelete.Items = append(apisToDelete.Items, x)
		}
	}

	// all the items to create
	for _, x := range apisToCreateOrUpdate.Items {
		create := true
		for _, y := range apisToUpdate.Items {
			if x.Name == y.Name {
				create = false
				break
			}
		}

		if create {
			apisToCreate.Items = append(apisToCreate.Items, x)
		}
	}

	// create new endpoints first
	for _, a := range apisToCreate.Items {
		if err := r.Create(ctx, &a); err != nil {
			log.Error(err, "unable to update api")
		}
	}

	// update second
	for _, a := range apisToUpdate.Items {
		apiDefToUpdate := v1alpha1.ApiDefinition{}
		err := r.Get(ctx, types.NamespacedName{Name: a.Name, Namespace: a.Namespace}, &apiDefToUpdate)
		if err != nil {
			log.Error(err, "unable to get api to update")
			continue
		}
		apiDefToUpdate.Spec = a.Spec
		if err := r.Update(ctx, &apiDefToUpdate); err != nil {
			log.Error(err, "unable to update api")
		}
	}

	// delete last - just in-case something renamed
	for _, a := range apisToDelete.Items {
		if err := r.Delete(ctx, &a); err != nil {
			log.Error(err, "unable to update api")
		}
	}

	return ctrl.Result{}, nil
}

func (r *IngressReconciler) buildAPIName(nameSpace, name, hostName, path string) string {
	return fmt.Sprintf("%s-%s-%s", nameSpace, name, shortHash(hostName+path))
}

func shortHash(txt string) string {
	h := sha256.New()
	h.Write([]byte(txt))
	return fmt.Sprintf("%x", h.Sum(nil))[:9]
}

func (r *IngressReconciler) ingressClassEventFilter() predicate.Predicate {
	isOurIngress := func(annotations map[string]string) bool {
		if ingressClass, ok := annotations[ingressClassAnnotationKey]; !ok {
			r.Log.Info("test ingress class")
			// if there is no ingress class - it's prob not for us
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
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.Ingress{}).
		WithEventFilter(r.ingressClassEventFilter()).
		Complete(r)
}
