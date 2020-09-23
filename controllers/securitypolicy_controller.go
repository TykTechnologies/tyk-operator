/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"reflect"
	"time"

	"k8s.io/client-go/tools/record"

	tykv1 "github.com/TykTechnologies/tyk-operator/api/v1"
	"github.com/TykTechnologies/tyk-operator/internal/universal_client"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
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

	policyID := req.NamespacedName

	log.Info("fetching SecurityPolicy instance")

	desired := &tykv1.SecurityPolicy{}
	if err := r.Get(ctx, policyID, desired); err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("SecurityPolicy resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get SecurityPolicy")
		return ctrl.Result{}, err
	}

	const securityPolicyFinalzerName = "finalizers.tyk.io/securitypolicy"
	if desired.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, if it does not have our finalizer,
		// then lets add the finalizer and update the object.
		if !containsString(desired.ObjectMeta.Finalizers, securityPolicyFinalzerName) {
			desired.ObjectMeta.Finalizers = append(desired.ObjectMeta.Finalizers, securityPolicyFinalzerName)
			if err := r.Update(ctx, desired); err != nil {
				return reconcile.Result{}, err
			}
		}
	} else {
		// The object is being deleted
		if containsString(desired.ObjectMeta.Finalizers, securityPolicyFinalzerName) {
			// our finalizer is present, so lets handle our external dependency

			if err := r.UniversalClient.SecurityPolicy().Delete(desired.Status.ID); err != nil {
				if err.Error() != "API Returned error: Not Found" {
					return reconcile.Result{Requeue: false}, err
				}
			}

			_ = r.UniversalClient.HotReload()

			// remove our finalizer from the list and update it.
			desired.ObjectMeta.Finalizers = removeString(desired.ObjectMeta.Finalizers, securityPolicyFinalzerName)
			if err := r.Update(ctx, desired); err != nil {
				return reconcile.Result{}, nil
			}
		}

		// Our finalizer has finished, so the reconciler can do nothing.
		return reconcile.Result{}, nil
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

			apiDef, err := r.UniversalClient.Api().Get(api.Status.Id)
			if err != nil {
				log.Error(err, "api doesnt exist")
				return ctrl.Result{Requeue: true}, err
			}

			desired.Spec.AccessRightsArray[i].APIID = apiDef.APIID
			desired.Spec.AccessRightsArray[i].APIName = apiDef.Name
			desired.Spec.OrgID = apiDef.OrgID

			if err := r.Update(ctx, desired); err != nil {
				log.Error(err, "unable to update apiId in access rights array")
				return ctrl.Result{Requeue: true}, err
			}
		}
	}

	// CREATE
	if desired.Status.ID == "" {
		// policy doesn't exist?
		log.Info("creating policy", "policyID", policyID.String())

		internalID, err := r.UniversalClient.SecurityPolicy().Create(&desired.Spec)
		if err != nil {
			log.Error(err, "unable to create SecurityPolicy")
			return ctrl.Result{RequeueAfter: time.Second * 5}, err
		}

		_ = r.UniversalClient.HotReload()

		desired.Status.ID = internalID
		err = r.Status().Update(ctx, desired)
		if err != nil {
			log.Error(err, "Failed to update ApiDef status")
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	// UPDATE
	actual, err := r.UniversalClient.SecurityPolicy().Get(desired.Status.ID)
	if err != nil {
		log.Error(err, "something fucked")
		return ctrl.Result{Requeue: true}, err
	}
	if !reflect.DeepEqual(desired.Spec, actual) {
		log.Info("updating SecurityPolicy")
		desired.Spec.ID = desired.Status.ID
		desired.Spec.MID = desired.Status.ID
		err := r.UniversalClient.SecurityPolicy().Update(&desired.Spec)
		if err != nil {
			log.Error(err, "unable to update SecurityPolicy")
			return ctrl.Result{Requeue: true}, err
		}

		_ = r.UniversalClient.HotReload()
	}

	return ctrl.Result{}, nil
}

func (r *SecurityPolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tykv1.SecurityPolicy{}).
		Complete(r)
}
