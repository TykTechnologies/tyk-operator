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
	"time"

	tykv1 "github.com/TykTechnologies/tyk-operator/api/v1"
	"github.com/TykTechnologies/tyk-operator/internal/gateway_client"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
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
			if err := r.UniversalClient.Api.Delete(apiIDEncode(apiID.String())); err != nil {
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

	newSpec := &apiDef.Spec

	// TODO: this belongs in webhook or CR will be wrong
	newSpec.APIID = apiIDEncode(apiID.String())
	r.applyDefaults(newSpec)

	// find the api definition object
	found := &tykv1.APIDefinitionSpec{}
	for _, api := range allAPIs {
		if api.APIID == apiIDEncode(apiID.String()) {
			found = &api
		}
	}

	// we didn't find it, so let's create it
	if found.APIID == "" {
		log.Info("creating api", "decodedID", apiID.String(), "encodedID", apiIDEncode(apiID.String()))
		_, err := r.UniversalClient.Api.Create(newSpec)
		if err != nil {
			log.Error(err, "unable to create API Definition")
			return ctrl.Result{RequeueAfter: time.Second * 5}, err
		}

		err = r.UniversalClient.HotReload()
		if err != nil {
			// TODO: what should we actually do here?
			log.Error(err, "unable to hotreload, but API created. Inconsistent state")
			return ctrl.Result{Requeue: false}, err
		}

		return ctrl.Result{}, nil
	}

	// we found it, so let's update it
	log.Info("updating api", "decoded", apiID.String(), "encoded", apiIDEncode(apiID.String()))
	err = r.UniversalClient.Api.Update(newSpec)
	if err != nil {
		log.Error(err, "unable to update API Definition")
		return ctrl.Result{Requeue: true}, err
	}

	err = r.UniversalClient.HotReload()
	if err != nil {
		// TODO: what should we actually do here?
		log.Error(err, "unable to hotreload, but API successfully updated. Inconsistent state")
		return ctrl.Result{Requeue: false}, err
	}

	return ctrl.Result{}, nil
}

func (r *ApiDefinitionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tykv1.ApiDefinition{}).
		Complete(r)
}

func (r *ApiDefinitionReconciler) applyDefaults(spec *tykv1.APIDefinitionSpec) {
	if len(spec.VersionData.Versions) == 0 {
		defaultVersionData := tykv1.VersionData{
			NotVersioned:   true,
			DefaultVersion: "Default",
			Versions: map[string]tykv1.VersionInfo{
				"Default": {
					Name:                        "Default",
					Expires:                     "",
					Paths:                       tykv1.VersionInfoPaths{},
					UseExtendedPaths:            false,
					ExtendedPaths:               tykv1.ExtendedPathsSet{},
					GlobalHeaders:               nil,
					GlobalHeadersRemove:         nil,
					GlobalResponseHeaders:       nil,
					GlobalResponseHeadersRemove: nil,
					IgnoreEndpointCase:          false,
					GlobalSizeLimit:             0,
					OverrideTarget:              "",
				},
			},
		}

		spec.VersionData = defaultVersionData
	}
}
