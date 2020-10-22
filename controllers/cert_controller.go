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

	"github.com/TykTechnologies/tyk-operator/pkg/cert"

	"github.com/TykTechnologies/tyk-operator/internal/universal_client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	certFinalizerName = "finalizers.tyk.io/certs"
)

// CertReconciler reconciles a Cert object
type CertReconciler struct {
	client.Client
	Log             logr.Logger
	Scheme          *runtime.Scheme
	UniversalClient universal_client.UniversalClient
}

// +kubebuilder:rbac:groups=tyk.tyk.io,resources=certs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=tyk.tyk.io,resources=certs/status,verbs=get;update;patch

func (r *CertReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("cert", req.NamespacedName)

	//namespacedName := req.NamespacedName

	desired := &v1.Secret{}
	if err := r.Get(ctx, req.NamespacedName, desired); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err) // Ignore not-found errors
	}

	// If object is being deleted
	if !desired.ObjectMeta.DeletionTimestamp.IsZero() {
		// If our finalizer is present, need to delete from Tyk still
		if containsString(desired.ObjectMeta.Finalizers, certFinalizerName) {

			certPemBytes, ok := desired.Data["tls.crt"]
			if !ok {
				return ctrl.Result{}, nil
			}

			certFingerPrint := universal_client.GetOrganizationID(r.UniversalClient) + cert.CalculateFingerPrint(certPemBytes)

			err := r.UniversalClient.Certificate().Delete(certFingerPrint)
			if err != nil {
				log.Info(err.Error())
				return ctrl.Result{RequeueAfter: time.Second * 5}, err
			}

			err = r.UniversalClient.HotReload()
			if err != nil {
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
	if !containsString(desired.ObjectMeta.Finalizers, certFinalizerName) {
		desired.ObjectMeta.Finalizers = append(desired.ObjectMeta.Finalizers, certFinalizerName)
		err := r.Update(ctx, desired)
		// Return either way because the update will
		// issue a requeue anyway
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	tlsKey, ok := desired.Data["tls.key"]
	if !ok {
		// cert doesn't exist yet
		log.Info("missing key, requeue as maybe it's being created")
		return reconcile.Result{Requeue: true}, nil
	}
	tlsCrt, ok := desired.Data["tls.crt"]
	if !ok {
		// cert doesn't exist yet
		log.Info("missing cert, requeue as maybe it's being created")
		return reconcile.Result{Requeue: true}, nil
	}

	certID, err := universal_client.UploadCertificate(r.UniversalClient, tlsKey, tlsCrt)
	if err != nil {
		return reconcile.Result{Requeue: true}, err
	}

	log.Info("uploaded certificate to Tyk", "certID", certID)

	return reconcile.Result{}, nil
}

// https://sdk.operatorframework.io/docs/building-operators/golang/tutorial/#resources-watched-by-the-controller
func (r *CertReconciler) SetupWithManager(mgr ctrl.Manager) error {
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

func (r *CertReconciler) ignoreNonTLSPredicate() predicate.Predicate {

	isTLSType := func(jsBytes []byte) bool {
		tlsType := "kubernetes.io/tls"
		secret := mySecretType{}
		json.Unmarshal(jsBytes, &secret)

		// if Update
		if secret.MetaNew.Type != "" {
			return secret.MetaNew.Type == tlsType
		}
		// then it's a create / delete op
		return secret.Meta.Type == tlsType
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
			// TODO: THIS IS BUGGED. Need to find out if it is a secret being deleted that we care about
			eBytes, _ := json.Marshal(e)

			return isTLSType(eBytes)
		},
	}
}
