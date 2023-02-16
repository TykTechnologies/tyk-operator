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

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	util "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/TykTechnologies/tyk-operator/api/model"
	tykv1alpha1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	tykclient "github.com/TykTechnologies/tyk-operator/pkg/client"
	"github.com/TykTechnologies/tyk-operator/pkg/client/klient"
	"github.com/TykTechnologies/tyk-operator/pkg/environmet"
	"github.com/TykTechnologies/tyk-operator/pkg/keys"
	v1 "k8s.io/api/core/v1"
)

// TykOASApiDefinitionReconciler reconciles a TykOASApiDefinition object
type TykOASApiDefinitionReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	Env    environmet.Env
}

//+kubebuilder:rbac:groups=tyk.tyk.io,resources=tykoasapidefinitions,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=tyk.tyk.io,resources=tykoasapidefinitions/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=tyk.tyk.io,resources=tykoasapidefinitions/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the TykOASApiDefinition object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *TykOASApiDefinitionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("tykoasapidefinition", req.NamespacedName)

	log.Info("Reconciling TykOASApiDefinition")

	tykOASDef := &tykv1alpha1.TykOASApiDefinition{}

	err := r.Client.Get(ctx, req.NamespacedName, tykOASDef)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	_, ctx, err = HttpContext(ctx, r.Client, r.Env, tykOASDef, log)
	if err != nil {
		return ctrl.Result{}, err
	}

	_, err = util.CreateOrUpdate(ctx, r.Client, tykOASDef, func() error {
		if !tykOASDef.ObjectMeta.DeletionTimestamp.IsZero() {
			err := r.delete(ctx, tykOASDef)
			if err != nil {
				return err
			}

			util.RemoveFinalizer(tykOASDef, keys.TykOASApiDefinitionFinalizerName)

			return nil
		}

		util.AddFinalizer(tykOASDef, keys.TykOASApiDefinitionFinalizerName)

		id, err := r.createOrUpdateTykOAS(ctx, *tykOASDef)
		if err != nil {
			return err
		}

		if tykOASDef.Status.ApiID == "" {
			tykOASDef.Status.ApiID = id

			err = r.Client.Status().Update(ctx, tykOASDef)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		log.Error(err, "Failed to create/update TykOASApiDefinition")
		return ctrl.Result{RequeueAfter: queueAfter}, err
	}

	return ctrl.Result{}, nil
}

func (r *TykOASApiDefinitionReconciler) createOrUpdateTykOAS(ctx context.Context, tykOASDef tykv1alpha1.TykOASApiDefinition) (string, error) {
	r.Log.Info("Creating/Updating OAS API on Tyk")

	var cm v1.ConfigMap

	cm_name := types.NamespacedName{Name: tykOASDef.Spec.OASRef.Name, Namespace: tykOASDef.Spec.OASRef.Namespace}

	err := r.Client.Get(ctx, cm_name, &cm)
	if err != nil {
		return "", err
	}

	data := cm.Data[tykOASDef.Spec.OASRef.KeyName]
	if data == "" {
		err = errors.New("OAS Spec is empty")
	}

	id := tykOASDef.Status.ApiID
	if id == "" {
		var result *model.Result

		result, err = klient.Universal.OAS().Create(ctx, data)

		id = result.Meta
	} else {
		_, err = klient.Universal.OAS().Update(ctx, id, data)
	}
	if err != nil {
		r.Log.Error(err, "Failed to create/update OAS API Definition")
		return "", err
	}

	r.Log.Info("Successfully created/updated OAS API on Tyk", "id", id)

	return id, nil
}

func (r *TykOASApiDefinitionReconciler) delete(ctx context.Context, tykOASDef *tykv1alpha1.TykOASApiDefinition) error {
	r.Log.Info("Deleting OAS API on Tyk")

	if util.ContainsFinalizer(tykOASDef, keys.TykOASApiDefinitionFinalizerName) {
		err := isLinkedToPolicies(ctx, r.Client, r.Log, tykOASDef.Status.LinkedByPolicies)
		if err != nil {
			return err
		}

		_, err = klient.Universal.OAS().Delete(ctx, tykOASDef.Status.ApiID)
		if err != nil {
			if tykclient.IsNotFound(err) {
				r.Log.Info(
					"Ignoring nonexistent OAS ApiDefinition on delete",
					"api_id", tykOASDef.Status.ApiID,
				)

				return nil
			}
			r.Log.Error(err, "Failed to delete OAS API Definition")
			return err
		}

		r.Log.Info("Successfully deleted OAS API on Tyk")
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TykOASApiDefinitionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	pred := predicate.GenerationChangedPredicate{}

	return ctrl.NewControllerManagedBy(mgr).
		For(&tykv1alpha1.TykOASApiDefinition{}).
		WithEventFilter(pred).
		Complete(r)
}
