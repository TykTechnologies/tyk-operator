package integration

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/TykTechnologies/tyk-operator/api/model"
	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/google/uuid"
	"github.com/matryer/is"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestApiDefinitionCreate(t *testing.T) {
	var (
		apiDefWithJSONValidationName = "apidef-json-validation"
		apiDefListenPath             = "/validation"
		defaultVersion               = "Default"
		errorResponseCode            = 422
		defaultTimeout               = 1 * time.Minute
	)

	eps := &model.ExtendedPathsSet{
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

	adCreate := features.New("ApiDefinition").
		Setup(func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
			testNS := ctx.Value(ctxNSKey).(string) //nolint:errcheck
			is := is.New(t)

			// Create ApiDefinition with JSON Schema Validation support.
			_, err := createTestAPIDef(ctx, testNS, func(apiDef *v1alpha1.ApiDefinition) {
				apiDef.Name = apiDefWithJSONValidationName
				apiDef.Spec.Proxy = model.Proxy{
					ListenPath:      apiDefListenPath,
					TargetURL:       "http://httpbin.org",
					StripListenPath: true,
				}
				apiDef.Spec.VersionData.DefaultVersion = defaultVersion
				apiDef.Spec.VersionData.NotVersioned = true
				apiDef.Spec.VersionData.Versions = map[string]model.VersionInfo{
					defaultVersion: {Name: defaultVersion, UseExtendedPaths: true, ExtendedPaths: eps},
				}
			}, envConf)
			is.NoErr(err) // failed to create apiDefinition

			return ctx
		}).
		Assess("ApiDefinition must have ValidateJSON field",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				is := is.New(t)
				client := cfg.Client()

				testNS := ctx.Value(ctxNSKey).(string) //nolint:errcheck
				desiredApiDef := v1alpha1.ApiDefinition{
					ObjectMeta: metav1.ObjectMeta{Name: apiDefWithJSONValidationName, Namespace: testNS},
				}

				err := wait.For(conditions.New(client.Resources()).ResourceMatch(&desiredApiDef, func(object k8s.Object) bool {
					apiDef := object.(*v1alpha1.ApiDefinition) //nolint:errcheck
					// 'validate_json' field must exist in the ApiDefinition object.
					return len(apiDef.Spec.VersionData.Versions[defaultVersion].ExtendedPaths.ValidateJSON) == 1
				}), wait.WithTimeout(defaultTimeout))
				is.NoErr(err)

				return ctx
			}).
		Assess("ApiDefinition must verify user requests",
			func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
				is := is.New(t)

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
					is.NoErr(err)
					req.Header.Add("Content-type", "application/json")

					// Since the following request does not match with the JSON Validation Schema,
					// the response status code must be 422 as indicated in the ErrorResponseCode of the ValidatePathMeta.
					resp, err := hc.Do(req)
					is.NoErr(err)

					if resp.StatusCode != errorResponseCode {
						return false, nil
					}

					return true, nil
				}, wait.WithTimeout(defaultTimeout))
				is.NoErr(err)

				return ctx
			}).Feature()

	testenv.Test(t, adCreate)
}

func TestApiDefinitionCreateWhitelist(t *testing.T) {
	var (
		apiDefWithWhitelist      = "apidef-whitelist"
		apiDefListenPath         = "/test"
		defaultVersion           = "Default"
		errForbiddenResponseCode = 403
		defaultTimeout           = 1 * time.Minute
	)

	const whiteListedPath = "/whitelisted"
	eps := &model.ExtendedPathsSet{
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

	adCreate := features.New("ApiDefinition").
		Setup(func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
			testNS := ctx.Value(ctxNSKey).(string) //nolint:errcheck
			is := is.New(t)

			// Create ApiDefinition with whitelist extended path
			_, err := createTestAPIDef(ctx, testNS, func(apiDef *v1alpha1.ApiDefinition) {
				apiDef.Name = apiDefWithWhitelist
				apiDef.Spec.Proxy = model.Proxy{
					ListenPath:      apiDefListenPath,
					TargetURL:       "http://httpbin.org",
					StripListenPath: true,
				}
				apiDef.Spec.VersionData.DefaultVersion = defaultVersion
				apiDef.Spec.VersionData.NotVersioned = true
				apiDef.Spec.VersionData.Versions = map[string]model.VersionInfo{
					defaultVersion: {Name: defaultVersion, UseExtendedPaths: true, ExtendedPaths: eps},
				}
			}, envConf)
			is.NoErr(err) // failed to create apiDefinition

			return ctx
		}).
		Assess("ApiDefinition must have whitelist field",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				is := is.New(t)
				client := cfg.Client()

				testNS := ctx.Value(ctxNSKey).(string) //nolint:errcheck
				desiredApiDef := v1alpha1.ApiDefinition{
					ObjectMeta: metav1.ObjectMeta{Name: apiDefWithWhitelist, Namespace: testNS},
				}

				err := wait.For(conditions.New(client.Resources()).ResourceMatch(&desiredApiDef, func(object k8s.Object) bool {
					apiDef := object.(*v1alpha1.ApiDefinition) //nolint:errcheck
					return len(apiDef.Spec.VersionData.Versions[defaultVersion].ExtendedPaths.WhiteList) == 1
				}), wait.WithTimeout(defaultTimeout))
				is.NoErr(err)

				return ctx
			}).
		Assess("ApiDefniition should allow traffic to whitelisted route",
			func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
				is := is.New(t)

				err := wait.For(func() (done bool, err error) {
					hc := &http.Client{}

					req, err := http.NewRequest(
						http.MethodGet,
						fmt.Sprintf("%s%s", gatewayLocalhost, apiDefListenPath+whiteListedPath),
						nil,
					)
					is.NoErr(err)

					resp, err := hc.Do(req)
					is.NoErr(err)

					if resp.StatusCode == errForbiddenResponseCode {
						return false, nil
					}

					return true, nil
				}, wait.WithTimeout(defaultTimeout))
				is.NoErr(err)
				return ctx
			}).
		Assess("ApiDefinition must not allow traffic to non-whitelisted route",
			func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
				is := is.New(t)

				err := wait.For(func() (done bool, err error) {
					hc := &http.Client{}

					req, err := http.NewRequest(
						http.MethodGet,
						fmt.Sprintf("%s%s", gatewayLocalhost, apiDefListenPath),
						nil,
					)
					is.NoErr(err)
					req.Header.Add("Content-type", "application/json")

					resp, err := hc.Do(req)
					is.NoErr(err)

					if resp.StatusCode != errForbiddenResponseCode {
						return false, nil
					}

					return true, nil
				}, wait.WithTimeout(defaultTimeout))
				is.NoErr(err)

				return ctx
			}).Feature()

	testenv.Test(t, adCreate)
}

func TestApiDefinitionCreateBlackList(t *testing.T) {
	var (
		apiDefWithBlacklist      = "apidef-blacklist"
		apiDefListenPath         = "/test"
		defaultVersion           = "Default"
		errForbiddenResponseCode = 403
		defaultTimeout           = 1 * time.Minute
	)

	const blackListedPath = "/blacklisted"
	eps := &model.ExtendedPathsSet{
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

	adCreate := features.New("ApiDefinition").
		Setup(func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
			testNS := ctx.Value(ctxNSKey).(string) //nolint:errcheck
			is := is.New(t)

			// Create ApiDefinition with whitelist extended path
			_, err := createTestAPIDef(ctx, testNS, func(apiDef *v1alpha1.ApiDefinition) {
				apiDef.Name = apiDefWithBlacklist
				apiDef.Spec.Proxy = model.Proxy{
					ListenPath:      apiDefListenPath,
					TargetURL:       "http://httpbin.org",
					StripListenPath: true,
				}
				apiDef.Spec.VersionData.DefaultVersion = defaultVersion
				apiDef.Spec.VersionData.NotVersioned = true
				apiDef.Spec.VersionData.Versions = map[string]model.VersionInfo{
					defaultVersion: {Name: defaultVersion, UseExtendedPaths: true, ExtendedPaths: eps},
				}
			}, envConf)
			is.NoErr(err) // failed to create apiDefinition

			return ctx
		}).
		Assess("ApiDefinition must have blacklist field",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				is := is.New(t)
				client := cfg.Client()

				testNS := ctx.Value(ctxNSKey).(string) //nolint:errcheck
				desiredApiDef := v1alpha1.ApiDefinition{
					ObjectMeta: metav1.ObjectMeta{Name: apiDefWithBlacklist, Namespace: testNS},
				}

				err := wait.For(conditions.New(client.Resources()).ResourceMatch(&desiredApiDef, func(object k8s.Object) bool {
					apiDef := object.(*v1alpha1.ApiDefinition) //nolint:errcheck
					return len(apiDef.Spec.VersionData.Versions[defaultVersion].ExtendedPaths.BlackList) == 1
				}), wait.WithTimeout(defaultTimeout))
				is.NoErr(err)

				return ctx
			}).
		Assess("ApiDefniition should forbid traffic to blacklist route",
			func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
				is := is.New(t)

				err := wait.For(func() (done bool, err error) {
					hc := &http.Client{}

					req, err := http.NewRequest(
						http.MethodGet,
						fmt.Sprintf("%s%s", gatewayLocalhost, apiDefListenPath+blackListedPath),
						nil,
					)
					is.NoErr(err)

					resp, err := hc.Do(req)
					is.NoErr(err)

					if resp.StatusCode != errForbiddenResponseCode {
						return false, nil
					}

					return true, nil
				}, wait.WithTimeout(defaultTimeout))
				is.NoErr(err)
				return ctx
			}).
		Assess("ApiDefinition must allow traffic to non-blacklisted route",
			func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
				is := is.New(t)

				err := wait.For(func() (done bool, err error) {
					hc := &http.Client{}

					req, err := http.NewRequest(
						http.MethodGet,
						fmt.Sprintf("%s%s", gatewayLocalhost, apiDefListenPath),
						nil,
					)
					is.NoErr(err)
					req.Header.Add("Content-type", "application/json")

					resp, err := hc.Do(req)
					is.NoErr(err)

					if resp.StatusCode == errForbiddenResponseCode {
						return false, nil
					}

					return true, nil
				}, wait.WithTimeout(defaultTimeout))
				is.NoErr(err)

				return ctx
			}).Feature()

	testenv.Test(t, adCreate)
}

func TestApiDefinitionCreateIgnored(t *testing.T) {
	var (
		apiDefWithWhitelist      = "apidef-ignored"
		apiDefListenPath         = "/test"
		defaultVersion           = "Default"
		errForbiddenResponseCode = 403
		defaultTimeout           = 1 * time.Minute
	)

	const whiteListedPath = "/whitelisted"
	const ignoredPath = "/ignored"
	eps := &model.ExtendedPathsSet{
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

	adCreate := features.New("ApiDefinition").
		Setup(func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
			testNS := ctx.Value(ctxNSKey).(string) //nolint:errcheck
			is := is.New(t)

			// Create ApiDefinition with whitelist + ingored extended path
			_, err := createTestAPIDef(ctx, testNS, func(apiDef *v1alpha1.ApiDefinition) {
				apiDef.Name = apiDefWithWhitelist
				apiDef.Spec.Proxy = model.Proxy{
					ListenPath:      apiDefListenPath,
					TargetURL:       "http://httpbin.org",
					StripListenPath: true,
				}
				apiDef.Spec.VersionData.DefaultVersion = defaultVersion
				apiDef.Spec.VersionData.NotVersioned = true
				apiDef.Spec.VersionData.Versions = map[string]model.VersionInfo{
					defaultVersion: {Name: defaultVersion, UseExtendedPaths: true, ExtendedPaths: eps},
				}
			}, envConf)
			is.NoErr(err) // failed to create apiDefinition

			return ctx
		}).
		Assess("ApiDefinition must have ignored field",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				is := is.New(t)
				client := cfg.Client()

				testNS := ctx.Value(ctxNSKey).(string) //nolint:errcheck
				desiredApiDef := v1alpha1.ApiDefinition{
					ObjectMeta: metav1.ObjectMeta{Name: apiDefWithWhitelist, Namespace: testNS},
				}

				err := wait.For(conditions.New(client.Resources()).ResourceMatch(&desiredApiDef, func(object k8s.Object) bool {
					apiDef := object.(*v1alpha1.ApiDefinition) //nolint:errcheck
					return len(apiDef.Spec.VersionData.Versions[defaultVersion].ExtendedPaths.Ignored) == 1
				}), wait.WithTimeout(defaultTimeout))
				is.NoErr(err)

				return ctx
			}).
		Assess("ApiDefniition should allow traffic to ignored route",
			func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
				is := is.New(t)

				err := wait.For(func() (done bool, err error) {
					hc := &http.Client{}

					req, err := http.NewRequest(
						http.MethodGet,
						fmt.Sprintf("%s%s", gatewayLocalhost, apiDefListenPath+ignoredPath),
						nil,
					)
					is.NoErr(err)

					resp, err := hc.Do(req)
					is.NoErr(err)

					if resp.StatusCode == errForbiddenResponseCode {
						return false, nil
					}

					return true, nil
				}, wait.WithTimeout(defaultTimeout))
				is.NoErr(err)
				return ctx
			}).
		Assess("ApiDefinition must not allow traffic to other non whitelisted routes",
			func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
				is := is.New(t)

				err := wait.For(func() (done bool, err error) {
					hc := &http.Client{}

					req, err := http.NewRequest(
						http.MethodGet,
						fmt.Sprintf("%s%s", gatewayLocalhost, apiDefListenPath+"/randomNonWhiteListedPath"),
						nil,
					)
					is.NoErr(err)
					req.Header.Add("Content-type", "application/json")

					resp, err := hc.Do(req)
					is.NoErr(err)

					if resp.StatusCode != errForbiddenResponseCode {
						return false, nil
					}

					return true, nil
				}, wait.WithTimeout(defaultTimeout))
				is.NoErr(err)

				return ctx
			}).Feature()

	testenv.Test(t, adCreate)
}

func TestAPIDefinition_GraphQL_ExecutionMode(t *testing.T) {
	createAPI := features.New("Create GraphQL API").
		Assess("validate_executionMode", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			testNS := ctx.Value(ctxNSKey).(string) //nolint:errcheck
			is := is.New(t)

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
					ReturnErr:     true,
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
					_, err := createTestAPIDef(ctx, testNS, func(ad *v1alpha1.ApiDefinition) {
						ad.Name = fmt.Sprintf("%s-%s", ad.Name, uuid.New().String())
						ad.Spec.Name = ad.Name
						ad.Spec.GraphQL = &model.GraphQLConfig{
							Enabled:       true,
							ExecutionMode: model.GraphQLExecutionMode(tc.ExecutionMode),
						}
					}, c)

					is.Equal(tc.ReturnErr, err != nil)
				})
			}

			return ctx
		}).Feature()

	testenv.Test(t, createAPI)
}
