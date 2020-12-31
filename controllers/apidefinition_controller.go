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
	"time"

	"github.com/TykTechnologies/tyk-operator/pkg/cert"
	"github.com/TykTechnologies/tyk-operator/pkg/keys"

	tykv1alpha1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/pkg/universal_client"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	util "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
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

// +kubebuilder:rbac:groups=tyk.tyk.io,resources=apidefinitions,verbs=get;list;watch;create;update;patch;delete;deletecollection
// +kubebuilder:rbac:groups=tyk.tyk.io,resources=apidefinitions/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get

func (r *ApiDefinitionReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	namespacedName := req.NamespacedName

	log := r.Log.WithValues("ApiDefinition", namespacedName.String())

	desired := &tykv1alpha1.ApiDefinition{}
	if err := r.Get(ctx, req.NamespacedName, desired); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err) // Ignore not-found errors
	}

	if desired.GetLabels()["template"] == "true" {
		log.Info("Syncing template", "template", desired.Name)
		err := r.syncTemplate(ctx, req.Namespace, desired)
		if err != nil {
			log.Error(err, "Failed to sync template")
			return ctrl.Result{}, err
		}
		log.Info("Synced template", "template", desired.Name)
		return ctrl.Result{}, nil
	}

	// If object is being deleted
	if !desired.ObjectMeta.DeletionTimestamp.IsZero() {
		log.Info("resource being deleted")
		// If our finalizer is present, need to delete from Tyk still
		if util.ContainsFinalizer(desired, keys.ApiDefFinalizerName) {
			log.Info("checking linked security policies")
			policies, err := r.UniversalClient.SecurityPolicy().All()
			if err != nil && !universal_client.IsTODO(err) {
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

			log.Info("deleting api")
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

			log.Info("removing finalizer")
			util.RemoveFinalizer(desired, keys.ApiDefFinalizerName)
			if err := r.Update(ctx, desired); err != nil {
				return reconcile.Result{}, err
			}
		}
		log.Info("done")
		return reconcile.Result{}, nil
	}

	if !util.ContainsFinalizer(desired, keys.ApiDefFinalizerName) {
		log.Info("adding finalizer")
		desired.ObjectMeta.Finalizers = append(desired.ObjectMeta.Finalizers, keys.ApiDefFinalizerName)
		err := r.Update(ctx, desired)
		// Return either way because the update will
		// issue a requeue anyway
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	for _, certName := range desired.Spec.CertificateSecretNames {
		secret := v1.Secret{}
		err := r.Get(ctx, types.NamespacedName{Name: certName, Namespace: namespacedName.Namespace}, &secret)
		if err != nil {
			log.Error(err, "requeueing because secret not found")
			return reconcile.Result{}, err
		}

		pemCrtBytes, ok := secret.Data["tls.crt"]
		if !ok {
			log.Error(err, "requeueing because cert not found in secret")
			return reconcile.Result{}, err
		}

		pemKeyBytes, ok := secret.Data["tls.key"]
		if !ok {
			log.Error(err, "requeueing because key not found in secret")
			return reconcile.Result{}, err
		}

		tykCertID := r.UniversalClient.Organization().GetID() + cert.CalculateFingerPrint(pemCrtBytes)
		exists := r.UniversalClient.Certificate().Exists(tykCertID)
		if !exists {
			// upload the certificate
			tykCertID, err = r.UniversalClient.Certificate().Upload(pemKeyBytes, pemCrtBytes)
			if err != nil {
				return reconcile.Result{Requeue: true}, err
			}
		}

		desired.Spec.Certificates = []string{tykCertID}
		break
	}

	desired.Spec.CertificateSecretNames = nil

	//  If this is not set, means it is a new object, set it first
	if desired.Status.ApiID == "" {
		err := r.UniversalClient.Api().Create(
			encodeNS(req.NamespacedName.String()),
			&desired.Spec,
		)
		if err != nil {
			log.Error(err, "Failed to create api definition")
			return ctrl.Result{}, err
		}
		desired.Status.ApiID = desired.Spec.APIID
		err = r.Status().Update(ctx, desired)
		if err != nil {
			log.Error(err, "Could not update Status ID")
		}
		r.UniversalClient.HotReload()
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	log.Info("Updating ApiDefinition")
	desired.Spec.APIID = desired.Status.ApiID
	err := r.UniversalClient.Api().Update(&desired.Spec)
	if err != nil {
		if err != nil {
			log.Error(err, "Failed to update api definition")
			return ctrl.Result{}, err
		}
	}
	r.UniversalClient.HotReload()
	log.Info("done")
	return ctrl.Result{}, nil
}

// This triggers an update to all ingress resources that have template
// annotation matching a.Name.
//
// We return nil when a is being deleted and do nothing.
func (r *ApiDefinitionReconciler) syncTemplate(ctx context.Context, ns string, a *tykv1alpha1.ApiDefinition) error {
	if !a.DeletionTimestamp.IsZero() {
		return nil
	}
	ls := v1beta1.IngressList{}
	err := r.List(ctx, &ls,
		client.InNamespace(ns),
	)
	if err != nil {
		return client.IgnoreNotFound(err)
	}
	for _, v := range ls.Items {
		if v.GetAnnotations()[keys.IngressTemplateAnnotationKey] == a.Name {
			key := client.ObjectKey{
				Namespace: v.GetNamespace(),
				Name:      v.GetName(),
			}
			r.Log.Info("Updating ingress " + key.String())
			if v.Labels == nil {
				v.Labels = make(map[string]string)
			}
			v.Labels[keys.IngressTaintLabelKey] = time.Now().String()
			err = r.Update(ctx, &v)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *ApiDefinitionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tykv1alpha1.ApiDefinition{}).
		Owns(&v1.Secret{}).
		Complete(r)
}
