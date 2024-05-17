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
	"fmt"
	"strings"

	"github.com/buger/jsonparser"

	"github.com/TykTechnologies/tyk-operator/api/model"
	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/pkg/environment"
	"github.com/TykTechnologies/tyk-operator/pkg/keys"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	util "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// IngressReconciler watches and reconciles Ingress objects
type IngressReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Env      environment.Env
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;create;list;watch;delete;deletecollection
// +kubebuilder:rbac:groups=tyk.tyk.io,resources=tykoasapidefinitions,verbs=get;list;watch;create;update;patch;delete;deletecollection

// Reconcile perform reconciliation logic for Ingress resource that is managed
// by the operator.
func (r *IngressReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := r.Log.WithValues("name", req.NamespacedName)

	desired := &networkingv1.Ingress{}
	if err := r.Get(ctx, req.NamespacedName, desired); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// set context for all api calls inside this reconciliation loop
	env, ctx, err := HttpContext(ctx, r.Client, &r.Env, desired, l)
	if err != nil {
		return ctrl.Result{}, err
	}

	op, err := util.CreateOrUpdate(ctx, r.Client, desired, func() error {
		if !desired.DeletionTimestamp.IsZero() {
			if util.ContainsFinalizer(desired, keys.IngressFinalizerName) {
				util.RemoveFinalizer(desired, keys.IngressFinalizerName)
			}
			return nil
		}

		if !util.ContainsFinalizer(desired, keys.IngressFinalizerName) {
			util.AddFinalizer(desired, keys.IngressFinalizerName)
		}

		return nil
	})
	if err != nil {
		l.Error(err, "failed to update Ingress object", "Op", op)
		return ctrl.Result{}, err
	}

	// Check if template kind, tyk.io/template-kind, is specified in the Ingress annotation.
	// Template Kind can take two values; TykOasApiDefinition or ApiDefinition. By default, Tyk Operator creates
	// ApiDefinition templates if tyk.io/template-kind isn't specified to support backward compatibility.
	tplKind, tplKindExists := desired.Annotations[keys.IngressTemplateKindAnnotation]
	if !tplKindExists || strings.TrimSpace(tplKind) == "" {
		tplKind = v1alpha1.KindApiDefinition
	}

	tplName := desired.Annotations[keys.IngressTemplateAnnotation]
	tplMeta := types.NamespacedName{Name: tplName, Namespace: req.Namespace}

	// Reconcile TykOasApiDefinition for each Rule and Path defined in Ingress.
	if tplKind == v1alpha1.KindTykOasApiDefinition {
		if err = r.reconcileOasApiTemplate(ctx, l, tplMeta, desired); err != nil {
			return ctrl.Result{}, err
		}

		l.Info("Successfully synced Ingress by using TykOasApiDefinition template")

		if err = r.deleteOrphanAPI(ctx, l, nil, desired); err != nil {
			l.Info("failed to delete orphan ApiDefinition CRs", "err", err)
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	// Reconcile ApiDefinition for each Rule and Path defined in Ingress.
	classicApiTpl := r.keyless()

	if tplMeta.Name != "" {
		l.Info("Reconciling ApiDefinition for Ingress", "Template ApiDefinition", tplMeta.String())

		if err = r.Get(ctx, tplMeta, classicApiTpl); err != nil {
			return ctrl.Result{}, err
		}
	} else {
		l.Info("Reconciling ApiDefinition for Ingress without template")
	}

	if !desired.DeletionTimestamp.IsZero() {
		l.Info("Deleted Ingress resource")
		return ctrl.Result{}, nil
	}

	if err = r.reconcileClassicApiDefinition(ctx, l, classicApiTpl, req.Namespace, desired, &env); err != nil {
		l.Error(err, "failed to reconcile classic ApiDefinition")
		return ctrl.Result{}, err
	}

	l.Info("Successfully synced Ingress by using ApiDefinition template")

	if err = r.deleteOrphanOasApis(ctx, l, nil, desired); err != nil {
		l.Info("failed to delete orphan TykOasApiDefinition CRs", "error", err)
		return ctrl.Result{}, err
	}

	if err = r.deleteOrphanOasCMs(ctx, l, nil, desired); err != nil {
		l.Info("failed to delete orphan ConfigMap templates", "error", err)
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *IngressReconciler) keyless() *v1alpha1.ApiDefinition {
	useKeyless := true
	active := true

	return &v1alpha1.ApiDefinition{
		Spec: v1alpha1.APIDefinitionSpec{
			APIDefinitionSpec: model.APIDefinitionSpec{
				Name:             "default-keyless",
				Protocol:         "http",
				UseKeylessAccess: &useKeyless,
				Active:           &active,
				Proxy: model.Proxy{
					TargetURL: "http://example.com",
				},
				VersionData: model.VersionData{
					NotVersioned: true,
				},
			},
		},
	}
}

func (r *IngressReconciler) reconcileClassicApiDefinition(
	ctx context.Context, lg logr.Logger,
	template *v1alpha1.ApiDefinition,
	ns string,
	desired *networkingv1.Ingress,
	env *environment.Env,
) error {
	var existingHashes []string

	for _, rule := range desired.Spec.Rules {
		for _, p := range rule.HTTP.Paths {
			hash := shortHash(rule.Host + p.Path)
			name := r.buildAPIName(ns, desired.Name, hash)
			api := &v1alpha1.ApiDefinition{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: ns,
				},
			}

			lg.Info("sync classic ApiDefinition", "ApiDefinition Name", name)

			op, err := util.CreateOrUpdate(ctx, r.Client, api, func() error {
				api.SetLabels(map[string]string{
					keys.IngressLabel: desired.Name,
					keys.APIDefLabel:  hash,
				})
				api.Spec = *template.Spec.DeepCopy()
				api.Spec.Name = name

				if api.Spec.OrgID == nil {
					api.Spec.OrgID = new(string)
					api.Spec.OrgID = &template.Status.OrgID
				} else if *api.Spec.OrgID == "" {
					api.Spec.OrgID = &template.Status.OrgID
				}

				if api.Spec.Proxy.ListenPath == nil {
					api.Spec.Proxy.ListenPath = new(string)
				}

				*api.Spec.Proxy.ListenPath = p.Path
				svc := p.Backend.Service
				api.Spec.Proxy.TargetURL = fmt.Sprintf("http://%s.%s.svc:%d", svc.Name,
					ns, svc.Port.Number)

				if rule.Host != "" {
					if api.Spec.Domain == nil {
						api.Spec.Domain = new(string)
					}

					*api.Spec.Domain = r.translateHost(rule.Host)
				}

				if env.Ingress.HTTPPort != 0 {
					api.Spec.ListenPort = env.Ingress.HTTPPort
				}

				if !strings.Contains(p.Path, ".well-known/acme-challenge") && !strings.Contains(svc.Name, "cm-acme-http-solver") {
					for _, tls := range desired.Spec.TLS {
						for _, host := range tls.Hosts {
							if rule.Host == host {
								api.Spec.Protocol = "https"
								api.Spec.CertificateSecretNames = []string{
									tls.SecretName,
								}
								api.Spec.ListenPort = env.Ingress.HTTPSPort
							}
						}
					}
				} else {
					// for the acme challenge
					stripListenPath := false
					preserveHostHeader := true

					api.Spec.Proxy.StripListenPath = &stripListenPath
					api.Spec.Proxy.PreserveHostHeader = &preserveHostHeader
				}

				return util.SetControllerReference(desired, api, r.Scheme)
			})
			if err != nil {
				lg.Error(err, "failed to sync api definition", "name", name, "op", op)
				return nil
			}

			lg.Info("successful sync api definition", "name", name, "op", op)

			existingHashes = append(existingHashes, hash)
		}
	}

	return r.deleteOrphanAPI(ctx, lg, existingHashes, desired)
}

func (r *IngressReconciler) translateHost(host string) string {
	return strings.Replace(host, "*", "{?:[^.]+}", 1)
}

func (r *IngressReconciler) deleteOrphanAPI(
	ctx context.Context,
	l logr.Logger,
	existingHashes []string,
	ing *networkingv1.Ingress,
) error {
	s := labels.NewSelector()

	exists, err := labels.NewRequirement(keys.APIDefLabel, selection.Exists, []string{})
	if err != nil {
		return err
	}

	s = s.Add(*exists)

	if existingHashes != nil {
		notIn, err := labels.NewRequirement(keys.APIDefLabel, selection.NotIn, existingHashes)
		if err != nil {
			return err
		}

		s = s.Add(*notIn)
	}

	name, err := labels.NewRequirement(keys.IngressLabel, selection.DoubleEquals, []string{ing.Name})
	if err != nil {
		return err
	}

	s = s.Add(*name)

	l.Info("Deleting orphan ApiDefinitions", "selector", s, "count", len(existingHashes))
	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		return r.DeleteAllOf(ctx, &v1alpha1.ApiDefinition{}, &client.DeleteAllOfOptions{
			ListOptions: client.ListOptions{
				LabelSelector: s,
				Namespace:     ing.Namespace,
			},
			DeleteOptions: client.DeleteOptions{},
		})
	})
}

func (r *IngressReconciler) buildAPIName(nameSpace, name, hash string) string {
	return fmt.Sprintf("%s-%s-%s", nameSpace, name, hash)
}

func shortHash(txt string) string {
	h := sha256.New()
	h.Write([]byte(txt))

	return fmt.Sprintf("%x", h.Sum(nil))[:9]
}

func (r *IngressReconciler) ingressClassEventFilter() predicate.Predicate {
	watch := keys.DefaultIngressClassAnnotationValue

	if override := r.Env.IngressClass; override != "" {
		watch = override
	}

	isOurIngress := func(o runtime.Object) bool {
		switch e := o.(type) {
		case *networkingv1.Ingress:
			return e.GetAnnotations()[keys.IngressClassAnnotation] == watch
		default:
			return false
		}
	}

	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return isOurIngress(e.Object)
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return isOurIngress(e.ObjectNew)
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return isOurIngress(e.Object)
		},
	}
}

// SetupWithManager initializes ingress controller manager
func (r *IngressReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&networkingv1.Ingress{}).
		Owns(&v1alpha1.ApiDefinition{}).
		WithEventFilter(r.ingressClassEventFilter()).
		Complete(r)
}

// reconcileOasApiTemplate reconciles TykOasApiDefinition CRs based on the template specified in the Ingress
// resource via its label.
// During reconciliation of TykOasApiDefinitions, the reconciler creates a ConfigMap and a TykOasApiDefinition
// resource on k8s based on each unique combination in Ingress Rules.
// Each combination comes from each path defined in ingress rules. For example, there are 2 Ingress rule defined
// where each of the rule contains 3 Ingress path, the reconciler will create 2 * 3 = 6 TykOasApiDefinition
// and ConfigMaps.
func (r *IngressReconciler) reconcileOasApiTemplate(
	ctx context.Context,
	l logr.Logger,
	tplMeta types.NamespacedName,
	ing *networkingv1.Ingress,
) error {
	tplCm := &corev1.ConfigMap{}
	tplOasApi := &v1alpha1.TykOasApiDefinition{}

	if tplMeta.Name != "" {
		l.Info("Reconciling TykOasApiDefinition for Ingress", "Template TykOasApiDefinition", tplMeta.String())

		if err := r.Get(ctx, tplMeta, tplOasApi); err != nil {
			return err
		}

		// Validate that TykOasApiDefinition template includes a config map reference.
		if tplOasApi.Spec.TykOAS.ConfigmapRef.Name == "" {
			return fmt.Errorf("TykOasApiDefinition template does not refer to a ConfigMap, template: %v", tplMeta.String())
		}

		// Obtain metadata of the template ConfigMap which will include the base Tyk OAS Definition for this Ingress.
		tplCmMeta := types.NamespacedName{
			Name:      tplOasApi.Spec.TykOAS.ConfigmapRef.Name,
			Namespace: tplOasApi.Spec.TykOAS.ConfigmapRef.Namespace,
		}
		if tplCmMeta.Namespace == "" {
			tplCmMeta.Namespace = tplOasApi.ObjectMeta.Namespace
		}

		// Fetch template ConfigMap from k8s.
		if err := r.Get(ctx, tplCmMeta, tplCm); err != nil {
			return err
		}
	} else {
		l.Info("Reconciling TykOasApiDefinition for Ingress without template")
		tplOasApi, tplCm = defaultOasIngressTemplate()
	}

	tykOasDocStr, exists := tplCm.Data[tplOasApi.Spec.TykOAS.ConfigmapRef.KeyName]
	if !exists {
		return fmt.Errorf(
			"ConfigMap %s does not include the key '%s' specified in template TykOasApiDefinition %s\n",
			objMetaToStr(tplCm), tplOasApi.Spec.TykOAS.ConfigmapRef.KeyName, objMetaToStr(tplOasApi),
		)
	}

	var existingHashes []string

	for _, rule := range ing.Spec.Rules {
		for _, httpPath := range rule.HTTP.Paths {
			hash := shortHash(rule.Host + httpPath.Path)
			name := fmt.Sprintf("oas-%v", r.buildAPIName(ing.Namespace, ing.Name, hash))
			listenPathValue := httpPath.Path
			upstreamURL := fmt.Sprintf(
				"http://%s.%s.svc:%d", httpPath.Backend.Service.Name, ing.Namespace, httpPath.Backend.Service.Port.Number,
			)

			customDomain := ""
			if rule.Host != "" {
				customDomain = r.translateHost(rule.Host)
			}

			listenPathStrip := "true"
			// preserveHostHeader is not supported in Tyk OAS API Definition at the moment.
			// For more details,
			//	please see: https://tyk.io/docs/getting-started/using-oas-definitions/oas-reference/#api-level-global-features
			// preserveHostHeader := false

			var certificateSecretNamesByte []byte

			if !strings.Contains(httpPath.Path, ".well-known/acme-challenge") &&
				!strings.Contains(httpPath.Backend.Service.Name, "cm-acme-http-solver") {
				for _, tls := range ing.Spec.TLS {
					for _, host := range tls.Hosts {
						if rule.Host == host {
							certificateSecretNamesByte = append(certificateSecretNamesByte, []byte(tls.SecretName)...)
						}
					}
				}
			} else {
				listenPathStrip = "false"
				// preserveHostHeader = true
			}

			// Reconcile new config map
			newConfigMap := corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: ing.Namespace,
				},
			}

			l.Info(
				"Creating a ConfigMap for TykOasApiDefinition in Ingress",
				"ConfigMap", objMetaToStr(&newConfigMap),
				"Template ConfigMap", objMetaToStr(tplCm),
			)

			op, err := util.CreateOrUpdate(ctx, r.Client, &newConfigMap, func() error {
				newConfigMap.SetLabels(map[string]string{
					keys.IngressLabel:             ing.Name,
					keys.TykOASApiDefinitionLabel: hash,
				})

				newConfigMap.Data = tplCm.Data
				tykOasDoc := []byte(tykOasDocStr)

				var err error
				tykOasDoc, err = setOasInfoName(tykOasDoc, strToJsonStrByte(name))
				if err != nil {
					l.Info("Failed to set info.name of OAS API Definition in ConfigMap",
						"ConfigMap", objMetaToStr(&newConfigMap),
						"info.name", name, "error", err,
					)
					return err
				}

				if customDomain != "" {
					tykOasDoc, err = setOasCustomDomain(tykOasDoc, strToJsonStrByte(customDomain), []byte("true"))
					if err != nil {
						l.Info("Failed to set server.customDomain OAS API Definition in ConfigMap",
							"ConfigMap", objMetaToStr(&newConfigMap),
							"server.customDomain.name", customDomain, "error", err,
						)

						return err
					}
				}

				if len(certificateSecretNamesByte) > 0 {
					// it is not supported until OAS support fetching from k8s secrets
					tykOasDoc, err = jsonparser.Set(tykOasDoc, certificateSecretNamesByte, ServerCustomDomainCertsKeys...)
					if err != nil {
						l.Info("Failed to set customDomain.certificates of OAS API Definition in ConfigMap",
							"ConfigMap", objMetaToStr(&newConfigMap),
							"error", err,
						)

						return err
					}
				}

				tykOasDoc, err = setOasListenPath(tykOasDoc, strToJsonStrByte(listenPathValue))
				if err != nil {
					l.Info("Failed to set listenPath.value in ConfigMap",
						"ConfigMap", objMetaToStr(&newConfigMap),
						"server.listenPath.value", listenPathValue, "error", err,
					)

					return err
				}

				tykOasDoc, err = setListenPathStrip(tykOasDoc, []byte(listenPathStrip))
				if err != nil {
					l.Info("Failed to set listenPath.strip in ConfigMap",
						"ConfigMap", objMetaToStr(&newConfigMap),
						"server.listenPath.strip", listenPathStrip, "error", err,
					)

					return err
				}

				tykOasDoc, err = setUpstreamURL(tykOasDoc, strToJsonStrByte(upstreamURL))
				if err != nil {
					l.Info("Failed to set upstream.url in ConfigMap",
						"ConfigMap", objMetaToStr(&newConfigMap),
						"upstream.url", upstreamURL, "error", err,
					)

					return err
				}

				newConfigMap.Data = map[string]string{
					tplOasApi.Spec.TykOAS.ConfigmapRef.KeyName: string(tykOasDoc),
				}

				return util.SetControllerReference(ing, &newConfigMap, r.Scheme)
			})
			if err != nil {
				l.Error(err,
					"failed to sync ConfigMap",
					"ConfigMap", objMetaToStr(&newConfigMap),
					"op", op,
				)

				return err
			}

			l.Info(
				"Successfully created a ConfigMap",
				"ConfigMap", objMetaToStr(&newConfigMap),
				"Template ConfigMap", objMetaToStr(tplCm),
			)

			// Reconcile new TykOasApiDefinition CR
			newOasApi := v1alpha1.TykOasApiDefinition{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: ing.Namespace,
				},
			}

			l.Info(
				"Creating a TykOasApiDefinition for Ingress",
				"TykOasApiDefinition", objMetaToStr(&newOasApi),
				"Template", objMetaToStr(tplOasApi),
			)

			op, err = util.CreateOrUpdate(ctx, r.Client, &newOasApi, func() error {
				newOasApi.SetLabels(map[string]string{
					keys.IngressLabel:             ing.Name,
					keys.TykOASApiDefinitionLabel: hash,
				})

				newOasApi.Spec.TykOAS.ConfigmapRef.Name = newConfigMap.Name
				newOasApi.Spec.TykOAS.ConfigmapRef.Namespace = newConfigMap.Namespace
				newOasApi.Spec.TykOAS.ConfigmapRef.KeyName = tplOasApi.Spec.TykOAS.ConfigmapRef.KeyName

				return util.SetControllerReference(ing, &newOasApi, r.Scheme)
			})
			if err != nil {
				l.Error(err, "failed to sync TykOasApiDefinition",
					"TykOasApiDefinition", objMetaToStr(&newOasApi),
					"op", op,
				)

				return err
			}

			l.Info(
				"Successfully created a TykOasApiDefinition",
				"TykOasApiDefinition", objMetaToStr(&newOasApi),
				"Template", objMetaToStr(tplOasApi),
			)

			existingHashes = append(existingHashes, hash)
		}
	}

	err := r.deleteOrphanOasApis(ctx, l, existingHashes, ing)
	if err != nil {
		return err
	}

	return r.deleteOrphanOasCMs(ctx, l, existingHashes, ing)
}

// deleteOrphanOasApis deletes TykOasApiDefinition CRs that are created by Ingress, which have outdated hashes.
// The 'tyk.io/tykoasapidefinition' label in TykOasApiDefinition CRs created by Ingress indicates the hash
// of a unique combination of each HTTP path in each Ingress Rule.
// Therefore, deleteOrphanOasApis deletes TykOasApiDefinitions that are not part of the newly reconciled Ingress object.
// For example, if the host is updated, the Ingress controller will create a new TykOasApiDefinition and its ConfigMap
// based on this new host. Meanwhile, deleteOrphanOasApis method deletes the TykOasApiDefinition that contains old host
// information as it is outdated and redundant.
func (r *IngressReconciler) deleteOrphanOasApis(
	ctx context.Context,
	l logr.Logger,
	existingHashes []string,
	ing *networkingv1.Ingress,
) error {
	s := labels.NewSelector()

	oasApiLabelExists, err := labels.NewRequirement(keys.TykOASApiDefinitionLabel, selection.Exists, []string{})
	if err != nil {
		return err
	}

	if existingHashes != nil {
		notIn, err := labels.NewRequirement(keys.TykOASApiDefinitionLabel, selection.NotIn, existingHashes)
		if err != nil {
			return err
		}

		s = s.Add(*notIn)
	}

	name, err := labels.NewRequirement(keys.IngressLabel, selection.DoubleEquals, []string{ing.Name})
	if err != nil {
		return err
	}

	s = s.Add(*oasApiLabelExists)
	s = s.Add(*name)

	l.Info("Deleting orphan TykOasApiDefinitions", "selector", s, "count", len(existingHashes))

	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		return r.DeleteAllOf(ctx, &v1alpha1.TykOasApiDefinition{}, &client.DeleteAllOfOptions{
			ListOptions: client.ListOptions{
				LabelSelector: s,
				Namespace:     ing.Namespace,
			},
		})
	})
}

// deleteOrphanOasCMs deletes redundant ConfigMaps that are created by Ingress for TykOasApiDefinition.
// The 'tyk.io/tykoasapidefinition' label in ConfigMaps created by Ingress indicates the hash of a unique combination
// of each HTTP path in each Ingress Rule. Therefore, deleteOrphanOasCMs deletes ConfigMaps that are not part of the
// newly reconciled Ingress object.
// For example, if the host is updated, the Ingress controller will create a new TykOasApiDefinition and its ConfigMap
// based on this new host. Meanwhile, deleteOrphanOasCMs method deletes the ConfigMap that contains old host
// information as it is outdated and redundant.
func (r *IngressReconciler) deleteOrphanOasCMs(
	ctx context.Context,
	l logr.Logger,
	existingHashes []string,
	ing *networkingv1.Ingress,
) error {
	s := labels.NewSelector()

	oasApiLabelExists, err := labels.NewRequirement(keys.TykOASApiDefinitionLabel, selection.Exists, []string{})
	if err != nil {
		return err
	}

	if existingHashes != nil {
		notIn, err := labels.NewRequirement(keys.TykOASApiDefinitionLabel, selection.NotIn, existingHashes)
		if err != nil {
			return err
		}

		s = s.Add(*notIn)
	}

	name, err := labels.NewRequirement(keys.IngressLabel, selection.DoubleEquals, []string{ing.Name})
	if err != nil {
		return err
	}

	s = s.Add(*oasApiLabelExists)
	s = s.Add(*name)

	l.Info("Deleting orphan ConfigMaps", "selector", s, "count", len(existingHashes))

	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		return r.DeleteAllOf(ctx, &corev1.ConfigMap{}, &client.DeleteAllOfOptions{
			ListOptions: client.ListOptions{
				LabelSelector: s,
				Namespace:     ing.Namespace,
			},
		})
	})
}

// setUpstreamURL sets required upstream field ('upstream.url').
func setUpstreamURL(tykOasDoc, upstreamURL []byte) ([]byte, error) {
	return jsonparser.Set(tykOasDoc, upstreamURL, UpstreamURLKeys...)
}

// setListenPathStrip sets required listenPath field ('server.listenPath.strip').
func setListenPathStrip(tykOasDoc, listenPathStrip []byte) ([]byte, error) {
	return jsonparser.Set(tykOasDoc, listenPathStrip, ServerListenpathStripKeys...)
}

// setOasListenPath sets 'server.listenPath.value' field in OAS API Definition based on the given given arguments.
func setOasListenPath(tykOasDoc, listenPath []byte) ([]byte, error) {
	return jsonparser.Set(tykOasDoc, listenPath, ServerListenpathValueKeys...)
}

// setOasCustomDomain sets required customDomain fields ('server.customDomain.name' and 'server.customDomain.enabled')
// in OAS API Definition based on the given given arguments.
func setOasCustomDomain(tykOASDoc, customDomain, enabled []byte) ([]byte, error) {
	var err error

	tykOASDoc, err = jsonparser.Set(tykOASDoc, customDomain, ServerCustomDomainNameKeys...)
	if err != nil {
		return nil, err
	}

	tykOASDoc, err = jsonparser.Set(tykOASDoc, enabled, ServerCustomDomainEnabledKeys...)
	if err != nil {
		return nil, err
	}

	return tykOASDoc, nil
}

// setOasInfoName sets OAS APIs info.name field.
func setOasInfoName(tykOASDoc, infoName []byte) ([]byte, error) {
	return jsonparser.Set(tykOASDoc, infoName, InfoNameKeys...)
}
