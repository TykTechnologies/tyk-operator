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
	"bytes"
	"context"
	"encoding/json"
	"os"
	"time"

	tykv1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/pkg/environmet"
	"github.com/TykTechnologies/tyk-operator/pkg/universal_client"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const securityPolicyFinalzerName = "finalizers.tyk.io/securitypolicy"

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

	log.Info("Reconciling SecurityPolicy instance")

	// Lookup policy object
	desired := &tykv1.SecurityPolicy{}
	if err := r.Get(ctx, req.NamespacedName, desired); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err) // Ignore not-found errors
	}

	// If object is being deleted
	if !desired.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is being deleted
		if containsString(desired.ObjectMeta.Finalizers, securityPolicyFinalzerName) {
			// our finalizer is present, so lets handle our external dependency
			if err := r.UniversalClient.SecurityPolicy().Delete(desired.Status.PolID); err != nil {
				log.Error(err, "unable to delete policy", "nameSpacedName", policyNamespacedName, "polId", desired.Status.PolID)
				return reconcile.Result{}, nil
			}

			err := r.UniversalClient.HotReload()
			if err != nil {
				log.Error(err, "unable to hot reload after policy delete", "nameSpacedName", policyNamespacedName, "polId", desired.Status.PolID)
			}

			// remove our finalizer from the list and update it.
			desired.ObjectMeta.Finalizers = removeString(desired.ObjectMeta.Finalizers, securityPolicyFinalzerName)
			if err := r.Update(ctx, desired); err != nil {
				log.Error(err, "Error deleting finalizer from Policy")
				return reconcile.Result{}, err
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

	if desired.Spec.ID == "" {
		desired.Spec.ID = encodeNS(policyNamespacedName)
	}
	if desired.Spec.OrgID == "" {
		desired.Spec.OrgID = os.Getenv(environmet.TykORG)
	}

	for i := 0; i < len(desired.Spec.AccessRightsArray); i++ {
		a := &desired.Spec.AccessRightsArray[i]
		apiNamespace := a.Namespace
		api := &tykv1.ApiDefinition{}
		if err := r.Get(ctx, types.NamespacedName{Name: a.Name, Namespace: apiNamespace}, api); err != nil {
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
		// update the access right with the api definition details
		apiDef, err := r.UniversalClient.Api().Get(api.Status.ApiID)
		if err != nil {
			log.Error(err, "Failed to find the APIDefinition")
			return ctrl.Result{Requeue: true}, err
		}
		a.APIID = apiDef.APIID
		a.APIName = apiDef.Name
		a.Name = ""
		a.Namespace = ""
	}
	if !r.PolicyChanged(log, desired) {
		log.Info("Nothing changed")
		return ctrl.Result{}, nil
	}
	desired.Spec.UpdateAccessRights()

	// if "Status.PolID" not there, add and save it, this is new object.
	if desired.Status.PolID == "" {
		// we are creating a new policy object
		err := r.UniversalClient.SecurityPolicy().Create(&desired.Spec)
		if err != nil {
			log.Error(err, "Failed to create policy ")
			return ctrl.Result{}, err
		}
		desired.Status.PolID = desired.Spec.MID
		log.Info("successful created a policy", "MID", desired.Spec.MID, "ID", desired.Status.PolID)
		err = r.Status().Update(ctx, desired)
		if err != nil {
			log.Error(err, "Could not update Status ID")
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	// we are updating a policy
	desired.Spec.MID = desired.Status.PolID
	err := r.UniversalClient.SecurityPolicy().Update(&desired.Spec)
	if err != nil {
		log.Error(err, "Failed to update policy resource", "ID", desired.Status.PolID)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	log.Info("Successful updated policy", "ID", desired.Status.PolID)
	return ctrl.Result{}, nil
}

// PolicyChanged returns true if there was any changes in the policy object.
func (r *SecurityPolicyReconciler) PolicyChanged(log logr.Logger, def *tykv1.SecurityPolicy) bool {
	if def.Status.PolID == "" {
		return true
	}
	def.Spec.MID = def.Status.PolID
	pol, err := r.UniversalClient.SecurityPolicy().Get(def.Status.PolID)
	if err != nil {
		return true
	}
	a, _ := json.Marshal(def.Spec)
	pol.AccessRights = nil
	b, _ := json.Marshal(pol)
	return !bytes.Equal(a, b)
}

func (r *SecurityPolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tykv1.SecurityPolicy{}).
		Complete(r)
}
