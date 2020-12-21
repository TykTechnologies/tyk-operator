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
	"os"

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
	util "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const policyFinalizer = "finalizers.tyk.io/securitypolicy"

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

// Reconcile reconciles SecurityPolicy custom resources
func (r *SecurityPolicyReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("SecurityPolicy", req.NamespacedName.String())

	ns := req.NamespacedName.String()
	log.Info("Reconciling SecurityPolicy instance")

	// Lookup policy object
	policy := &tykv1.SecurityPolicy{}
	if err := r.Get(ctx, req.NamespacedName, policy); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	op, err := util.CreateOrUpdate(ctx, r.Client, policy, func() error {
		if !policy.ObjectMeta.DeletionTimestamp.IsZero() {
			if util.ContainsFinalizer(policy, policyFinalizer) {
				util.RemoveFinalizer(policy, policyFinalizer)
				if err := r.delete(ctx, policy.Status.PolID); err != nil {
					if universal_client.IsNotFound(err) {
						r.Log.Info("Policy not found")
						return nil
					}
					r.Log.Error(err, "Failed to delete resource")
					return err
				}
				r.Log.Info("Successfully deleted Policy")
			}
			return nil
		}
		if !util.ContainsFinalizer(policy, policyFinalizer) {
			util.AddFinalizer(policy, policyFinalizer)
		}
		if policy.Spec.ID == "" {
			policy.Spec.ID = encodeNS(ns)
		}
		if policy.Spec.OrgID == "" {
			policy.Spec.OrgID = os.Getenv(environmet.TykORG)
		}
		// update access rights
		r.Log.Info("updating access rights")
		for i := 0; i < len(policy.Spec.AccessRightsArray); i++ {
			a := &policy.Spec.AccessRightsArray[i]
			err := r.updateAccess(ctx, a)
			if err != nil {
				if errors.IsNotFound(err) {
					r.Log.Info("APIDefinition resource was not found",
						"Name", a.Name,
						"Namespace", a.Namespace,
					)
					return nil
				}
				return err
			}
		}
		if policy.Status.PolID == "" {
			r.Log.Info("Creating  policy")
			err := r.create(ctx, &policy.Spec)
			if err != nil {
				r.Log.Error(err, "Failed to create policy")
				return err
			}
			r.Log.Info("Successful created Policy")
			policy.Status.PolID = policy.Spec.MID
			return r.Status().Update(ctx, policy)
		}
		r.Log.Info("Updating  policy")
		policy.Spec.MID = policy.Status.PolID
		err := r.update(ctx, &policy.Spec)
		if err != nil {
			r.Log.Error(err, "Failed to update policy")
			return err
		}
		r.Log.Info("Successfully updated Policy")
		return nil
	})
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		r.Log.Error(err, "Failed create or update security policy ", "Op", op)
		return ctrl.Result{}, err
	}
	r.Log.Info("Nothing changed", "Op", op)
	return ctrl.Result{}, nil
}

func (r *SecurityPolicyReconciler) updateAccess(ctx context.Context,
	a *tykv1.AccessDefinition) error {
	api := &tykv1.ApiDefinition{}
	if err := r.Get(ctx, types.NamespacedName{Name: a.Name, Namespace: a.Namespace}, api); err != nil {
		return err
	}
	def, err := r.UniversalClient.Api().Get(api.Status.ApiID)
	if err != nil {
		return err
	}
	a.APIID = def.APIID
	a.APIName = def.Name
	return nil
}

// deletes policy with id and ensure hot reload is called after the operation
func (r *SecurityPolicyReconciler) delete(ctx context.Context, id string) error {
	defer func() {
		r.UniversalClient.HotReload()
	}()
	return r.UniversalClient.SecurityPolicy().Delete(id)
}

// update updates policy using spec and ensure hot reload is called.
func (r *SecurityPolicyReconciler) update(ctx context.Context, spec *tykv1.SecurityPolicySpec) error {
	defer func() {
		r.UniversalClient.HotReload()
	}()
	return r.UniversalClient.SecurityPolicy().Update(spec)
}

func (r *SecurityPolicyReconciler) create(ctx context.Context, spec *tykv1.SecurityPolicySpec) error {
	defer func() {
		r.UniversalClient.HotReload()
	}()
	return r.UniversalClient.SecurityPolicy().Create(spec)
}

func (r *SecurityPolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tykv1.SecurityPolicy{}).
		Complete(r)
}
