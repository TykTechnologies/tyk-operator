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

	tykv1 "github.com/TykTechnologies/tyk-operator/api/v1"
	"github.com/TykTechnologies/tyk-operator/internal/universal_client"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// ApiDefinitionReconciler reconciles a ApiDefinition object
type ApiDefinitionReconciler struct {
	client.Client
	Log             logr.Logger
	Scheme          *runtime.Scheme
	UniversalClient universal_client.UniversalClient
	Recorder        record.EventRecorder
}

// +kubebuilder:rbac:groups=tyk.tyk.io,resources=apidefinitions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=tyk.tyk.io,resources=apidefinitions/status,verbs=get;update;patch
func (r *ApiDefinitionReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	apiID := req.NamespacedName
	apiIDEncoded := apiIDEncode(apiID.String())

	log := r.Log.WithValues("ApiDefinition", apiID.String())

	log.Info("fetching apidefinition instance")
	desired := &tykv1.ApiDefinition{}
	if err := r.Get(ctx, apiID, desired); err != nil {
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
	r.Recorder.Event(desired, "Normal", "ApiDefinition", "Reconciling")

	const apiDefFinalizerName = "finalizers.tyk.io/apidefinition"
	if desired.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, if it does not have our finalizer,
		// then lets add the finalizer and update the object.
		if !containsString(desired.ObjectMeta.Finalizers, apiDefFinalizerName) {
			desired.ObjectMeta.Finalizers = append(desired.ObjectMeta.Finalizers, apiDefFinalizerName)
			if err := r.Update(ctx, desired); err != nil {
				return reconcile.Result{}, err
			}
		}
	} else {
		// The object is being deleted
		if containsString(desired.ObjectMeta.Finalizers, apiDefFinalizerName) {
			// our finalizer is present, so lets handle our external dependency

			// TODO: check for any security policies that grant access to this API Definition.
			// If any policies grant access to this resource, return error and requeue
			// We need to keep doing this till:
			// 1. the policy(ies) are deleted
			// 2. the policy is edited and no longer grants access to this API

			err := r.UniversalClient.Api().Delete(apiIDEncoded)
			if err != nil {
				log.Error(err, "unable to delete api", "api_id", desired.Status.Id)
			}

			err = r.UniversalClient.HotReload()
			if err != nil {
				log.Error(err, "unable to hot reload", "api_id", desired.Status.Id)
			}

			// remove our finalizer from the list and update it.
			desired.ObjectMeta.Finalizers = removeString(desired.ObjectMeta.Finalizers, apiDefFinalizerName)
			if err := r.Update(ctx, desired); err != nil {
				return reconcile.Result{}, nil
			}
		}

		// Our finalizer has finished, so the reconciler can do nothing.
		return reconcile.Result{}, nil
	}

	newSpec := &desired.Spec

	// TODO: this belongs in webhook or CR will be wrong
	// we only care about this for OSS
	newSpec.APIID = apiIDEncoded
	r.applyDefaults(newSpec)

	_, err := universal_client.CreateOrUpdateAPI(r.UniversalClient, newSpec)
	if err != nil {
		log.Error(err, "createOrUpdate failure")
		r.Recorder.Event(desired, "Warning", "ApiDefinition", "Create or Update API Definition")
		return ctrl.Result{Requeue: true}, nil
	}

	desired.Status.Id = newSpec.APIID
	err = r.Status().Update(ctx, desired)
	if err != nil {
		log.Error(err, "Failed to update ApiDefinition status")
		r.Recorder.Event(desired, "Warning", "ApiDefinition", "Unable to update status")
		return ctrl.Result{}, nil
	}

	r.Recorder.Event(desired, "Normal", "ApiDefinition", "Done")

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
