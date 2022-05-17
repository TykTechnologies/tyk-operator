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

	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/TykTechnologies/tyk-operator/pkg/keys"
	util "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/TykTechnologies/tyk-operator/pkg/environmet"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	tykv1alpha1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	graphQlMerge "github.com/jensneuse/graphql-go-tools/pkg/federation/sdlmerge"
)

const SubgraphField = ".subgraphs_refs.name"

var ErrSuperGraphReference = errors.New("supergraph is referenced in apiDefinition")

// SuperGraphReconciler reconciles a SuperGraph object
type SuperGraphReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	Env    environmet.Env
}

//+kubebuilder:rbac:groups=tyk.tyk.io,resources=supergraphs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=tyk.tyk.io,resources=supergraphs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=tyk.tyk.io,resources=supergraphs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *SuperGraphReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	namespacedName := req.NamespacedName
	log := r.Log.WithValues("SuperGraph", namespacedName.String())
	log.Info("Reconciling SuperGraph instance")

	desired := &tykv1alpha1.SuperGraph{}

	if err := r.Get(ctx, req.NamespacedName, desired); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err) // Ignore not-found errors
	}

	// supergraph is marked for deletion
	if !desired.ObjectMeta.DeletionTimestamp.IsZero() {
		// Check if supergraph is referenced in any apidefinition.
		apiDefList := &tykv1alpha1.ApiDefinitionList{}
		listOps := &client.ListOptions{
			FieldSelector: fields.OneTermEqualSelector(GraphKey, desired.GetName()),
			Namespace:     desired.GetNamespace(),
		}

		if err := r.List(ctx, apiDefList, listOps); err != nil {
			return ctrl.Result{}, err
		}

		if len(apiDefList.Items) != 0 {
			return ctrl.Result{}, ErrSuperGraphReference
		}

		util.RemoveFinalizer(desired, keys.SuperGraphFinalizerName)

		if err := r.Update(ctx, desired); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	var sdls []string

	for _, subGraphRef := range desired.Spec.SubgraphsRefs {
		subGraph := &tykv1alpha1.SubGraph{}
		if err := r.Client.Get(ctx, types.NamespacedName{
			Name:      subGraphRef.Name,
			Namespace: subGraphRef.Namespace,
		}, subGraph); err != nil {
			return ctrl.Result{}, err
		}

		sdls = append(sdls, subGraph.Spec.Subgraph.SDL)
	}

	mergedSdl, err := graphQlMerge.MergeSDLs(sdls...)
	if err != nil {
		return ctrl.Result{}, err
	}

	desired.Spec.MergedSDL = mergedSdl

	if !util.ContainsFinalizer(desired, keys.SuperGraphFinalizerName) {
		util.AddFinalizer(desired, keys.SuperGraphFinalizerName)
	}

	err = r.Update(ctx, desired)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *SuperGraphReconciler) findObjectsForSubGraph(subGraph client.Object) []reconcile.Request {
	attachedSupergraphDeployments := &tykv1alpha1.SuperGraphList{}
	listOps := &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(SubgraphField, subGraph.GetName()),
	}

	if err := r.List(context.TODO(), attachedSupergraphDeployments, listOps); err != nil {
		return []reconcile.Request{}
	}

	requests := make([]reconcile.Request, len(attachedSupergraphDeployments.Items))
	for i, item := range attachedSupergraphDeployments.Items { //nolint
		requests[i] = reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      item.GetName(),
				Namespace: item.GetNamespace(),
			},
		}
	}

	return requests
}

// SetupWithManager sets up the controller with the Manager.
func (r *SuperGraphReconciler) SetupWithManager(mgr ctrl.Manager) error {
	err := mgr.GetFieldIndexer().IndexField(
		context.Background(),
		&tykv1alpha1.SuperGraph{},
		SubgraphField,
		func(object client.Object) []string {
			sg := object.(*tykv1alpha1.SuperGraph) //nolint
			if len(sg.Spec.SubgraphsRefs) == 0 {
				return nil
			}

			keyIndexes := []string{}
			for _, ref := range sg.Spec.SubgraphsRefs {
				keyIndexes = append(keyIndexes, ref.Name)
			}

			return keyIndexes
		})
	if err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&tykv1alpha1.SuperGraph{}).
		Watches(
			&source.Kind{Type: &tykv1alpha1.SubGraph{}},
			handler.EnqueueRequestsFromMapFunc(r.findObjectsForSubGraph),
			builder.WithPredicates(r.ignoreSubGraphCreationEvents()),
		).
		Complete(r)
}

func (r *SuperGraphReconciler) ignoreSubGraphCreationEvents() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return false
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return true
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return true
		},
	}
}
