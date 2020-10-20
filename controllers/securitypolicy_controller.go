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
	"time"

	tykv1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/internal/universal_client"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// SecurityPolicyReconciler reconciles a SecurityPolicy object
type SecurityPolicyReconciler struct {
	client.Client
	Log             logr.Logger
	Scheme          *runtime.Scheme
	UniversalClient universal_client.UniversalClient
	Recorder        record.EventRecorder
}

// +kubebuilder:rbac:groups=tyk.tyk.io,resources=securitypolicies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=tyk.tyk.io,resources=securitypolicies/status,verbs=get;update;patch

func (r *SecurityPolicyReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("SecurityPolicy", req.NamespacedName.String())

	policyNamespacedName := req.NamespacedName.String()

	log.Info("fetching SecurityPolicy instance")

	// Lookup policy object
	desired := &tykv1.SecurityPolicy{}
	if err := r.Get(ctx, req.NamespacedName, desired); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err) // Ignore not-found errors
	}
	r.Recorder.Event(desired, "Normal", "SecurityPolicy", "Reconciling")
	const securityPolicyFinalzerName = "finalizers.tyk.io/securitypolicy"

	// If object is being deleted
	if !desired.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is being deleted
		if containsString(desired.ObjectMeta.Finalizers, securityPolicyFinalzerName) {
			// our finalizer is present, so lets handle our external dependency

			if err := r.UniversalClient.SecurityPolicy().Delete(policyNamespacedName); err != nil {
				log.Error(err, "unable to delete policy", "nameSpacedName", policyNamespacedName)
				return reconcile.Result{Requeue: true}, err
			}

			err := r.UniversalClient.HotReload()
			if err != nil {
				log.Error(err, "unable to hot reload after policy delete", "nameSpacedName", policyNamespacedName)
			}

			// remove our finalizer from the list and update it.
			desired.ObjectMeta.Finalizers = removeString(desired.ObjectMeta.Finalizers, securityPolicyFinalzerName)
			if err := r.Update(ctx, desired); err != nil {
				return reconcile.Result{}, nil
			}
		}

		// Our finalizer has finished, so the reconciler can do nothing.
		return reconcile.Result{}, nil
	}

	// If finalizer not present, add it; This is a new object
	if !containsString(desired.ObjectMeta.Finalizers, securityPolicyFinalzerName) {
		desired.ObjectMeta.Finalizers = append(desired.ObjectMeta.Finalizers, securityPolicyFinalzerName)
		err := r.Update(ctx, desired)
		// Return either way because the update will
		// issue a requeue anyway
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	for i, accessRight := range desired.Spec.AccessRightsArray {
		apiNamespace := accessRight.Namespace
		apiName := accessRight.Name

		api := &tykv1.ApiDefinition{}
		if err := r.Get(ctx, types.NamespacedName{Name: apiName, Namespace: apiNamespace}, api); err != nil {
			if errors.IsNotFound(err) {
				// Request object not found, could have been deleted after reconcile request.
				// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
				// Return and don't requeue
				log.Info("ApiDefinition resource not found. Unable to attach to SecurityPolicy. ReQueue")
				return ctrl.Result{RequeueAfter: time.Second * 5}, err
			}
			// Error reading the access_right object - requeue the request.
			log.Error(err, "Failed to get APIDefinition to attach to SecurityPolicy")
			return ctrl.Result{RequeueAfter: time.Second * 5}, err
		}

		// We have the apiDefinition resource
		// if it doesn't match
		if desired.Spec.AccessRightsArray[i].APIID == "" ||
			desired.Spec.OrgID == "" ||
			desired.Spec.AccessRightsArray[i].APIName != api.Spec.Name {

			apiDef, err := r.UniversalClient.Api().Get(api.Status.ApiID)
			if err != nil || apiDef == nil {
				log.Error(err, "api doesnt exist")
				return ctrl.Result{Requeue: true}, err
			}

			desired.Spec.AccessRightsArray[i].APIID = apiDef.APIID
			desired.Spec.AccessRightsArray[i].APIName = apiDef.Name
			desired.Spec.OrgID = apiDef.OrgID

			if err := r.Update(ctx, desired); err != nil {
				log.Error(err, "unable to update apiId in access rights array")
			}
			return ctrl.Result{Requeue: true}, client.IgnoreNotFound(err)
		}
	}

	createdPol, err := universal_client.CreateOrUpdatePolicy(r.UniversalClient, &desired.Spec, policyNamespacedName)
	if err != nil {
		log.Error(err, "createOrUpdatePolicy failure")
		r.Recorder.Event(desired, "Warning", "SecurityPolicy", "Create or Update Security Policy")
		return ctrl.Result{Requeue: true}, nil
	}

	// if pol_id not there, add it, this is new object.
	if desired.Status.PolID == "" {
		desired.Status.PolID = createdPol.MID
		if err = r.Status().Update(ctx, desired); err != nil {
			log.Error(err, "Could not update ID")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *SecurityPolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tykv1.SecurityPolicy{}).
		Complete(r)
}
