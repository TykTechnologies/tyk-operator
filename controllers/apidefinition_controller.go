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
	"crypto/sha256"
	"encoding/hex"
	"reflect"
	"time"

	tykv1alpha1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/internal/universal_client"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	apiDefFinalizerName = "finalizers.tyk.io/apidefinition"
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
	namespacedName := req.NamespacedName

	log := r.Log.WithValues("ApiDefinition", namespacedName.String())

	log.Info("fetching apidefinition instance")
	desired := &tykv1alpha1.ApiDefinition{}
	if err := r.Get(ctx, req.NamespacedName, desired); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err) // Ignore not-found errors
	}
	r.Recorder.Event(desired, "Normal", "ApiDefinition", "Reconciling")

	// If object is being deleted
	if !desired.ObjectMeta.DeletionTimestamp.IsZero() {

		// If our finalizer is present, need to delete from Tyk still
		if containsString(desired.ObjectMeta.Finalizers, apiDefFinalizerName) {

			policies, err := r.UniversalClient.SecurityPolicy().All()
			if err != nil {
				log.Info(err.Error())
				return ctrl.Result{RequeueAfter: time.Second * 5}, err
			}

			for _, policy := range policies {
				for _, right := range policy.AccessRightsArray {
					if right.APIID == desired.Status.ApiID {
						log.Info("unable to delete api due to security policy dependency",
							"api", namespacedName.String(),
							"policy", apiIDDecode(policy.ID),
						)
						return ctrl.Result{RequeueAfter: time.Second * 5}, err
					}
				}
			}

			err = r.UniversalClient.Api().Delete(desired.Status.ApiID)
			if err != nil {
				log.Error(err, "unable to delete api", "api_id", desired.Status.ApiID)
				return ctrl.Result{}, err
			}

			err = r.UniversalClient.HotReload()
			if err != nil {
				log.Error(err, "unable to hot reload", "api_id", desired.Status.ApiID)
				return ctrl.Result{}, err
			}

			// remove our finalizer from the list and update it.
			desired.ObjectMeta.Finalizers = removeString(desired.ObjectMeta.Finalizers, apiDefFinalizerName)
			if err := r.Update(ctx, desired); err != nil {
				return reconcile.Result{}, err
			}
		}

		return reconcile.Result{}, nil
	}

	// If finalizer not present, add it; This is a new object
	if !containsString(desired.ObjectMeta.Finalizers, apiDefFinalizerName) {
		desired.ObjectMeta.Finalizers = append(desired.ObjectMeta.Finalizers, apiDefFinalizerName)
		err := r.Update(ctx, desired)
		// Return either way because the update will
		// issue a requeue anyway
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	for _, certID := range desired.Spec.CertificateSecretNames {
		secret := v1.Secret{}
		err := r.Get(ctx, types.NamespacedName{Name: certID, Namespace: namespacedName.Namespace}, &secret)
		if err != nil {
			log.Error(err, "requeueing because secret not found")
			return reconcile.Result{}, err
		}

		cert, ok := secret.Data["tls.crt"]
		println(string(cert))
		if !ok {
			log.Error(err, "requeueing because cert not found in secret")
			return reconcile.Result{}, err
		}

		tykCertID := universal_client.GetOrganizationID(r.UniversalClient) + r.generateCertId(cert)
		_, err = universal_client.GetCertificate(r.UniversalClient, tykCertID)

		desired.Spec.Certificates = append(desired.Spec.Certificates, tykCertID)
	}

	desired.Spec.CertificateSecretNames = nil

	r.applyDefaults(&desired.Spec)

	//  If this is not set, means it is a new object, set it first
	if desired.Status.ApiID == "" {

		// If directly specified in the spec, this refers to an existing API definition
		// Otherwise, we use the B64 encoded namespace name as the custom API ID
		if desired.Spec.APIID != "" {
			desired.Status.ApiID = desired.Spec.APIID
		} else {
			desired.Status.ApiID = apiIDEncode(req.NamespacedName.String())
		}

		err := r.Status().Update(ctx, desired)
		if err != nil {
			log.Error(err, "Could not update Status ID")
		}

		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	desired.Spec.APIID = desired.Status.ApiID
	if err := universal_client.CreateOrUpdateAPI(r.UniversalClient, &desired.Spec); err != nil {
		log.Error(err, "createOrUpdate failure")
		r.Recorder.Event(desired, "Warning", "ApiDefinition", "Create or Update API Definition")
		return ctrl.Result{Requeue: true}, nil
	}

	return ctrl.Result{}, nil
}

func (r *ApiDefinitionReconciler) generateCertId(cert []byte) string {
	certSHA := sha256.Sum256(cert)
	return hex.EncodeToString(certSHA[:])
}

func (r *ApiDefinitionReconciler) ensureIngress(ctx context.Context, log logr.Logger, desired *v1beta1.Ingress) (reconcile.Result, error) {
	actual := &v1beta1.Ingress{}
	err := r.Get(ctx, types.NamespacedName{
		Name:      desired.Name,
		Namespace: desired.Namespace,
	}, actual)
	if err != nil {
		if errors.IsNotFound(err) {
			// create the ingress
			err = r.Create(ctx, desired)

			if err != nil {
				// Ingress creation failed
				log.Error(err, "Failed to create Ingress")
				return ctrl.Result{}, err
			}
			return reconcile.Result{Requeue: true}, nil
		}
		return reconcile.Result{}, err
	}

	if !reflect.DeepEqual(desired.Spec, actual.Spec) {
		desired.ObjectMeta = actual.ObjectMeta
		err = r.Update(context.TODO(), desired)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}

func (r *ApiDefinitionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tykv1alpha1.ApiDefinition{}).
		Complete(r)
}

func (r *ApiDefinitionReconciler) applyDefaults(spec *tykv1alpha1.APIDefinitionSpec) {
	if len(spec.VersionData.Versions) == 0 {
		defaultVersionData := tykv1alpha1.VersionData{
			NotVersioned:   true,
			DefaultVersion: "Default",
			Versions: map[string]tykv1alpha1.VersionInfo{
				"Default": {
					Name:                        "Default",
					Expires:                     "",
					Paths:                       tykv1alpha1.VersionInfoPaths{},
					UseExtendedPaths:            false,
					ExtendedPaths:               tykv1alpha1.ExtendedPathsSet{},
					GlobalHeaders:               nil,
					GlobalHeadersRemove:         nil,
					GlobalResponseHeaders:       nil,
					GlobalResponseHeadersRemove: nil,
					IgnoreEndpointCase:          false,
					GlobalSizeLimit:             0,
				},
			},
		}

		spec.VersionData = defaultVersionData
	}
}
