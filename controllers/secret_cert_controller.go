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
	"encoding/json"
	"time"

	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/pkg/cert"
	"github.com/TykTechnologies/tyk-operator/pkg/universal_client"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	certFinalizerName = "finalizers.tyk.io/certs"
	secretType        = "kubernetes.io/tls"
)

// CertReconciler reconciles a Cert object
type SecretCertReconciler struct {
	client.Client
	Log             logr.Logger
	Scheme          *runtime.Scheme
	UniversalClient universal_client.UniversalClient
}

// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete

func (r *SecretCertReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("cert", req.NamespacedName)

	log.Info("getting secret resource")
	desired := &v1.Secret{}
	if err := r.Get(ctx, req.NamespacedName, desired); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err) // Ignore not-found errors
	}

	// If object is being deleted
	if !desired.ObjectMeta.DeletionTimestamp.IsZero() {
		log.Info("secret being deleted")
		// If our finalizer is present, need to delete from Tyk still
		if containsString(desired.ObjectMeta.Finalizers, certFinalizerName) {
			log.Info("running finalizer logic")

			certPemBytes, ok := desired.Data["tls.crt"]
			if !ok {
				return ctrl.Result{}, nil
			}

			orgID := universal_client.GetOrganizationID(r.UniversalClient)
			certFingerPrint := cert.CalculateFingerPrint(certPemBytes)

			certID := orgID + certFingerPrint

			log.Info("deleting certificate from tyk certificate manager", "orgID", orgID, "fingerprint", certFingerPrint)
			if err := r.UniversalClient.Certificate().Delete(certID); err != nil {
				log.Error(err, "unable to delete certificate")
				return ctrl.Result{RequeueAfter: time.Second * 5}, err
			}

			if err := r.UniversalClient.HotReload(); err != nil {
				return ctrl.Result{}, err
			}

			log.Info("removing finalizer from secret")
			desired.ObjectMeta.Finalizers = removeString(desired.ObjectMeta.Finalizers, certFinalizerName)
			if err := r.Update(ctx, desired); err != nil {
				return ctrl.Result{}, err
			}
		}

		log.Info("secret successfully deleted")
		return ctrl.Result{}, nil
	}

	log.Info("checking secret type is tls")
	if desired.Type != secretType {
		// it's not for us
		return ctrl.Result{}, nil
	}

	log.Info("ensuring tls.key is present")
	tlsKey, ok := desired.Data["tls.key"]
	if !ok {
		// cert doesn't exist yet
		log.Info("missing tls.key, we don't care about it yet")
		return ctrl.Result{}, nil
	}
	log.Info("ensuring tls.key is present")
	tlsCrt, ok := desired.Data["tls.crt"]
	if !ok {
		// cert doesn't exist yet
		log.Info("missing tls.crt, we don't care about it yet")
		return ctrl.Result{}, nil
	}

	// all apidefinitions in current namespace
	apiDefList := v1alpha1.ApiDefinitionList{}
	opts := []client.ListOption{
		client.InNamespace(req.Namespace),
	}
	if err := r.List(ctx, &apiDefList, opts...); err != nil {
		log.Info("unable to list api definitions")
		return ctrl.Result{}, err
	}

	if len(apiDefList.Items) == 0 {
		log.Info("no apidefinitions in namespace")
		return ctrl.Result{}, nil
	}

	ret := true
	for _, apiDef := range apiDefList.Items {
		if containsString(apiDef.Spec.CertificateSecretNames, req.Name) {
			log.Info("apidefinition references this secret", "apiid", apiDef.Status.ApiID)
			ret = false
		}
	}

	if ret {
		log.Info("no apidefinitions reference this secret")
		return ctrl.Result{}, nil
	}

	// If finalizer not present, add it; This is a new object
	if !containsString(desired.ObjectMeta.Finalizers, certFinalizerName) {
		log.Info("adding finalizer for cleanup")

		desired.ObjectMeta.Finalizers = append(desired.ObjectMeta.Finalizers, certFinalizerName)
		err := r.Update(ctx, desired)

		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	certID, err := universal_client.UploadCertificate(r.UniversalClient, tlsKey, tlsCrt)
	if err != nil {
		return ctrl.Result{Requeue: true}, err
	}

	log.Info("uploaded certificate to Tyk", "certID", certID)
	for _, apiDef := range apiDefList.Items {
		if containsString(apiDef.Spec.CertificateSecretNames, req.Name) {
			log.Info("replacing certificate", "apiID", apiDef.Status.ApiID, "certID", certID)

			apiDefObj, _ := r.UniversalClient.Api().Get(apiDef.Status.ApiID)
			apiDefObj.Certificates = []string{}
			apiDefObj.Certificates = append(apiDefObj.Certificates, certID)
			r.UniversalClient.Api().Update(apiDef.Status.ApiID, apiDefObj)

			// TODO: we only care about 1 secret - we don't need to support multiple for mvp
			break
		}
	}

	return ctrl.Result{}, nil
}

// https://sdk.operatorframework.io/docs/building-operators/golang/tutorial/#resources-watched-by-the-controller
func (r *SecretCertReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.Secret{}).
		WithEventFilter(r.ignoreNonTLSPredicate()).
		Complete(r)
}

type mySecretType struct {
	Meta struct {
		Type string `json:"type"`
	} `json:"Meta"`
	MetaNew struct {
		Type string `json:"type"`
	} `json:"MetaNew"`
}

func (r *SecretCertReconciler) ignoreNonTLSPredicate() predicate.Predicate {

	isTLSType := func(jsBytes []byte) bool {
		secret := mySecretType{}
		json.Unmarshal(jsBytes, &secret)

		// if Update
		if secret.MetaNew.Type != "" {
			return secret.MetaNew.Type == secretType
		}
		// then it's a create / delete op
		return secret.Meta.Type == secretType
	}

	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			//obj := runtime.Object
			//return e.Meta.Type == "kubernetes.io/tls"
			eBytes, _ := json.Marshal(e)
			return isTLSType(eBytes)
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			eBytes, _ := json.Marshal(e)
			return isTLSType(eBytes)
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return true
		},
	}
}
