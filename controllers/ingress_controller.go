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
	"errors"
	"fmt"

	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/pkg/universal_client"
	"github.com/go-logr/logr"
	"k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const ingressFinalizerName = "finalizers.tyk.io/ingress"

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

	// FINALIZER LOGIC =================================================================================================

	// If object is being deleted
	if !desired.ObjectMeta.DeletionTimestamp.IsZero() {
		// If our finalizer is present, need to delete from Tyk still
		if containsString(desired.ObjectMeta.Finalizers, ingressFinalizerName) {
			// TODO: Logic to delete ALL apis created by this ingress resource
		}

		return ctrl.Result{}, nil
	}

	// If finalizer not present, add it; This is a new object
	if !containsString(desired.ObjectMeta.Finalizers, ingressFinalizerName) {
		desired.ObjectMeta.Finalizers = append(desired.ObjectMeta.Finalizers, ingressFinalizerName)
		err := r.Update(ctx, desired)
		// Return either way because the update will
		// issue a requeue anyway
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// /FINALIZER LOGIC ================================================================================================

	if desired.Spec.IngressClassName == nil {
		log.Info("IngressClassName should not be nil")
		return ctrl.Result{}, errors.New("IngressClassName should not be nil")
	}

	ingressClassList := &v1beta1.IngressClassList{}
	opts := []client.ListOption{
		client.InNamespace(req.Namespace),
	}
	err := r.List(ctx, ingressClassList, opts...)
	if err != nil {
		println(err.Error())
		return ctrl.Result{}, err
	}

	println(ingressClassList.Kind)

	ingressClass := &v1beta1.IngressClass{}
	if err := r.Get(ctx, types.NamespacedName{Name: *desired.Spec.IngressClassName, Namespace: req.Namespace}, ingressClass); err != nil {
		println(err.Error())
		return ctrl.Result{}, err
	}

	if ingressClass.Spec.Controller != "tyk-operator" {
		log.Info(fmt.Sprintf("this ingress class is not for us %s", ingressClass.Spec.Controller))
		return ctrl.Result{}, nil
	}

	if ingressClass.Spec.Parameters == nil {
		// without any params - this just a passthrough API?

		return ctrl.Result{}, nil
	}

	if ingressClass.Spec.Parameters.Kind != "ApiDefinition" {
		return ctrl.Result{}, fmt.Errorf("unknown Kind %s", ingressClass.Spec.Parameters.Kind)
	}

	if *ingressClass.Spec.Parameters.APIGroup != "tyk.tyk.io" {
		return ctrl.Result{}, fmt.Errorf("unknown APIGroup %s", *ingressClass.Spec.Parameters.APIGroup)
	}

	// we have parameters - as such, we should ensure that there is an api definition resource
	template := &v1alpha1.ApiDefinition{}
	err = r.Get(ctx, types.NamespacedName{Name: ingressClass.Spec.Parameters.Name, Namespace: req.Namespace}, template)
	if err != nil {
		log.Error(err, "error getting api definition to use as a template", "name", ingressClass.Spec.Parameters.Name)
		return ctrl.Result{}, err
	}

	ingressInNS := &v1beta1.IngressList{}
	opts = []client.ListOption{
		client.InNamespace(req.NamespacedName.Namespace),
	}
	if err := r.List(ctx, ingressInNS, opts...); err != nil {
		return ctrl.Result{}, err
	}
	//for _, item := range ingressInNS.Items {
	//	for _, rule := range item.Spec.Rules {
	//		for _, path := range rule.HTTP.Paths {
	//
	//		}
	//	}
	//}

	apiDefinitionsInNS := &v1alpha1.ApiDefinitionList{}
	opts = []client.ListOption{
		client.InNamespace(req.NamespacedName.Namespace),
	}
	if err := r.List(ctx, apiDefinitionsInNS, opts...); err != nil {
		return ctrl.Result{}, err
	}

	//apiDefinitionsToDelete := &v1alpha1.ApiDefinitionList{}
	apiDefinitionsToCreateOrUpdate := &v1alpha1.ApiDefinitionList{}

	for _, rule := range desired.Spec.Rules {
		// TODO: precise match only for time being. Need to convert K8s * syntax to Tyk host regexp.
		hostName := rule.Host

		for _, p := range rule.HTTP.Paths {
			api := template.DeepCopy()
			api.Name = fmt.Sprintf("%s %s %s #%s", namespacedName.Namespace, namespacedName.Name, p.Path, ingressClass.Spec.Parameters.Name)
			api.Spec.Proxy.ListenPath = p.Path
			api.Spec.Proxy.TargetURL = fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", p.Backend.ServiceName, namespacedName.Namespace, p.Backend.ServicePort.IntValue())
			api.Spec.Domain = hostName

			apiDefinitionsToCreateOrUpdate.Items = append(apiDefinitionsToCreateOrUpdate.Items, *api)
		}
	}

	return ctrl.Result{}, nil
}

func (r *IngressReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.Ingress{}).
		Complete(r)
}
