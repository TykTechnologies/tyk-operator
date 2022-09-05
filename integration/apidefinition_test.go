package integration

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/TykTechnologies/tyk-operator/pkg/cert"
	"github.com/TykTechnologies/tyk-operator/pkg/client/klient"
	"github.com/google/uuid"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	v1 "k8s.io/api/core/v1"

	"github.com/TykTechnologies/tyk-operator/api/model"
	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	pkgclient "github.com/TykTechnologies/tyk-operator/pkg/client"
	"github.com/TykTechnologies/tyk-operator/pkg/environmet"
	"github.com/matryer/is"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestApiDefinitionJSONSchemaValidation(t *testing.T) {
	var (
		apiDefWithJSONValidationName = "apidef-json-validation"
		apiDefListenPath             = "/validation"
		defaultVersion               = "Default"
		errorResponseCode            = 422
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
				}), wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
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
				}, wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
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
				}), wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
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
				}, wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
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
				}, wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
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
				}), wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
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
				}, wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
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
				}, wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
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
				}), wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
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
				}, wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
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
				}, wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
				is.NoErr(err)

				return ctx
			}).Feature()

	testenv.Test(t, adCreate)
}

func TestApiDefinitionCertificatePinning(t *testing.T) {
	var (
		apiDefPinning = "apidef-certificate-pinning"

		apiDefPinningViaSecret           = "apidef-with-secret"
		apiDefPinningViaSecretListenPath = "/secret"

		invalidApiDef           = "invalid-proxy-apidef"
		invalidApiDefListenPath = "/invalid"

		secretName = "secret"

		publicKeyID = "test-public-key-id"
		pubKeyPem   = []byte(`-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAhOQnpezrwA0vHzf47Pa+
O84fWue/562TqQrVirtf+3fsGQd3MmwnId+ksAGQvWN4M1/hSelYJb246pFqGB7t
+ZI+vjBYH4/J6CiFsKwzusqkSF63ftQh8Ox0OasB9HvRlOPHT/B5Dskh8HNiJ+1l
ExSZEaO9zsQ9wO62bsGHsMX/UP3VQByXLVBZu0DMKsl2hGaUNy9+LgZv4/iVpWDP
Q1+khpfxP9x1H+mMlUWBgYPq7jG5ceTbltIoF/sUQPNR+yKIBSnuiISXFHO9HEnk
5ph610hWmVQKIrCAPsAUMM9m6+iDb64NjrMjWV/bkm36r+FBMz9L8HfEB4hxlwwg
5QIDAQAB
-----END PUBLIC KEY-----`)
	)

	adCreate := features.New("Create ApiDefinition objects for Certificate Pinning").
		Setup(func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
			testNS := ctx.Value(ctxNSKey).(string) //nolint:errcheck
			client := envConf.Client()
			is := is.New(t)

			// Create an ApiDefinition with Certificate Pinning using 'pinned_public_keys' field.
			// It contains a dummy public key which is not valid to be used.
			_, err := createTestAPIDef(ctx, testNS, func(apiDef *v1alpha1.ApiDefinition) {
				apiDef.Name = apiDefPinning
				apiDef.Spec.PinnedPublicKeys = map[string]string{"*": publicKeyID}
			}, envConf)
			is.NoErr(err)

			// Create a secret to store the public key of httpbin.org.
			secret := &v1.Secret{
				Type: v1.SecretTypeTLS,
				ObjectMeta: metav1.ObjectMeta{
					Name:      secretName,
					Namespace: testNS,
				},
				Data: map[string][]byte{"tls.crt": pubKeyPem, "tls.key": []byte("")},
			}

			err = client.Resources(testNS).Create(ctx, secret)
			is.NoErr(err)

			// For all domains (`*`), use following secret that contains the public key of the httpbin.org
			// So, if you make any requests to any addresses except httpbin.org, we should get proxy errors because
			// of pinned public key.
			publicKeySecrets := map[string]string{"*": secretName}

			// Create an ApiDefinition with Certificate Pinning using Kubernetes Secret object.
			_, err = createTestAPIDef(ctx, testNS, func(apiDef *v1alpha1.ApiDefinition) {
				apiDef.Name = apiDefPinningViaSecret
				apiDef.Spec.Name = "valid"
				apiDef.Spec.Proxy = model.Proxy{
					ListenPath:      apiDefPinningViaSecretListenPath,
					TargetURL:       "https://httpbin.org/",
					StripListenPath: true,
				}
				apiDef.Spec.PinnedPublicKeysRefs = publicKeySecrets
			}, envConf)
			is.NoErr(err)

			// Create an invalid ApiDefinition with Certificate Pinning using Kubernetes Secret object.
			// Although this ApiDefinition has a Public Key of httpbin.org for all domains, this ApiDefinition will try
			// to reach github.com, which must fail due to proxy error.
			_, err = createTestAPIDef(ctx, testNS, func(apiDef *v1alpha1.ApiDefinition) {
				apiDef.Name = invalidApiDef
				apiDef.Spec.Proxy = model.Proxy{
					ListenPath:      invalidApiDefListenPath,
					TargetURL:       "https://github.com/",
					StripListenPath: true,
				}
				apiDef.Spec.Name = "invalid"
				apiDef.Spec.PinnedPublicKeysRefs = publicKeySecrets
			}, envConf)
			is.NoErr(err)

			return ctx
		}).
		Assess("ApiDefinition must have Certificate Pinning field defined",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				is := is.New(t)
				client := cfg.Client()

				testNS := ctx.Value(ctxNSKey).(string) //nolint:errcheck
				desiredApiDef := v1alpha1.ApiDefinition{
					ObjectMeta: metav1.ObjectMeta{Name: apiDefPinning, Namespace: testNS},
				}

				err := wait.For(conditions.New(client.Resources()).ResourceMatch(&desiredApiDef, func(object k8s.Object) bool {
					apiDef := object.(*v1alpha1.ApiDefinition) //nolint:errcheck

					if apiDef.Spec.PinnedPublicKeys == nil {
						t.Log("PinnedPublicKeys field is undefined.")
						return false
					}

					// 'pinned_public_keys' field must exist in the ApiDefinition object.
					val, ok := apiDef.Spec.PinnedPublicKeys["*"]
					if !ok {
						t.Log("cannot find a public key for the domain '*'")
						return false
					}

					return val == publicKeyID
				}), wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
				is.NoErr(err)

				return ctx
			}).
		Assess("Allow making requests based on the pinned public key defined via secret",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				is := is.New(t)
				err := wait.For(func() (done bool, err error) {
					hc := &http.Client{}
					req, err := http.NewRequest(
						http.MethodGet,
						fmt.Sprintf("%s%s", gatewayLocalhost, apiDefPinningViaSecretListenPath),
						nil,
					)
					is.NoErr(err)

					resp, err := hc.Do(req)
					is.NoErr(err)

					if resp.StatusCode != 200 {
						t.Log("expected to access httpbin.org since it is pinned via public key.")
						return false, nil
					}

					return true, nil
				}, wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
				is.NoErr(err)

				return ctx
			}).
		Assess("Prevent making requests to disallowed addresses based on the pinned public key defined via secret",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				is := is.New(t)
				err := wait.For(func() (done bool, err error) {
					hc := &http.Client{}
					req, err := http.NewRequest(
						http.MethodGet,
						fmt.Sprintf("%s%s", gatewayLocalhost, invalidApiDefListenPath),
						nil,
					)
					is.NoErr(err)

					resp, err := hc.Do(req)
					is.NoErr(err)

					if resp.StatusCode == 200 {
						t.Log("unexpected access to invalid address")
						return false, nil
					}

					return true, nil
				}, wait.WithInterval(defaultWaitInterval), wait.WithTimeout(defaultWaitTimeout))
				is.NoErr(err)

				return ctx
			}).
		Feature()

	testenv.Test(t, adCreate)
}

func TestApiDefinitionUpstreamCertificates(t *testing.T) {
	var (
		apiDefUpstreamCerts = "apidef-upstream-certs"
		defaultVersion      = "Default"
		opNs                = "tyk-operator-system"
		certName            = "test-tls-secret-name"
		tykConnectionURL    = ""
	)

	mode := os.Getenv("TYK_MODE")

	switch mode {
	case "pro":
		tykConnectionURL = adminLocalhost
	case "ce":
		tykConnectionURL = gatewayLocalhost
	}

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

			// Create ApiDefinition with Upstream certificate
			_, err := createTestAPIDef(ctx, testNS, func(apiDef *v1alpha1.ApiDefinition) {
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

				err := client.Resources(testNS).Get(ctx, certName, testNS, &tlsSecret)
				is.NoErr(err)

				certPemBytes, ok := tlsSecret.Data["tls.crt"]
				if !ok {
					is.Fail()
				}
				certFingerPrint := cert.CalculateFingerPrint(certPemBytes)

				opConfSecret := v1.Secret{}
				err = client.Resources(opNs).Get(ctx, "tyk-operator-conf", opNs, &opConfSecret)
				is.NoErr(err)

				tykAuth, ok := opConfSecret.Data["TYK_AUTH"]
				if !ok {
					is.Fail()
				}

				tykOrg, ok := opConfSecret.Data["TYK_ORG"]
				if !ok {
					is.Fail()
				}

				calculatedCertID := string(tykOrg) + certFingerPrint

				err = wait.For(func() (done bool, err error) {
					env := environmet.Env{}
					env.Mode = v1alpha1.OperatorContextMode(mode)
					env.Org = string(tykOrg)
					env.Auth = string(tykAuth)
					env.URL = tykConnectionURL

					pkgContext := pkgclient.Context{
						Env: env,
						Log: log.NullLogger{},
					}

					// validate certificate was created
					reqContext := pkgclient.SetContext(context.Background(), pkgContext)
					exists := klient.Universal.Certificate().Exists(reqContext, calculatedCertID)

					is.True(exists)

					return true, nil
				}, wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
				is.NoErr(err)

				return ctx
			}).Feature()

	testenv.Test(t, adCreate)
}

func TestApiDefinitionClientMTLS(t *testing.T) {
	var (
		apiDefClientMTLSWithCert    = "apidef-client-mtls-with-cert"
		apiDefClientMTLSWithoutCert = "apidef-client-mtls-without-cert"
		defaultVersion              = "Default"
		opNs                        = "tyk-operator-system"
		certName                    = "test-tls-secret-name"
		tykConnectionURL            = ""
		tykOrg                      = ""
		tykAuth                     = ""
	)

	type ContextKey string
	var certIDCtxKey ContextKey = "certID"

	mode := os.Getenv("TYK_MODE")

	switch mode {
	case "pro":
		tykConnectionURL = adminLocalhost
	case "ce":
		tykConnectionURL = gatewayLocalhost
	}

	testWithCert := features.New("Client MTLS with certificate").
		Setup(func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
			client := envConf.Client()
			is := is.New(t)
			opConfSecret := v1.Secret{}

			err := client.Resources(opNs).Get(ctx, "tyk-operator-conf", opNs, &opConfSecret)
			is.NoErr(err)

			data, ok := opConfSecret.Data["TYK_AUTH"]
			if !ok {
				is.Fail()
			}

			tykAuth = string(data)

			data, ok = opConfSecret.Data["TYK_ORG"]
			if !ok {
				is.Fail()
			}

			tykOrg = string(data)

			return ctx
		}).
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

			// Create ApiDefinition with Client certificate
			_, err := createTestAPIDef(ctx, testNS, func(apiDef *v1alpha1.ApiDefinition) {
				apiDef.Name = apiDefClientMTLSWithCert
				apiDef.Spec.UseMutualTLSAuth = true
				apiDef.Spec.ClientCertificateRefs = []string{certName}
				apiDef.Spec.VersionData.DefaultVersion = defaultVersion
				apiDef.Spec.VersionData.NotVersioned = true
				apiDef.Spec.VersionData.Versions = map[string]model.VersionInfo{
					defaultVersion: {Name: defaultVersion},
				}
			}, envConf)
			is.NoErr(err) // failed to create apiDefinition

			return ctx
		}).Assess("Certificate from secret must be uploaded",
		func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			is := is.New(t)
			client := cfg.Client()
			testNS := ctx.Value(ctxNSKey).(string) //nolint:errcheck

			tlsSecret := v1.Secret{} //nolint:errcheck

			err := client.Resources(testNS).Get(ctx, certName, testNS, &tlsSecret)
			is.NoErr(err)

			certPemBytes, ok := tlsSecret.Data["tls.crt"]
			if !ok {
				is.Fail()
			}

			certFingerPrint := cert.CalculateFingerPrint(certPemBytes)
			calculatedCertID := tykOrg + certFingerPrint

			err = wait.For(func() (done bool, err error) {
				env := environmet.Env{}
				env.Mode = v1alpha1.OperatorContextMode(mode)
				env.Org = string(tykOrg)
				env.Auth = string(tykAuth)
				env.URL = tykConnectionURL

				pkgContext := pkgclient.Context{
					Env: env,
					Log: log.NullLogger{},
				}

				// validate certificate was created
				reqContext := pkgclient.SetContext(context.Background(), pkgContext)
				exists := klient.Universal.Certificate().Exists(reqContext, calculatedCertID)

				if !exists {
					return false, errors.New("Certificate is not created yet")
				}

				return true, nil
			}, wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
			is.NoErr(err)

			ctx = context.WithValue(ctx, certIDCtxKey, calculatedCertID)

			return ctx
		}).
		Assess("API must have client certificate field defined",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				is := is.New(t)
				client := cfg.Client()
				testNS := ctx.Value(ctxNSKey).(string) //nolint:errcheck

				certID := ctx.Value(certIDCtxKey)

				err := wait.For(func() (done bool, err error) {
					env := environmet.Env{}
					env.Mode = v1alpha1.OperatorContextMode(mode)
					env.Org = string(tykOrg)
					env.Auth = string(tykAuth)
					env.URL = tykConnectionURL

					pkgContext := pkgclient.Context{
						Env: env,
						Log: log.NullLogger{},
					}

					reqContext := pkgclient.SetContext(context.Background(), pkgContext)

					// validate client certificate field was set
					var apiDefCRD v1alpha1.ApiDefinition

					err = client.Resources().Get(ctx, apiDefClientMTLSWithCert, testNS, &apiDefCRD)
					is.NoErr(err)

					apiDef, err := klient.Universal.Api().Get(reqContext, apiDefCRD.Status.ApiID)
					if err != nil {
						return false, errors.New("API is not created yet")
					}

					is.True(len(apiDef.ClientCertificates) == 1)
					is.True(apiDef.ClientCertificates[0] == certID)

					return true, nil
				}, wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
				is.NoErr(err)

				return ctx
			}).Feature()

	testWithoutCert := features.New("Client MTLS without certs").
		Setup(func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
			testNS := ctx.Value(ctxNSKey).(string) //nolint:errcheck
			is := is.New(t)

			// Create ApiDefinition with Upstream certificate
			_, err := createTestAPIDef(ctx, testNS, func(apiDef *v1alpha1.ApiDefinition) {
				apiDef.Name = apiDefClientMTLSWithoutCert
				apiDef.Spec.UseMutualTLSAuth = true
				apiDef.Spec.ClientCertificateRefs = []string{certName}
				apiDef.Spec.VersionData.DefaultVersion = defaultVersion
				apiDef.Spec.VersionData.NotVersioned = true
				apiDef.Spec.VersionData.Versions = map[string]model.VersionInfo{
					defaultVersion: {Name: defaultVersion},
				}
			}, envConf)
			is.NoErr(err) // failed to create apiDefinition

			return ctx
		}).
		Assess("API should be created even though certs doesn't exists",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				is := is.New(t)
				client := cfg.Client()
				testNS := ctx.Value(ctxNSKey).(string) //nolint:errcheck

				opConfSecret := v1.Secret{}
				err := client.Resources(opNs).Get(ctx, "tyk-operator-conf", opNs, &opConfSecret)
				is.NoErr(err)

				tykAuth, ok := opConfSecret.Data["TYK_AUTH"]
				if !ok {
					is.Fail()
				}

				tykOrg, ok := opConfSecret.Data["TYK_ORG"]
				if !ok {
					is.Fail()
				}

				err = wait.For(func() (done bool, err error) {
					env := environmet.Env{}
					env.Mode = v1alpha1.OperatorContextMode(mode)
					env.Org = string(tykOrg)
					env.Auth = string(tykAuth)
					env.URL = tykConnectionURL

					pkgContext := pkgclient.Context{
						Env: env,
						Log: log.NullLogger{},
					}

					reqContext := pkgclient.SetContext(context.Background(), pkgContext)

					// validate api Def was created without certificate
					var apiDefCRD v1alpha1.ApiDefinition

					err = client.Resources().Get(ctx, apiDefClientMTLSWithoutCert, testNS, &apiDefCRD)
					is.NoErr(err)

					apiDef, err := klient.Universal.Api().Get(reqContext, apiDefCRD.Status.ApiID)
					is.NoErr(err)

					is.True(len(apiDef.ClientCertificates) == 0)

					return true, nil
				}, wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
				is.NoErr(err)

				return ctx
			}).Feature()

	testenv.Test(t, testWithCert)
	testenv.Test(t, testWithoutCert)
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

					t.Log("Error=", err)
					is.Equal(tc.ReturnErr, err != nil)
				})
			}

			return ctx
		}).Feature()

	testenv.Test(t, createAPI)
}
