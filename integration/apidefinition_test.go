package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/TykTechnologies/tyk-operator/api/model"
	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/pkg/cert"
	"github.com/matryer/is"
	"io"
	v1 "k8s.io/api/core/v1"
	"net/http"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
	"testing"
)

//func TestApiDefinitionCreate(t *testing.T) {
//	var (
//		apiDefWithJSONValidationName = "apidef-json-validation"
//		apiDefListenPath             = "/validation"
//		defaultVersion               = "Default"
//		errorResponseCode            = 422
//		defaultTimeout               = 1 * time.Minute
//	)
//
//	eps := &model.ExtendedPathsSet{
//		ValidateJSON: []model.ValidatePathMeta{{
//			ErrorResponseCode: errorResponseCode,
//			Path:              "/get",
//			Method:            http.MethodGet,
//			Schema: &model.MapStringInterfaceType{Unstructured: unstructured.Unstructured{
//				Object: map[string]interface{}{
//					"properties": map[string]interface{}{
//						"key": map[string]interface{}{
//							"type":      "string",
//							"minLength": 2,
//						},
//					},
//				},
//			}},
//		}},
//	}
//
//	adCreate := features.New("ApiDefinition").
//		Setup(func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
//			testNS := ctx.Value(ctxNSKey).(string) //nolint:errcheck
//			is := is.New(t)
//
//			// Create ApiDefinition with JSON Schema Validation support.
//			_, err := createTestAPIDef(ctx, testNS, func(apiDef *v1alpha1.ApiDefinition) {
//				apiDef.Name = apiDefWithJSONValidationName
//				apiDef.Spec.Proxy = model.Proxy{
//					ListenPath:      apiDefListenPath,
//					TargetURL:       "http://httpbin.org",
//					StripListenPath: true,
//				}
//				apiDef.Spec.VersionData.DefaultVersion = defaultVersion
//				apiDef.Spec.VersionData.NotVersioned = true
//				apiDef.Spec.VersionData.Versions = map[string]model.VersionInfo{
//					defaultVersion: {Name: defaultVersion, UseExtendedPaths: true, ExtendedPaths: eps},
//				}
//			}, envConf)
//			is.NoErr(err) // failed to create apiDefinition
//
//			return ctx
//		}).
//		Assess("ApiDefinition must have ValidateJSON field",
//			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
//				is := is.New(t)
//				client := cfg.Client()
//
//				testNS := ctx.Value(ctxNSKey).(string) //nolint:errcheck
//				desiredApiDef := v1alpha1.ApiDefinition{
//					ObjectMeta: metav1.ObjectMeta{Name: apiDefWithJSONValidationName, Namespace: testNS},
//				}
//
//				err := wait.For(conditions.New(client.Resources()).ResourceMatch(&desiredApiDef, func(object k8s.Object) bool {
//					apiDef := object.(*v1alpha1.ApiDefinition) //nolint:errcheck
//					// 'validate_json' field must exist in the ApiDefinition object.
//					return len(apiDef.Spec.VersionData.Versions[defaultVersion].ExtendedPaths.ValidateJSON) == 1
//				}), wait.WithTimeout(defaultTimeout))
//				is.NoErr(err)
//
//				return ctx
//			}).
//		Assess("ApiDefinition must verify user requests",
//			func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
//				is := is.New(t)
//
//				err := wait.For(func() (done bool, err error) {
//					hc := &http.Client{}
//
//					// invalidJSONBody does not meet the requirements of the Schema because
//					// Schema requires the "key" field to have a length of 2 at least.
//					invalidJSONBody := strings.NewReader(`{"key": "a"}`)
//
//					req, err := http.NewRequest(
//						http.MethodGet,
//						fmt.Sprintf("%s%s/get", gatewayLocalhost, apiDefListenPath),
//						invalidJSONBody,
//					)
//					is.NoErr(err)
//					req.Header.Add("Content-type", "application/json")
//
//					// Since the following request does not match with the JSON Validation Schema,
//					// the response status code must be 422 as indicated in the ErrorResponseCode of the ValidatePathMeta.
//					resp, err := hc.Do(req)
//					is.NoErr(err)
//
//					if resp.StatusCode != errorResponseCode {
//						return false, nil
//					}
//
//					return true, nil
//				}, wait.WithTimeout(defaultTimeout))
//				is.NoErr(err)
//
//				return ctx
//			}).Feature()
//
//	testenv.Test(t, adCreate)
//}
//
//func TestApiDefinitionCreateWhitelist(t *testing.T) {
//	var (
//		apiDefWithWhitelist      = "apidef-whitelist"
//		apiDefListenPath         = "/test"
//		defaultVersion           = "Default"
//		errForbiddenResponseCode = 403
//		defaultTimeout           = 1 * time.Minute
//	)
//
//	const whiteListedPath = "/whitelisted"
//	eps := &model.ExtendedPathsSet{
//		WhiteList: []model.EndPointMeta{{
//			Path:       whiteListedPath,
//			IgnoreCase: true,
//			MethodActions: map[string]model.EndpointMethodMeta{
//				"GET": {
//					Action: "no_action",
//					Code:   200, Data: "",
//					Headers: make(map[string]string),
//				},
//			},
//		}},
//	}
//
//	adCreate := features.New("ApiDefinition").
//		Setup(func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
//			testNS := ctx.Value(ctxNSKey).(string) //nolint:errcheck
//			is := is.New(t)
//
//			// Create ApiDefinition with whitelist extended path
//			_, err := createTestAPIDef(ctx, testNS, func(apiDef *v1alpha1.ApiDefinition) {
//				apiDef.Name = apiDefWithWhitelist
//				apiDef.Spec.Proxy = model.Proxy{
//					ListenPath:      apiDefListenPath,
//					TargetURL:       "http://httpbin.org",
//					StripListenPath: true,
//				}
//				apiDef.Spec.VersionData.DefaultVersion = defaultVersion
//				apiDef.Spec.VersionData.NotVersioned = true
//				apiDef.Spec.VersionData.Versions = map[string]model.VersionInfo{
//					defaultVersion: {Name: defaultVersion, UseExtendedPaths: true, ExtendedPaths: eps},
//				}
//			}, envConf)
//			is.NoErr(err) // failed to create apiDefinition
//
//			return ctx
//		}).
//		Assess("ApiDefinition must have whitelist field",
//			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
//				is := is.New(t)
//				client := cfg.Client()
//
//				testNS := ctx.Value(ctxNSKey).(string) //nolint:errcheck
//				desiredApiDef := v1alpha1.ApiDefinition{
//					ObjectMeta: metav1.ObjectMeta{Name: apiDefWithWhitelist, Namespace: testNS},
//				}
//
//				err := wait.For(conditions.New(client.Resources()).ResourceMatch(&desiredApiDef, func(object k8s.Object) bool {
//					apiDef := object.(*v1alpha1.ApiDefinition) //nolint:errcheck
//					return len(apiDef.Spec.VersionData.Versions[defaultVersion].ExtendedPaths.WhiteList) == 1
//				}), wait.WithTimeout(defaultTimeout))
//				is.NoErr(err)
//
//				return ctx
//			}).
//		Assess("ApiDefniition should allow traffic to whitelisted route",
//			func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
//				is := is.New(t)
//
//				err := wait.For(func() (done bool, err error) {
//					hc := &http.Client{}
//
//					req, err := http.NewRequest(
//						http.MethodGet,
//						fmt.Sprintf("%s%s", gatewayLocalhost, apiDefListenPath+whiteListedPath),
//						nil,
//					)
//					is.NoErr(err)
//
//					resp, err := hc.Do(req)
//					is.NoErr(err)
//
//					if resp.StatusCode == errForbiddenResponseCode {
//						return false, nil
//					}
//
//					return true, nil
//				}, wait.WithTimeout(defaultTimeout))
//				is.NoErr(err)
//				return ctx
//			}).
//		Assess("ApiDefinition must not allow traffic to non-whitelisted route",
//			func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
//				is := is.New(t)
//
//				err := wait.For(func() (done bool, err error) {
//					hc := &http.Client{}
//
//					req, err := http.NewRequest(
//						http.MethodGet,
//						fmt.Sprintf("%s%s", gatewayLocalhost, apiDefListenPath),
//						nil,
//					)
//					is.NoErr(err)
//					req.Header.Add("Content-type", "application/json")
//
//					resp, err := hc.Do(req)
//					is.NoErr(err)
//
//					if resp.StatusCode != errForbiddenResponseCode {
//						return false, nil
//					}
//
//					return true, nil
//				}, wait.WithTimeout(defaultTimeout))
//				is.NoErr(err)
//
//				return ctx
//			}).Feature()
//
//	testenv.Test(t, adCreate)
//}
//
//func TestApiDefinitionCreateBlackList(t *testing.T) {
//	var (
//		apiDefWithBlacklist      = "apidef-blacklist"
//		apiDefListenPath         = "/test"
//		defaultVersion           = "Default"
//		errForbiddenResponseCode = 403
//		defaultTimeout           = 1 * time.Minute
//	)
//
//	const blackListedPath = "/blacklisted"
//	eps := &model.ExtendedPathsSet{
//		BlackList: []model.EndPointMeta{{
//			Path:       blackListedPath,
//			IgnoreCase: true,
//			MethodActions: map[string]model.EndpointMethodMeta{
//				"GET": {
//					Action: "no_action",
//					Code:   200, Data: "",
//					Headers: make(map[string]string),
//				},
//			},
//		}},
//	}
//
//	adCreate := features.New("ApiDefinition").
//		Setup(func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
//			testNS := ctx.Value(ctxNSKey).(string) //nolint:errcheck
//			is := is.New(t)
//
//			// Create ApiDefinition with whitelist extended path
//			_, err := createTestAPIDef(ctx, testNS, func(apiDef *v1alpha1.ApiDefinition) {
//				apiDef.Name = apiDefWithBlacklist
//				apiDef.Spec.Proxy = model.Proxy{
//					ListenPath:      apiDefListenPath,
//					TargetURL:       "http://httpbin.org",
//					StripListenPath: true,
//				}
//				apiDef.Spec.VersionData.DefaultVersion = defaultVersion
//				apiDef.Spec.VersionData.NotVersioned = true
//				apiDef.Spec.VersionData.Versions = map[string]model.VersionInfo{
//					defaultVersion: {Name: defaultVersion, UseExtendedPaths: true, ExtendedPaths: eps},
//				}
//			}, envConf)
//			is.NoErr(err) // failed to create apiDefinition
//
//			return ctx
//		}).
//		Assess("ApiDefinition must have blacklist field",
//			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
//				is := is.New(t)
//				client := cfg.Client()
//
//				testNS := ctx.Value(ctxNSKey).(string) //nolint:errcheck
//				desiredApiDef := v1alpha1.ApiDefinition{
//					ObjectMeta: metav1.ObjectMeta{Name: apiDefWithBlacklist, Namespace: testNS},
//				}
//
//				err := wait.For(conditions.New(client.Resources()).ResourceMatch(&desiredApiDef, func(object k8s.Object) bool {
//					apiDef := object.(*v1alpha1.ApiDefinition) //nolint:errcheck
//					return len(apiDef.Spec.VersionData.Versions[defaultVersion].ExtendedPaths.BlackList) == 1
//				}), wait.WithTimeout(defaultTimeout))
//				is.NoErr(err)
//
//				return ctx
//			}).
//		Assess("ApiDefniition should forbid traffic to blacklist route",
//			func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
//				is := is.New(t)
//
//				err := wait.For(func() (done bool, err error) {
//					hc := &http.Client{}
//
//					req, err := http.NewRequest(
//						http.MethodGet,
//						fmt.Sprintf("%s%s", gatewayLocalhost, apiDefListenPath+blackListedPath),
//						nil,
//					)
//					is.NoErr(err)
//
//					resp, err := hc.Do(req)
//					is.NoErr(err)
//
//					if resp.StatusCode != errForbiddenResponseCode {
//						return false, nil
//					}
//
//					return true, nil
//				}, wait.WithTimeout(defaultTimeout))
//				is.NoErr(err)
//				return ctx
//			}).
//		Assess("ApiDefinition must allow traffic to non-blacklisted route",
//			func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
//				is := is.New(t)
//
//				err := wait.For(func() (done bool, err error) {
//					hc := &http.Client{}
//
//					req, err := http.NewRequest(
//						http.MethodGet,
//						fmt.Sprintf("%s%s", gatewayLocalhost, apiDefListenPath),
//						nil,
//					)
//					is.NoErr(err)
//					req.Header.Add("Content-type", "application/json")
//
//					resp, err := hc.Do(req)
//					is.NoErr(err)
//
//					if resp.StatusCode == errForbiddenResponseCode {
//						return false, nil
//					}
//
//					return true, nil
//				}, wait.WithTimeout(defaultTimeout))
//				is.NoErr(err)
//
//				return ctx
//			}).Feature()
//
//	testenv.Test(t, adCreate)
//}
//
//func TestApiDefinitionCreateIgnored(t *testing.T) {
//	var (
//		apiDefWithWhitelist      = "apidef-ignored"
//		apiDefListenPath         = "/test"
//		defaultVersion           = "Default"
//		errForbiddenResponseCode = 403
//		defaultTimeout           = 1 * time.Minute
//	)
//
//	const whiteListedPath = "/whitelisted"
//	const ignoredPath = "/ignored"
//	eps := &model.ExtendedPathsSet{
//		WhiteList: []model.EndPointMeta{{
//			Path:       whiteListedPath,
//			IgnoreCase: true,
//			MethodActions: map[string]model.EndpointMethodMeta{
//				"GET": {
//					Action: "no_action",
//					Code:   200, Data: "",
//					Headers: make(map[string]string),
//				},
//			},
//		}},
//		Ignored: []model.EndPointMeta{{
//			Path:       ignoredPath,
//			IgnoreCase: true,
//			MethodActions: map[string]model.EndpointMethodMeta{
//				"GET": {
//					Action: "no_action",
//					Code:   200, Data: "",
//					Headers: make(map[string]string),
//				},
//			},
//		}},
//	}
//
//	adCreate := features.New("ApiDefinition").
//		Setup(func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
//			testNS := ctx.Value(ctxNSKey).(string) //nolint:errcheck
//			is := is.New(t)
//
//			// Create ApiDefinition with whitelist + ingored extended path
//			_, err := createTestAPIDef(ctx, testNS, func(apiDef *v1alpha1.ApiDefinition) {
//				apiDef.Name = apiDefWithWhitelist
//				apiDef.Spec.Proxy = model.Proxy{
//					ListenPath:      apiDefListenPath,
//					TargetURL:       "http://httpbin.org",
//					StripListenPath: true,
//				}
//				apiDef.Spec.VersionData.DefaultVersion = defaultVersion
//				apiDef.Spec.VersionData.NotVersioned = true
//				apiDef.Spec.VersionData.Versions = map[string]model.VersionInfo{
//					defaultVersion: {Name: defaultVersion, UseExtendedPaths: true, ExtendedPaths: eps},
//				}
//			}, envConf)
//			is.NoErr(err) // failed to create apiDefinition
//
//			return ctx
//		}).
//		Assess("ApiDefinition must have ignored field",
//			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
//				is := is.New(t)
//				client := cfg.Client()
//
//				testNS := ctx.Value(ctxNSKey).(string) //nolint:errcheck
//				desiredApiDef := v1alpha1.ApiDefinition{
//					ObjectMeta: metav1.ObjectMeta{Name: apiDefWithWhitelist, Namespace: testNS},
//				}
//
//				err := wait.For(conditions.New(client.Resources()).ResourceMatch(&desiredApiDef, func(object k8s.Object) bool {
//					apiDef := object.(*v1alpha1.ApiDefinition) //nolint:errcheck
//					return len(apiDef.Spec.VersionData.Versions[defaultVersion].ExtendedPaths.Ignored) == 1
//				}), wait.WithTimeout(defaultTimeout))
//				is.NoErr(err)
//
//				return ctx
//			}).
//		Assess("ApiDefniition should allow traffic to ignored route",
//			func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
//				is := is.New(t)
//
//				err := wait.For(func() (done bool, err error) {
//					hc := &http.Client{}
//
//					req, err := http.NewRequest(
//						http.MethodGet,
//						fmt.Sprintf("%s%s", gatewayLocalhost, apiDefListenPath+ignoredPath),
//						nil,
//					)
//					is.NoErr(err)
//
//					resp, err := hc.Do(req)
//					is.NoErr(err)
//
//					if resp.StatusCode == errForbiddenResponseCode {
//						return false, nil
//					}
//
//					return true, nil
//				}, wait.WithTimeout(defaultTimeout))
//				is.NoErr(err)
//				return ctx
//			}).
//		Assess("ApiDefinition must not allow traffic to other non whitelisted routes",
//			func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
//				is := is.New(t)
//
//				err := wait.For(func() (done bool, err error) {
//					hc := &http.Client{}
//
//					req, err := http.NewRequest(
//						http.MethodGet,
//						fmt.Sprintf("%s%s", gatewayLocalhost, apiDefListenPath+"/randomNonWhiteListedPath"),
//						nil,
//					)
//					is.NoErr(err)
//					req.Header.Add("Content-type", "application/json")
//
//					resp, err := hc.Do(req)
//					is.NoErr(err)
//
//					if resp.StatusCode != errForbiddenResponseCode {
//						return false, nil
//					}
//
//					return true, nil
//				}, wait.WithTimeout(defaultTimeout))
//				is.NoErr(err)
//
//				return ctx
//			}).Feature()
//
//	testenv.Test(t, adCreate)
//}

func TestApiDefinitionUpstreamCertificates(t *testing.T) {
	var (
		apiDefUpstreamCerts = "apidef-upstream-certs"
		defaultVersion      = "Default"

		//defaultTimeout = 3 * time.Minute
	)

	adCreate := features.New("Create an ApiDefinition for Upstream TLS").
		Setup(func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {

			testNS := ctx.Value(ctxNSKey).(string) //nolint: errcheck
			is := is.New(t)
			t.Log(testNS)
			_, err := createTestTlsSecret(ctx, testNS, nil, envConf)
			is.NoErr(err)

			return ctx
		}).
		Setup(func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
			testNS := ctx.Value(ctxNSKey).(string) //nolint:errcheck
			is := is.New(t)

			//_, err := createTestTlsSecret(ctx, testNS, func(secret *v1.Secret) {}, envConf)
			//is.NoErr(err) // failed to create apiDefinition

			// Create ApiDefinition with Certificate Pinning.
			_, err := createTestAPIDef(ctx, testNS, func(apiDef *v1alpha1.ApiDefinition) {
				certName := "test-tls-secret-name"
				//apiDef.Spec.OrgID = "test-org"
				apiDef.Name = apiDefUpstreamCerts
				apiDef.Spec.UpstreamCertificateRefs = map[string]string{
					"*": certName,
				}
				apiDef.Spec.VersionData.DefaultVersion = defaultVersion
				apiDef.Spec.VersionData.NotVersioned = true
				apiDef.Spec.VersionData.Versions = map[string]model.VersionInfo{
					defaultVersion: {Name: defaultVersion},
				}
			}, envConf)
			is.NoErr(err) // failed to create apiDefinition

			return ctx
		}).
		Assess("ApiDefinition must have upstream field defined",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				is := is.New(t)
				client := cfg.Client()
				testNS := ctx.Value(ctxNSKey).(string) //nolint:errcheck

				tlsSecret := v1.Secret{} //nolint:errcheck

				err2 := client.Resources(testNS).Get(ctx, "test-tls-secret-name", testNS, &tlsSecret)
				is.NoErr(err2)

				certPemBytes, ok := tlsSecret.Data["tls.crt"]
				if !ok {
					is.Fail()
				}
				certFingerPrint := cert.CalculateFingerPrint(certPemBytes)

				opConfSecret := v1.Secret{}
				opNs := "tyk-operator-system"
				err3 := client.Resources(opNs).Get(ctx, "tyk-operator-conf", opNs, &opConfSecret)
				is.NoErr(err3)

				tykAuth, ok := opConfSecret.Data["TYK_AUTH"]
				if !ok {
					is.Fail()
				}

				tykOrg, ok := opConfSecret.Data["TYK_ORG"]
				if !ok {
					is.Fail()
				}

				calculatedCertID := string(tykOrg) + certFingerPrint
				t.Log(fmt.Sprintf("certId is %s", calculatedCertID))
				//existsTest := klient.Universal.Certificate().Exists(ctx, calculatedCertID)

				err := wait.For(func() (done bool, err error) {
					hc := &http.Client{}

					req, err := http.NewRequest(
						http.MethodGet,
						fmt.Sprintf("%s/api/certs/?certId=%s&org_id=%s", dashboardLocalHost, calculatedCertID, string(tykOrg)),
						nil,
					)
					is.NoErr(err)
					req.Header.Add("Content-type", "application/json")
					req.Header.Add("authorization", string(tykAuth))

					resp, err := hc.Do(req)
					is.NoErr(err)

					response, err := io.ReadAll(resp.Body)

					if err != nil {
						return false, nil
					}

					certResponse := struct {
						Certs []string `json:"certs"`
						Pages int      `json:"pages"`
					}{}
					err = json.Unmarshal(response, &certResponse)
					if err != nil {
						return false, nil
					}

					if len(certResponse.Certs) != 1 {
						return false, nil
					}
					return true, nil

				})
				is.NoErr(err)
				return ctx
			}).Feature()

	testenv.Test(t, adCreate)
}
