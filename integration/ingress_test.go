package integration

import (
	"context"
	"fmt"
	"testing"

	"github.com/buger/jsonparser"

	ctrl "sigs.k8s.io/controller-runtime"
	cr "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"

	"github.com/TykTechnologies/tyk-operator/controllers"

	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	tykClient "github.com/TykTechnologies/tyk-operator/pkg/client"
	"github.com/TykTechnologies/tyk-operator/pkg/client/klient"
	"github.com/TykTechnologies/tyk-operator/pkg/environment"
	"github.com/TykTechnologies/tyk-operator/pkg/keys"
	"github.com/matryer/is"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/version"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

// minOasGwVersion represents the minimum Tyk Gateway version required to use the new Tyk OAS API Definition.
var minOasGwVersion = version.MustParseGeneric("5.3.0")

// isSupportedTykVersion verifies that if the current version of Tyk, which will be referred from tyk-operator-conf
// secret, is a valid version that includes stable OAS.
func isSupportedTykVersion(t *testing.T, tykEnv *environment.Env) {
	v, err := version.ParseGeneric(tykEnv.TykVersion)
	if err != nil {
		t.Fatal("failed to parse Tyk Version")
	}

	if !v.AtLeast(minOasGwVersion) {
		t.Skip("Tyk OAS API Definition requires at least Tyk v5.3.0")
	}
}

func TestIngressOas(t *testing.T) {
	var (
		eval = is.New(t)

		// ingress is the ingress object that will be reconciled in this test case. During the reconciliation
		// of the ingress, as it will include 2 rules, 2 TykOasApiDefinition CR and 2 ConfigMaps must be created.
		ingress *networkingv1.Ingress
		// ingHost represents the host that one of the rules in the ingress will include.
		ingHost = "cool.host"
		// ingPath represents the path that all rules in the ingress will include.
		ingPath = "/test"
		// ingSvcName represents the service name that all rules in the ingress will include.
		ingSvcName = "anisebackendsvc"
		// ingSvcPort represents the service port number that all rules in the ingress will include.
		ingSvcPort int32 = 8080

		// tykCtx is a context.Context holding credentials of Tyk which allows Tyk Client to talk with Tyk installation.
		tykCtx context.Context

		// Template TykOasApiDefinition and ConfigMap
		tplOasApi *v1alpha1.TykOasApiDefinition
		tplOasCm  *v1.ConfigMap

		// tplOasData refers to the template OAS API Definition (in []byte representing JSON file) which is stored
		// in template ConfigMap
		tplOasData []byte

		// allOasApis is an array of all TykOasApiDefinitions in the current namespace. It is helper to reduce
		// number of requests sent to k8s api.
		allOasApis v1alpha1.TykOasApiDefinitionList

		// expOasStatus defines how newly created TykOasApiDefinitions' status field should look like
		expOasStatus expectedOasStatus
		r            controllers.IngressReconciler
	)

	testOasCreatedViaIngress := features.New("Creating TykOasApiDefinition via Ingress").
		Setup(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			// Obtain Environment configuration to be able to connect Tyk.
			tykEnv, err := generateEnvConfig(ctx, c)
			eval.NoErr(err)

			// Since TykOasApiDefinition will be used in the test, let's skip testing Tyk versions
			// that do not support TykOasApiDefinition - it means skipping running these tests on
			// versions < 5.3.0
			isSupportedTykVersion(t, &tykEnv)

			testNS, ok := ctx.Value(ctxNSKey).(string)
			eval.True(ok)

			tplOasApi, tplOasCm, err = createTestOASApi(ctx, testNS, c, "", map[string]string{
				keys.TykOasApiDefinitionTemplateLabel: "true",
			})
			eval.NoErr(err)
			tplOasData = []byte(tplOasCm.Data[tplOasApi.Spec.TykOAS.ConfigmapRef.KeyName])

			tykCtx = tykClient.SetContext(context.Background(), tykClient.Context{
				Env: tykEnv,
				Log: log.NullLogger{},
			})

			// Create ApiDefinition Reconciler.
			cl, err := createTestClient(c.Client())
			eval.NoErr(err)

			r = controllers.IngressReconciler{
				Client: cl,
				Log:    log.NullLogger{},
				Scheme: cl.Scheme(),
				Env:    tykEnv,
			}

			return ctx
		}).
		Assess("Creating an Ingress by referencing TykOasApiDefinition template",
			func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
				testNS, ok := ctx.Value(ctxNSKey).(string)
				eval.True(ok)

				ingress = newIngress(
					ingTplMeta{name: tplOasApi.Name, kind: v1alpha1.KindTykOasApiDefinition},
					"ingress", testNS, ingHost, ingPath, ingSvcName, ingSvcPort,
				)

				err := c.Client().Resources(testNS).Create(ctx, ingress)
				eval.NoErr(err)

				expOasStatus.upstreamURL = fmt.Sprintf("http://%s.%s.svc:%d", ingSvcName, testNS, ingSvcPort)
				expOasStatus.listenPath = ingPath
				expOasStatus.status = v1alpha1.Successful

				return ctx
			}).
		Assess("Creating an Ingress must create one or more TykOasApiDefinition and ConfigMap based on template on k8s",
			func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
				testNS, ok := ctx.Value(ctxNSKey).(string)
				eval.True(ok)

				err := wait.For(
					conditions.New(c.Client().Resources(testNS)).ResourceMatch(ingress, func(object k8s.Object) bool {
						_, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: cr.ObjectKeyFromObject(ingress)})
						if err != nil {
							t.Logf("Failed to reconcile Ingress, err: %v", err)
						}

						return err == nil
					}), wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
				eval.NoErr(err)

				ls := metav1.LabelSelector{
					MatchLabels: map[string]string{
						keys.IngressLabel: ingress.Name,
					},
				}

				configMaps := v1.ConfigMapList{}

				err = wait.For(func() (done bool, err error) {
					err = c.Client().Resources(testNS).List(ctx, &configMaps, func(lo *metav1.ListOptions) {
						lo.LabelSelector = labels.Set(ls.MatchLabels).String()
					})
					if err != nil {
						t.Logf("failed to list ConfigMaps, err: %v", err)
						return false, nil
					}

					if len(configMaps.Items) != 2 {
						t.Logf("unexpected number of ConfigMap created, got %d", len(configMaps.Items))
						return false, nil
					}

					return true, nil
				}, wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
				eval.NoErr(err)

				oasApis := v1alpha1.TykOasApiDefinitionList{}
				err = wait.For(func() (done bool, err error) {
					err = c.Client().Resources(testNS).List(ctx, &oasApis, func(lo *metav1.ListOptions) {
						lo.LabelSelector = labels.Set(ls.MatchLabels).String()
					})
					if err != nil {
						t.Logf("failed to list TykOasApiDefinitions, err: %v", err)
						return false, nil
					}

					if len(oasApis.Items) != 2 {
						t.Logf("unexpected number of TykOasApiDefinition created, got %d", len(oasApis.Items))
						return false, nil
					}

					return true, nil
				}, wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
				eval.NoErr(err)

				exists := checkOasApisReferExistingConfigMaps(oasApis.Items, configMaps.Items)
				eval.True(exists)

				for _, oasApi := range oasApis.Items {
					eval.NoErr(waitForTykResourceCreation(c, &oasApi))
					eval.NoErr(checkOasApiStatus(&oasApi, expOasStatus))
				}

				allOasApis = oasApis

				return ctx
			}).
		Assess("Created OAS APIs on Tyk must reflect template and Ingress spec",
			func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
				for _, oasApi := range allOasApis.Items {
					err := wait.For(func() (done bool, err error) {
						tykOasApi, err := klient.Universal.TykOAS().Get(tykCtx, oasApi.Status.ID)
						if err != nil {
							t.Logf("failed to get TykOasApiDefinition with ID %v from Tyk, err: %v", oasApi.Status.ID, err)
							return false, nil
						}

						eval.NoErr(validateStrField([]byte(tykOasApi), oasApi.Name, controllers.InfoNameKeys))
						eval.NoErr(validateStrField(
							[]byte(tykOasApi), expOasStatus.listenPath, controllers.ServerListenpathValueKeys),
						)
						eval.NoErr(validateStrField(
							[]byte(tykOasApi), expOasStatus.upstreamURL, controllers.UpstreamURLKeys),
						)
						eval.NoErr(validateDomainName([]byte(tykOasApi), oasApi.Status.Domain))
						eval.NoErr(validateOasAuthentication([]byte(tykOasApi), tplOasData))

						return true, nil
					}, wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
					eval.NoErr(err)
				}

				return ctx
			}).
		Assess("Updating hash of Ingress (host or path) must NOT create a new TykOasApiDefinition and ConfigMap",
			func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
				testNS, ok := ctx.Value(ctxNSKey).(string)
				eval.True(ok)

				currIngress := ingress
				eval.NoErr(c.Client().Resources(testNS).Get(ctx, currIngress.Name, currIngress.Namespace, currIngress))

				const randomListenPath = "/somerandompath"
				currIngress.Spec.Rules[0].HTTP.Paths[0].Path = randomListenPath
				eval.NoErr(c.Client().Resources(testNS).Update(ctx, currIngress))

				err := wait.For(
					conditions.New(c.Client().Resources(testNS)).ResourceMatch(ingress, func(object k8s.Object) bool {
						_, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: cr.ObjectKeyFromObject(ingress)})
						if err != nil {
							t.Logf("Failed to reconcile Ingress, err: %v", err)
						}

						return err == nil
					}), wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
				eval.NoErr(err)

				ls := metav1.LabelSelector{
					MatchLabels: map[string]string{
						keys.IngressLabel: ingress.Name,
					},
				}

				configMaps := v1.ConfigMapList{}
				err = wait.For(func() (done bool, err error) {
					err = c.Client().Resources(testNS).List(ctx, &configMaps, func(lo *metav1.ListOptions) {
						lo.LabelSelector = labels.Set(ls.MatchLabels).String()
					})
					if err != nil {
						t.Logf("failed to list ConfigMaps, err: %v", err)
						return false, nil
					}

					if len(configMaps.Items) != 2 {
						t.Logf(
							"updating the path in ingress must NOT increase the total number of ConfigMaps",
						)
						return false, nil
					}

					return true, nil
				}, wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
				eval.NoErr(err)

				oasApis := v1alpha1.TykOasApiDefinitionList{}
				err = wait.For(func() (done bool, err error) {
					err = c.Client().Resources(testNS).List(ctx, &oasApis, func(lo *metav1.ListOptions) {
						lo.LabelSelector = labels.Set(ls.MatchLabels).String()
					})
					if err != nil {
						t.Logf("failed to list TykOasApiDefinitions, err: %v", err)
						return false, nil
					}

					if len(oasApis.Items) != 2 {
						t.Log(
							"updating the path in ingress must NOT increase the total number of TykOasApiDefinition CRs",
						)
						return false, nil
					}

					var existingOas *v1alpha1.TykOasApiDefinition = nil

					for i := range oasApis.Items {
						if oasApis.Items[i].Status.ListenPath == randomListenPath {
							existingOas = &oasApis.Items[i]
						}
					}

					if existingOas == nil {
						t.Log("Failed to find TykOasApiDefinition CR with new listen path")
						return false, nil
					}

					return true, nil
				}, wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
				eval.NoErr(err)

				return ctx
			}).
		Assess("Adding a new path to Ingress must create a new TykOasApiDefinition and ConfigMap",
			func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
				testNS, ok := ctx.Value(ctxNSKey).(string)
				eval.True(ok)

				currIngress := ingress
				eval.NoErr(c.Client().Resources(testNS).Get(ctx, currIngress.Name, currIngress.Namespace, currIngress))
				pt := networkingv1.PathTypeExact

				currIngress.Spec.Rules[0].HTTP.Paths = append(
					currIngress.Spec.Rules[0].HTTP.Paths,
					networkingv1.HTTPIngressPath{
						Path:     "/newpath",
						PathType: &pt,
						Backend: networkingv1.IngressBackend{
							Service: &networkingv1.IngressServiceBackend{
								Name: "burakssvc",
								Port: networkingv1.ServiceBackendPort{Number: int32(5050)},
							},
						},
					},
				)

				eval.NoErr(c.Client().Resources(testNS).Update(ctx, currIngress))

				err := wait.For(
					conditions.New(c.Client().Resources(testNS)).ResourceMatch(ingress, func(object k8s.Object) bool {
						_, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: cr.ObjectKeyFromObject(ingress)})
						if err != nil {
							t.Logf("Failed to reconcile Ingress, err: %v", err)
						}

						return err == nil
					}), wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
				eval.NoErr(err)

				ls := metav1.LabelSelector{
					MatchLabels: map[string]string{
						keys.IngressLabel: ingress.Name,
					},
				}

				configMaps := v1.ConfigMapList{}
				err = wait.For(func() (done bool, err error) {
					err = c.Client().Resources(testNS).List(ctx, &configMaps, func(lo *metav1.ListOptions) {
						lo.LabelSelector = labels.Set(ls.MatchLabels).String()
					})
					if err != nil {
						t.Logf("failed to list ConfigMaps, err: %v", err)
						return false, nil
					}

					if len(configMaps.Items) != 3 {
						return false, fmt.Errorf(
							"adding a new path to ingress must increase the total number of ConfigMaps",
						)
					}

					return true, nil
				}, wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
				eval.NoErr(err)

				oasApis := v1alpha1.TykOasApiDefinitionList{}
				err = wait.For(func() (done bool, err error) {
					err = c.Client().Resources(testNS).List(ctx, &oasApis, func(lo *metav1.ListOptions) {
						lo.LabelSelector = labels.Set(ls.MatchLabels).String()
					})
					if err != nil {
						t.Logf("failed to list TykOasApiDefinitions, err: %v", err)
						return false, nil
					}

					if len(oasApis.Items) != 3 {
						return false, fmt.Errorf(
							"adding a new path to ingress must increase the total number of TykOasApiDefinition CRs",
						)
					}

					return true, nil
				}, wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
				eval.NoErr(err)

				return ctx
			}).
		Assess("Deleting a path in Ingress must delete its TykOasApiDefinition and ConfigMap",
			func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
				testNS, ok := ctx.Value(ctxNSKey).(string)
				eval.True(ok)

				currIngress := ingress
				eval.NoErr(c.Client().Resources(testNS).Get(ctx, currIngress.Name, currIngress.Namespace, currIngress))

				deletedPath := currIngress.Spec.Rules[0].HTTP.Paths[0].Path
				var newPaths []networkingv1.HTTPIngressPath
				for i := 1; i < len(currIngress.Spec.Rules[0].HTTP.Paths); i++ {
					newPaths = append(newPaths, currIngress.Spec.Rules[0].HTTP.Paths[i])
				}

				currIngress.Spec.Rules[0].HTTP.Paths = newPaths
				eval.NoErr(c.Client().Resources(testNS).Update(ctx, currIngress))

				err := wait.For(
					conditions.New(c.Client().Resources(testNS)).ResourceMatch(ingress, func(object k8s.Object) bool {
						_, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: cr.ObjectKeyFromObject(ingress)})
						if err != nil {
							t.Logf("Failed to reconcile Ingress, err: %v", err)
						}

						return err == nil
					}), wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
				eval.NoErr(err)

				ls := metav1.LabelSelector{
					MatchLabels: map[string]string{
						keys.IngressLabel: ingress.Name,
					},
				}

				configMaps := v1.ConfigMapList{}
				err = wait.For(func() (done bool, err error) {
					err = c.Client().Resources(testNS).List(ctx, &configMaps, func(lo *metav1.ListOptions) {
						lo.LabelSelector = labels.Set(ls.MatchLabels).String()
					})
					if err != nil {
						t.Logf("failed to list ConfigMaps, err: %v", err)
						return false, nil
					}

					if len(configMaps.Items) != 2 {
						return false, fmt.Errorf(
							"deleting a path in ingress must decrease the total number of ConfigMaps",
						)
					}

					return true, nil
				}, wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
				eval.NoErr(err)

				oasApis := v1alpha1.TykOasApiDefinitionList{}
				err = wait.For(func() (done bool, err error) {
					err = c.Client().Resources(testNS).List(ctx, &oasApis, func(lo *metav1.ListOptions) {
						lo.LabelSelector = labels.Set(ls.MatchLabels).String()
					})
					if err != nil {
						t.Logf("failed to list TykOasApiDefinitions, err: %v", err)
						return false, nil
					}

					if len(oasApis.Items) != 2 {
						return false, fmt.Errorf(
							"deleting a path in ingress must decrease the total number of TykOasApiDefinition CRs",
						)
					}

					for i := range oasApis.Items {
						if oasApis.Items[i].Status.ListenPath == deletedPath {
							t.Logf("TykOasApiDefinition CR with listenPath: %v must be deleted", deletedPath)
							return false, nil
						}
					}

					return true, nil
				}, wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
				eval.NoErr(err)

				return ctx
			}).
		Assess("Deleting an Ingress must delete ConfigMap and TykOasApiDefinition from k8s",
			func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
				testNS, ok := ctx.Value(ctxNSKey).(string)
				eval.True(ok)

				eval.NoErr(c.Client().Resources(testNS).Delete(ctx, ingress))

				ls := metav1.LabelSelector{
					MatchLabels: map[string]string{
						keys.IngressLabel: ingress.Name,
					},
				}

				configMaps := v1.ConfigMapList{}
				err := wait.For(func() (done bool, err error) {
					err = c.Client().Resources(testNS).List(ctx, &configMaps, func(lo *metav1.ListOptions) {
						lo.LabelSelector = labels.Set(ls.MatchLabels).String()
					})
					if err != nil {
						t.Logf("failed to list ConfigMaps, err: %v", err)
						return false, nil
					}

					if len(configMaps.Items) != 0 {
						return false, fmt.Errorf("only template ConfigMap must remain")
					}

					return true, nil
				}, wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
				eval.NoErr(err)

				oasApis := v1alpha1.TykOasApiDefinitionList{}
				err = wait.For(func() (done bool, err error) {
					err = c.Client().Resources(testNS).List(ctx, &oasApis, func(lo *metav1.ListOptions) {
						lo.LabelSelector = labels.Set(ls.MatchLabels).String()
					})
					if err != nil {
						t.Logf("failed to list TykOasApiDefinitions, err: %v", err)
						return false, nil
					}

					if len(oasApis.Items) != 0 {
						return false, fmt.Errorf("only template TykOasApiDefinition must remain")
					}

					return true, nil
				}, wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
				eval.NoErr(err)

				return ctx
			}).
		Feature()

	testenv.Test(t, testOasCreatedViaIngress)
}

func TestIngressUpdateTplCm(t *testing.T) {
	var (
		eval = is.New(t)

		// ingressWithTpl is the ingress object that will be reconciled in this test case. During the reconciliation
		// of the ingress, as it will include 2 rules, 2 TykOasApiDefinition CR and 2 ConfigMaps must be created.
		// It will point to dummy TykOasApiDefinition and ConfigMap templates.
		ingressWithTpl *networkingv1.Ingress
		// ingHost represents the host that one of the rules in the ingress will include.
		ingHost = "cool.host"
		// ingPath represents the path that all rules in the ingress will include.
		ingPath = "/test"
		// ingSvcName represents the service name that all rules in the ingress will include.
		ingSvcName = "anisebackendsvc"
		// ingSvcPort represents the service port number that all rules in the ingress will include.
		ingSvcPort int32 = 8080

		// tykCtx is a context.Context holding credentials of Tyk which allows Tyk Client to talk with Tyk installation.
		tykCtx context.Context

		// Template TykOasApiDefinition and ConfigMap
		tplOasApi *v1alpha1.TykOasApiDefinition
		tplOasCm  *v1.ConfigMap
		// tplOasData refers to the template OAS API Definition (in []byte representing JSON file) which is stored
		// in template ConfigMap
		tplOasData   []byte
		baseAuthMode = []byte("\"oauth_key\"")

		// expOasStatus defines how newly created TykOasApiDefinitions' status field should look like
		expOasStatus expectedOasStatus

		r controllers.IngressReconciler
	)

	testUpdateIngressTemplate := features.New("Creating template TykOasApiDefinition and ConfigMap and Ingress").
		Setup(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			// Obtain Environment configuration to be able to connect Tyk.
			tykEnv, err := generateEnvConfig(ctx, c)
			eval.NoErr(err)

			// Since TykOasApiDefinition will be used in the test, let's skip testing Tyk versions
			// that do not support TykOasApiDefinition - it means skipping running these tests on
			// versions < 5.3.0
			isSupportedTykVersion(t, &tykEnv)

			testNS, ok := ctx.Value(ctxNSKey).(string)
			eval.True(ok)

			tplOasApi, tplOasCm, err = createTestOASApi(ctx, testNS, c, "", map[string]string{
				keys.TykOasApiDefinitionTemplateLabel: "true",
			})
			eval.NoErr(err)
			tplOasData = []byte(tplOasCm.Data[tplOasApi.Spec.TykOAS.ConfigmapRef.KeyName])

			ingressWithTpl = newIngress(
				ingTplMeta{name: tplOasApi.Name, kind: v1alpha1.KindTykOasApiDefinition},
				"ingress", testNS, ingHost, ingPath, ingSvcName, ingSvcPort,
			)
			eval.NoErr(c.Client().Resources(testNS).Create(ctx, ingressWithTpl))

			// Create ApiDefinition Reconciler.
			cl, err := createTestClient(c.Client())
			eval.NoErr(err)

			r = controllers.IngressReconciler{
				Client: cl,
				Log:    log.NullLogger{},
				Scheme: cl.Scheme(),
				Env:    tykEnv,
			}

			expOasStatus.upstreamURL = fmt.Sprintf("http://%s.%s.svc:%d", ingSvcName, testNS, ingSvcPort)
			expOasStatus.listenPath = ingPath
			expOasStatus.status = v1alpha1.Successful

			tykCtx = tykClient.SetContext(context.Background(), tykClient.Context{
				Env: tykEnv,
				Log: log.NullLogger{},
			})

			return ctx
		}).
		Assess("Update ConfigMap template",
			func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
				testNS, ok := ctx.Value(ctxNSKey).(string)
				eval.True(ok)

				// Update templates
				var err error
				tplOasData, err = jsonparser.Set(
					tplOasData, baseAuthMode, controllers.ServerAuthenticationBaseIdentityProviderKeys...,
				)
				eval.NoErr(err)

				tplOasCm.Data[tplOasApi.Spec.TykOAS.ConfigmapRef.KeyName] = string(tplOasData)
				eval.NoErr(c.Client().Resources(testNS).Update(ctx, tplOasCm))

				err = wait.For(
					conditions.New(c.Client().Resources(testNS)).
						ResourceMatch(ingressWithTpl, func(object k8s.Object) bool {
							_, err := r.Reconcile(ctx, ctrl.Request{
								NamespacedName: cr.ObjectKeyFromObject(ingressWithTpl),
							})
							if err != nil {
								t.Logf("Failed to reconcile Ingress, err: %v", err)
							}

							return err == nil
						}), wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
				eval.NoErr(err)

				return ctx
			}).
		Assess("Updating ConfigMap template should update Tyk OAS APIs on Tyk",
			func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
				testNS, ok := ctx.Value(ctxNSKey).(string)
				eval.True(ok)

				ls := metav1.LabelSelector{
					MatchLabels: map[string]string{
						keys.IngressLabel: ingressWithTpl.Name,
					},
				}

				oasApis := v1alpha1.TykOasApiDefinitionList{}
				err := wait.For(func() (done bool, err error) {
					err = c.Client().Resources(testNS).List(ctx, &oasApis, func(lo *metav1.ListOptions) {
						lo.LabelSelector = labels.Set(ls.MatchLabels).String()
					})
					if err != nil {
						t.Logf("failed to list TykOasApiDefinitions, err: %v", err)
						return false, nil
					}

					if len(oasApis.Items) != 2 {
						t.Logf("unexpected number of TykOasApiDefinition created, got %d", len(oasApis.Items))
						return false, nil
					}

					return true, nil
				}, wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
				eval.NoErr(err)

				for i := range oasApis.Items {
					err = wait.For(func() (done bool, err error) {
						tykOasApi, err := klient.Universal.TykOAS().Get(tykCtx, oasApis.Items[i].Status.ID)
						if err != nil {
							t.Logf(
								"failed to get TykOasApiDefinition with ID %v from Tyk, err: %v",
								oasApis.Items[i].Status.ID, err,
							)
							return false, nil
						}

						eval.NoErr(validateStrField([]byte(tykOasApi), oasApis.Items[i].Name, controllers.InfoNameKeys))
						eval.NoErr(validateStrField(
							[]byte(tykOasApi), expOasStatus.listenPath, controllers.ServerListenpathValueKeys),
						)
						eval.NoErr(validateStrField(
							[]byte(tykOasApi), expOasStatus.upstreamURL, controllers.UpstreamURLKeys),
						)
						eval.NoErr(validateDomainName([]byte(tykOasApi), oasApis.Items[i].Status.Domain))
						eval.NoErr(validateOasAuthentication([]byte(tykOasApi), tplOasData))

						return true, nil
					}, wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
					eval.NoErr(err)
				}

				return ctx
			}).
		Feature()

	testenv.Test(t, testUpdateIngressTemplate)
}

func TestIngressUpdateTplKind(t *testing.T) {
	var (
		eval = is.New(t)

		allTykOasApis []v1alpha1.TykOasApiDefinition
		// ingressWithoutTpl is the ingress object that will be reconciled in this test case. During the reconciliation
		// of the ingress, as it will include 2 rules, 2 TykOasApiDefinition CR and 2 ConfigMaps must be created.
		ingressWithoutTpl *networkingv1.Ingress
		// ingHost represents the host that one of the rules in the ingress will include.
		ingHost = "cool.host"
		// ingPath represents the path that all rules in the ingress will include.
		ingPath = "/tplkindtest"
		// ingSvcName represents the service name that all rules in the ingress will include.
		ingSvcName = "anisebackendsvc"
		// ingSvcPort represents the service port number that all rules in the ingress will include.
		ingSvcPort int32 = 8080

		// tykCtx is a context.Context holding credentials of Tyk which allows Tyk Client to talk with Tyk installation.
		tykCtx context.Context

		// expOasStatus defines how newly created TykOasApiDefinitions' status field should look like
		expOasStatus expectedOasStatus

		r controllers.IngressReconciler
	)

	testUpdateIngressTemplateKind := features.New("Creating Ingress without template").
		Setup(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			// Obtain Environment configuration to be able to connect Tyk.
			tykEnv, err := generateEnvConfig(ctx, c)
			eval.NoErr(err)

			// Since TykOasApiDefinition will be used in the test, let's skip testing Tyk versions
			// that do not support TykOasApiDefinition - it means skipping running these tests on
			// versions < 5.3.0
			isSupportedTykVersion(t, &tykEnv)

			testNS, ok := ctx.Value(ctxNSKey).(string)
			eval.True(ok)

			ingressWithoutTpl = newIngress(
				ingTplMeta{kind: v1alpha1.KindTykOasApiDefinition},
				"ingress-without-template", testNS, ingHost, ingPath, ingSvcName, ingSvcPort,
			)
			eval.NoErr(c.Client().Resources(testNS).Create(ctx, ingressWithoutTpl))

			// Create ApiDefinition Reconciler.
			cl, err := createTestClient(c.Client())
			eval.NoErr(err)

			r = controllers.IngressReconciler{
				Client: cl,
				Log:    log.NullLogger{},
				Scheme: cl.Scheme(),
				Env:    tykEnv,
			}

			err = wait.For(conditions.New(c.Client().Resources(testNS)).
				ResourceMatch(ingressWithoutTpl, func(object k8s.Object) bool {
					_, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: cr.ObjectKeyFromObject(ingressWithoutTpl)})
					if err != nil {
						t.Logf("Failed to reconcile Ingress, err: %v", err)
					}

					return err == nil
				}), wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
			eval.NoErr(err)

			expOasStatus.upstreamURL = fmt.Sprintf("http://%s.%s.svc:%d", ingSvcName, testNS, ingSvcPort)
			expOasStatus.listenPath = ingPath
			expOasStatus.status = v1alpha1.Successful

			tykCtx = tykClient.SetContext(context.Background(), tykClient.Context{
				Env: tykEnv,
				Log: log.NullLogger{},
			})

			return ctx
		}).
		Assess("Ingress without template must create TykOasApiDefinition and ConfigMaps on k8s",
			func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
				testNS, ok := ctx.Value(ctxNSKey).(string)
				eval.True(ok)

				ls := metav1.LabelSelector{
					MatchLabels: map[string]string{
						keys.IngressLabel: ingressWithoutTpl.Name,
					},
				}

				oasApis := v1alpha1.TykOasApiDefinitionList{}
				err := wait.For(func() (done bool, err error) {
					err = c.Client().Resources(testNS).List(ctx, &oasApis, func(lo *metav1.ListOptions) {
						lo.LabelSelector = labels.Set(ls.MatchLabels).String()
					})
					if err != nil {
						t.Logf("failed to list TykOasApiDefinitions, err: %v", err)
						return false, nil
					}

					if len(oasApis.Items) != 2 {
						t.Logf("unexpected number of TykOasApiDefinition CR created, got %d", len(oasApis.Items))
						return false, nil
					}

					return true, nil
				}, wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
				eval.NoErr(err)
				allTykOasApis = oasApis.Items

				configMaps := v1.ConfigMapList{}
				err = wait.For(func() (done bool, err error) {
					err = c.Client().Resources(testNS).List(ctx, &configMaps, func(lo *metav1.ListOptions) {
						lo.LabelSelector = labels.Set(ls.MatchLabels).String()
					})
					if err != nil {
						t.Logf("failed to list ConfigMaps, err: %v", err)
						return false, nil
					}

					if len(configMaps.Items) != 2 {
						t.Logf("unexpected number of ConfigMap created, got %d", len(configMaps.Items))
						return false, nil
					}

					return true, nil
				}, wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
				eval.NoErr(err)

				return ctx
			}).
		Assess("Ingress without template must create Tyk OAS APIs on Tyk",
			func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
				for i := range allTykOasApis {
					err := wait.For(func() (done bool, err error) {
						tykOasApi, err := klient.Universal.TykOAS().Get(tykCtx, allTykOasApis[i].Status.ID)
						if err != nil {
							t.Logf(
								"failed to get TykOasApiDefinition with ID %v from Tyk, err: %v",
								allTykOasApis[i].Status.ID, err,
							)
							return false, nil
						}

						eval.NoErr(validateStrField([]byte(tykOasApi), allTykOasApis[i].Name, controllers.InfoNameKeys))
						eval.NoErr(validateStrField(
							[]byte(tykOasApi), expOasStatus.listenPath, controllers.ServerListenpathValueKeys),
						)
						eval.NoErr(validateStrField(
							[]byte(tykOasApi), expOasStatus.upstreamURL, controllers.UpstreamURLKeys),
						)

						eval.NoErr(validateDomainName([]byte(tykOasApi), allTykOasApis[i].Status.Domain))

						return true, nil
					}, wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
					eval.NoErr(err)
				}

				return ctx
			}).
		Assess("Updating Ingress's kind must update Tyk OAS APIs to classic ones on k8s",
			func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
				testNS, ok := ctx.Value(ctxNSKey).(string)
				eval.True(ok)

				currIng := networkingv1.Ingress{}
				eval.NoErr(c.Client().Resources(testNS).Get(ctx, ingressWithoutTpl.Name, testNS, &currIng))

				currKind, exists := currIng.GetAnnotations()[keys.IngressTemplateKindAnnotation]
				eval.True(exists)
				eval.True(currKind == v1alpha1.KindTykOasApiDefinition)

				currIng.GetAnnotations()[keys.IngressTemplateKindAnnotation] = v1alpha1.KindApiDefinition
				eval.NoErr(c.Client().Resources(testNS).Update(ctx, &currIng))

				err := wait.For(
					conditions.New(c.Client().Resources(testNS)).ResourceMatch(&currIng, func(object k8s.Object) bool {
						_, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: cr.ObjectKeyFromObject(&currIng)})
						if err != nil {
							t.Logf("Failed to reconcile Ingress, err: %v", err)
						}

						return err == nil
					}), wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
				eval.NoErr(err)

				ls := metav1.LabelSelector{
					MatchLabels: map[string]string{
						keys.IngressLabel: ingressWithoutTpl.Name,
					},
				}

				oasApis := v1alpha1.TykOasApiDefinitionList{}
				err = wait.For(func() (done bool, err error) {
					err = c.Client().Resources(testNS).List(ctx, &oasApis, func(lo *metav1.ListOptions) {
						lo.LabelSelector = labels.Set(ls.MatchLabels).String()
					})
					if err != nil {
						t.Logf("failed to list TykOasApiDefinitions, err: %v", err)
						return false, nil
					}

					if len(oasApis.Items) != 0 {
						t.Logf("all TykOasApiDefinition CRs must be deleted, found %d", len(oasApis.Items))
						return false, nil
					}

					return true, nil
				}, wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
				eval.NoErr(err)

				configMaps := v1.ConfigMapList{}
				err = wait.For(func() (done bool, err error) {
					err = c.Client().Resources(testNS).List(ctx, &configMaps, func(lo *metav1.ListOptions) {
						lo.LabelSelector = labels.Set(ls.MatchLabels).String()
					})
					if err != nil {
						t.Logf("failed to list ConfigMaps, err: %v", err)
						return false, nil
					}

					if len(configMaps.Items) != 0 {
						t.Logf("all ConfigMaps must be deleted, found %d", len(configMaps.Items))
						return false, nil
					}

					return true, nil
				}, wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
				eval.NoErr(err)

				classicApis := v1alpha1.ApiDefinitionList{}
				err = wait.For(func() (done bool, err error) {
					err = c.Client().Resources(testNS).List(ctx, &classicApis, func(lo *metav1.ListOptions) {
						lo.LabelSelector = labels.Set(ls.MatchLabels).String()
					})
					if err != nil {
						t.Logf("failed to list ApiDefinitionList, err: %v", err)
						return false, nil
					}

					if len(classicApis.Items) != 2 {
						t.Logf(
							"all TykOasApiDefinition CRs must be deployed as ApiDefinition, found %d",
							len(oasApis.Items),
						)
						return false, nil
					}

					return true, nil
				}, wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
				eval.NoErr(err)

				for i := range classicApis.Items {
					err = wait.For(func() (done bool, err error) {
						_, err = klient.Universal.Api().Get(tykCtx, classicApis.Items[i].Status.ApiID)
						if err != nil {
							t.Logf("failed to fetch ApiDefinition from Tyk, err: %v", err)
							return false, nil
						}

						return true, nil
					}, wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
					eval.NoErr(err)
				}

				return ctx
			}).
		Feature()

	testenv.Test(t, testUpdateIngressTemplateKind)
}
