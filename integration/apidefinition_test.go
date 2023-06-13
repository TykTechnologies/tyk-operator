package integration

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/TykTechnologies/tyk-operator/api/model"
	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/controllers"
	"github.com/TykTechnologies/tyk-operator/pkg/cert"
	tykClient "github.com/TykTechnologies/tyk-operator/pkg/client"
	"github.com/TykTechnologies/tyk-operator/pkg/client/klient"
	"github.com/TykTechnologies/tyk-operator/pkg/environmet"
	"github.com/google/uuid"
	"github.com/matryer/is"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/version"
	ctrl "sigs.k8s.io/controller-runtime"
	cr "sigs.k8s.io/controller-runtime/pkg/client"
	util "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

const (
	errFailedToGetApiDefCRMsg  = "failed to get ApiDefinition CR"
	errFailedToGetApiDefTykMsg = "failed to get ApiDefinition from Tyk"
)

// deleteApiDefinitionFromTyk sends a Tyk API call to delete ApiDefinition with given ID.
func deleteApiDefinitionFromTyk(ctx context.Context, id string) error {
	err := wait.For(func() (done bool, err error) {
		_, err = klient.Universal.Api().Delete(ctx, id)
		if err != nil {
			return false, err
		}

		err = klient.Universal.HotReload(ctx)
		if err != nil {
			return false, err
		}

		return true, nil
	}, wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))

	return err
}

// TestDeletingNonexistentAPI tests if deleting nonexistent resources cause an error or not on k8s level.
// Assume that the user deleted ApiDefinition resource from Tyk instead of deleting it via kubectl.
// This will create a drift between Tyk and K8s. Deleting the same resource from k8s shouldn't cause any
// external API errors such as 404.
func TestDeletingNonexistentAPI(t *testing.T) {
	var (
		eval   = is.New(t)
		tykCtx context.Context
		tykEnv environmet.Env
	)

	testDeletingNonexistentAPIs := features.New("Deleting Nonexistent APIs from k8s").
		Setup(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			opConfSecret := v1.Secret{}

			err := c.Client().Resources(opNs).Get(ctx, operatorSecret, opNs, &opConfSecret)
			eval.NoErr(err)

			// Obtain Environment configuration to be able to connect Tyk.
			tykEnv, err = generateEnvConfig(&opConfSecret)
			eval.NoErr(err)

			tykCtx = tykClient.SetContext(context.Background(), tykClient.Context{
				Env: tykEnv,
				Log: log.NullLogger{},
			})

			return ctx
		}).
		Assess("Delete nonexistent ApiDefinition from k8s successfully",
			func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
				// To begin with, delete the ApiDefinition from Tyk, which is the wrong thing to do because it'll
				// cause a drift between k8s and Tyk. Now, deleting ApiDefinition CR from k8s,
				// `kubectl delete tykapis <resource_name>`, must be handled gracefully.
				testNs, ok := ctx.Value(ctxNSKey).(string)
				eval.True(ok)

				// First, create the ApiDefinition
				apiDefCR, err := createTestAPIDef(ctx, c, testNs, func(apiDef *v1alpha1.ApiDefinition) {})
				eval.NoErr(err)

				err = waitForTykResourceCreation(c, apiDefCR)
				eval.NoErr(err)

				err = deleteApiDefinitionFromTyk(tykCtx, apiDefCR.Status.ApiID)
				eval.NoErr(err)

				err = c.Client().Resources(testNs).Delete(ctx, apiDefCR)
				eval.NoErr(err)

				err = wait.For(
					conditions.New(c.Client().Resources()).ResourceDeleted(apiDefCR),
					wait.WithTimeout(defaultWaitTimeout),
					wait.WithInterval(defaultWaitInterval),
				)
				eval.NoErr(err)

				return ctx
			}).Feature()

	testenv.Test(t, testDeletingNonexistentAPIs)
}

// TestReconcileNonexistentAPI tests whether reconciliation finishes successfully if ApiDefinition
// does not exist on Tyk. Reconciliation logic must handle API calls to Tyk Gateway / Dashboard for
// nonexistent ApiDefinitions and create it if needed. So that, the k8s remains as source of truth.
func TestReconcileNonexistentAPI(t *testing.T) {
	var (
		eval   = is.New(t)
		tykCtx context.Context
		tykEnv environmet.Env
		r      controllers.ApiDefinitionReconciler
	)

	testReconcilingNonexistentAPIs := features.New("Reconciling Nonexistent ApiDefinition CRs").
		Setup(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			opConfSecret := v1.Secret{}

			err := c.Client().Resources(opNs).Get(ctx, operatorSecret, opNs, &opConfSecret)
			eval.NoErr(err)

			// Obtain Environment configuration to be able to connect Tyk.
			tykEnv, err = generateEnvConfig(&opConfSecret)
			eval.NoErr(err)

			tykCtx = tykClient.SetContext(context.Background(), tykClient.Context{
				Env: tykEnv,
				Log: log.NullLogger{},
			})

			// Create ApiDefinition Reconciler.
			cl, err := createTestClient(c.Client())
			eval.NoErr(err)

			r = controllers.ApiDefinitionReconciler{
				Client: cl,
				Log:    log.NullLogger{},
				Scheme: cl.Scheme(),
				Env:    tykEnv,
			}

			return ctx
		}).
		Assess("Create a drift between Tyk and k8s by deleting an ApiDefinition from Tyk",
			func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
				// To begin with, we should create a drift between Tyk and K8s. In order to do that
				// first create an ApiDefinition, then delete it from Tyk via Tyk API calls. The next
				// reconciliation request must understand nonexistent entity and create it from scratch.
				testNs, ok := ctx.Value(ctxNSKey).(string)
				eval.True(ok)

				// First, create the ApiDefinition.
				apiDefCR, err := createTestAPIDef(ctx, c, testNs, nil)
				eval.NoErr(err)

				err = waitForTykResourceCreation(c, apiDefCR)
				eval.NoErr(err)

				// Here, we create a drift between Tyk and k8s. Although the resource is created on k8s,
				// we manually delete it from Tyk. Since the k8s is unaware of this change, this scenario
				// causes drift between them.
				err = deleteApiDefinitionFromTyk(tykCtx, apiDefCR.Status.ApiID)
				eval.NoErr(err)

				// Ensure that the resource does not exist on Tyk.
				err = wait.For(func() (done bool, err error) {
					_, err = klient.Universal.Api().Get(tykCtx, apiDefCR.Status.ApiID)
					if err != nil {
						return true, nil
					}

					// TODO(buraksekili): API should return 404 in case of nonexistent resource deletion.
					// Because the response may contain other types of errors.
					//	if tykClient.IsNotFound(err) {
					//		return true, nil
					//	}

					return false, err
				}, wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
				eval.NoErr(err)

				// Now, send a reconciliation request to Operator. In the next reconciliation request, the operator
				// must understand the change between Tyk and K8s and create nonexistent ApiDefinition again based
				// on k8s state.
				err = wait.For(func() (done bool, err error) {
					_, err = r.Reconcile(ctx, ctrl.Request{NamespacedName: cr.ObjectKeyFromObject(apiDefCR)})
					return err == nil, err
				}, wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
				eval.NoErr(err)

				// Ensure that the resource is recreated after reconciliation.
				err = wait.For(func() (done bool, err error) {
					_, err = klient.Universal.Api().Get(tykCtx, apiDefCR.Status.ApiID)
					return err == nil, err
				}, wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
				eval.NoErr(err)

				return ctx
			}).Feature()

	testenv.Test(t, testReconcilingNonexistentAPIs)
}

// TestApiDefinitionUpdate tests if changes in ApiDefinition CR updates corresponding ApiDefinition on Tyk.
func TestApiDefinitionUpdate(t *testing.T) {
	var (
		eval        = is.New(t)
		tykCtx      context.Context
		tykEnv      environmet.Env
		updatedName = "updatedName"
	)

	testApiDefinitionUpdate := features.New("Updating ApiDefinition CRs on k8s").
		Setup(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			opConfSecret := v1.Secret{}

			err := c.Client().Resources(opNs).Get(ctx, "tyk-operator-conf", opNs, &opConfSecret)
			eval.NoErr(err)

			// Obtain Environment configuration to be able to connect Tyk.
			tykEnv, err = generateEnvConfig(&opConfSecret)
			eval.NoErr(err)

			tykCtx = tykClient.SetContext(context.Background(), tykClient.Context{
				Env: tykEnv,
				Log: log.NullLogger{},
			})

			return ctx
		}).
		Assess("Update ApiDefinition and check changes in k8s and Tyk",
			func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
				testNs, ok := ctx.Value(ctxNSKey).(string)
				eval.True(ok)

				// First, create the ApiDefinition
				apiDefCR, err := createTestAPIDef(ctx, c, testNs, func(apiDef *v1alpha1.ApiDefinition) {})
				eval.NoErr(err)

				err = waitForTykResourceCreation(c, apiDefCR)
				eval.NoErr(err)

				apiDefCR.Spec.Name = updatedName

				// Update ApiDefinition
				err = c.Client().Resources(opNs).Update(ctx, apiDefCR)
				eval.NoErr(err)

				// Ensure that k8s state is updated.
				err = wait.For(
					conditions.New(c.Client().Resources()).ResourceMatch(apiDefCR, func(object k8s.Object) bool {
						apiDefObj, ok := object.(*v1alpha1.ApiDefinition)
						eval.True(ok)

						return apiDefObj.Spec.Name == updatedName
					}),
					wait.WithTimeout(defaultWaitTimeout),
					wait.WithInterval(defaultWaitInterval),
				)
				eval.NoErr(err)

				// Ensure that Tyk is updated
				err = wait.For(func() (done bool, err error) {
					apiDefOnTyk, err := klient.Universal.Api().Get(tykCtx, apiDefCR.Status.ApiID)
					if err != nil {
						return false, err
					}

					if apiDefOnTyk.Name != updatedName {
						return false, fmt.Errorf(
							"ApiDefinition is not updated properly, expected %v, got %v",
							updatedName, apiDefOnTyk.Name,
						)
					}

					return true, nil
				}, wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
				eval.NoErr(err)

				return ctx
			}).Feature()

	testenv.Test(t, testApiDefinitionUpdate)
}

func TestApiDefinitionJSONSchemaValidation(t *testing.T) {
	var (
		eval                         = is.New(t)
		apiDefWithJSONValidationName = "apidef-json-validation"
		apiDefListenPath             = "/validation"
		defaultVersion               = "Default"
		errorResponseCode            = 422
		eps                          = &model.ExtendedPathsSet{
			ValidateJSON: []model.ValidatePathMeta{{
				ErrorResponseCode: errorResponseCode,
				Path:              "/get",
				Method:            http.MethodGet,
				Schema: &model.MapStringInterfaceType{Unstructured: unstructured.Unstructured{
					Object: map[string]interface{}{
						"properties": map[string]interface{}{
							"key": map[string]interface{}{
								"type":      "string",
								"minLength": 2,
							},
						},
					},
				}},
			}},
		}
	)

	adCreate := features.New("ApiDefinition JSON Schema Validation").
		Setup(func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
			testNS, ok := ctx.Value(ctxNSKey).(string)
			eval.True(ok)

			// Create ApiDefinition with JSON Schema Validation support.
			apiDef, err := createTestAPIDef(ctx, envConf, testNS, func(apiDef *v1alpha1.ApiDefinition) {
				useExtendedPaths := true

				apiDef.Name = apiDefWithJSONValidationName
				apiDef.Spec.Proxy = model.Proxy{
					ListenPath: &apiDefListenPath,
					TargetURL:  "http://httpbin.org",
				}
				apiDef.Spec.VersionData.DefaultVersion = defaultVersion
				apiDef.Spec.VersionData.NotVersioned = true
				apiDef.Spec.VersionData.Versions = map[string]model.VersionInfo{
					defaultVersion: {Name: defaultVersion, UseExtendedPaths: &useExtendedPaths, ExtendedPaths: eps},
				}
			})
			eval.NoErr(err) // failed to create apiDefinition

			err = waitForTykResourceCreation(envConf, apiDef)
			eval.NoErr(err)

			return ctx
		}).
		Assess("ApiDefinition must verify user requests",
			func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
				err := wait.For(func() (done bool, err error) {
					hc := &http.Client{}

					// invalidJSONBody does not meet the requirements of the Schema because
					// Schema requires the "key" field to have a length of 2 at least.
					invalidJSONBody := strings.NewReader(`{"key": "a"}`)

					req, err := http.NewRequest(
						http.MethodGet,
						fmt.Sprintf("%s%s/get", gatewayLocalhost, apiDefListenPath),
						invalidJSONBody,
					)
					eval.NoErr(err)
					req.Header.Add("Content-type", "application/json")

					// Since the following request does not match with the JSON Validation Schema,
					// the response status code must be 422 as indicated in the ErrorResponseCode of the ValidatePathMeta.
					resp, err := hc.Do(req)
					eval.NoErr(err)

					if resp.StatusCode != errorResponseCode {
						return false, nil
					}

					return true, nil
				}, wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
				eval.NoErr(err)

				return ctx
			}).
		Feature()

	testenv.Test(t, adCreate)
}

func TestApiDefinitionCreateWhitelist(t *testing.T) {
	var (
		eval                     = is.New(t)
		whiteListedPath          = "/whitelisted"
		apiDefWithWhitelist      = "apidef-whitelist"
		apiDefListenPath         = "/test-whitelist"
		defaultVersion           = "Default"
		errForbiddenResponseCode = 403
		eps                      = &model.ExtendedPathsSet{
			WhiteList: []model.EndPointMeta{{
				Path:       whiteListedPath,
				IgnoreCase: true,
				MethodActions: map[string]model.EndpointMethodMeta{
					"GET": {
						Action: "no_action",
						Code:   200, Data: "",
						Headers: make(map[string]string),
					},
				},
			}},
		}
	)

	adCreate := features.New("ApiDefinition whitelisting").
		Setup(func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
			testNS, ok := ctx.Value(ctxNSKey).(string)
			eval.True(ok)

			// Create ApiDefinition with whitelist extended path
			apiDef, err := createTestAPIDef(ctx, envConf, testNS, func(apiDef *v1alpha1.ApiDefinition) {
				useExtendedPaths := true

				apiDef.Name = apiDefWithWhitelist
				apiDef.Spec.Proxy = model.Proxy{
					ListenPath: &apiDefListenPath,
					TargetURL:  "http://httpbin.org",
				}
				apiDef.Spec.VersionData.DefaultVersion = defaultVersion
				apiDef.Spec.VersionData.NotVersioned = true
				apiDef.Spec.VersionData.Versions = map[string]model.VersionInfo{
					defaultVersion: {Name: defaultVersion, UseExtendedPaths: &useExtendedPaths, ExtendedPaths: eps},
				}
			})
			eval.NoErr(err) // failed to create apiDefinition

			err = waitForTykResourceCreation(envConf, apiDef)
			eval.NoErr(err)

			return ctx
		}).
		Assess("ApiDefinition should allow traffic to whitelisted route",
			func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
				err := wait.For(func() (done bool, err error) {
					hc := &http.Client{}

					req, err := http.NewRequest(
						http.MethodGet,
						fmt.Sprintf("%s%s", gatewayLocalhost, apiDefListenPath+whiteListedPath),
						nil,
					)
					eval.NoErr(err)

					resp, err := hc.Do(req)
					eval.NoErr(err)

					if resp.StatusCode == errForbiddenResponseCode {
						return false, nil
					}

					return true, nil
				}, wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
				eval.NoErr(err)
				return ctx
			}).
		Assess("ApiDefinition must not allow traffic to non-whitelisted route",
			func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
				err := wait.For(func() (done bool, err error) {
					hc := &http.Client{}

					req, err := http.NewRequest(
						http.MethodGet,
						fmt.Sprintf("%s%s", gatewayLocalhost, apiDefListenPath),
						nil,
					)
					eval.NoErr(err)
					req.Header.Add("Content-type", "application/json")

					resp, err := hc.Do(req)
					eval.NoErr(err)

					if resp.StatusCode != errForbiddenResponseCode {
						return false, nil
					}

					return true, nil
				}, wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
				eval.NoErr(err)

				return ctx
			}).Feature()

	testenv.Test(t, adCreate)
}

func TestApiDefinitionCreateBlackList(t *testing.T) {
	var (
		eval                     = is.New(t)
		blackListedPath          = "/blacklisted"
		apiDefWithBlacklist      = "apidef-blacklist"
		apiDefListenPath         = "/test-blacklist"
		defaultVersion           = "Default"
		errForbiddenResponseCode = 403
		eps                      = &model.ExtendedPathsSet{
			BlackList: []model.EndPointMeta{{
				Path:       blackListedPath,
				IgnoreCase: true,
				MethodActions: map[string]model.EndpointMethodMeta{
					"GET": {
						Action: "no_action",
						Code:   200, Data: "",
						Headers: make(map[string]string),
					},
				},
			}},
		}
	)

	adCreate := features.New("ApiDefinition").
		Setup(func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
			testNS, ok := ctx.Value(ctxNSKey).(string)
			eval.True(ok)

			// Create ApiDefinition with whitelist extended path
			apiDef, err := createTestAPIDef(ctx, envConf, testNS, func(apiDef *v1alpha1.ApiDefinition) {
				useExtendedPaths := true

				apiDef.Name = apiDefWithBlacklist
				apiDef.Spec.Proxy = model.Proxy{
					ListenPath: &apiDefListenPath,
					TargetURL:  "http://httpbin.org",
				}
				apiDef.Spec.VersionData.DefaultVersion = defaultVersion
				apiDef.Spec.VersionData.NotVersioned = true
				apiDef.Spec.VersionData.Versions = map[string]model.VersionInfo{
					defaultVersion: {Name: defaultVersion, UseExtendedPaths: &useExtendedPaths, ExtendedPaths: eps},
				}
			})
			eval.NoErr(err) // failed to create apiDefinition

			err = waitForTykResourceCreation(envConf, apiDef)
			eval.NoErr(err)

			return ctx
		}).
		Assess("ApiDefinition should forbid traffic to blacklist route",
			func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
				err := wait.For(func() (done bool, err error) {
					hc := &http.Client{}

					req, err := http.NewRequest(
						http.MethodGet,
						fmt.Sprintf("%s%s", gatewayLocalhost, apiDefListenPath+blackListedPath),
						nil,
					)
					eval.NoErr(err)

					resp, err := hc.Do(req)
					eval.NoErr(err)

					if resp.StatusCode != errForbiddenResponseCode {
						return false, nil
					}

					return true, nil
				}, wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
				eval.NoErr(err)
				return ctx
			}).
		Assess("ApiDefinition must allow traffic to non-blacklisted route",
			func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
				err := wait.For(func() (done bool, err error) {
					hc := &http.Client{}

					req, err := http.NewRequest(
						http.MethodGet,
						fmt.Sprintf("%s%s", gatewayLocalhost, apiDefListenPath),
						nil,
					)
					eval.NoErr(err)
					req.Header.Add("Content-type", "application/json")

					resp, err := hc.Do(req)
					eval.NoErr(err)

					if resp.StatusCode == errForbiddenResponseCode {
						return false, nil
					}

					return true, nil
				}, wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
				eval.NoErr(err)

				return ctx
			}).Feature()

	testenv.Test(t, adCreate)
}

func TestApiDefinitionCreateIgnored(t *testing.T) {
	var (
		eval                     = is.New(t)
		whiteListedPath          = "/whitelisted"
		ignoredPath              = "/ignored"
		apiDefWithWhitelist      = "apidef-ignored"
		apiDefListenPath         = "/test-ignored"
		defaultVersion           = "Default"
		errForbiddenResponseCode = 403
		eps                      = &model.ExtendedPathsSet{
			WhiteList: []model.EndPointMeta{{
				Path:       whiteListedPath,
				IgnoreCase: true,
				MethodActions: map[string]model.EndpointMethodMeta{
					"GET": {
						Action: "no_action",
						Code:   200, Data: "",
						Headers: make(map[string]string),
					},
				},
			}},
			Ignored: []model.EndPointMeta{{
				Path:       ignoredPath,
				IgnoreCase: true,
				MethodActions: map[string]model.EndpointMethodMeta{
					"GET": {
						Action: "no_action",
						Code:   200, Data: "",
						Headers: make(map[string]string),
					},
				},
			}},
		}
	)

	adCreate := features.New("ApiDefinition").
		Setup(func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
			testNS, ok := ctx.Value(ctxNSKey).(string)
			eval.True(ok)

			// Create ApiDefinition with whitelist + ignored extended path
			_, err := createTestAPIDef(ctx, envConf, testNS, func(apiDef *v1alpha1.ApiDefinition) {
				useExtendedPaths := true

				apiDef.Name = apiDefWithWhitelist
				apiDef.Spec.Proxy = model.Proxy{
					ListenPath: &apiDefListenPath,
					TargetURL:  "http://httpbin.org",
				}
				apiDef.Spec.VersionData.DefaultVersion = defaultVersion
				apiDef.Spec.VersionData.NotVersioned = true
				apiDef.Spec.VersionData.Versions = map[string]model.VersionInfo{
					defaultVersion: {Name: defaultVersion, UseExtendedPaths: &useExtendedPaths, ExtendedPaths: eps},
				}
			})
			eval.NoErr(err) // failed to create apiDefinition

			apiDef := v1alpha1.ApiDefinition{
				ObjectMeta: metav1.ObjectMeta{
					Name:      apiDefWithWhitelist,
					Namespace: testNS,
				},
			}
			err = waitForTykResourceCreation(envConf, &apiDef)
			eval.NoErr(err)

			return ctx
		}).
		Assess("ApiDefinition should allow traffic to ignored route",
			func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
				err := wait.For(func() (done bool, err error) {
					hc := &http.Client{}

					req, err := http.NewRequest(
						http.MethodGet,
						fmt.Sprintf("%s%s", gatewayLocalhost, apiDefListenPath+ignoredPath),
						nil,
					)
					eval.NoErr(err)

					resp, err := hc.Do(req)
					eval.NoErr(err)

					if resp.StatusCode == errForbiddenResponseCode {
						return false, nil
					}

					return true, nil
				}, wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
				eval.NoErr(err)

				return ctx
			}).
		Assess("ApiDefinition must not allow traffic to other non whitelisted routes",
			func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
				err := wait.For(func() (done bool, err error) {
					hc := &http.Client{}

					req, err := http.NewRequest(
						http.MethodGet,
						fmt.Sprintf("%s%s", gatewayLocalhost, apiDefListenPath+"/randomNonWhiteListedPath"),
						nil,
					)
					eval.NoErr(err)
					req.Header.Add("Content-type", "application/json")

					resp, err := hc.Do(req)
					eval.NoErr(err)

					if resp.StatusCode != errForbiddenResponseCode {
						return false, nil
					}

					return true, nil
				}, wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
				eval.NoErr(err)

				return ctx
			}).Feature()

	testenv.Test(t, adCreate)
}

func TestApiDefinitionCertificatePinning(t *testing.T) {
	var (
		eval   = is.New(t)
		tykCtx context.Context
		apiDef *v1alpha1.ApiDefinition
	)

	adCreate := features.New("Create ApiDefinition objects for Certificate Pinning").
		Setup(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			testNS, ok := ctx.Value(ctxNSKey).(string)
			eval.True(ok)

			opConfSecret := v1.Secret{}
			err := c.Client().Resources(opNs).Get(ctx, operatorSecret, opNs, &opConfSecret)
			eval.NoErr(err)

			// Obtain Environment configuration to be able to connect Tyk.
			tykEnv, err := generateEnvConfig(&opConfSecret)
			eval.NoErr(err)

			secret, err := createTestTlsSecret(ctx, testNS, c, nil)
			eval.NoErr(err)

			// For all domains (`*`), use following secret that contains the public key of the httpbin.org
			// So, if you make any requests to any addresses except httpbin.org, we should get proxy errors because
			// of pinned public key.
			publicKeySecrets := map[string]string{"*": secret.Name}

			// Create an ApiDefinition with Certificate Pinning using Kubernetes Secret object.
			apiDef, err = createTestAPIDef(ctx, c, testNS, func(apiDef *v1alpha1.ApiDefinition) {
				apiDef.Spec.PinnedPublicKeysRefs = publicKeySecrets
			})
			eval.NoErr(err)

			tykCtx = tykClient.SetContext(context.Background(), tykClient.Context{
				Env: tykEnv,
				Log: log.NullLogger{},
			})

			return ctx
		}).
		Assess("Ensure that the secret is created on Tyk and linked to ApiDefinition",
			func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
				// Wait until ApiDefinition is created on Tyk.
				err := waitForTykResourceCreation(c, apiDef)
				eval.NoErr(err)

				err = wait.For(
					conditions.New(c.Client().Resources()).ResourceMatch(apiDef, func(object k8s.Object) bool {
						apiDefObj, ok := object.(*v1alpha1.ApiDefinition)
						eval.True(ok)

						tykCertID, exists := apiDefObj.Spec.PinnedPublicKeys["*"]
						eval.True(exists)

						if !klient.Universal.Certificate().Exists(tykCtx, tykCertID) {
							t.Logf("failed to access certificate with ID %v on Tyk", tykCertID)
							return false
						}

						apiDefOnTyk, err := klient.Universal.Api().Get(tykCtx, apiDefObj.Status.ApiID)
						eval.NoErr(err)

						certIdOfApi, exists := apiDefOnTyk.PinnedPublicKeys["*"]
						eval.True(exists)

						if certIdOfApi != tykCertID {
							t.Logf(
								"The cert ID linked to ApiDefinition is wrong, expected %v, got %v",
								tykCertID,
								certIdOfApi,
							)

							eval.True(certIdOfApi != tykCertID)
						}

						return true
					}),
					wait.WithTimeout(defaultWaitTimeout),
					wait.WithInterval(defaultWaitInterval),
				)
				eval.NoErr(err)

				return ctx
			}).
		Feature()

	testenv.Test(t, adCreate)
}

func TestApiDefinitionUpstreamCertificates(t *testing.T) {
	var (
		eval                = is.New(t)
		apiDefUpstreamCerts = "apidef-upstream-certs"
		defaultVersion      = "Default"
		certName            = "test-tls-secret-name"

		tykEnv environmet.Env
		reqCtx context.Context
	)

	adCreate := features.New("Create an ApiDefinition for Upstream TLS").
		Setup(func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
			testNS, ok := ctx.Value(ctxNSKey).(string)
			eval.True(ok)

			opConfSecret := v1.Secret{}
			err := envConf.Client().Resources(opNs).Get(ctx, operatorSecret, opNs, &opConfSecret)
			eval.NoErr(err)

			// Obtain Environment configuration to be able to connect Tyk.
			tykEnv, err = generateEnvConfig(&opConfSecret)
			eval.NoErr(err)

			_, err = createTestTlsSecret(ctx, testNS, envConf, nil)
			eval.NoErr(err)

			reqCtx = tykClient.SetContext(context.Background(), tykClient.Context{
				Env: tykEnv,
				Log: log.NullLogger{},
			})

			return ctx
		}).
		Setup(func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
			testNS, ok := ctx.Value(ctxNSKey).(string)
			eval.True(ok)

			// Create ApiDefinition with Upstream certificate
			_, err := createTestAPIDef(ctx, envConf, testNS, func(apiDef *v1alpha1.ApiDefinition) {
				apiDef.Name = apiDefUpstreamCerts
				apiDef.Spec.UpstreamCertificateRefs = map[string]string{
					"*": certName,
				}
				apiDef.Spec.VersionData.DefaultVersion = defaultVersion
				apiDef.Spec.VersionData.NotVersioned = true
				apiDef.Spec.VersionData.Versions = map[string]model.VersionInfo{
					defaultVersion: {Name: defaultVersion},
				}
			})
			eval.NoErr(err) // failed to create apiDefinition

			return ctx
		}).
		Assess("ApiDefinition must have upstream field defined",
			func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
				testNS, ok := ctx.Value(ctxNSKey).(string)
				eval.True(ok)

				tlsSecret := v1.Secret{}
				err := c.Client().Resources(testNS).Get(ctx, certName, testNS, &tlsSecret)
				eval.NoErr(err)

				certPemBytes, ok := tlsSecret.Data["tls.crt"]
				eval.True(ok)

				certFingerPrint, err := cert.CalculateFingerPrint(certPemBytes)
				eval.NoErr(err)

				calculatedCertID := tykEnv.Org + certFingerPrint

				err = wait.For(func() (done bool, err error) {
					// validate certificate was created on Tyk
					exists := klient.Universal.Certificate().Exists(reqCtx, calculatedCertID)
					if !exists {
						return false, errors.New("certificate is not created yet")
					}

					return true, nil
				}, wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
				eval.NoErr(err)

				return ctx
			}).Feature()

	testenv.Test(t, adCreate)
}

func TestApiCertificates(t *testing.T) {
	eval := is.New(t)
	apiDef := &v1alpha1.ApiDefinition{}

	f := features.New("API Definition Certificates").Setup(
		func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
			testNS, ok := ctx.Value(ctxNSKey).(string)
			eval.True(ok)

			secret, err := createTestTlsSecret(ctx, testNS, envConf, nil)
			eval.NoErr(err)

			apiDef, err = createTestAPIDef(ctx, envConf, testNS, func(apiDef *v1alpha1.ApiDefinition) {
				apiDef.Name = "test-certificates"
				apiDef.Spec.Name = apiDef.Name
				apiDef.Spec.CertificateSecretNames = []string{secret.Name}
			})
			eval.NoErr(err)

			return ctx
		}).Assess("API Definition has the certificate id",
		func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
			err := wait.For(conditions.New(envConf.Client().Resources()).ResourceMatch(apiDef, func(object k8s.Object) bool {
				api, ok := object.(*v1alpha1.ApiDefinition)
				if !ok {
					return false
				}
				return api.Spec.Certificates != nil && len(api.Spec.Certificates) > 0
			}), wait.WithInterval(defaultWaitInterval), wait.WithTimeout(defaultWaitTimeout))
			eval.NoErr(err)

			return ctx
		}).Assess("Certificate is created on Tyk",
		func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
			opConfSecret := v1.Secret{}
			err := envConf.Client().Resources(opNs).Get(ctx, operatorSecret, opNs, &opConfSecret)
			eval.NoErr(err)

			// Obtain Environment configuration to be able to connect Tyk.
			tykEnv, err := generateEnvConfig(&opConfSecret)
			eval.NoErr(err)

			tykCtx := tykClient.SetContext(context.Background(), tykClient.Context{
				Env: tykEnv,
				Log: log.NullLogger{},
			})

			exists := klient.Universal.Certificate().Exists(tykCtx, apiDef.Spec.Certificates[0])
			eval.True(exists)

			return ctx
		}).Feature()

	testenv.Test(t, f)
}

func TestApiDefinitionBasicAuth(t *testing.T) {
	var (
		apiDefBasicAuth = "apidef-basic-authentication"
		defaultVersion  = "Default"

		eval   = is.New(t)
		reqCtx context.Context
		tykEnv environmet.Env
	)

	testBasicAuth := features.New("Basic authentication").
		Setup(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			opConfSecret := v1.Secret{}

			err := c.Client().Resources(opNs).Get(ctx, operatorSecret, opNs, &opConfSecret)
			eval.NoErr(err)

			// Obtain Environment configuration to be able to connect Tyk.
			tykEnv, err = generateEnvConfig(&opConfSecret)
			eval.NoErr(err)

			reqCtx = tykClient.SetContext(context.Background(), tykClient.Context{
				Env: tykEnv,
				Log: log.NullLogger{},
			})

			return ctx
		}).
		Setup(func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
			testNS, ok := ctx.Value(ctxNSKey).(string)
			eval.True(ok)

			// Create ApiDefinition with Basic Authentication
			_, err := createTestAPIDef(ctx, envConf, testNS, func(apiDef *v1alpha1.ApiDefinition) {
				useBasicAuth := true

				apiDef.Name = apiDefBasicAuth
				apiDef.Spec.UseBasicAuth = &useBasicAuth
				apiDef.Spec.VersionData.DefaultVersion = defaultVersion
				apiDef.Spec.VersionData.NotVersioned = true
				apiDef.Spec.VersionData.Versions = map[string]model.VersionInfo{
					defaultVersion: {Name: defaultVersion},
				}
			})
			eval.NoErr(err) // failed to create apiDefinition

			return ctx
		}).
		Assess("API must have basic authentication enabled",
			func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
				testNS, ok := ctx.Value(ctxNSKey).(string)
				eval.True(ok)

				var apiDef *model.APIDefinitionSpec

				err := wait.For(func() (done bool, err error) {
					// validate basic authentication field was set
					var apiDefCRD v1alpha1.ApiDefinition

					err = c.Client().Resources().Get(ctx, apiDefBasicAuth, testNS, &apiDefCRD)
					if err != nil {
						return false, err
					}

					apiDef, err = klient.Universal.Api().Get(reqCtx, apiDefCRD.Status.ApiID)
					if err != nil {
						t.Error("API is not created yet")
						return false, nil
					}

					return true, nil
				}, wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
				eval.NoErr(err)

				eval.True(apiDef.UseBasicAuth != nil)
				eval.True(*apiDef.UseBasicAuth)

				return ctx
			}).Feature()

	testenv.Test(t, testBasicAuth)
}

func TestApiDefinitionBaseIdentityProviderWithMultipleAuthTypes(t *testing.T) {
	var (
		apiDefBasicAndMTLSAuth = "apidef-basic-and-mtls-authentication"
		defaultVersion         = "Default"

		eval   = is.New(t)
		reqCtx context.Context
		tykEnv environmet.Env
	)

	testBasicAuth := features.New("Base Identity Provider for Basic Auth and mTLS").
		Setup(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			opConfSecret := v1.Secret{}

			err := c.Client().Resources(opNs).Get(ctx, operatorSecret, opNs, &opConfSecret)
			eval.NoErr(err)

			// Obtain Environment configuration to be able to connect Tyk.
			tykEnv, err = generateEnvConfig(&opConfSecret)
			eval.NoErr(err)

			reqCtx = tykClient.SetContext(context.Background(), tykClient.Context{
				Env: tykEnv,
				Log: log.NullLogger{},
			})

			return ctx
		}).
		Setup(func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
			testNS, ok := ctx.Value(ctxNSKey).(string)
			eval.True(ok)

			// Create ApiDefinition with Basic Authentication and mTLS enabled with BasicAuthUser as base identity provider
			_, err := createTestAPIDef(ctx, envConf, testNS, func(apiDef *v1alpha1.ApiDefinition) {
				useBasicAuth := true
				useMutualTLSAuth := true

				apiDef.Name = apiDefBasicAndMTLSAuth
				apiDef.Spec.UseBasicAuth = &useBasicAuth
				apiDef.Spec.UseMutualTLSAuth = &useMutualTLSAuth
				apiDef.Spec.BaseIdentityProvidedBy = "basic_auth_user"
				apiDef.Spec.VersionData.DefaultVersion = defaultVersion
				apiDef.Spec.VersionData.NotVersioned = true
				apiDef.Spec.VersionData.Versions = map[string]model.VersionInfo{
					defaultVersion: {Name: defaultVersion},
				}
			})
			eval.NoErr(err) // failed to create apiDefinition

			return ctx
		}).
		Assess("API must have base identity provider set",
			func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
				testNS, ok := ctx.Value(ctxNSKey).(string)
				eval.True(ok)

				var apiDef *model.APIDefinitionSpec

				err := wait.For(func() (done bool, err error) {
					// validate base identity provider and all authentication fields
					var apiDefCRD v1alpha1.ApiDefinition

					err = c.Client().Resources().Get(ctx, apiDefBasicAndMTLSAuth, testNS, &apiDefCRD)
					if err != nil {
						return false, err
					}

					apiDef, err = klient.Universal.Api().Get(reqCtx, apiDefCRD.Status.ApiID)
					if err != nil {
						t.Logf("API is not created yet on Tyk")
						return false, nil
					}

					return true, nil
				}, wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
				eval.NoErr(err)

				eval.True(apiDef.UseBasicAuth != nil)
				eval.True(*apiDef.UseBasicAuth)

				eval.True(apiDef.UseMutualTLSAuth != nil)
				eval.True(*apiDef.UseMutualTLSAuth)

				eval.Equal(apiDef.BaseIdentityProvidedBy, model.AuthTypeEnum("basic_auth_user"))

				return ctx
			}).Feature()

	testenv.Test(t, testBasicAuth)
}

func TestApiDefinitionClientMTLS(t *testing.T) {
	type ContextKey string

	var (
		apiDefClientMTLSWithCert    = "apidef-client-mtls-with-cert"
		apiDefClientMTLSWithoutCert = "apidef-client-mtls-without-cert"
		defaultVersion              = "Default"
		certName                    = "test-tls-secret-name"

		certIDCtxKey ContextKey = "certID"
		tykEnv       environmet.Env
		reqCtx       context.Context
		eval         = is.New(t)
	)

	testWithCert := features.New("Client MTLS with certificate").
		Setup(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			opConfSecret := v1.Secret{}
			err := c.Client().Resources(opNs).Get(ctx, operatorSecret, opNs, &opConfSecret)
			eval.NoErr(err)

			tykEnv, err = generateEnvConfig(&opConfSecret)
			eval.NoErr(err)

			reqCtx = tykClient.SetContext(context.Background(), tykClient.Context{
				Env: tykEnv,
				Log: log.NullLogger{},
			})

			return ctx
		}).
		Setup(func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
			testNS, ok := ctx.Value(ctxNSKey).(string)
			eval.True(ok)

			_, err := createTestTlsSecret(ctx, testNS, envConf, nil)
			eval.NoErr(err)

			return ctx
		}).
		Setup(func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
			testNS, ok := ctx.Value(ctxNSKey).(string)
			eval.True(ok)

			// Create ApiDefinition with Client certificate
			_, err := createTestAPIDef(ctx, envConf, testNS, func(apiDef *v1alpha1.ApiDefinition) {
				useMutualTLS := true

				apiDef.Name = apiDefClientMTLSWithCert
				apiDef.Spec.UseMutualTLSAuth = &useMutualTLS
				apiDef.Spec.ClientCertificateRefs = []string{certName}
				apiDef.Spec.VersionData.DefaultVersion = defaultVersion
				apiDef.Spec.VersionData.NotVersioned = true
				apiDef.Spec.VersionData.Versions = map[string]model.VersionInfo{
					defaultVersion: {Name: defaultVersion},
				}
			})
			eval.NoErr(err)

			return ctx
		}).Assess("Certificate from secret must be uploaded",
		func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			testNS, ok := ctx.Value(ctxNSKey).(string)
			eval.True(ok)

			tlsSecret := v1.Secret{}

			err := c.Client().Resources(testNS).Get(ctx, certName, testNS, &tlsSecret)
			eval.NoErr(err)

			certPemBytes, ok := tlsSecret.Data["tls.crt"]
			eval.True(ok)

			certFingerPrint, _ := cert.CalculateFingerPrint(certPemBytes)
			calculatedCertID := tykEnv.Org + certFingerPrint

			err = wait.For(func() (done bool, err error) {
				// validate certificate was created
				exists := klient.Universal.Certificate().Exists(reqCtx, calculatedCertID)
				if !exists {
					t.Log("certificate is not created yet")
					return false, nil
				}

				return true, nil
			}, wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
			eval.NoErr(err)

			ctx = context.WithValue(ctx, certIDCtxKey, calculatedCertID)

			return ctx
		}).
		Assess("API must have client certificate field defined",
			func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
				testNS, ok := ctx.Value(ctxNSKey).(string)
				eval.True(ok)

				certID := ctx.Value(certIDCtxKey)

				var apiDef *model.APIDefinitionSpec

				err := wait.For(func() (done bool, err error) {
					// validate client certificate field was set
					var apiDefCRD v1alpha1.ApiDefinition

					err = c.Client().Resources().Get(ctx, apiDefClientMTLSWithCert, testNS, &apiDefCRD)
					if err != nil {
						return false, err
					}

					apiDef, err = klient.Universal.Api().Get(reqCtx, apiDefCRD.Status.ApiID)
					if err != nil {
						return false, errors.New("API is not created yet")
					}

					if len(apiDef.ClientCertificates) == 0 {
						return false, errors.New("Client certificate field is not set yet")
					}

					return true, nil
				}, wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
				eval.NoErr(err)

				eval.True(len(apiDef.ClientCertificates) == 1)
				eval.True(apiDef.ClientCertificates[0] == certID)

				return ctx
			}).Feature()

	testWithoutCert := features.New("Client MTLS without certs").
		Setup(func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
			testNS, ok := ctx.Value(ctxNSKey).(string)
			eval.True(ok)

			// Create ApiDefinition with Upstream certificate
			_, err := createTestAPIDef(ctx, envConf, testNS, func(apiDef *v1alpha1.ApiDefinition) {
				useMutualTLS := true

				apiDef.Name = apiDefClientMTLSWithoutCert
				apiDef.Spec.UseMutualTLSAuth = &useMutualTLS
				apiDef.Spec.ClientCertificateRefs = []string{certName}
				apiDef.Spec.VersionData.DefaultVersion = defaultVersion
				apiDef.Spec.VersionData.NotVersioned = true
				apiDef.Spec.VersionData.Versions = map[string]model.VersionInfo{
					defaultVersion: {Name: defaultVersion},
				}
			})
			eval.NoErr(err)

			return ctx
		}).
		Assess("API should be created even though certs doesn't exists",
			func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
				testNS, ok := ctx.Value(ctxNSKey).(string)
				eval.True(ok)

				opConfSecret := v1.Secret{}
				err := c.Client().Resources(opNs).Get(ctx, operatorSecret, opNs, &opConfSecret)
				eval.NoErr(err)

				var apiDef *model.APIDefinitionSpec
				err = wait.For(func() (done bool, err error) {
					// validate api Def was created without certificate
					var apiDefCRD v1alpha1.ApiDefinition

					err = c.Client().Resources().Get(ctx, apiDefClientMTLSWithoutCert, testNS, &apiDefCRD)
					if err != nil {
						t.Logf("%v, err: %v", errFailedToGetApiDefCRMsg, err)
						return false, nil
					}

					apiDef, err = klient.Universal.Api().Get(reqCtx, apiDefCRD.Status.ApiID)
					if err != nil {
						t.Logf("%v, err: %v", errFailedToGetApiDefTykMsg, err)
						return false, nil
					}

					return true, nil
				}, wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
				eval.NoErr(err)

				eval.True(len(apiDef.ClientCertificates) == 0)

				return ctx
			}).Feature()

	testenv.Test(t, testWithCert)
	testenv.Test(t, testWithoutCert)
}

func TestAPIDefinition_GraphQL_ExecutionMode(t *testing.T) {
	eval := is.New(t)

	createAPI := features.New("Create GraphQL API").
		Assess("validate_executionMode", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			testNS, ok := ctx.Value(ctxNSKey).(string)
			eval.True(ok)

			tests := map[string]struct {
				ExecutionMode string
				ReturnErr     bool
			}{
				"invalid execution mode": {
					ExecutionMode: "invalid",
					ReturnErr:     true,
				},
				"empty execution mode": {
					ExecutionMode: "",
					ReturnErr:     false,
				},
				"valid execution engine mode": {
					ExecutionMode: "executionEngine",
					ReturnErr:     false,
				},
				"valid proxy only mode": {
					ExecutionMode: "proxyOnly",
					ReturnErr:     false,
				},
			}

			for n, tc := range tests {
				t.Run(n, func(t *testing.T) {
					_, err := createTestAPIDef(ctx, c, testNS, func(ad *v1alpha1.ApiDefinition) {
						ad.Name = fmt.Sprintf("%s-%s", ad.Name, uuid.New().String())
						ad.Spec.Name = ad.Name
						ad.Spec.GraphQL = &model.GraphQLConfig{
							Enabled:       true,
							ExecutionMode: model.GraphQLExecutionMode(tc.ExecutionMode),
						}
					})

					t.Log("Error=", err)
					eval.Equal(tc.ReturnErr, err != nil)
				})
			}

			return ctx
		}).Feature()

	testenv.Test(t, createAPI)
}

func TestApiDefinitionSubGraphExecutionMode(t *testing.T) {
	const supportedMajorTykVersion = uint(4)

	var (
		tykEnv          = environmet.Env{}
		majorTykVersion = supportedMajorTykVersion
		r               = controllers.ApiDefinitionReconciler{}
	)

	gqlSubGraph := features.New("GraphQL SubGraph Execution mode").
		Setup(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			eval := is.New(t)
			opConfSecret := v1.Secret{}

			err := c.Client().Resources(opNs).Get(ctx, operatorSecret, opNs, &opConfSecret)
			eval.NoErr(err)

			// Obtain Environment configuration to be able to connect Tyk.
			tykEnv, err = generateEnvConfig(&opConfSecret)
			eval.NoErr(err)

			v, err := version.ParseGeneric(tykEnv.TykVersion)
			eval.NoErr(err)
			if v.Major() < 4 {
				majorTykVersion = v.Major()
				t.Skip("GraphQL Federation is not available on Tyk v3")
			}

			// Create ApiDefinition Reconciler.
			cl, err := createTestClient(c.Client())
			eval.NoErr(err)
			r = controllers.ApiDefinitionReconciler{
				Client: cl,
				Log:    log.NullLogger{},
				Scheme: cl.Scheme(),
				Env:    tykEnv,
			}

			return ctx
		}).
		Assess("ApiDefinition must include SubGraph CR details",
			func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
				eval := is.New(t)

				testNs, ok := ctx.Value(ctxNSKey).(string)
				eval.True(ok)

				// Generate ApiDefinition CR and create it.
				api := generateApiDef(testNs, func(definition *v1alpha1.ApiDefinition) {
					graphRef := testSubGraphCRMetaName

					definition.Spec.GraphQL = &model.GraphQLConfig{
						GraphRef:      &graphRef,
						ExecutionMode: model.SubGraphExecutionMode,
						Version:       "1",
					}
				})
				_, err := util.CreateOrUpdate(ctx, r.Client, api, func() error {
					return nil
				})
				eval.NoErr(err)

				// Generate SubGraph CR and create it.
				sg := generateSubGraphCR(testNs, nil)
				_, err = util.CreateOrUpdate(ctx, r.Client, sg, func() error {
					return nil
				})
				eval.NoErr(err)

				// Wait for reconciliation; so that, the ApiDefinition is updated according to linked SubGraph CR.
				err = wait.For(
					conditions.New(c.Client().Resources()).ResourceMatch(api, func(object k8s.Object) bool {
						_, err = r.Reconcile(ctx, ctrl.Request{NamespacedName: cr.ObjectKeyFromObject(api)})
						return err == nil
					}),
					wait.WithTimeout(defaultWaitTimeout),
					wait.WithInterval(defaultWaitInterval),
				)
				eval.NoErr(err)

				// After reconciliation, check that ApiDefinition CR is updated properly.
				err = wait.For(
					conditions.New(c.Client().Resources()).ResourceMatch(api, func(object k8s.Object) bool {
						apiDefObj, ok := object.(*v1alpha1.ApiDefinition)
						eval.True(ok)

						return apiDefObj.Spec.GraphQL != nil && apiDefObj.Spec.GraphQL.GraphRef != nil &&
							*apiDefObj.Spec.GraphQL.GraphRef == testSubGraphCRMetaName && apiDefObj.Spec.GraphQL.Schema != nil &&
							*apiDefObj.Spec.GraphQL.Schema == testSubGraphSchema &&
							apiDefObj.Spec.GraphQL.Subgraph.SDL == testSubGraphSDL &&
							apiDefObj.Status.LinkedToSubgraph == testSubGraphCRMetaName
					}),
					wait.WithTimeout(defaultWaitTimeout),
					wait.WithInterval(defaultWaitInterval),
				)
				eval.NoErr(err)

				// After reconciliation, check that SubGraph CR is updated properly.
				err = wait.For(
					conditions.New(c.Client().Resources()).ResourceMatch(sg, func(object k8s.Object) bool {
						sgObj, ok := object.(*v1alpha1.SubGraph)
						eval.True(ok)

						return sgObj.Status.LinkedByAPI == controllers.EncodeNS(cr.ObjectKeyFromObject(api).String())
					}),
					wait.WithTimeout(defaultWaitTimeout),
					wait.WithInterval(defaultWaitInterval),
				)
				eval.NoErr(err)

				return ctx
			}).
		Assess("another ApiDefinition must not use already linked SubGraph CR",
			func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
				eval := is.New(t)

				testNs, ok := ctx.Value(ctxNSKey).(string)
				eval.True(ok)

				// Generate another ApiDefinition CR and create it.
				api := generateApiDef(testNs, func(definition *v1alpha1.ApiDefinition) {
					graphRef := testSubGraphCRMetaName

					definition.ObjectMeta = metav1.ObjectMeta{
						Name: "another-api", Namespace: testNs,
					}
					definition.Spec.GraphQL = &model.GraphQLConfig{
						GraphRef:      &graphRef,
						ExecutionMode: model.SubGraphExecutionMode,
						Version:       "1",
					}
				})
				_, err := util.CreateOrUpdate(ctx, r.Client, api, func() error {
					return nil
				})
				eval.NoErr(err)

				// Reconciliation must fail because the SubGraph is already linked by another ApiDefinition since linking
				// multiple ApiDefinition to one SubGraph CR is forbidden.
				err = wait.For(conditions.New(c.Client().Resources()).ResourceMatch(api, func(object k8s.Object) bool {
					_, err = r.Reconcile(ctx, ctrl.Request{NamespacedName: cr.ObjectKeyFromObject(api)})
					return errors.Is(err, controllers.ErrMultipleLinkSubGraph)
				}),
					wait.WithTimeout(defaultWaitTimeout),
					wait.WithInterval(defaultWaitInterval),
				)
				eval.NoErr(err)

				return ctx
			}).
		Assess("update ApiDefinition GraphRef to another SubGraph CR",
			func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
				eval := is.New(t)

				testNs, ok := ctx.Value(ctxNSKey).(string)
				eval.True(ok)

				const (
					newSgName = "new-sg"
					newSDL    = "newSDL"
					newSchema = "newSchema"
				)

				// Generate another SubGraph CR and create it.
				sg := generateSubGraphCR(testNs, func(subGraph *v1alpha1.SubGraph) {
					subGraph.ObjectMeta.Name = newSgName
					subGraph.Spec.SDL = newSDL
					subGraph.Spec.Schema = newSchema
				})
				_, err := util.CreateOrUpdate(ctx, r.Client, sg, func() error {
					return nil
				})
				eval.NoErr(err)

				// Get ApiDefinition and update it based on the new SubGraph CR information.
				api := &v1alpha1.ApiDefinition{ObjectMeta: metav1.ObjectMeta{Name: testApiDef, Namespace: testNs}}
				err = wait.For(
					conditions.New(c.Client().Resources()).ResourceMatch(api, func(object k8s.Object) bool {
						return true
					}),
					wait.WithTimeout(defaultWaitTimeout),
					wait.WithInterval(defaultWaitInterval),
				)
				eval.NoErr(err)

				_, err = util.CreateOrUpdate(ctx, r.Client, api, func() error {
					graphRef := newSgName

					api.Spec.GraphQL.GraphRef = &graphRef
					return nil
				})
				eval.NoErr(err)

				// Wait for reconciliation; so that, the ApiDefinition is updated according to new SubGraph CR.
				err = wait.For(
					conditions.New(c.Client().Resources()).ResourceMatch(api, func(object k8s.Object) bool {
						_, err = r.Reconcile(ctx, ctrl.Request{NamespacedName: cr.ObjectKeyFromObject(api)})
						return err == nil
					}),
					wait.WithTimeout(defaultWaitTimeout),
					wait.WithInterval(defaultWaitInterval),
				)
				eval.NoErr(err)

				// After successful reconciliation, check that ApiDefinition CR is updated properly.
				err = wait.For(
					conditions.New(c.Client().Resources()).ResourceMatch(api, func(object k8s.Object) bool {
						apiDefObj, ok := object.(*v1alpha1.ApiDefinition)
						eval.True(ok)

						return apiDefObj.Spec.GraphQL != nil && apiDefObj.Spec.GraphQL.GraphRef != nil &&
							*apiDefObj.Spec.GraphQL.GraphRef == newSgName && apiDefObj.Spec.GraphQL.Schema != nil &&
							*apiDefObj.Spec.GraphQL.Schema == newSchema &&
							apiDefObj.Spec.GraphQL.Subgraph.SDL == newSDL &&
							apiDefObj.Status.LinkedToSubgraph == newSgName
					}),
					wait.WithTimeout(defaultWaitTimeout),
					wait.WithInterval(defaultWaitInterval),
				)
				eval.NoErr(err)

				// After successful reconciliation, check that SubGraph CR is updated properly.
				err = wait.For(
					conditions.New(c.Client().Resources()).ResourceMatch(sg, func(object k8s.Object) bool {
						sgObj, ok := object.(*v1alpha1.SubGraph)
						eval.True(ok)

						return sgObj.Status.LinkedByAPI == controllers.EncodeNS(cr.ObjectKeyFromObject(api).String())
					}),
					wait.WithTimeout(defaultWaitTimeout),
					wait.WithInterval(defaultWaitInterval),
				)
				eval.NoErr(err)

				return ctx
			}).
		Assess("remove GraphRef from ApiDefinition",
			func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
				eval := is.New(t)

				testNs, ok := ctx.Value(ctxNSKey).(string)
				eval.True(ok)

				// Get ApiDefinition and remove GraphRef from it.
				api := &v1alpha1.ApiDefinition{ObjectMeta: metav1.ObjectMeta{Name: testApiDef, Namespace: testNs}}
				err := wait.For(
					conditions.New(c.Client().Resources()).ResourceMatch(api, func(object k8s.Object) bool {
						return true
					}),
					wait.WithTimeout(defaultWaitTimeout),
					wait.WithInterval(defaultWaitInterval),
				)
				eval.NoErr(err)

				_, err = util.CreateOrUpdate(ctx, r.Client, api, func() error {
					api.Spec.GraphQL.GraphRef = nil
					return nil
				})
				eval.NoErr(err)

				// Wait for reconciliation; so that, the ApiDefinition is updated according to new SubGraph CR.
				err = wait.For(
					conditions.New(c.Client().Resources()).ResourceMatch(api, func(object k8s.Object) bool {
						_, err = r.Reconcile(ctx, ctrl.Request{NamespacedName: cr.ObjectKeyFromObject(api)})
						return err == nil
					}),
					wait.WithTimeout(defaultWaitTimeout),
					wait.WithInterval(defaultWaitInterval),
				)
				eval.NoErr(err)

				const (
					newSDL    = "newSDL"
					newSchema = "newSchema"
				)

				// After successful reconciliation, check that ApiDefinition CR is updated properly.
				err = wait.For(
					conditions.New(c.Client().Resources()).ResourceMatch(api, func(object k8s.Object) bool {
						apiDefObj := &v1alpha1.ApiDefinition{}
						err = r.Get(ctx, cr.ObjectKeyFromObject(api), apiDefObj)
						if err != nil {
							return false
						}

						return apiDefObj.Spec.GraphQL != nil && (apiDefObj.Spec.GraphQL.GraphRef == nil ||
							*apiDefObj.Spec.GraphQL.GraphRef == "") && apiDefObj.Spec.GraphQL.Schema != nil &&
							*apiDefObj.Spec.GraphQL.Schema == newSchema &&
							apiDefObj.Spec.GraphQL.Subgraph.SDL == newSDL &&
							apiDefObj.Status.LinkedToSubgraph == ""
					}),
					wait.WithTimeout(defaultWaitTimeout),
					wait.WithInterval(defaultWaitInterval),
				)
				eval.NoErr(err)

				return ctx
			}).
		Feature()

	testenv.Finish(func(ctx context.Context, c *envconf.Config) (context.Context, error) {
		if t.Skipped() || majorTykVersion < supportedMajorTykVersion {
			return ctx, nil
		}

		eval := is.New(t)
		testNs, ok := ctx.Value(ctxNSKey).(string)
		eval.True(ok)

		err := r.DeleteAllOf(ctx, &v1alpha1.ApiDefinition{}, cr.InNamespace(testNs))
		if err != nil {
			return ctx, err
		}

		return ctx, r.DeleteAllOf(ctx, &v1alpha1.SubGraph{}, cr.InNamespace(testNs))
	})

	testenv.Test(t, gqlSubGraph)
}
