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

	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/TykTechnologies/tyk-operator/pkg/universal_client"
	"github.com/go-logr/logr"
	"k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ingressFinalizerName               = "finalizers.tyk.io/ingress"
	ingressClassAnnotationKey          = "kubernetes.io/ingress.class"
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

	if ingressClassValue, ok := desired.Annotations[ingressClassAnnotationKey]; !ok {
		log.Info("not for us")
		return ctrl.Result{}, nil
	} else {
		if ingressClassValue != defaultIngressClassAnnotationValue {
			log.Info("not for us")
			return ctrl.Result{}, nil
		}
	}

	// FINALIZER LOGIC =================================================================================================

	// If object is being deleted
	if !desired.ObjectMeta.DeletionTimestamp.IsZero() {
		// If our finalizer is present, need to delete from Tyk still
		if containsString(desired.ObjectMeta.Finalizers, ingressFinalizerName) {
			log.Info("resource is being deleted - executing finalizer logic")
			// TODO: Logic to delete ALL apis managed by this ingress resource

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

	//// we have parameters - as such, we should ensure that there is an api definition resource
	//template := &v1alpha1.ApiDefinition{}
	//err = r.Get(ctx, types.NamespacedName{Name: ingressClass.Spec.Parameters.Name, Namespace: req.Namespace}, template)
	//if err != nil {
	//	log.Error(err, "error getting api definition to use as a template", "name", ingressClass.Spec.Parameters.Name)
	//	return ctrl.Result{}, err
	//}
	//
	//ingressInNS := &v1beta1.IngressList{}
	//opts = []client.ListOption{
	//	client.InNamespace(req.NamespacedName.Namespace),
	//}
	//if err := r.List(ctx, ingressInNS, opts...); err != nil {
	//	return ctrl.Result{}, err
	//}
	////for _, item := range ingressInNS.Items {
	////	for _, rule := range item.Spec.Rules {
	////		for _, path := range rule.HTTP.Paths {
	////
	////		}
	////	}
	////}
	//
	//apiDefinitionsInNS := &v1alpha1.ApiDefinitionList{}
	//opts = []client.ListOption{
	//	client.InNamespace(req.NamespacedName.Namespace),
	//}
	//if err := r.List(ctx, apiDefinitionsInNS, opts...); err != nil {
	//	return ctrl.Result{}, err
	//}
	//
	////apiDefinitionsToDelete := &v1alpha1.ApiDefinitionList{}
	//apiDefinitionsToCreateOrUpdate := &v1alpha1.ApiDefinitionList{}
	//
	//for _, rule := range desired.Spec.Rules {
	//	// TODO: precise match only for time being. Need to convert K8s * syntax to Tyk host regexp.
	//	hostName := rule.Host
	//
	//	for _, p := range rule.HTTP.Paths {
	//		api := template.DeepCopy()
	//		api.Name = fmt.Sprintf("%s %s %s #%s", namespacedName.Namespace, namespacedName.Name, p.Path, ingressClass.Spec.Parameters.Name)
	//		api.Spec.Proxy.ListenPath = p.Path
	//		api.Spec.Proxy.TargetURL = fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", p.Backend.ServiceName, namespacedName.Namespace, p.Backend.ServicePort.IntValue())
	//		api.Spec.Domain = hostName
	//
	//		apiDefinitionsToCreateOrUpdate.Items = append(apiDefinitionsToCreateOrUpdate.Items, *api)
	//	}
	//}

	return ctrl.Result{}, nil
}

func (r *IngressReconciler) ingressClassEventFilter() predicate.Predicate {
	isOurIngress := func(annotations map[string]string) bool {
		if ingressClass, ok := annotations[ingressClassAnnotationKey]; !ok {
			r.Log.Info("test ingress class")
			// if there is no ingress class - we can be greedy. :D
			return true
		} else if ingressClass == defaultIngressClassAnnotationValue {
			// if the ingress class is `tyk` it's for us
			return true
		}

		// TODO: custom ingress class annotation

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
