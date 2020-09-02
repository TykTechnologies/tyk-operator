/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"encoding/json"

	"github.com/TykTechnologies/tyk-operator/internal/gateway_client"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	tykv1 "github.com/TykTechnologies/tyk-operator/api/v1"
)

// ApiDefinitionReconciler reconciles a ApiDefinition object
type ApiDefinitionReconciler struct {
	client.Client
	Log             logr.Logger
	Scheme          *runtime.Scheme
	UniversalClient *gateway_client.Client
}

// +kubebuilder:rbac:groups=tyk.tyk.io,resources=apidefinitions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=tyk.tyk.io,resources=apidefinitions/status,verbs=get;update;patch
func (r *ApiDefinitionReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	apiID := req.NamespacedName

	log := r.Log.WithValues("apidefinition", apiID)

	log.Info("fetching apidefinition instance")
	apiDef := &tykv1.ApiDefinition{}
	if err := r.Get(ctx, apiID, apiDef); err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("Tyk API Definition resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get API Definition")
		return ctrl.Result{}, err
	}

	const apiDefFinalizerName = "finalizers.tyk.io/apidefinition"
	if apiDef.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, if it does not have our finalizer,
		// then lets add the finalizer and update the object.
		if !containsString(apiDef.ObjectMeta.Finalizers, apiDefFinalizerName) {
			apiDef.ObjectMeta.Finalizers = append(apiDef.ObjectMeta.Finalizers, apiDefFinalizerName)
			if err := r.Update(ctx, apiDef); err != nil {
				return reconcile.Result{}, err
			}
		}
	} else {
		// The object is being deleted
		if containsString(apiDef.ObjectMeta.Finalizers, apiDefFinalizerName) {
			// our finalizer is present, so lets handle our external dependency
			if err := r.UniversalClient.Api.Delete(apiID.String()); err != nil {
				return reconcile.Result{Requeue: true}, err
			}

			_ = r.UniversalClient.HotReload()

			// remove our finalizer from the list and update it.
			apiDef.ObjectMeta.Finalizers = removeString(apiDef.ObjectMeta.Finalizers, apiDefFinalizerName)
			if err := r.Update(context.Background(), apiDef); err != nil {
				return reconcile.Result{}, err
			}
		}

		// Our finalizer has finished, so the reconciler can do nothing.
		return reconcile.Result{}, nil
	}

	allAPIs, err := r.UniversalClient.Api.All()
	if err != nil {
		log.Error(err, "unable to get all apis")
		return ctrl.Result{}, err
	}

	log.Info("getting all apis from gateway")
	jsBytes, _ := json.MarshalIndent(allAPIs, "", "  ")
	log.Info(string(jsBytes))

	log.Info("getting api cr")
	jsBytes, _ = json.MarshalIndent(apiDef, "", "  ")
	log.Info(string(jsBytes))

	if err := r.UniversalClient.HotReload(); err != nil {
		log.Error(err, "unable to hot reload")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *ApiDefinitionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tykv1.ApiDefinition{}).
		Complete(r)
}
