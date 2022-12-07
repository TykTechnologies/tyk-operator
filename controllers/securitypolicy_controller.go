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

		if policy.Status.PolID == "" {
			newSpec, err := r.create(ctx, policy)
			if err != nil {
				return err
			}

			policy.Spec = *newSpec
			return nil
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

// spec returns a copy of SecurityPolicySpec with AccessRightsArray updated. As a result, each AccessRightsArray
// element contains the K8s details for each ApiDefinition. It returns an error if ApiDefinition does not exist
// in the K8s environment.
func (r *SecurityPolicyReconciler) spec(
	ctx context.Context,
	s *tykv1.SecurityPolicySpec,
) (*tykv1.SecurityPolicySpec, error) {
	spec := s.DeepCopy()
	spec.Context = nil

	for i := 0; i < len(spec.AccessRightsArray); i++ {
		err := r.updateAccess(ctx, spec.AccessRightsArray[i])
		if err != nil {
			return nil, err
		}
	}

	return spec, nil
}

// updateAccess updates given AccessDefinition's APIID and APIName fields based on ApiDefinition CR that is referred
// in the AccessDefinition. So that, it includes k8s details of the referred ApiDefinitions.
func (r *SecurityPolicyReconciler) updateAccess(ctx context.Context, ad *tykv1.AccessDefinition) error {
	api := &tykv1.ApiDefinition{}
	if err := r.Get(ctx, types.NamespacedName{Name: ad.Name, Namespace: ad.Namespace}, api); err != nil {
		if errors.IsNotFound(err) {
			r.Log.Info("ApiDefinition resource not found. Unable to attach to SecurityPolicy. ReQueue",
				"Name", ad.Name,
				"Namespace", ad.Namespace,
			)

			return err
		}

		r.Log.Error(err, "Failed to get APIDefinition to attach to SecurityPolicy")

		return err
	}

	if api.Status.ApiID == "" {
		r.Log.Error(
			opclient.ErrNotFound,
			"ApiDefinition does not exist on Tyk", "ApiDefinition", client.ObjectKeyFromObject(api),
		)

		return opclient.ErrNotFound
	}

	ad.APIID = api.Status.ApiID
	ad.APIName = api.Spec.Name

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

	err = r.updateLinkedAPI(ctx, policy, func(ads *tykv1.ApiDefinitionStatus, ns model.Target) {
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

	spec, err := r.spec(ctx, &policy.Spec)
	if err != nil {
		return err
	}

	// If SecurityPolicy does not exist on Tyk Side, Tyk Operator must create a Security
	// Policy on Tyk based on k8s state. So, unintended deletions from Dashboard can be avoided.
	_, err = klient.Universal.Portal().Policy().Get(ctx, policy.Status.PolID)
	if err == nil {
		err = klient.Universal.Portal().Policy().Update(ctx, spec)
		if err != nil {
			r.Log.Error(err, "Failed to update policy")
			return err
		}
	} else {
		err = klient.Universal.Portal().Policy().Create(ctx, spec)
		if err != nil {
			r.Log.Error(err, "failed to re-create Policy",
				"SecurityPolicy", client.ObjectKeyFromObject(policy),
			)

			return err
		}
	}

	err = r.updateLinkedAPI(ctx, policy, func(ads *tykv1.ApiDefinitionStatus, s model.Target) {
		ads.LinkedByPolicies = addTarget(ads.LinkedByPolicies, s)
	})

	if err != nil {
		return err
	}

	klient.Universal.HotReload(ctx)
	r.Log.Info("Successfully updated Policy")

	return nil
}

func (r *SecurityPolicyReconciler) create(
	ctx context.Context,
	policy *tykv1.SecurityPolicy,
) (*tykv1.SecurityPolicySpec, error) {
	r.Log.Info("Creating a policy")

	// Check if policy exists. During migration, policy exists on the Dashboard but not in the k8s environment.
	// If policy does not exist on Tyk side, create it. Otherwise, get AccessRightsArray from SecurityPolicy CR
	// and update existing policy's spec.
	spec, err := klient.Universal.Portal().Policy().Get(ctx, policy.Spec.ID)
	if err != nil || spec == nil || spec.MID == "" {
		spec, err = r.spec(ctx, &policy.Spec)
		if err != nil {
			return nil, err
		}

		r.Log.Info("Creating a new policy")

		err = klient.Universal.Portal().Policy().Create(ctx, spec)
		if err != nil {
			r.Log.Error(err, "Failed to create policy", "policy", client.ObjectKeyFromObject(policy))

			return nil, err
		}

		policy.Spec.MID = spec.MID
	} else {
		spec.AccessRightsArray = policy.Spec.AccessRightsArray

		spec, err = r.spec(ctx, spec)
		if err != nil {
			return nil, err
		}

		spec.ID = policy.Spec.ID

		if err := klient.Universal.Portal().Policy().Update(ctx, spec); err != nil {
			r.Log.Error(
				err,
				"Failed to update policy",
				"Policy", client.ObjectKeyFromObject(policy),
			)

			return nil, err
		}
	}

	err = r.updateLinkedAPI(ctx, policy, func(ads *tykv1.ApiDefinitionStatus, target model.Target) {
		ads.LinkedByPolicies = addTarget(ads.LinkedByPolicies, target)
	})
	if err != nil {
		r.Log.Error(err,
			"failed to update linkedAPI status",
			"policy", client.ObjectKeyFromObject(policy),
		)

		return nil, err
	}

	r.Log.Info("Successfully created Policy")

	policy.Status.PolID = policy.Spec.MID

	err = r.Status().Update(ctx, policy)
	if err != nil {
		return nil, err
	}

	return spec, nil
}

// updateLinkedAPI updates the status of api definitions associated with this policy.
func (r *SecurityPolicyReconciler) updateLinkedAPI(ctx context.Context, policy *tykv1.SecurityPolicy,
	fn func(*tykv1.ApiDefinitionStatus, model.Target),
) error {
	r.Log.Info("Updating linked api definitions")

	ns := model.Target{
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
