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
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/TykTechnologies/tyk-operator/pkg/cert"
	"github.com/TykTechnologies/tyk-operator/pkg/keys"

	tykv1alpha1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/pkg/universal_client"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	util "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const queueAfter = time.Second * 5

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
	log.Info("Reconciling ApiDefinition instance")
	desired := &tykv1alpha1.ApiDefinition{}
	if err := r.Get(ctx, req.NamespacedName, desired); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err) // Ignore not-found errors
	}

	if desired.GetLabels()["template"] == "true" {
		log.Info("Syncing template", "template", desired.Name)
		res, err := r.syncTemplate(ctx, req.Namespace, desired)
		if err != nil {
			log.Error(err, "Failed to sync template")
			return res, err
		}
		log.Info("Synced template", "template", desired.Name)
		return ctrl.Result{}, nil
	}

	var queue bool
	var queueA time.Duration
	_, err := util.CreateOrUpdate(ctx, r.Client, desired, func() error {
		if !desired.ObjectMeta.DeletionTimestamp.IsZero() {
			e, err := r.delete(ctx, desired)
			queueA = e
			return err
		}
		if desired.Spec.APIID == "" {
			desired.Spec.APIID = encodeNS(req.NamespacedName.String())
		}
		util.AddFinalizer(desired, keys.ApiDefFinalizerName)
		for _, certName := range desired.Spec.CertificateSecretNames {
			secret := v1.Secret{}
			err := r.Get(ctx, types.NamespacedName{Name: certName, Namespace: namespacedName.Namespace}, &secret)
			if err != nil {
				log.Error(err, "requeueing because secret not found")
				return err
			}
			pemCrtBytes, ok := secret.Data["tls.crt"]
			if !ok {
				log.Error(err, "requeueing because cert not found in secret")
				return err
			}

			pemKeyBytes, ok := secret.Data["tls.key"]
			if !ok {
				log.Error(err, "requeueing because key not found in secret")
				return err
			}

			tykCertID := r.UniversalClient.Organization().GetID() + cert.CalculateFingerPrint(pemCrtBytes)
			exists := r.UniversalClient.Certificate().Exists(tykCertID)
			if !exists {
				// upload the certificate
				tykCertID, err = r.UniversalClient.Certificate().Upload(pemKeyBytes, pemCrtBytes)
				if err != nil {
					queue = true
					return err
				}
			}
			desired.Spec.Certificates = []string{tykCertID}
			break
		}
		desired.Spec.CertificateSecretNames = nil
		err := r.updateLinkedPolicies(ctx, desired, func(sps *tykv1alpha1.SecurityPolicyStatus, s string) {
			sps.LinkedAPI = addString(sps.LinkedAPI, s)
		})
		if err != nil {
			log.Error(err, "Failed to update linked policies")
			return err
		}
		//  If this is not set, means it is a new object, set it first
		if desired.Status.ApiID == "" {
			err := r.UniversalClient.Api().Create(&desired.Spec)
			if err != nil {
				log.Error(err, "Failed to create api definition")
				return err
			}
			desired.Status.ApiID = desired.Spec.APIID
			err = r.Status().Update(ctx, desired)
			if err != nil {
				log.Error(err, "Could not update Status ID")
			}
			r.UniversalClient.HotReload()
			return client.IgnoreNotFound(err)
		}
		log.Info("Updating ApiDefinition")
		desired.Spec.APIID = desired.Status.ApiID
		err = r.UniversalClient.Api().Update(&desired.Spec)
		if err != nil {
			if err != nil {
				log.Error(err, "Failed to update api definition")
				return err
			}
		}
		r.UniversalClient.HotReload()
		return nil
	})
	if err == nil {
		log.Info("Completed reconciling ApiDefinition isnatnce")
	}
	return ctrl.Result{Requeue: queue, RequeueAfter: queueA}, err
}

// This triggers an update to all ingress resources that have template
// annotation matching a.Name.
//
// We return nil when a is being deleted and do nothing.
func (r *ApiDefinitionReconciler) syncTemplate(ctx context.Context, ns string, a *tykv1alpha1.ApiDefinition) (ctrl.Result, error) {
	if !a.DeletionTimestamp.IsZero() {
		if util.ContainsFinalizer(a, keys.ApiDefTemplateFinalizerName) {
			ls := v1beta1.IngressList{}
			err := r.List(ctx, &ls,
				client.InNamespace(ns),
			)
			if err != nil {
				if !errors.IsNotFound(err) {
					return ctrl.Result{}, err
				}
			}
			var refs []string
			for _, v := range ls.Items {
				if v.GetAnnotations()[keys.IngressTemplateAnnotation] == a.Name {
					refs = append(refs, v.GetName())
				}
			}
			if len(refs) > 0 {
				return ctrl.Result{RequeueAfter: time.Second * 5}, fmt.Errorf("Can't delete %s %v depends on it", a.Name, refs)
			}
			util.RemoveFinalizer(a, keys.ApiDefTemplateFinalizerName)
			return ctrl.Result{}, r.Update(ctx, a)
		}
		return ctrl.Result{}, nil
	}
	if !util.ContainsFinalizer(a, keys.ApiDefTemplateFinalizerName) {
		util.AddFinalizer(a, keys.ApiDefTemplateFinalizerName)
		return ctrl.Result{}, r.Update(ctx, a)
	}
	ls := v1beta1.IngressList{}
	err := r.List(ctx, &ls,
		client.InNamespace(ns),
	)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	for _, v := range ls.Items {
		if v.GetAnnotations()[keys.IngressTemplateAnnotation] == a.Name {
			key := client.ObjectKey{
				Namespace: v.GetNamespace(),
				Name:      v.GetName(),
			}
			r.Log.Info("Updating ingress " + key.String())
			if v.Labels == nil {
				v.Labels = make(map[string]string)
			}
			v.Labels[keys.IngressTaintLabel] = strconv.FormatInt(time.Now().UnixNano(), 10)
			err = r.Update(ctx, &v)
			if err != nil {
				return ctrl.Result{}, err
			}
		}
	}
	return ctrl.Result{}, nil
}

func (r *ApiDefinitionReconciler) delete(ctx context.Context, desired *tykv1alpha1.ApiDefinition) (time.Duration, error) {
	r.Log.Info("resource being deleted")
	// If our finalizer is present, need to delete from Tyk still
	if util.ContainsFinalizer(desired, keys.ApiDefFinalizerName) {
		if err := r.checkLinkedPolicies(ctx, desired); err != nil {
			return queueAfter, err
		}
		r.Log.Info("deleting api")
		err := r.UniversalClient.Api().Delete(desired.Status.ApiID)
		if err != nil {
			r.Log.Error(err, "unable to delete api", "api_id", desired.Status.ApiID)
			return 0, err
		}
		err = r.UniversalClient.HotReload()
		if err != nil {
			r.Log.Error(err, "unable to hot reload", "api_id", desired.Status.ApiID)
			return 0, err
		}
		err = r.updateLinkedPolicies(ctx, desired, func(sps *tykv1alpha1.SecurityPolicyStatus, s string) {
			sps.LinkedAPI = removeString(sps.LinkedAPI, s)
		})
		if err != nil {
			r.Log.Error(err, "Failed to update linked policies")
			return 0, err
		}
		r.Log.Info("removing finalizer")
		util.RemoveFinalizer(desired, keys.ApiDefFinalizerName)
	}
	return 0, nil
}

// checkLinkedPolicies checks if there are any policies that are still linking
// to this api defintion resource.
func (r *ApiDefinitionReconciler) checkLinkedPolicies(ctx context.Context, a *tykv1alpha1.ApiDefinition) error {
	r.Log.Info("checking linked security policies")
	if len(a.Status.LinkedPolicies) == 0 {
		return nil
	}
	for _, n := range a.Status.LinkedPolicies {
		p := strings.Split(n, string(types.Separator))
		if len(p) != 2 {
			err := fmt.Errorf("malformed linked_policy expected namespace/name format got %s", n)
			r.Log.Error(err, "Failed to parse lined_policies")
			return err
		}
		ns := types.NamespacedName{Namespace: p[0], Name: p[1]}
		var policy tykv1alpha1.SecurityPolicy
		if err := r.Get(ctx, ns, &policy); err == nil {
			return fmt.Errorf("unable to delete api due to security policy dependency=%s", n)
		}
	}
	return nil
}

// updateLinkedPolicies ensure that all policies needed by this api denition are
// updated.
func (r *ApiDefinitionReconciler) updateLinkedPolicies(ctx context.Context, a *tykv1alpha1.ApiDefinition,
	fn func(*tykv1alpha1.SecurityPolicyStatus, string),
) error {
	r.Log.Info("Updating linked policies")
	ns := (types.NamespacedName{Namespace: a.Namespace, Name: a.Name}).String()
	names := map[string]struct{}{}
	for _, x := range a.Spec.JWTDefaultPolicies {
		names[x] = struct{}{}
	}
	for _, x := range a.Spec.JWTScopeToPolicyMapping {
		names[x] = struct{}{}
	}
	if len(names) == 0 {
		return nil
	}
	replace := map[string]string{}
	for n := range names {
		p := strings.Split(n, string(types.Separator))
		if len(p) != 2 {
			replace[n] = n
			continue
		}
		api := &tykv1alpha1.SecurityPolicy{}
		if err := r.Get(ctx, types.NamespacedName{Namespace: p[0], Name: p[1]}, api); err != nil {
			r.Log.Error(err, "Failed to get linked api definition")
			return err
		}
		x := api.Status.DeepCopy()
		fn(&api.Status, ns)
		if !equality.Semantic.DeepEqual(x, api.Status) {
			if err := r.Status().Update(ctx, api); err != nil {
				r.Log.Error(err, "Failed to update linked security policy")
				return err
			}
		}
		replace[n] = api.Status.PolID
	}
	for x := range a.Spec.JWTDefaultPolicies {
		a.Spec.JWTDefaultPolicies[x] = replace[a.Spec.JWTDefaultPolicies[x]]
	}
	for k, x := range a.Spec.JWTScopeToPolicyMapping {
		a.Spec.JWTScopeToPolicyMapping[k] = replace[x]
	}
	return nil
}

// SetupWithManager initializes the api definition controller.
func (r *ApiDefinitionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tykv1alpha1.ApiDefinition{}).
		Owns(&v1.Secret{}).
		Complete(r)
}
