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

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	util "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/TykTechnologies/tyk-operator/api/model"
	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	tykv1alpha1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/pkg/environmet"
	"github.com/TykTechnologies/tyk-operator/pkg/keys"
)

// APIDescriptionReconciler reconciles a APIDescription object
type APIDescriptionReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	Env    environmet.Env
}

//+kubebuilder:rbac:groups=tyk.tyk.io,resources=apidescriptions,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=tyk.tyk.io,resources=apidescriptions/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=tyk.tyk.io,resources=apidescriptions/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the APIDescription object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *APIDescriptionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("APICatalogue", req.NamespacedName.String())

	log.Info("Reconciling APICatalogue instance")
	desired := &tykv1alpha1.APIDescription{}
	if err := r.Get(ctx, req.NamespacedName, desired); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err) // Ignore not-found errors
	}
	// set context for all api calls inside this reconciliation loop
	env, ctx := httpContext(ctx, r.Client, r.Env, desired, log)
	_, err := util.CreateOrUpdate(ctx, r.Client, desired, func() error {
		if !desired.ObjectMeta.DeletionTimestamp.IsZero() {
			return r.delete(ctx, desired, env)
		}
		util.AddFinalizer(desired, keys.APIDescriptionFinalizerName)
		return nil
	})
	return ctrl.Result{}, err
}

func (r *APIDescriptionReconciler) delete(
	ctx context.Context,
	desired *v1alpha1.APIDescription,
	env environmet.Env,
) error {
	// we find all api catalogues referencing this and update it to reflect the
	// change
	var ls v1alpha1.APICatalogueList
	err := r.List(ctx, &ls, &client.ListOptions{
		Namespace:     desired.Namespace,
		FieldSelector: fields.OneTermEqualSelector("spec.org_id", env.Org),
	})
	if err != nil {
		return err
	}
	ta := model.Target{
		Name:      desired.Name,
		Namespace: desired.Namespace,
	}
	for _, catalogue := range ls.Items {
		if err := r.updateCatalogue(ctx, &catalogue, ta); err != nil {
			return err
		}
	}
	util.RemoveFinalizer(desired, keys.APIDescriptionFinalizerName)
	return nil
}

func (r *APIDescriptionReconciler) updateCatalogue(
	ctx context.Context,
	catalogue *v1alpha1.APICatalogue,
	target model.Target,
) error {
	for _, desc := range catalogue.Spec.APIDescriptionList {
		if desc.Equal(target) {
			// Update this catalogue
			catalogue.Spec.APIDescriptionList =
				removeTarget(catalogue.Spec.APIDescriptionList, target)
			return r.Update(ctx, catalogue)
		}
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *APIDescriptionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tykv1alpha1.APIDescription{}).
		Complete(r)
}
