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
	"fmt"
	"time"

	"github.com/TykTechnologies/tyk-operator/api/model"
	tykv1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	opclient "github.com/TykTechnologies/tyk-operator/pkg/client"
	"github.com/TykTechnologies/tyk-operator/pkg/client/klient"
	"github.com/TykTechnologies/tyk-operator/pkg/environmet"
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
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Env      environmet.Env
	Recorder record.EventRecorder
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
	// set context for all api calls inside this reconciliation loop
	env, ctx, err := HttpContext(ctx, r.Client, r.Env, policy, log)
	if err != nil {
		return ctrl.Result{}, err
	}

	var reqA time.Duration

	_, err = util.CreateOrUpdate(ctx, r.Client, policy, func() error {
		if !policy.ObjectMeta.DeletionTimestamp.IsZero() {
			if util.ContainsFinalizer(policy, policyFinalizer) {
				return r.delete(ctx, policy)
			}
			return nil
		}

		util.AddFinalizer(policy, policyFinalizer)

		if policy.Spec.ID == "" {
			policy.Spec.ID = EncodeNS(ns)
		}

		if policy.Spec.OrgID == "" {
			policy.Spec.OrgID = env.Org
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
	} else {
		reqA = queueAfter
	}

	return ctrl.Result{RequeueAfter: reqA}, err
}

// returns a copy of SecurityPolicySpec with AccessRightsArray updated
func (r *SecurityPolicyReconciler) spec(ctx context.Context, policy *tykv1.SecurityPolicy) (*tykv1.SecurityPolicySpec, error) {
	spec := policy.Spec.DeepCopy()
	spec.Context = nil

	for i := 0; i < len(spec.AccessRightsArray); i++ {
		err := r.updateAccess(ctx, spec.AccessRightsArray[i])
		if err != nil {
			return nil, err
		}
	}

	return spec, nil
}

func (r *SecurityPolicyReconciler) updateAccess(ctx context.Context,
	a *tykv1.AccessDefinition,
) error {
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

	if api.Status.ApiID == "" {
		return opclient.ErrNotFound
	}

	a.APIID = api.Status.ApiID
	a.APIName = api.Spec.Name

	return nil
}

func (r *SecurityPolicyReconciler) delete(ctx context.Context, policy *tykv1.SecurityPolicy) error {
	r.Log.Info("Deleting policy")

	all, err := klient.Universal.Portal().Catalogue().Get(ctx)
	if err != nil {
		return err
	}

	for _, v := range all.APIS {
		if v.PolicyID == policy.Status.PolID {
			return fmt.Errorf("cannot delete policy due to catalogue %q dependency", all.Id)
		}
	}

	util.RemoveFinalizer(policy, policyFinalizer)

	if err := klient.Universal.Portal().Policy().Delete(ctx, policy.Status.PolID); err != nil {
		if opclient.IsNotFound(err) {
			r.Log.Info("Policy not found")
			return nil
		}

		r.Log.Error(err, "Failed to delete resource")

		return err
	}

	err = r.updateStatusOfLinkedAPIs(ctx, policy, true)
	if err != nil {
		return err
	}

	r.Log.Info("Successfully deleted Policy")

	return nil
}

func (r *SecurityPolicyReconciler) update(ctx context.Context, policy *tykv1.SecurityPolicy) error {
	r.Log.Info("Updating policy")

	policy.Spec.MID = policy.Status.PolID

	spec, err := r.spec(ctx, policy)
	if err != nil {
		return err
	}

	err = klient.Universal.Portal().Policy().Update(ctx, spec)
	if err != nil {
		r.Log.Error(err, "Failed to update policy")
		return err
	}

	err = r.updateStatusOfLinkedAPIs(ctx, policy, false)
	if err != nil {
		return err
	}

	klient.Universal.HotReload(ctx)
	r.Log.Info("Successfully updated Policy")

	return r.updatePolicyStatus(ctx, policy)
}

func (r *SecurityPolicyReconciler) create(ctx context.Context, policy *tykv1.SecurityPolicy) error {
	r.Log.Info("Creating  policy")

	spec, err := r.spec(ctx, policy)
	if err != nil {
		return err
	}

	// Check if policy exists. During migration, policy exists on the Dashboard but not in the k8s environment. Therefore,
	// although policy.status.ID is an empty string, which triggers policy create API call, we cannot create a policy on
	// the dashboard due to duplicated policy name. To resolve this problem, check if policy exists before creating it.
	// If the policy exists, just update it with spec. Otherwise, create it.
	if _, err = klient.Universal.Portal().Policy().Get(ctx, spec.ID); err != nil {
		err = klient.Universal.Portal().Policy().Create(ctx, spec)
		if err != nil {
			r.Log.Error(err, "Failed to create policy")

			return err
		}
	} else {
		if err := klient.Universal.Portal().Policy().Update(ctx, spec); err != nil {
			r.Log.Error(
				err,
				"Failed to update policy",
				"Name", spec.Name,
			)
		}
	}

	err = r.updateStatusOfLinkedAPIs(ctx, policy, false)
	if err != nil {
		r.Log.Error(err, "failed to update linkedAPI status")
	}

	r.Log.Info("Successfully created Policy")

	policy.Spec.MID = spec.MID

	return r.updatePolicyStatus(ctx, policy)
}

// updatePolicyStatus updates the status of the policy.
func (r *SecurityPolicyReconciler) updatePolicyStatus(ctx context.Context, policy *tykv1.SecurityPolicy) error {
	r.Log.Info("Updating policy status")

	policy.Status.PolID = policy.Spec.MID

	if policy.Spec.AccessRightsArray != nil && len(policy.Spec.AccessRightsArray) > 0 {
		policy.Status.LinkedAPIs = make([]model.Target, 0)
	} else {
		policy.Status.LinkedAPIs = nil
	}

	for _, v := range policy.Spec.AccessRightsArray {
		target := model.Target{Name: v.Name, Namespace: v.Namespace}

		policy.Status.LinkedAPIs = append(policy.Status.LinkedAPIs, target)
	}

	return r.Status().Update(ctx, policy)
}

// updateStatusOfLinkedAPIs updates the status of api definitions associated with this
// policy.
func (r *SecurityPolicyReconciler) updateStatusOfLinkedAPIs(ctx context.Context, policy *tykv1.SecurityPolicy,
	policyDeleted bool,
) error {
	r.Log.Info("Updating linked api definitions")

	ns := model.Target{
		Namespace: policy.Namespace, Name: policy.Name,
	}

	// Remove links from api definitions
	for _, t := range policy.Status.LinkedAPIs {
		api := &tykv1.ApiDefinition{}

		if err := r.Get(ctx, types.NamespacedName{Name: t.Name, Namespace: t.Namespace}, api); err != nil {
			r.Log.Error(err, "Failed to remove link from api definition")

			return err
		}

		api.Status.LinkedByPolicies = removeTarget(api.Status.LinkedByPolicies, ns)

		if err := r.Status().Update(ctx, api); err != nil {
			r.Log.Error(err, "Failed to remove link from api definition")

			return err
		}
	}

	for _, a := range policy.Spec.AccessRightsArray {
		api := &tykv1.ApiDefinition{}

		if err := r.Get(ctx, types.NamespacedName{Name: a.Name, Namespace: a.Namespace}, api); err != nil {
			r.Log.Error(err, "Failed to get linked api definition")

			return err
		}

		if policyDeleted {
			api.Status.LinkedByPolicies = removeTarget(api.Status.LinkedByPolicies, ns)
		} else {
			api.Status.LinkedByPolicies = addTarget(api.Status.LinkedByPolicies, ns)
		}

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
