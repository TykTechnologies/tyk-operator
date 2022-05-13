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

	"github.com/TykTechnologies/tyk-operator/pkg/keys"
	util "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"k8s.io/apimachinery/pkg/fields"

	tykv1alpha1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"

	"github.com/go-logr/logr"

	"github.com/TykTechnologies/tyk-operator/pkg/environmet"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	ErrSubgrapReferenceExistsInGraph  = errors.New("subgraph is referenced in supergraph")
	ErrSubgrapReferenceExistsInApiDef = errors.New("subgraph is referenced in apiDefinition")
)

// SubGraphReconciler reconciles a SubGraph object
type SubGraphReconciler struct {
	client.Client
	Log    logr.Logger
	Env    environmet.Env
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=tyk.tyk.io,resources=subgraphs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=tyk.tyk.io,resources=supergraphs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=tyk.tyk.io,resources=apidefinitions,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=tyk.tyk.io,resources=subgraphs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=tyk.tyk.io,resources=subgraphs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *SubGraphReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	namespacedName := req.NamespacedName
	log := r.Log.WithValues("SubGraph", namespacedName.String())
	log.Info("Reconciling SuperGraph instance")

	desired := &tykv1alpha1.SubGraph{}
	if err := r.Get(ctx, req.NamespacedName, desired); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err) // Ignore not-found errors
	}

	// subgraph is marked for deletion
	if !desired.ObjectMeta.DeletionTimestamp.IsZero() {
		// Check if subgraph is referenced in any supergraph.
		superGraphList := &tykv1alpha1.SuperGraphList{}
		listOps := &client.ListOptions{
			FieldSelector: fields.OneTermEqualSelector(SubgraphField, desired.GetName()),
			Namespace:     desired.GetNamespace(),
		}
		if err := r.List(ctx, superGraphList, listOps); err != nil {
			return ctrl.Result{}, err
		}

		if len(superGraphList.Items) != 0 {
			return ctrl.Result{}, ErrSubgrapReferenceExistsInGraph
		}

		// Check if subgraph is referenced in any apidefinition.
		apiDefinitionList := &tykv1alpha1.ApiDefinitionList{}
		listOps = &client.ListOptions{
			FieldSelector: fields.OneTermEqualSelector(GraphKey, desired.GetName()),
			Namespace:     desired.GetNamespace(),
		}
		if err := r.List(ctx, apiDefinitionList, listOps); err != nil {
			return ctrl.Result{}, err
		}

		if len(apiDefinitionList.Items) != 0 {
			return ctrl.Result{}, ErrSubgrapReferenceExistsInApiDef
		}

		util.RemoveFinalizer(desired, keys.SubGraphFinalizerName)

		if err := r.Update(ctx, desired); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	if !util.ContainsFinalizer(desired, keys.SubGraphFinalizerName) {
		util.AddFinalizer(desired, keys.SubGraphFinalizerName)
	}

	err := r.Update(ctx, desired)
	return ctrl.Result{}, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *SubGraphReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tykv1alpha1.SubGraph{}).
		Complete(r)
}
