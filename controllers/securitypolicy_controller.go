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

		if policy.Spec.ID == nil || *policy.Spec.ID == "" {
			if policy.Spec.ID == nil {
				policy.Spec.ID = new(string)
			}

			polID := EncodeNS(ns)
			policy.Spec.ID = &polID
		}

		if policy.Spec.OrgID == nil || *policy.Spec.OrgID == "" {
			if policy.Spec.OrgID == nil {
				policy.Spec.OrgID = new(string)
			}
		}

		orgID := env.Org
		policy.Spec.OrgID = &orgID

		if policy.Status.PolID == "" {
			return r.create(ctx, policy)
		}

		newSpec, err := r.update(ctx, policy)
		if err != nil {
			return err
		}

		policy.Spec.SecurityPolicySpec = *newSpec
		return nil
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

		if spec.AccessRights == nil {
			spec.AccessRights = map[string]model.AccessDefinition{}
		}

		// Set AccessRights for Tyk OSS.
		spec.AccessRights[*spec.AccessRightsArray[i].APIID] = *spec.AccessRightsArray[i]
	}

	return spec, nil
}

// updateAccess updates given AccessDefinition's APIID and APIName fields based on ApiDefinition CR that is referred
// in the AccessDefinition. So that, it includes k8s details of the referred ApiDefinitions.
func (r *SecurityPolicyReconciler) updateAccess(ctx context.Context, ad *model.AccessDefinition) error {
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
			"Failed to find ApiDefinition on Tyk", "ApiDefinition", client.ObjectKeyFromObject(api),
		)

		return opclient.ErrNotFound
	}

	if ad.APIID == nil {
		ad.APIID = new(string)
	}

	if ad.APIName == nil {
		ad.APIName = new(string)
	}

	*ad.APIID = api.Status.ApiID
	*ad.APIName = api.Spec.Name

	return nil
}

func (r *SecurityPolicyReconciler) delete(ctx context.Context, policy *tykv1.SecurityPolicy) error {
	r.Log.Info("Deleting a policy", "policy", client.ObjectKeyFromObject(policy))

	if r.Env.Mode == "pro" {
		all, err := klient.Universal.Portal().Catalogue().Get(ctx)
		if err != nil {
			return err
		}

		for i := 0; i < len(all.APIS); i++ {
			if all.APIS[i].PolicyID == policy.Status.PolID {
				return fmt.Errorf("cannot delete policy due to catalogue %q dependency", all.Id)
			}
		}
	}

	util.RemoveFinalizer(policy, policyFinalizer)

	_, errTyk := klient.Universal.Portal().Policy().Get(ctx, policy.Status.PolID)
	if !opclient.IsNotFound(errTyk) {
		err := klient.Universal.Portal().Policy().Delete(ctx, policy.Status.PolID)
		if err != nil {
			r.Log.Error(err, "Failed to delete SecurityPolicy from Tyk",
				"Policy", client.ObjectKeyFromObject(policy).String(),
			)

			return err
		}

		err = klient.Universal.HotReload(ctx)
		if err != nil {
			r.Log.Error(err, "Failed to hot-reload Tyk after deleting a Policy",
				"Policy", client.ObjectKeyFromObject(policy),
			)

			return err
		}
	}

	err := r.updateStatusOfLinkedAPIs(ctx, policy, true)
	if err != nil {
		return err
	}

	r.Log.Info("Successfully deleted Policy")

	return nil
}

func (r *SecurityPolicyReconciler) update(ctx context.Context,
	policy *tykv1.SecurityPolicy,
) (*model.SecurityPolicySpec, error) {
	r.Log.Info("Updating SecurityPolicy", "Policy ID", policy.Status.PolID)

	if policy.Spec.MID == nil {
		policy.Spec.MID = new(string)
	}

	*policy.Spec.MID = policy.Status.PolID

	if policy.Spec.ID == nil {
		policy.Spec.ID = new(string)
	}

	*policy.Spec.ID = policy.Status.PolID

	spec, err := r.spec(ctx, &policy.Spec)
	if err != nil {
		return nil, err
	}

	// If SecurityPolicy does not exist on Tyk Side, Tyk Operator must create a Security
	// Policy on Tyk based on k8s state. So, unintended deletions from Dashboard can be avoided.
	specTyk, err := klient.Universal.Portal().Policy().Get(ctx, policy.Status.PolID)
	if err == nil {
		if isSame(policy.Status.LatestCRDSpecHash, spec) && isSame(policy.Status.LatestTykSpecHash, specTyk) {
			// TODO(buraksekili): needs refactoring - no need for code duplication.
			err = r.updateStatusOfLinkedAPIs(ctx, policy, false)
			if err != nil {
				return nil, err
			}

			return &spec.SecurityPolicySpec, r.updatePolicyStatus(ctx, policy, nil)
		}

		err = klient.Universal.Portal().Policy().Update(ctx, spec)
		if err != nil {
			r.Log.Error(err, "Failed to update policy on Tyk")
			return nil, err
		}
	} else {
		if opclient.IsNotFound(err) {
			err = klient.Universal.Portal().Policy().Create(ctx, spec)
			if err != nil {
				r.Log.Error(err, "Failed to re-create Policy on Tyk",
					"SecurityPolicy", client.ObjectKeyFromObject(policy),
				)

				return nil, err
			}

			if policy.Spec.MID == nil {
				policy.Spec.MID = new(string)
			}

			*policy.Spec.MID = *spec.MID
			policy.Status.PolID = *spec.MID
		} else {
			r.Log.Error(err, "Failed to get Policy from Tyk", err)

			return nil, err
		}
	}

	err = klient.Universal.HotReload(ctx)
	if err != nil {
		r.Log.Error(err, "Failed to hot-reload Tyk after updating the Policy",
			"SecurityPolicy", client.ObjectKeyFromObject(policy),
		)

		return nil, err
	}

	err = r.updateStatusOfLinkedAPIs(ctx, policy, false)
	if err != nil {
		return nil, err
	}

	polOnTyk, _ := klient.Universal.Portal().Policy().Get(ctx, *policy.Spec.MID) //nolint:errcheck

	r.Log.Info("Successfully updated Policy")

	return &spec.SecurityPolicySpec, r.updatePolicyStatus(ctx, policy, func(status *tykv1.SecurityPolicyStatus) {
		status.LatestTykSpecHash = calculateHash(polOnTyk)
		status.LatestCRDSpecHash = calculateHash(spec)
	})
}

func (r *SecurityPolicyReconciler) create(ctx context.Context, policy *tykv1.SecurityPolicy) error {
	r.Log.Info("Creating a policy")

	spec, err := r.spec(ctx, &policy.Spec)
	if err != nil {
		return err
	}

	// Check if policy exists. During migration, policy exists on the Dashboard but not in the k8s environment.
	// If policy does not exist on Tyk side, create it. Otherwise, update it based on Kubernetes spec because
	// creating a Policy with duplicated name causes API call errors.
	existingSpec, err := klient.Universal.Portal().Policy().Get(ctx, *policy.Spec.ID)
	if err != nil || existingSpec == nil || existingSpec.MID == nil || *existingSpec.MID == "" {
		r.Log.Info("Creating a new policy")

		err = klient.Universal.Portal().Policy().Create(ctx, spec)
		if err != nil {
			r.Log.Error(
				err,
				"Failed to create policy on Tyk",
				"Policy", client.ObjectKeyFromObject(policy),
			)

			return err
		}
	} else {
		if spec.MID == nil {
			spec.MID = new(string)
		}

		*spec.MID = *existingSpec.MID

		err = klient.Universal.Portal().Policy().Update(ctx, spec)
		if err != nil {
			r.Log.Error(
				err,
				"Failed to update policy on Tyk",
				"Policy", client.ObjectKeyFromObject(policy),
			)

			return err
		}
	}

	err = klient.Universal.HotReload(ctx)
	if err != nil {
		r.Log.Error(err, "Failed to hot-reload Tyk after creating a Policy",
			"Policy", client.ObjectKeyFromObject(policy),
		)

		return err
	}

	r.Log.Info("Successfully created Policy")

	if policy.Spec.MID == nil {
		policy.Spec.MID = new(string)
	}

	*policy.Spec.MID = *spec.MID
	*policy.Spec.ID = *spec.MID

	err = r.updateStatusOfLinkedAPIs(ctx, policy, false)
	if err != nil {
		r.Log.Error(err,
			"failed to update linkedAPI status",
			"Policy", client.ObjectKeyFromObject(policy),
		)

		return err
	}

	polOnTyk, _ := klient.Universal.Portal().Policy().Get(ctx, *spec.MID) //nolint:errcheck

	return r.updatePolicyStatus(ctx, policy, func(status *tykv1.SecurityPolicyStatus) {
		status.LatestTykSpecHash = calculateHash(polOnTyk)
		status.LatestCRDSpecHash = calculateHash(spec)
	})
}

// updatePolicyStatus updates the status of the policy.
func (r *SecurityPolicyReconciler) updatePolicyStatus(
	ctx context.Context,
	policy *tykv1.SecurityPolicy,
	fn func(status *tykv1.SecurityPolicyStatus),
) error {
	r.Log.Info("Updating policy status")

	if policy.Spec.MID != nil {
		policy.Status.PolID = *policy.Spec.MID
	}

	if policy.Spec.AccessRightsArray != nil && len(policy.Spec.AccessRightsArray) > 0 {
		policy.Status.LinkedAPIs = make([]model.Target, 0)
	} else {
		policy.Status.LinkedAPIs = nil
	}

	for _, v := range policy.Spec.AccessRightsArray {
		namespace := v.Namespace
		target := model.Target{Name: v.Name, Namespace: &namespace}

		policy.Status.LinkedAPIs = append(policy.Status.LinkedAPIs, target)
	}

	if fn != nil {
		fn(&policy.Status)
	}

	return r.Status().Update(ctx, policy)
}

// updateStatusOfLinkedAPIs updates the status of api definitions associated with this
// policy.
func (r *SecurityPolicyReconciler) updateStatusOfLinkedAPIs(ctx context.Context, policy *tykv1.SecurityPolicy,
	policyDeleted bool,
) error {
	r.Log.Info("Updating linked api definitions")

	namespace := policy.Namespace

	target := model.Target{
		Namespace: &namespace, Name: policy.Name,
	}

	// Remove links from api definitions
	for _, t := range policy.Status.LinkedAPIs {
		api := &tykv1.ApiDefinition{}

		namespace := ""
		if t.Namespace != nil {
			namespace = *t.Namespace
		}

		if err := r.Get(ctx, types.NamespacedName{Name: t.Name, Namespace: namespace}, api); err != nil {
			r.Log.Error(err, "Failed to get the linked API", "api", t.String())

			return err
		}

		api.Status.LinkedByPolicies = removeTarget(api.Status.LinkedByPolicies, target)

		if err := r.Status().Update(ctx, api); err != nil {
			r.Log.Error(err, "Failed to update status of linked api definition", "api", t.String())

			return err
		}
	}

	for _, a := range policy.Spec.AccessRightsArray {
		api := &tykv1.ApiDefinition{}

		name := types.NamespacedName{Name: a.Name, Namespace: a.Namespace}

		if err := r.Get(ctx, types.NamespacedName{Name: a.Name, Namespace: a.Namespace}, api); err != nil {
			r.Log.Error(err, "Failed to get linked api definition", "api", name)

			return err
		}

		if policyDeleted {
			api.Status.LinkedByPolicies = removeTarget(api.Status.LinkedByPolicies, target)
		} else {
			api.Status.LinkedByPolicies = addTarget(api.Status.LinkedByPolicies, target)
		}

		if err := r.Status().Update(ctx, api); err != nil {
			r.Log.Error(err, "Failed to update status of linked api definition", "api", name)

			return err
		}

		// update the JWT default policy of the api
		apiOnTyk, err := klient.Universal.Api().Get(ctx, EncodeNS(target.String()))
		if err != nil {
			r.Log.Error(
				err, "Failed to get ApiDefinition on Tyk",
				"ApiDefinition", target.String(),
			)
			return err
		}
		if apiOnTyk.JWTDefaultPolicies == nil || len(apiOnTyk.JWTDefaultPolicies) == 0 {
			apiOnTyk.JWTDefaultPolicies = make([]string, 0)
		}
		AddUniqueElement(&apiOnTyk.JWTDefaultPolicies, *policy.Spec.MID)
		_, err = klient.Universal.Api().Update(ctx, apiOnTyk)
		if err != nil {
			r.Log.Error(
				err, "Failed to update ApiDefinition on Tyk",
				"ApiDefinition", target.String(),
			)
			return err
		}
		r.Log.Info("Successfully updated JWT default policies", "ApiDefinition", target.String())
	}

	return nil
}

// SetupWithManager initializes the security policy controller.
func (r *SecurityPolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tykv1.SecurityPolicy{}).
		Complete(r)
}
