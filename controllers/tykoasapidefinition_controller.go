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
	"fmt"
	"strconv"
	"time"

	tykClient "github.com/TykTechnologies/tyk-operator/pkg/client"

	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	util "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/pkg/cert"
	"github.com/TykTechnologies/tyk-operator/pkg/client/klient"
	"github.com/TykTechnologies/tyk-operator/pkg/environment"
	"github.com/TykTechnologies/tyk-operator/pkg/keys"
	"github.com/buger/jsonparser"
	"github.com/go-logr/logr"
)

const (
	TykOASConfigMapKey  = "spec.tykOAS.configmapRef.name"
	TykOASExtenstionStr = "x-tyk-api-gateway"
)

var (
	InfoNameKeys                                 = []string{TykOASExtenstionStr, "info", "name"}
	InfoStateActiveKeys                          = []string{TykOASExtenstionStr, "info", "state", "active"}
	UpstreamURLKeys                              = []string{TykOASExtenstionStr, "upstream", "url"}
	ServerListenpathStripKeys                    = []string{TykOASExtenstionStr, "server", "listenPath", "strip"}
	ServerListenpathValueKeys                    = []string{TykOASExtenstionStr, "server", "listenPath", "value"}
	ServerCustomDomainNameKeys                   = []string{TykOASExtenstionStr, "server", "customDomain", "name"}
	ServerCustomDomainEnabledKeys                = []string{TykOASExtenstionStr, "server", "customDomain", "enabled"}
	ServerCustomDomainCertsKeys                  = []string{TykOASExtenstionStr, "server", "customDomain", "certificates"}
	InfoVersioningKeys                           = []string{TykOASExtenstionStr, "info", "versioning"}
	ServerAuthenticationBaseIdentityProviderKeys = []string{
		TykOASExtenstionStr, "server", "authentication", "baseIdentityProvider",
	}
)

// TykOasApiDefinitionReconciler reconciles a TykOasApiDefinition object
type TykOasApiDefinitionReconciler struct {
	client.Client
	Log    logr.Logger
	Env    environment.Env
	Scheme *runtime.Scheme
}

// TykOasApiDefinitionVersion refrences a version of the Base TykOasAPI.
type TykOasApiDefinitionVersion struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// TykOasApiDefinitionVersioning represents the version information for a base TykOasApiDefinition object
type TykOasApiDefinitionVersioning struct {
	Default  string                        `json:"default"`
	Enabled  bool                          `json:"enabled"`
	Key      string                        `json:"key"`
	Location v1alpha1.LocationType         `json:"location"`
	Name     string                        `json:"name"`
	Versions []*TykOasApiDefinitionVersion `json:"versions"`
}

//+kubebuilder:rbac:groups=tyk.tyk.io,resources=tykoasapidefinitions,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=tyk.tyk.io,resources=tykoasapidefinitions/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=tyk.tyk.io,resources=tykoasapidefinitions/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *TykOasApiDefinitionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("TykOasApiDefinition", req.NamespacedName.String())

	log.Info("Reconciling Tyk OAS instance")

	var reqA time.Duration
	var apiID string
	var markForDeletion bool

	// Lookup Tyk OAS object
	tykOAS := &v1alpha1.TykOasApiDefinition{}
	if err := r.Get(ctx, req.NamespacedName, tykOAS); err != nil {
		if k8serrors.IsNotFound(err) {
			log.Info("failed to get TykOasApiDefinition CR", "err", err.Error())
		}

		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	env, ctx, err := HttpContext(ctx, r.Client, &r.Env, tykOAS, log)
	if err != nil {
		return ctrl.Result{}, err
	}

	if tykOAS.GetLabels()[keys.TykOasApiDefinitionTemplateLabel] == "true" {
		log.Info("Reconciling TykOasApiDefinition Template for Ingress")

		if !tykOAS.ObjectMeta.DeletionTimestamp.IsZero() {
			return ctrl.Result{}, r.deleteOasTpl(ctx, log, tykOAS)
		}

		if !util.ContainsFinalizer(tykOAS, keys.ApiDefTemplateFinalizerName) {
			util.AddFinalizer(tykOAS, keys.ApiDefTemplateFinalizerName)
			return ctrl.Result{}, r.Update(ctx, tykOAS)
		}

		if err = r.reconcileOasTpl(ctx, tykOAS); err != nil {
			return ctrl.Result{}, err
		}

		tykOAS.Status.LatestTransaction = v1alpha1.TransactionInfo{
			Time: metav1.Now(),
		}
		tykOAS.Status.IngressTemplate = true

		if err = r.Status().Update(ctx, tykOAS); err != nil {
			return ctrl.Result{}, err
		}

		log.Info("Successfully reconciled Tyk OasApiDefinition template for Ingress")

		return ctrl.Result{}, nil
	}

	if tykOAS.Status.GetIsVersionedAPI() {
		err := r.reconcileVersionedOasApi(ctx, log, tykOAS)
		if err != nil {
			log.Error(err, "failed to handle versioned Tyk OAS API")

			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	var cmHash string

	_, err = util.CreateOrUpdate(ctx, r.Client, tykOAS, func() error {
		if !tykOAS.ObjectMeta.DeletionTimestamp.IsZero() {
			markForDeletion = true

			if err := r.reconcileDeletingBaseOasAPIVersions(ctx, log, tykOAS); err != nil {
				return err
			}

			return reconcileDelete(ctx, log, tykOAS)
		}

		util.AddFinalizer(tykOAS, keys.TykOASFinalizerName)

		apiID, cmHash, err = r.createOrUpdateTykOASAPI(ctx, tykOAS, log, &env)
		if err != nil {
			log.Error(err, "Failed to create/update Tyk OAS API")

			return err
		}

		return nil
	})

	if !markForDeletion {
		var transactionInfo v1alpha1.TransactionInfo

		transactionInfo.Time = metav1.Now()
		if err == nil {
			transactionInfo.Status = v1alpha1.Successful
		} else {
			reqA = queueAfter

			transactionInfo.Status = v1alpha1.Failed
			transactionInfo.Error = err.Error()

			log.Error(err, "Error from create update")
		}

		tykOAS.Status.LatestTransaction = transactionInfo

		oasOnTyk, _ := klient.Universal.TykOAS().Get(ctx, apiID) //nolint
		tykOAS.Status.LatestTykSpecHash = calculateHash(oasOnTyk)
		tykOAS.Status.LatestCRDSpecHash = calculateHash(tykOAS.Spec)
		tykOAS.Status.LatestConfigMapHash = cmHash

		if err = r.updateStatus(ctx, tykOAS, apiID, log, false); err != nil {
			log.Error(err, "Failed to update status of Tyk OAS CRD")
			return ctrl.Result{RequeueAfter: reqA}, err
		}
	}

	if err := klient.Universal.HotReload(ctx); err != nil {
		log.Error(err, "Failed to reload gateway")
		return ctrl.Result{RequeueAfter: reqA}, err
	}

	log.Info("Completed reconciling Tyk OAS instance")

	return ctrl.Result{RequeueAfter: reqA}, err
}

// reconcileVersionedOasApi checks if a versioned API can be deleted and isn't a dangling Versioned API
// i.e an API that was previously part of the version list of a Base API.
func (r *TykOasApiDefinitionReconciler) reconcileVersionedOasApi(
	ctx context.Context,
	log logr.Logger,
	versionedTykOAS *v1alpha1.TykOasApiDefinition,
) error {
	if err := r.cleanUpDanglingVersionedOasAPIs(ctx, versionedTykOAS, log); err != nil {
		return err
	}

	if !versionedTykOAS.ObjectMeta.DeletionTimestamp.IsZero() {
		if versionedTykOAS.Status.GetIsDefaultVersion() {
			errMsg := "Cannot delete default API version. Remove it from base API's default first"

			versionedTykOAS.Status.LatestTransaction = v1alpha1.TransactionInfo{
				Time:   metav1.Now(),
				Status: v1alpha1.TransactionStatus(errMsg),
			}

			log.Info("error deleting versioned TykOasAPI", errMsg)

			if err := r.Status().Update(ctx, versionedTykOAS); err != nil {
				return err
			}

			return nil
		}

		var baseOasAPI v1alpha1.TykOasApiDefinition

		if err := r.Get(ctx, types.NamespacedName{
			Name:      versionedTykOAS.Status.GetBaseVersionName(),
			Namespace: versionedTykOAS.Status.GetBaseVersionNamespace(),
		}, &baseOasAPI); err != nil {
			if k8serrors.IsNotFound(err) {
				log.Info("failed to get TykOasApiDefinition CR", "err", err.Error())
			}

			return err
		}

		errMsg := "Cannot delete API. It is referenced as version in another Tyk OAS API. Please remove it before deleting."

		versionedTykOAS.Status.LatestTransaction = v1alpha1.TransactionInfo{
			Time:   metav1.Now(),
			Status: v1alpha1.TransactionStatus(errMsg),
		}

		log.Info("error deleting versioned TykOasAPI", errMsg)

		if err := r.Status().Update(ctx, versionedTykOAS); err != nil {
			return err
		}
	}

	return nil
}

func (r *TykOasApiDefinitionReconciler) cleanUpDanglingVersionedOasAPIs(
	ctx context.Context,
	versionedTykOAS *v1alpha1.TykOasApiDefinition,
	log logr.Logger,
) error {
	if versionedTykOAS.Status.GetBaseVersionName() == "" {
		return nil
	}

	var baseOasAPI v1alpha1.TykOasApiDefinition

	if err := r.Get(ctx, types.NamespacedName{
		Name:      versionedTykOAS.Status.GetBaseVersionName(),
		Namespace: versionedTykOAS.Status.GetBaseVersionNamespace(),
	}, &baseOasAPI); err != nil {
		if k8serrors.IsNotFound(err) {
			log.Info("base oas version not found, detaching version from the base OAS API.")

			versionedTykOAS.RemoveOASVersionStatus()

			if err := r.Status().Update(ctx, versionedTykOAS); err != nil {
				return err
			}

			return nil
		}

		return err
	}

	var oasAPIPresent bool

	for _, v := range baseOasAPI.Spec.Versioning.Versions {
		if v.TykOasApiDefinitionRef == versionedTykOAS.Name {
			oasAPIPresent = true
			break
		}
	}

	if !oasAPIPresent {
		versionedTykOAS.RemoveOASVersionStatus()

		if err := r.Status().Update(ctx, versionedTykOAS); err != nil {
			return err
		}
	}

	return nil
}

// handleVersioningTykOASAPI versions a Base TykOasAPI using refrenced TykOasAPIs.
func (r *TykOasApiDefinitionReconciler) handleVersioningTykOASAPI(
	ctx context.Context,
	tykBaseOASCrd *v1alpha1.TykOasApiDefinition,
	tykBaseOASDoc *string,
) error {
	var versions []*TykOasApiDefinitionVersion

	for _, v := range tykBaseOASCrd.Spec.Versioning.Versions {
		var versionOasAPI v1alpha1.TykOasApiDefinition

		if err := r.Get(ctx, types.NamespacedName{
			Name:      v.TykOasApiDefinitionRef,
			Namespace: v.Namespace,
		}, &versionOasAPI); err != nil {
			return err
		}

		if versionOasAPI.Status.VersioningStatus == nil {
			versionOasAPI.Status.NewVersioningStatus()
		}

		if tykBaseOASCrd.Spec.Versioning.Default == v.Name {
			versionOasAPI.Status.VersioningStatus.IsDefaultVersion = true
		}

		versionOasAPI.Status.SetBaseVersionName(tykBaseOASCrd.Name)
		versionOasAPI.Status.SetBaseVersionNamespace(&tykBaseOASCrd.Namespace)
		versionOasAPI.Status.SetIsVersionedAPI(true)

		version := TykOasApiDefinitionVersion{
			ID:   versionOasAPI.Status.ID,
			Name: v.Name,
		}

		versions = append(versions, &version)

		if err := r.Status().Update(ctx, &versionOasAPI); err != nil {
			return err
		}
	}

	versioning := TykOasApiDefinitionVersioning{
		Default:  tykBaseOASCrd.Spec.Versioning.Default,
		Enabled:  tykBaseOASCrd.Spec.Versioning.Enabled,
		Key:      tykBaseOASCrd.Spec.Versioning.Key,
		Location: *tykBaseOASCrd.Spec.Versioning.Location,
		Name:     tykBaseOASCrd.Spec.Versioning.Name,
		Versions: versions,
	}

	byteVersioning, err := json.Marshal(versioning)
	if err != nil {
		return err
	}

	byteData := []byte(*tykBaseOASDoc)

	if byteData, err = jsonparser.Set(byteData, byteVersioning, InfoVersioningKeys...); err != nil {
		return err
	}

	*tykBaseOASDoc = string(byteData)

	return nil
}

func (r *TykOasApiDefinitionReconciler) createOrUpdateTykOASAPI(
	ctx context.Context, tykOASCrd *v1alpha1.TykOasApiDefinition, log logr.Logger, env *environment.Env,
) (string, string, error) {
	var cm v1.ConfigMap

	configMapNs := tykOASCrd.Spec.TykOAS.ConfigmapRef.Namespace
	if configMapNs == "" {
		configMapNs = tykOASCrd.Namespace
	}

	ns := types.NamespacedName{
		Name:      tykOASCrd.Spec.TykOAS.ConfigmapRef.Name,
		Namespace: configMapNs,
	}

	err := r.Client.Get(ctx, ns, &cm)
	if err != nil {
		log.Error(err, "Failed to fetch config map")
		return "", "", err
	}

	tykOASDoc := cm.Data[tykOASCrd.Spec.TykOAS.ConfigmapRef.KeyName]
	cmHash := calculateHash(tykOASDoc)

	_, _, _, err = jsonparser.Get([]byte(tykOASDoc), TykOASExtenstionStr)
	if err != nil {
		errMsg := "invalid Tyk OAS APIDefinition. Failed to fetch value of `x-tyk-api-gateway` "
		log.Error(err, errMsg)

		return "", "", fmt.Errorf("%s: %s", errMsg, err.Error())
	}

	apiID, err := getAPIID(tykOASCrd, tykOASDoc)
	if err != nil {
		return "", "", err
	}

	if apiID == "" {
		apiID = EncodeNS(client.ObjectKeyFromObject(tykOASCrd).String())
	}

	if tykOASDoc, err = r.processClientCertificate(ctx, env, log, tykOASCrd, tykOASDoc); err != nil {
		log.Error(err, "failed to process client certificate")
		return "", "", err
	}

	if tykOASCrd.Spec.Versioning != nil && tykOASCrd.Spec.Versioning.Enabled {
		err := r.handleVersioningTykOASAPI(ctx, tykOASCrd, &tykOASDoc)
		if err != nil {
			log.Error(err, "Failed to version Tyk OAS API")

			return "", "", err
		}
	}

	exists := klient.Universal.TykOAS().Exists(ctx, apiID)
	if !exists {
		if err = klient.Universal.TykOAS().Create(ctx, apiID, tykOASDoc); err != nil {
			log.Error(err, "failed to create Tyk OAS API")
			return "", "", err
		}
	} else {
		err = r.tykUpdate(ctx, log, &updateReq{
			tykOasDoc:      tykOASDoc,
			apiID:          apiID,
			crd:            tykOASCrd,
			existingCmHash: cmHash,
		})
		if err != nil {
			log.Error(err, "failed to update Tyk OAS API on Tyk")
			return "", "", err
		}
	}

	return apiID, cmHash, nil
}

type updateReq struct {
	crd            *v1alpha1.TykOasApiDefinition
	tykOasDoc      string
	apiID          string
	existingCmHash string
}

func (r *TykOasApiDefinitionReconciler) tykUpdate(ctx context.Context, l logr.Logger, req *updateReq) error {
	if req == nil || req.crd == nil {
		return fmt.Errorf("invalid Tyk update request")
	}

	oasApiDefOnTyk, _ := klient.Universal.TykOAS().Get(ctx, req.apiID) //nolint
	// If we have same OAS API Definition on Tyk, we do not need to send Update to Tyk.
	// So, we can simply return to main reconciliation logic.
	if isSame(req.crd.Status.LatestTykSpecHash, oasApiDefOnTyk) &&
		isSame(req.crd.Status.LatestCRDSpecHash, req.crd.Spec) &&
		req.existingCmHash == req.crd.Status.LatestConfigMapHash {
		l.Info("No need to update the resource on Tyk side")

		return nil
	}

	if err := klient.Universal.TykOAS().Update(ctx, req.apiID, req.tykOasDoc); err != nil {
		l.Error(err, "failed to update Tyk OAS API")
		return err
	}

	return nil
}

func (r *TykOasApiDefinitionReconciler) processClientCertificate(ctx context.Context,
	env *environment.Env, log logr.Logger, tykOASCrd *v1alpha1.TykOasApiDefinition, tykOASDoc string,
) (string, error) {
	log.Info("Processing client certificate reference")

	clientCertEnabled := tykOASCrd.Spec.ClientCertificate.Enabled
	val := ""

	if clientCertEnabled {
		val = "true"
	} else {
		val = "false"
	}

	result, err := jsonparser.Set([]byte(tykOASDoc), []byte(val), OASClientCertEnabledPath...)
	if err != nil {
		return "", err
	}

	tykOASDoc = string(result)

	if clientCertEnabled {
		for _, secretName := range tykOASCrd.Spec.ClientCertificate.Allowlist {
			var secret v1.Secret
			secretNS := types.NamespacedName{Name: secretName, Namespace: tykOASCrd.Namespace}

			if err := r.Client.Get(ctx, secretNS, &secret); err != nil {
				log.Error(err, "failed to fetch secret", "secret", secretNS.String())
				return "", err
			}

			tlsCrt := secret.Data["tls.crt"]
			tlsKey := secret.Data["tls.key"]

			certID, err := cert.CalculateCertID(env.Org, tlsCrt)
			if err != nil {
				return "", err
			}

			if !isCertificateAlreadyUploaded(ctx, false, tlsCrt, env.Org) {
				if certID, err = klient.Universal.Certificate().Upload(ctx, tlsKey, tlsCrt); err != nil {
					return "", err
				}

				log.Info("Successfully uploaded certificate to Tyk", "certID", certID)
			}

			result, err := OASSetClientCertificatesAllowlist([]byte(tykOASDoc), certID)
			if err != nil {
				return "", err
			}

			tykOASDoc = string(result)
		}
	}

	return tykOASDoc, nil
}

func (r *TykOasApiDefinitionReconciler) updateStatus(
	ctx context.Context,
	tykOASCrd *v1alpha1.TykOasApiDefinition,
	apiID string,
	log logr.Logger,
	isIngTpl bool,
) error {
	log.Info("Updating status of Tyk OAS instance")

	tykOASCrd.Status.IngressTemplate = isIngTpl

	if tykOASCrd.Status.ID == "" {
		tykOASCrd.Status.ID = apiID
	}

	configMapNs := tykOASCrd.Spec.TykOAS.ConfigmapRef.Namespace
	if configMapNs == "" {
		configMapNs = tykOASCrd.Namespace
	}

	cmNS := types.NamespacedName{
		Name:      tykOASCrd.Spec.TykOAS.ConfigmapRef.Name,
		Namespace: configMapNs,
	}

	var cm v1.ConfigMap

	err := r.Client.Get(ctx, cmNS, &cm)
	if err != nil {
		log.Error(err, "Failed to fetch config map")

		tykOASCrd.Status.LatestTransaction.Status = v1alpha1.Failed
		tykOASCrd.Status.LatestTransaction.Error = err.Error()
		tykOASCrd.Status.LatestTransaction.Time = metav1.Now()
	} else {
		tykOASDoc := cm.Data[tykOASCrd.Spec.TykOAS.ConfigmapRef.KeyName]

		name, err := jsonparser.GetString([]byte(tykOASDoc), InfoNameKeys...)
		// do not throw error if field doesn't exist
		if err != nil && err != jsonparser.KeyPathNotFoundError {
			log.Error(err, "Failed to fetch name information from Tyk OAS document")
		} else {
			tykOASCrd.Status.Name = name
		}

		state, err := jsonparser.GetBoolean([]byte(tykOASDoc), InfoStateActiveKeys...)
		// do not throw error if field doesn't exist
		if err != nil && err != jsonparser.KeyPathNotFoundError {
			log.Error(err, "Failed to fetch state information from Tyk OAS document")
		} else {
			tykOASCrd.Status.Enabled = state
		}

		str, err := jsonparser.GetString([]byte(tykOASDoc), ServerCustomDomainNameKeys...)
		// do not throw error if field doesn't exist
		if err != nil && err != jsonparser.KeyPathNotFoundError {
			log.Error(err, "Failed to fetch domain information from Tyk OAS document")
		} else {
			tykOASCrd.Status.Domain = str
		}

		str, err = jsonparser.GetString([]byte(tykOASDoc), ServerListenpathValueKeys...)
		// do not throw error if field doesn't exist
		if err != nil && err != jsonparser.KeyPathNotFoundError {
			log.Error(err, "Failed to fetch listen path information from Tyk OAS document")
		} else {
			tykOASCrd.Status.ListenPath = str
		}

		str, err = jsonparser.GetString([]byte(tykOASDoc), UpstreamURLKeys...)
		// do not throw error if field doesn't exist
		if err != nil && err != jsonparser.KeyPathNotFoundError {
			log.Error(err, "Failed to fetch upstream url  information from Tyk OAS document")
		} else {
			tykOASCrd.Status.TargetURL = str
		}
	}

	return r.Client.Status().Update(ctx, tykOASCrd)
}

func getAPIID(tykOASCrd *v1alpha1.TykOasApiDefinition, tykOASDoc string) (string, error) {
	if tykOASCrd.Status.ID != "" {
		return tykOASCrd.Status.ID, nil
	}

	val, err := jsonparser.GetString([]byte(tykOASDoc), TykOASExtenstionStr, "info", "id")
	// do not throw error if id doesn't exist
	if err != nil && err != jsonparser.KeyPathNotFoundError {
		return "", err
	}

	return val, nil
}

func deleteOasApi(
	ctx context.Context,
	tykOASCrd *v1alpha1.TykOasApiDefinition,
) error {
	if tykOASCrd == nil {
		return nil
	}

	if tykOASCrd.Status.ID != "" {
		if err := klient.Universal.TykOAS().Delete(ctx, tykOASCrd.Status.ID); err != nil {
			return err
		}
	}

	util.RemoveFinalizer(tykOASCrd, keys.TykOASFinalizerName)

	return nil
}

func (r *TykOasApiDefinitionReconciler) reconcileDeletingBaseOasAPIVersions(
	ctx context.Context,
	log logr.Logger,
	tykOASCrd *v1alpha1.TykOasApiDefinition,
) error {
	if tykOASCrd == nil {
		return nil
	}

	if tykOASCrd.Spec.Versioning != nil && tykOASCrd.Spec.Versioning.Enabled {
		for _, v := range tykOASCrd.Spec.Versioning.Versions {
			var versionOasAPI v1alpha1.TykOasApiDefinition

			if err := r.Get(ctx, types.NamespacedName{
				Name:      v.TykOasApiDefinitionRef,
				Namespace: v.Namespace,
			}, &versionOasAPI); err != nil {
				if k8serrors.IsNotFound(err) {
					log.Info("failed to get versioned TykOasApiDefinition CR", "err", err.Error())

					return nil
				}

				return err
			}

			versionOasAPI.RemoveOASVersionStatus()

			if err := r.Status().Update(ctx, &versionOasAPI); err != nil {
				return err
			}
		}
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TykOasApiDefinitionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(context.Background(),
		&v1alpha1.TykOasApiDefinition{},
		TykOASConfigMapKey,
		func(o client.Object) []string {
			tykOAS := o.(*v1alpha1.TykOasApiDefinition) //nolint:errcheck
			return []string{tykOAS.Spec.TykOAS.ConfigmapRef.Name}
		}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.TykOasApiDefinition{}).
		Watches(
			&source.Kind{Type: &v1.ConfigMap{}},
			handler.EnqueueRequestsFromMapFunc(r.findOASApisDependentOnConfigmap),
			builder.WithPredicates(r.configmapEvents()),
		).
		Complete(r)
}

func (r *TykOasApiDefinitionReconciler) findOASApisDependentOnConfigmap(cm client.Object) []reconcile.Request {
	tykOASAPIs := &v1alpha1.TykOasApiDefinitionList{}
	listOps := &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(TykOASConfigMapKey, cm.GetName()),
	}

	if err := r.List(context.TODO(), tykOASAPIs, listOps); err != nil {
		return []reconcile.Request{}
	}

	requests := make([]reconcile.Request, len(tykOASAPIs.Items))
	for i, item := range tykOASAPIs.Items { //nolint
		requests[i] = reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      item.GetName(),
				Namespace: item.GetNamespace(),
			},
		}
	}

	return requests
}

func (r *TykOasApiDefinitionReconciler) configmapEvents() predicate.Predicate {
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

func (r *TykOasApiDefinitionReconciler) deleteOasTpl(
	ctx context.Context,
	l logr.Logger,
	tykOas *v1alpha1.TykOasApiDefinition,
) error {
	l.Info("TykOasApiDefinition is being deleted")

	// If there are no finalizers, no need to check for Ingress dependencies.
	if !util.ContainsFinalizer(tykOas, keys.ApiDefTemplateFinalizerName) {
		return nil
	}

	var ingList networkingv1.IngressList
	if err := r.List(ctx, &ingList, client.InNamespace(tykOas.Namespace)); err != nil {
		// If there are no ingress in the namespace where this TykOasApiDefinition template lives,
		// no need to return any errors. Otherwise, there is an error while fetching ingress list.
		if !k8serrors.IsNotFound(err) {
			return err
		}
	}

	var dependencies []string

	for i := range ingList.Items {
		if ingList.Items[i].GetAnnotations() != nil &&
			ingList.Items[i].GetAnnotations()[keys.IngressTemplateAnnotation] == tykOas.Name {
			dependencies = append(dependencies, ingList.Items[i].Name)
		}
	}

	if len(dependencies) > 0 {
		return fmt.Errorf("failed to delete TykOasApiDefinition as Ingress resources %+v depend on it", dependencies)
	}

	util.RemoveFinalizer(tykOas, keys.ApiDefTemplateFinalizerName)

	return r.Update(ctx, tykOas)
}

func (r *TykOasApiDefinitionReconciler) reconcileOasTpl(ctx context.Context, tpl *v1alpha1.TykOasApiDefinition) error {
	ingressList := networkingv1.IngressList{}

	err := r.List(ctx, &ingressList, client.InNamespace(tpl.Namespace))
	if err != nil {
		return client.IgnoreNotFound(err)
	}

	for i := range ingressList.Items {
		if ingressList.Items[i].GetAnnotations()[keys.IngressTemplateAnnotation] == tpl.Name {
			key := client.ObjectKey{
				Namespace: ingressList.Items[i].GetNamespace(),
				Name:      ingressList.Items[i].GetName(),
			}

			r.Log.Info("Updating ingress " + key.String())

			if ingressList.Items[i].Labels == nil {
				ingressList.Items[i].Labels = make(map[string]string)
			}

			ingressList.Items[i].Labels[keys.IngressTaintLabel] = strconv.FormatInt(time.Now().UnixNano(), 10)

			if err = r.Update(ctx, &ingressList.Items[i]); err != nil {
				return err
			}
		}
	}

	return nil
}

func reconcileDelete(ctx context.Context, l logr.Logger, oas *v1alpha1.TykOasApiDefinition) error {
	if oas == nil {
		return nil
	}

	if util.ContainsFinalizer(oas, keys.TykOASFinalizerName) {
		if len(oas.GetLinkedPolicies()) > 0 {
			err := fmt.Errorf(
				"failed to delete OAS API Definition %v due to SecurityPolicy dependency",
				objMetaToStr(oas),
			)
			l.Error(err, "Break the link between SecurityPolicy and TykOasApiDefinition "+
				"in order to delete TykOasApiDefinition resource",
			)

			return err
		}

		if oas.Status.ID != "" {
			if err := klient.Universal.TykOAS().Delete(ctx, oas.Status.ID); err != nil {
				if !tykClient.IsNotFound(err) {
					return err
				}

				l.Info("TykOasApiDefinition CR already deleted from Tyk", "ID", oas.Status.ID)
			}
		}

		util.RemoveFinalizer(oas, keys.TykOASFinalizerName)
	}

	return nil
}
