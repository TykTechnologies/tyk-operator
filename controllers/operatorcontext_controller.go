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

	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/pkg/keys"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	util "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// OperatorContextReconciler reconciles a OperatorContext object
type OperatorContextReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

//+kubebuilder:rbac:groups=tyk.tyk.io,resources=operatorcontexts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=tyk.tyk.io,resources=operatorcontexts/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=tyk.tyk.io,resources=operatorcontexts/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the OperatorContext object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *OperatorContextReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := r.Log.WithValues("OperatorContext", req.NamespacedName.String())

	logger.Info("Reconciling OperatorContext instance")

	var list v1alpha1.ApiDefinitionList
	var desired v1alpha1.OperatorContext

	if err := r.Get(ctx, req.NamespacedName, &desired); err != nil {
		logger.Error(err, "failed to fetch operator context object")

		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !desired.DeletionTimestamp.IsZero() {
		err := r.List(ctx, &list, client.InNamespace(req.Namespace))
		if err != nil {
			logger.Error(err, "failed to fetch apidefinitions in current namespace")

			return ctrl.Result{}, err
		}

		canBeDeleted := true

		for _, apiDef := range list.Items {
			if apiDef.Spec.Context != nil && apiDef.Spec.Context.Name == desired.Name {
				canBeDeleted = false
				break
			}
		}

		if !canBeDeleted {
			logger.Error(errors.New("cannot delete operator context while it is being referenced by other resources"), "failed to delete operator context")

			return ctrl.Result{Requeue: true}, nil
		}

		util.RemoveFinalizer(&desired, keys.OperatorContextFinalizerName)

		return ctrl.Result{}, r.Update(ctx, &desired)
	}

	if !util.ContainsFinalizer(&desired, keys.OperatorContextFinalizerName) {
		logger.Info("Obj doesn't have finalizer. Adding one")
		util.AddFinalizer(&desired, keys.OperatorContextFinalizerName)

		return ctrl.Result{}, r.Update(ctx, &desired)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *OperatorContextReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		// Uncomment the following line adding a pointer to an instance of the controlled resource as an argument
		For(&v1alpha1.OperatorContext{}).
		Complete(r)
}
