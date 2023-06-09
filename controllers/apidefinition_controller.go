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
	"encoding/base64"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/TykTechnologies/tyk-operator/api/model"
	tykv1alpha1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/pkg/cert"
	tykClient "github.com/TykTechnologies/tyk-operator/pkg/client"
	"github.com/TykTechnologies/tyk-operator/pkg/client/klient"
	"github.com/TykTechnologies/tyk-operator/pkg/environmet"
	"github.com/TykTechnologies/tyk-operator/pkg/keys"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	util "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	queueAfter = time.Second * 5
	GraphKey   = "graph_ref"
)

var ErrMultipleLinkSubGraph = errors.New("linking one SubGraph to multiple ApiDefinition is forbidden")

// ApiDefinitionReconciler reconciles a ApiDefinition object
type ApiDefinitionReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Env      environmet.Env
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=tyk.tyk.io,resources=apidefinitions,verbs=get;list;watch;create;update;patch;delete;deletecollection
// +kubebuilder:rbac:groups=tyk.tyk.io,resources=subgraphs,verbs=get;list;watch;create;update;patch;delete;deletecollection
// +kubebuilder:rbac:groups=tyk.tyk.io,resources=supergraphs,verbs=get;list;watch;create;update;patch;delete;deletecollection
// +kubebuilder:rbac:groups=tyk.tyk.io,resources=apidefinitions/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=tyk.tyk.io,resources=subgraphs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;update;create
// +kubebuilder:rbac:groups=coordination.k8s.io,resources=leases,verbs=get;list;create;update

func (r *ApiDefinitionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	namespacedName := req.NamespacedName
	log := r.Log.WithValues("ApiDefinition", namespacedName.String())

	log.Info("Reconciling ApiDefinition instance")

	desired := &tykv1alpha1.ApiDefinition{}

	if err := r.Get(ctx, req.NamespacedName, desired); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err) // Ignore not-found errors
	}

	upstreamRequestStruct := &tykv1alpha1.ApiDefinition{}
	desired.DeepCopyInto(upstreamRequestStruct)

	// set context for all api calls inside this reconciliation loop
	env, ctx, err := HttpContext(ctx, r.Client, r.Env, desired, log)
	if err != nil {
		return ctrl.Result{}, err
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

	var queueA time.Duration

	_, err = util.CreateOrUpdate(ctx, r.Client, desired, func() error {
		if !desired.ObjectMeta.DeletionTimestamp.IsZero() {
			e, err := r.delete(ctx, desired)
			queueA = e
			return err
		}

		if desired.Spec.APIID == nil || *desired.Spec.APIID == "" {
			apiID := EncodeNS(req.NamespacedName.String())
			upstreamRequestStruct.Spec.APIID = &apiID
		}

		if desired.Spec.OrgID == nil || *desired.Spec.OrgID == "" {
			orgID := env.Org
			upstreamRequestStruct.Spec.OrgID = &orgID
		}

		util.AddFinalizer(desired, keys.ApiDefFinalizerName)

		if err := r.processCertificateReferences(ctx, &env, log, upstreamRequestStruct); err != nil {
			return err
		}
		desired.Spec.Certificates = upstreamRequestStruct.Spec.Certificates

		r.processUpstreamCertificateReferences(ctx, &env, log, upstreamRequestStruct)
		desired.Spec.UpstreamCertificates = upstreamRequestStruct.Spec.UpstreamCertificates

		// Check Pinned Public keys
		r.processPinnedPublicKeyReferences(ctx, &env, log, upstreamRequestStruct)
		desired.Spec.PinnedPublicKeys = upstreamRequestStruct.Spec.PinnedPublicKeys

		if desired.Spec.UseMutualTLSAuth != nil && *desired.Spec.UseMutualTLSAuth {
			r.processClientCertificateReferences(ctx, &env, log, upstreamRequestStruct)
			desired.Spec.ClientCertificates = upstreamRequestStruct.Spec.ClientCertificates
		}

		// Check GraphQL Federation
		if desired.Spec.GraphQL != nil {
			switch desired.Spec.GraphQL.ExecutionMode {
			case model.SubGraphExecutionMode:
				err = r.processSubGraphExec(ctx, upstreamRequestStruct)
				if err != nil {
					return err
				}

				desired.Spec.GraphQL.Schema = new(string)

				*desired.Spec.GraphQL.Schema = *upstreamRequestStruct.Spec.GraphQL.Schema
				desired.Spec.GraphQL.Subgraph.SDL = upstreamRequestStruct.Spec.GraphQL.Subgraph.SDL
			case model.SuperGraphExecutionMode:
				err = r.processSuperGraphExec(ctx, upstreamRequestStruct)
				if err != nil {
					return err
				}

				desired.Spec.GraphQL.Schema = new(string)
				desired.Spec.GraphQL.Supergraph.MergedSDL = new(string)

				*desired.Spec.GraphQL.Schema = *upstreamRequestStruct.Spec.GraphQL.Schema
				*desired.Spec.GraphQL.Supergraph.MergedSDL = *upstreamRequestStruct.Spec.GraphQL.Supergraph.MergedSDL
			}
		}

		r.updateLinkedPolicies(ctx, upstreamRequestStruct)

		targets := upstreamRequestStruct.Spec.CollectLoopingTarget()
		if err := r.ensureTargets(ctx, upstreamRequestStruct.Namespace, targets); err != nil {
			return err
		}

		err := r.updateLoopingTargets(ctx, upstreamRequestStruct, targets)
		if err != nil {
			log.Error(err, "Failed to update looping targets")
			return err
		}

		upstreamRequestStruct.Spec.CollectLoopingTarget()

		//  If this is not set, means it is a new object, set it first
		if desired.Status.ApiID == "" {
			return r.create(ctx, upstreamRequestStruct)
		}

		return r.update(ctx, upstreamRequestStruct)
	})

	if err == nil {
		log.Info("Completed reconciling ApiDefinition instance")
	} else {
		queueA = queueAfter
	}

	return ctrl.Result{RequeueAfter: queueA}, err
}

func (r *ApiDefinitionReconciler) processClientCertificateReferences(
	ctx context.Context,
	env *environmet.Env,
	log logr.Logger,
	upstreamRequestStruct *tykv1alpha1.ApiDefinition,
) {
	log.Info("Processing client certificate references")

	if len(upstreamRequestStruct.Spec.ClientCertificateRefs) != 0 {
		clientCerts := make([]string, 0)

		for _, secretName := range upstreamRequestStruct.Spec.ClientCertificateRefs {
			tykCertID, err := r.checkSecretAndUpload(ctx, secretName, upstreamRequestStruct.Namespace, log, env)
			if err != nil {
				// we should log the missing secret, but we should still create the API definition
				log.Error(
					err,
					"failed to upload client certificate. Creating API Definition without uploading certificate",
					"secretName",
					secretName)

				continue
			}

			clientCerts = append(clientCerts, tykCertID)
		}

		upstreamRequestStruct.Spec.ClientCertificates = clientCerts
		upstreamRequestStruct.Spec.ClientCertificateRefs = nil
	}
}

func (r *ApiDefinitionReconciler) processCertificateReferences(
	ctx context.Context,
	env *environmet.Env,
	log logr.Logger,
	upstreamRequestStruct *tykv1alpha1.ApiDefinition,
) error {
	// we support only one certificate secret name for mvp
	if len(upstreamRequestStruct.Spec.CertificateSecretNames) != 0 {
		certName := upstreamRequestStruct.Spec.CertificateSecretNames[0]

		tykCertID, err := r.checkSecretAndUpload(ctx, certName, upstreamRequestStruct.Namespace, log, env)
		if err != nil {
			return err
		}

		upstreamRequestStruct.Spec.Certificates = []string{tykCertID}
	}
	// To prevent API object validation failures, set additional properties to nil.
	upstreamRequestStruct.Spec.CertificateSecretNames = nil

	return nil
}

func (r *ApiDefinitionReconciler) processPinnedPublicKeyReferences(
	ctx context.Context,
	env *environmet.Env,
	log logr.Logger,
	upstreamRequestStruct *tykv1alpha1.ApiDefinition,
) {
	if len(upstreamRequestStruct.Spec.PinnedPublicKeysRefs) != 0 {
		if upstreamRequestStruct.Spec.PinnedPublicKeys == nil {
			upstreamRequestStruct.Spec.PinnedPublicKeys = map[string]string{}
		}

		for domain, secretName := range upstreamRequestStruct.Spec.PinnedPublicKeysRefs {
			// Set the namespace for referenced secret to the current namespace where ApiDefinition lives.
			tykCertID, err := r.checkSecretAndUpload(ctx, secretName, upstreamRequestStruct.Namespace, log, env)
			if err != nil {
				// we should log the missing secret, but we should still create the API definition
				log.Error(
					err,
					"failed to upload pinned public key from the secret. Creating API Definition without uploading certificate",
					"secretName",
					secretName)

				continue
			}

			upstreamRequestStruct.Spec.PinnedPublicKeys[domain] = tykCertID
		}
	}

	upstreamRequestStruct.Spec.PinnedPublicKeysRefs = nil
}

func (r *ApiDefinitionReconciler) processUpstreamCertificateReferences(
	ctx context.Context,
	env *environmet.Env,
	log logr.Logger,
	upstreamRequestStruct *tykv1alpha1.ApiDefinition,
) {
	if len(upstreamRequestStruct.Spec.UpstreamCertificateRefs) != 0 {
		for domain, certName := range upstreamRequestStruct.Spec.UpstreamCertificateRefs {
			tykCertID, err := r.checkSecretAndUpload(ctx, certName, upstreamRequestStruct.Namespace, log, env)
			if err != nil {
				// we should log the missing secret, but we should still create the API definition
				log.Info(fmt.Sprintf("cert name %s is missing", certName), "error", err)
			} else {
				if upstreamRequestStruct.Spec.UpstreamCertificates == nil {
					upstreamRequestStruct.Spec.UpstreamCertificates = make(map[string]string)
				}
				upstreamRequestStruct.Spec.UpstreamCertificates[domain] = tykCertID
			}
		}
	}

	upstreamRequestStruct.Spec.UpstreamCertificateRefs = nil
}

func decodeID(encodedID string) (namespace, name string) {
	if encodedID == "" {
		return
	}

	b, err := base64.RawURLEncoding.DecodeString(encodedID)
	if err != nil {
		return
	}

	namespacedName := strings.Split(string(b), "/")
	if len(namespacedName) < 2 {
		return
	}

	return namespacedName[0], namespacedName[1]
}

func uploadCert(ctx context.Context, orgID string, pemKeyBytes, pemCrtBytes []byte) (tykCertID string, err error) {
	fingerprint, err := cert.CalculateFingerPrint(pemCrtBytes)
	if err != nil {
		return "", err
	}

	tykCertID = orgID + fingerprint
	exists := klient.Universal.Certificate().Exists(ctx, tykCertID)

	if !exists {
		// upload the certificate
		tykCertID, err = klient.Universal.Certificate().Upload(ctx, pemKeyBytes, pemCrtBytes)
		if err != nil {
			return "", err
		}
	}

	return tykCertID, nil
}

func (r *ApiDefinitionReconciler) checkSecretAndUpload(
	ctx context.Context,
	certName string,
	ns string,
	log logr.Logger,
	env *environmet.Env,
) (string, error) {
	secret := v1.Secret{}

	err := r.Get(ctx, types.NamespacedName{Name: certName, Namespace: ns}, &secret)
	if err != nil {
		log.Error(err, "requeueing because secret not found")
		return "", err
	}

	pemCrtBytes, ok := secret.Data["tls.crt"]
	if !ok {
		err = fmt.Errorf("%s", "requeueing because cert not found in secret")
		log.Error(err, "requeueing because cert not found in secret")

		return "", err
	}

	pemKeyBytes, ok := secret.Data["tls.key"]
	if !ok {
		err = fmt.Errorf("%s", "requeueing because key not found in secret")
		log.Error(err, "requeueing because key not found in secret")

		return "", err
	}

	return uploadCert(ctx, env.Org, pemKeyBytes, pemCrtBytes)
}

func (r *ApiDefinitionReconciler) create(ctx context.Context, desired *tykv1alpha1.ApiDefinition) error {
	r.Log.Info("Creating a new ApiDefinition",
		"ApiDefinition", client.ObjectKeyFromObject(desired).String(),
	)

	_, err := klient.Universal.Api().Create(ctx, &desired.Spec.APIDefinitionSpec)
	if err != nil {
		r.Log.Error(
			err,
			"Failed to create ApiDefinition on Tyk",
			"ApiDefinition", client.ObjectKeyFromObject(desired).String(),
		)

		return err
	}

	err = klient.Universal.HotReload(ctx)
	if err != nil {
		r.Log.Error(
			err,
			"Failed to hot-reload Tyk after creating the ApiDefinition",
			"ApiDefinition", client.ObjectKeyFromObject(desired).String(),
		)

		return err
	}

	apiOnTyk, _ := klient.Universal.Api().Get(ctx, *desired.Spec.APIID) //nolint:errcheck

	namespace := desired.Namespace
	target := model.Target{Namespace: &namespace, Name: desired.Name}

	err = r.updateStatus(
		ctx,
		desired.Namespace,
		target,
		false,
		func(status *tykv1alpha1.ApiDefinitionStatus) {
			status.ApiID = *desired.Spec.APIID
			status.LatestTykSpecHash = calculateHash(apiOnTyk)
			status.LatestCRDSpecHash = calculateHash(desired.Spec)
		},
	)
	if err != nil {
		r.Log.Error(
			err,
			"Failed to update Status ID",
			"ApiDefinition", client.ObjectKeyFromObject(desired).String(),
		)

		return err
	}

	return nil
}

func (r *ApiDefinitionReconciler) update(ctx context.Context, desired *tykv1alpha1.ApiDefinition) error {
	r.Log.Info("Updating ApiDefinition",
		"ApiDefinition", client.ObjectKeyFromObject(desired).String(),
	)

	apiDefOnTyk, err := klient.Universal.Api().Get(ctx, desired.Status.ApiID)
	if err != nil {
		_, err = klient.Universal.Api().Create(ctx, &desired.Spec.APIDefinitionSpec)
		if err != nil {
			r.Log.Error(
				err, "Failed to create ApiDefinition on Tyk",
				"ApiDefinition", client.ObjectKeyFromObject(desired).String(),
			)

			return err
		}
	} else {
		// If we have same ApiDefinition on Tyk, we do not need to send Update and Hot Reload requests
		// to Tyk. So, we can simply return to main reconciliation logic.
		if isSame(desired.Status.LatestTykSpecHash, apiDefOnTyk) && isSame(desired.Status.LatestCRDSpecHash, desired.Spec) {
			return nil
		}

		_, err = klient.Universal.Api().Update(ctx, &desired.Spec.APIDefinitionSpec)
		if err != nil {
			r.Log.Error(
				err, "Failed to update ApiDefinition on Tyk",
				"ApiDefinition", client.ObjectKeyFromObject(desired).String(),
			)

			return err
		}
	}

	err = klient.Universal.HotReload(ctx)
	if err != nil {
		r.Log.Error(
			err,
			"Failed to hot-reload Tyk after updating the ApiDefinition",
			"ApiDefinition", client.ObjectKeyFromObject(desired).String(),
		)

		return err
	}

	apiOnTyk, _ := klient.Universal.Api().Get(ctx, *desired.Spec.APIID) //nolint:errcheck

	namespace := desired.Namespace
	target := model.Target{Namespace: &namespace, Name: desired.Name}

	err = r.updateStatus(
		ctx,
		desired.Namespace,
		target,
		false,
		func(status *tykv1alpha1.ApiDefinitionStatus) {
			status.LatestTykSpecHash = calculateHash(apiOnTyk)
			status.LatestCRDSpecHash = calculateHash(desired.Spec)
		},
	)
	if err != nil {
		r.Log.Error(
			err,
			"Failed to update Status",
			"ApiDefinition", client.ObjectKeyFromObject(desired).String(),
		)

		return err
	}
	return nil
}

// This triggers an update to all ingress resources that have template
// annotation matching a.Name.
//
// We return nil when a is being deleted and do nothing.
func (r *ApiDefinitionReconciler) syncTemplate(ctx context.Context, ns string, a *tykv1alpha1.ApiDefinition) (ctrl.Result, error) {
	if !a.DeletionTimestamp.IsZero() {
		if util.ContainsFinalizer(a, keys.ApiDefTemplateFinalizerName) {
			ls := netv1.IngressList{}

			err := r.List(ctx, &ls,
				client.InNamespace(ns),
			)
			if err != nil {
				if !k8sErrors.IsNotFound(err) {
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
				return ctrl.Result{RequeueAfter: time.Second * 5},
					fmt.Errorf("Can't delete %s %v depends on it", a.Name, refs)
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

	ls := netv1.IngressList{}

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
	r.Log.Info("ApiDefinition being deleted",
		"ApiDefinition", client.ObjectKeyFromObject(desired).String(),
	)

	// If our finalizer is present, need to delete from Tyk still
	if util.ContainsFinalizer(desired, keys.ApiDefFinalizerName) {
		if err := r.checkLinkedPolicies(ctx, desired); err != nil {
			return queueAfter, err
		}

		if err := r.checkLoopingTargets(ctx, desired); err != nil {
			return queueAfter, err
		}

		namespace := desired.Namespace
		ns := model.Target{
			Name:      desired.Name,
			Namespace: &namespace,
		}

		for _, target := range desired.Status.LinkedToAPIs {
			err := r.updateStatus(ctx, desired.Namespace, target, true, func(ads *tykv1alpha1.ApiDefinitionStatus) {
				ads.LinkedByAPIs = removeTarget(ads.LinkedByAPIs, ns)
			})
			if err != nil {
				return queueAfter, err
			}
		}

		err := r.breakSubgraphLink(ctx, desired, true)
		if err != nil {
			return queueAfter, err
		}

		r.Log.Info("Deleting an ApiDefinition from Tyk", "ApiDefinition ID", desired.Status.ApiID)

		_, err = klient.Universal.Api().Delete(ctx, desired.Status.ApiID)
		if err != nil && tykClient.IsNotFound(err) {
			r.Log.Info(
				"Ignoring nonexistent ApiDefinition on delete",
				"api_id", desired.Status.ApiID,
				"err", err,
			)
		} else if err != nil {
			// If the ApiDefinition does not exist on Tyk, no need to reconcile with error.
			// Older versions of GW does not return 404 while deleting non-existent ApiDefinitions.
			// Therefore, check if ApiDefinition exists on Tyk before returning with error. If ApiDefinition
			// exists, which means Get call returns successful response, Operator should reconcile to complete
			// deletion of the ApiDefinition.
			_, errTyk := klient.Universal.Api().Get(ctx, desired.Status.ApiID)
			if errTyk == nil {
				r.Log.Error(
					err,
					"Failed to delete ApiDefinition from Tyk", "api_id", desired.Status.ApiID,
				)
				return queueAfter, err
			}
		}

		err = klient.Universal.HotReload(ctx)
		if err != nil {
			r.Log.Error(
				err,
				"Failed to hot-reload Tyk after deleting the ApiDefinition",
				"ApiDefinition", client.ObjectKeyFromObject(desired).String(),
			)

			return queueAfter, err
		}

		util.RemoveFinalizer(desired, keys.ApiDefFinalizerName)
	}

	r.Log.Info(
		"Deleted ApiDefinition successfully",
		"ApiDefinition", client.ObjectKeyFromObject(desired).String(),
	)

	return 0, nil
}

// checkLinkedPolicies checks if there are any policies that are still linking to this api definition resource.
func (r *ApiDefinitionReconciler) checkLinkedPolicies(ctx context.Context, a *tykv1alpha1.ApiDefinition) error {
	r.Log.Info("checking linked security policies")

	if len(a.Status.LinkedByPolicies) == 0 {
		return nil
	}

	for _, n := range a.Status.LinkedByPolicies {
		var api tykv1alpha1.SecurityPolicy

		if err := r.Get(ctx, n.NS(a.Namespace), &api); err == nil {
			return fmt.Errorf("unable to delete api due to security policy dependency=%s", n)
		}
	}

	return nil
}

func encodeIfNotBase64(s string) string {
	_, err := base64.RawURLEncoding.DecodeString(s)
	if err == nil {
		return s
	}

	return EncodeNS(s)
}

// updateLinkedPolicies ensure that all policies needed by this api definition are updated.
func (r *ApiDefinitionReconciler) updateLinkedPolicies(ctx context.Context, a *tykv1alpha1.ApiDefinition) {
	r.Log.Info("Updating linked policies")

	for k, x := range a.Spec.JWTScopeToPolicyMapping {
		a.Spec.JWTScopeToPolicyMapping[k] = encodeIfNotBase64(x)
	}

	allPolicies, _ := klient.Universal.Portal().Policy().All(ctx)
	policyIDList := make([]string, 0)
	nameMIDMap := make(map[string]string)
	for _, p := range allPolicies {
		nameMIDMap[p.Name] = *p.MID
	}
	for _, policyName := range a.Spec.JWTDefaultPolicies {
		if id, ok := nameMIDMap[policyName]; ok {
			AddUniqueElement(&policyIDList, id)
		}
	}
	a.Spec.JWTDefaultPolicies = policyIDList
}

// checkLoopingTargets Check if there is any other api's linking to a
func (r *ApiDefinitionReconciler) checkLoopingTargets(ctx context.Context, a *tykv1alpha1.ApiDefinition) error {
	r.Log.Info("checking linked api resources")

	if len(a.Status.LinkedByAPIs) == 0 {
		return nil
	}

	for _, n := range a.Status.LinkedByAPIs {
		var api tykv1alpha1.ApiDefinition

		if err := r.Get(ctx, n.NS(a.Namespace), &api); err == nil {
			return fmt.Errorf("unable to delete api due to being depended by =%s", n)
		}
	}

	return nil
}

func (r *ApiDefinitionReconciler) ensureTargets(
	ctx context.Context,
	ns string,
	targets []model.Target,
) error {
	for _, target := range targets {
		var api tykv1alpha1.ApiDefinition

		if err := r.Get(ctx, target.NS(ns), &api); err != nil {
			return err
		}
	}

	return nil
}

func (r *ApiDefinitionReconciler) updateLoopingTargets(ctx context.Context,
	a *tykv1alpha1.ApiDefinition, links []model.Target,
) error {
	r.Log.Info("Updating looping targets")

	if len(links) == 0 {
		return nil
	}

	namespace := a.Namespace
	ns := model.Target{
		Name:      a.Name,
		Namespace: &namespace,
	}

	for _, target := range links {
		err := r.updateStatus(ctx, a.Namespace, target, false, func(ads *tykv1alpha1.ApiDefinitionStatus) {
			ads.LinkedByAPIs = addTarget(ads.LinkedByAPIs, ns)
			sort.Slice(ads.LinkedByAPIs, func(i, j int) bool {
				return ads.LinkedByAPIs[i].String() < ads.LinkedByAPIs[j].String()
			})
		})
		if err != nil {
			return err
		}
	}

	// we need to update removed targets
	newTargets := make(map[string]model.Target)
	for _, v := range links {
		newTargets[v.String()] = v
	}

	for _, v := range a.Status.LinkedToAPIs {
		if _, ok := newTargets[v.String()]; !ok {
			err := r.updateStatus(ctx, a.Namespace, v, true, func(ads *tykv1alpha1.ApiDefinitionStatus) {
				ads.LinkedByAPIs = removeTarget(ads.LinkedByAPIs, ns)
				sort.Slice(ads.LinkedByAPIs, func(i, j int) bool {
					return ads.LinkedByAPIs[i].String() < ads.LinkedByAPIs[j].String()
				})
			})
			if err != nil {
				return err
			}
		}
	}

	a.Status.LinkedToAPIs = links

	return client.IgnoreNotFound(r.Status().Update(ctx, a))
}

func (r *ApiDefinitionReconciler) updateStatus(
	ctx context.Context,
	ns string,
	target model.Target,
	ignoreNotFound bool,
	fn func(*tykv1alpha1.ApiDefinitionStatus),
) error {
	var api tykv1alpha1.ApiDefinition
	if err := r.Get(ctx, target.NS(ns), &api); err != nil {
		if k8sErrors.IsNotFound(err) {
			if ignoreNotFound {
				return nil
			}
		}

		return fmt.Errorf("unable to get api %v %v", target.NS(ns), err)
	}

	fn(&api.Status)

	return r.Status().Update(ctx, &api)
}

// breakSubgraphLink breaks the link between given ApiDefinition and Subgraph object it refers to. If pass is specified,
// the function skips updating given ApiDefinition's status.
func (r *ApiDefinitionReconciler) breakSubgraphLink(
	ctx context.Context,
	desired *tykv1alpha1.ApiDefinition,
	pass bool,
) error {
	if desired.Status.LinkedToSubgraph == "" {
		return nil
	}

	subgraph := &tykv1alpha1.SubGraph{}
	subgraphNamespacedName := types.NamespacedName{
		Namespace: desired.ObjectMeta.Namespace,
		Name:      desired.Status.LinkedToSubgraph,
	}

	// If ApiDefinition is linked to Subgraph (through .status.subgraph_name field) even though it's
	// GraphRef is empty, remove the link between Subgraph CR and ApiDefinition CR respectively.
	err := r.Client.Get(ctx, subgraphNamespacedName, subgraph)
	if err != nil {
		r.Log.Error(err,
			"failed to break the link between ApiDefinition and Subgraph CRs",
			"couldn't find Subgraph", subgraphNamespacedName.String(),
		)

		return err
	}

	if subgraph.Status.LinkedByAPI != "" {
		subgraph.Status.LinkedByAPI = ""

		err = r.Status().Update(ctx, subgraph)
		if err != nil {
			r.Log.Error(err,
				"failed to update Subgraph status after removing ApiDefinition link",
				"Subgraph CR", subgraphNamespacedName.String(),
			)

			return err
		}
	}

	namespace := desired.Namespace
	target := model.Target{Namespace: &namespace, Name: desired.Name}

	if !pass {
		err = r.updateStatus(
			ctx,
			desired.ObjectMeta.Namespace,
			target,
			false,
			func(status *tykv1alpha1.ApiDefinitionStatus) {
				status.LinkedToSubgraph = ""
			},
		)

		if err != nil {
			r.Log.Error(err,
				"failed to update ApiDefinition status after removing Subgraph link",
				"ApiDefinition CR", client.ObjectKeyFromObject(desired).String(),
			)

			return err
		}
	}

	return nil
}

func (r *ApiDefinitionReconciler) processSubGraphExec(ctx context.Context, urs *tykv1alpha1.ApiDefinition) error {
	if urs == nil || urs.Spec.GraphQL == nil {
		return nil
	}

	if urs.Spec.GraphQL.GraphRef == nil || *urs.Spec.GraphQL.GraphRef == "" {
		err := r.breakSubgraphLink(ctx, urs, false)
		if err != nil {
			return err
		}

		r.Log.Info(
			"ApiDefinition has no Subgraph Reference",
			"ApiDefinition CR", fmt.Sprintf("%s/%s", urs.Namespace, urs.Name),
		)

		return nil
	}

	subgraph := &tykv1alpha1.SubGraph{}
	subgraphNamespacedName := types.NamespacedName{
		Namespace: urs.ObjectMeta.Namespace,
		Name:      *urs.Spec.GraphQL.GraphRef,
	}

	err := r.Client.Get(ctx, subgraphNamespacedName, subgraph)
	if err != nil {
		return err
	}

	// SubGraph can only refer to one ApiDefinition. There is one-to-one relationship between ApiDefinition
	// and SubGraph CRs. If another ApiDefinition tries to refer to the SubGraph that is already referred by
	// another ApiDefinition, we should return error to indicate that multiple linking is not allowed.
	if subgraph.Status.LinkedByAPI != "" &&
		subgraph.Status.LinkedByAPI != *urs.Spec.APIID {
		r.Log.Error(ErrMultipleLinkSubGraph, fmt.Sprintf(
			"failed to link ApiDefinition CR with SubGraph CR; SubGraph %q is already linked by %s",
			client.ObjectKeyFromObject(subgraph), subgraph.Status.LinkedByAPI,
		))

		return ErrMultipleLinkSubGraph
	}

	// If ApiDefinition refers to another Subgraph, the link between the previous Subgraph CR and
	// ApiDefinition CR must be broken before updating the current ApiDefinition CR based on the new
	// Subgraph CR.
	if urs.Status.LinkedToSubgraph != *urs.Spec.GraphQL.GraphRef {
		err = r.breakSubgraphLink(ctx, urs, false)
		if err != nil {
			return err
		}
	}

	schema := subgraph.Spec.Schema
	urs.Spec.GraphQL.Schema = &schema
	urs.Spec.GraphQL.Subgraph.SDL = subgraph.Spec.SDL

	namespace := urs.Namespace
	target := model.Target{Namespace: &namespace, Name: urs.Name}

	err = r.updateStatus(
		ctx,
		urs.ObjectMeta.Namespace,
		target,
		false,
		func(status *tykv1alpha1.ApiDefinitionStatus) {
			status.LinkedToSubgraph = subgraph.ObjectMeta.Name
		},
	)
	if err != nil {
		r.Log.Error(err,
			"failed to update ApiDefinition status after adding Subgraph link",
			"ApiDefinition CR", fmt.Sprintf("%s/%s", urs.Namespace, urs.Name),
		)

		return err
	}

	subgraph.Status.LinkedByAPI = *urs.Spec.APIID

	err = r.Status().Update(ctx, subgraph)
	if err != nil {
		r.Log.Error(err,
			"failed to update Subgraph status after adding ApiDefinition link",
			"Subgraph CR", subgraphNamespacedName.String(),
		)

		return err
	}

	return nil
}

func (r *ApiDefinitionReconciler) processSuperGraphExec(ctx context.Context, urs *tykv1alpha1.ApiDefinition) error {
	if urs.Spec.GraphQL.GraphRef == nil || *urs.Spec.GraphQL.GraphRef == "" {
		return errors.New("GraphRef is not set")
	}

	supergraph := &tykv1alpha1.SuperGraph{}

	err := r.Client.Get(ctx, types.NamespacedName{
		Namespace: urs.Namespace,
		Name:      *urs.Spec.GraphQL.GraphRef,
	}, supergraph)
	if err != nil {
		return err
	}

	for _, ref := range supergraph.Spec.SubgraphRefs {
		ns := ref.Namespace

		if ns == nil || *ns == "" {
			if ns == nil {
				ns = new(string)
			}

			*ns = supergraph.Namespace
		}

		subGraph := &tykv1alpha1.SubGraph{}

		err := r.Client.Get(ctx, types.NamespacedName{
			Name:      ref.Name,
			Namespace: *ns,
		}, subGraph)
		if err != nil {
			return err
		}

		apiNS, apiName := decodeID(subGraph.Status.LinkedByAPI)
		apiDef := &tykv1alpha1.ApiDefinition{}

		err = r.Client.Get(ctx, types.NamespacedName{Namespace: apiNS, Name: apiName}, apiDef)
		if err != nil {
			return err
		}

		urs.Spec.GraphQL.Supergraph.Subgraphs = append(urs.Spec.GraphQL.Supergraph.Subgraphs,
			model.GraphQLSubgraphEntity{
				APIID: subGraph.Status.LinkedByAPI,
				Name:  apiDef.Spec.Name,
				URL:   fmt.Sprintf("tyk://%s", apiDef.Name),
				SDL:   subGraph.Spec.SDL,
			})
	}

	schema := supergraph.Spec.Schema
	mergedSDL := supergraph.Spec.MergedSDL

	urs.Spec.GraphQL.Schema = &schema
	urs.Spec.GraphQL.Supergraph.MergedSDL = &mergedSDL

	return err
}

func (r *ApiDefinitionReconciler) findGraphsForApiDefinition(graph client.Object) []reconcile.Request {
	apiDefDeployments := &tykv1alpha1.ApiDefinitionList{}
	listOps := &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(GraphKey, graph.GetName()),
		Namespace:     graph.GetNamespace(),
	}

	if err := r.List(context.TODO(), apiDefDeployments, listOps); err != nil {
		return []reconcile.Request{}
	}

	requests := make([]reconcile.Request, len(apiDefDeployments.Items))
	for i, item := range apiDefDeployments.Items { //nolint
		requests[i] = reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      item.GetName(),
				Namespace: item.GetNamespace(),
			},
		}
	}

	return requests
}

// SetupWithManager initializes the api definition controller.
func (r *ApiDefinitionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	err := mgr.GetFieldIndexer().IndexField(
		context.Background(),
		&tykv1alpha1.ApiDefinition{},
		GraphKey,
		func(rawObj client.Object) []string {
			// Extract the ConfigMap name from the ConfigDeployment Spec, if one is provided
			apiDefDeployment, ok := rawObj.(*tykv1alpha1.ApiDefinition)
			if !ok {
				r.Log.Info("Not ApiDefinition")
				return nil
			}
			if apiDefDeployment.Spec.GraphQL == nil {
				return nil
			}

			if apiDefDeployment.Spec.GraphQL.GraphRef == nil || *apiDefDeployment.Spec.GraphQL.GraphRef == "" {
				return nil
			}

			return []string{*apiDefDeployment.Spec.GraphQL.GraphRef}
		})
	if err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		For(&tykv1alpha1.ApiDefinition{}).
		Owns(&v1.Secret{}).
		Watches(
			&source.Kind{Type: &tykv1alpha1.SubGraph{}},
			handler.EnqueueRequestsFromMapFunc(r.findGraphsForApiDefinition),
			builder.WithPredicates(r.ignoreGraphCreationEvents()),
		).
		Watches(
			&source.Kind{Type: &tykv1alpha1.SuperGraph{}},
			handler.EnqueueRequestsFromMapFunc(r.findGraphsForApiDefinition),
			builder.WithPredicates(r.ignoreGraphCreationEvents()),
		).
		Complete(r)
}

func (r *ApiDefinitionReconciler) ignoreGraphCreationEvents() predicate.Predicate {
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
