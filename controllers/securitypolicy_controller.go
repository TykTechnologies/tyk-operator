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
	"github.com/TykTechnologies/tyk-operator/pkg/client/universal"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	util "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const policyFinalizer = "finalizers.tyk.io/securitypolicy"

// SecurityPolicyReconciler reconciles a SecurityPolicy object
type SecurityPolicyReconciler struct {
	client.Client
	Log             logr.Logger
	Scheme          *runtime.Scheme
	UniversalClient universal.Client
	Recorder        record.EventRecorder
}

// +kubebuilder:rbac:groups=tyk.tyk.io,resources=securitypolicies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=tyk.tyk.io,resources=securitypolicies/status,verbs=get;update;patch

// Reconcile reconciles SecurityPolicy custom resources
func (r *SecurityPolicyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("SecurityPolicy", req.NamespacedName.String())

	ns := req.NamespacedName.String()
	log.Info("Reconciling SecurityPolicy instance")

	// Lookup policy object
	policy := &tykv1.SecurityPolicy{}
	if err := r.Get(ctx, req.NamespacedName, policy); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	var reqA time.Duration
	_, err := util.CreateOrUpdate(ctx, r.Client, policy, func() error {
		if !policy.ObjectMeta.DeletionTimestamp.IsZero() {
			if util.ContainsFinalizer(policy, policyFinalizer) {
				return r.delete(ctx, policy)
			}
			return nil
		}
		util.AddFinalizer(policy, policyFinalizer)
		if policy.Spec.ID == "" {
			policy.Spec.ID = encodeNS(ns)
		}
		// update access rights
		r.Log.Info("updating access rights")
		if policy.Status.PolID == "" {
			return r.create(ctx, policy)
		}
		return r.update(ctx, policy)
	})
	if err == nil {
		r.Log.Info("Completed reconciling SecurityPolicy instance")
	}
	return ctrl.Result{RequeueAfter: reqA}, err
}

// returns a copy of SecurityPolicySpec with AccessRightsArray updated
func (r *SecurityPolicyReconciler) spec(ctx context.Context, policy *tykv1.SecurityPolicy) (*tykv1.SecurityPolicySpec, error) {
	spec := policy.Spec.DeepCopy()
	for i := 0; i < len(spec.AccessRightsArray); i++ {
		err := r.updateAccess(ctx, spec.AccessRightsArray[i])
		if err != nil {
			return nil, err
		}
	}
	return spec, nil
}

func (r *SecurityPolicyReconciler) updateAccess(ctx context.Context,
	a *tykv1.AccessDefinition) error {
	api := &tykv1.ApiDefinition{}
	if err := r.Get(ctx, types.NamespacedName{Name: a.Name, Namespace: a.Namespace}, api); err != nil {
		if errors.IsNotFound(err) {
			r.Log.Info("ApiDefinition resource not found. Unable to attach to SecurityPolicy. ReQueue",
				"Name", a.Name,
				"Namespace", a.Namespace,
			)
			return err
		}
		r.Log.Error(err, "Failed to get APIDefinition to attach to SecurityPolicy")
		return err
	}
	def, err := r.UniversalClient.Api().Get(ctx, api.Status.ApiID)
	if err != nil {
		return err
	}
	a.APIID = def.APIID
	a.APIName = def.Name
	return nil
}

func (r *SecurityPolicyReconciler) delete(ctx context.Context, policy *tykv1.SecurityPolicy) error {
	r.Log.Info("Deleting policy")
	util.RemoveFinalizer(policy, policyFinalizer)
	if err := r.UniversalClient.SecurityPolicy().Delete(ctx, policy.Status.PolID); err != nil {
		if universal.IsNotFound(err) {
			r.Log.Info("Policy not found")
			return nil
		}
		r.Log.Error(err, "Failed to delete resource")
		return err
	}
	err := r.updateLinkedAPI(ctx, policy, func(ads *tykv1.ApiDefinitionStatus, ns tykv1.Target) {
		ads.LinkedByPolicies = removeTarget(ads.LinkedByPolicies, ns)
	})
	if err != nil {
		return err
	}
	r.Log.Info("Successfully deleted Policy")
	return nil
}

func (r *SecurityPolicyReconciler) update(ctx context.Context, policy *tykv1.SecurityPolicy) error {
	r.Log.Info("Updating  policy")
	policy.Spec.MID = policy.Status.PolID
	spec, err := r.spec(ctx, policy)
	if err != nil {
		return err
	}
	err = r.UniversalClient.SecurityPolicy().Update(ctx, spec)
	if err != nil {
		r.Log.Error(err, "Failed to update policy")
		return err
	}
	err = r.updateLinkedAPI(ctx, policy, func(ads *tykv1.ApiDefinitionStatus, s tykv1.Target) {
		ads.LinkedByPolicies = addTarget(ads.LinkedByPolicies, s)
	})
	if err != nil {
		return err
	}
	r.UniversalClient.HotReload()
	r.Log.Info("Successfully updated Policy")
	return nil
}

func (r *SecurityPolicyReconciler) create(ctx context.Context, policy *tykv1.SecurityPolicy) error {
	r.Log.Info("Creating  policy")
	spec, err := r.spec(ctx, policy)
	if err != nil {
		return err
	}
	err = r.UniversalClient.SecurityPolicy().Create(ctx, spec)
	if err != nil {
		r.Log.Error(err, "Failed to create policy")
		return err
	}
	err = r.updateLinkedAPI(ctx, policy, func(ads *tykv1.ApiDefinitionStatus, s tykv1.Target) {
		ads.LinkedByPolicies = addTarget(ads.LinkedByPolicies, s)
	})
	r.Log.Info("Successful created Policy")
	policy.Spec.MID = spec.MID
	policy.Status.PolID = policy.Spec.MID
	return r.Status().Update(ctx, policy)
}

// updateLinkedAPI updates the status of api definitions associated with this
// policy.
func (r *SecurityPolicyReconciler) updateLinkedAPI(ctx context.Context, policy *tykv1.SecurityPolicy,
	fn func(*tykv1.ApiDefinitionStatus, tykv1.Target),
) error {
	r.Log.Info("Updating linked api definitions")
	ns := tykv1.Target{
		Namespace: policy.Namespace, Name: policy.Name,
	}
	for _, a := range policy.Spec.AccessRightsArray {
		api := &tykv1.ApiDefinition{}
		if err := r.Get(ctx, types.NamespacedName{Name: a.Name, Namespace: a.Namespace}, api); err != nil {
			r.Log.Error(err, "Failed to get linked api definition")
			return err
		}
		fn(&api.Status, ns)
		if err := r.Status().Update(ctx, api); err != nil {
			r.Log.Error(err, "Failed to update linked api definition")
			return err
		}
	}
	return nil
}

// SetupWithManager initializes the security policy controller.
func (r *SecurityPolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tykv1.SecurityPolicy{}).
		Complete(r)
}
