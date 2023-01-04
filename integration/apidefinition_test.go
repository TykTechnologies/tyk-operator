package integration

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/TykTechnologies/tyk-operator/api/model"
	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/controllers"
	"github.com/TykTechnologies/tyk-operator/pkg/cert"
	pkgclient "github.com/TykTechnologies/tyk-operator/pkg/client"
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
			eval := is.New(t)

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
			eval.NoErr(err) // failed to create apiDefinition

			return ctx
		}).
		Assess("ApiDefinition must have ValidateJSON field",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				eval := is.New(t)
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
				eval.NoErr(err)

				return ctx
			}).
		Assess("ApiDefinition must verify user requests",
			func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
				eval := is.New(t)

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
			eval := is.New(t)

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
			eval.NoErr(err) // failed to create apiDefinition

			return ctx
		}).
		Assess("ApiDefinition must have whitelist field",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				eval := is.New(t)
				client := cfg.Client()

				testNS := ctx.Value(ctxNSKey).(string) //nolint:errcheck
				desiredApiDef := v1alpha1.ApiDefinition{
					ObjectMeta: metav1.ObjectMeta{Name: apiDefWithWhitelist, Namespace: testNS},
				}

				err := wait.For(conditions.New(client.Resources()).ResourceMatch(&desiredApiDef, func(object k8s.Object) bool {
					apiDef := object.(*v1alpha1.ApiDefinition) //nolint:errcheck
					return len(apiDef.Spec.VersionData.Versions[defaultVersion].ExtendedPaths.WhiteList) == 1
				}), wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
				eval.NoErr(err)

				return ctx
			}).
		Assess("ApiDefinition should allow traffic to whitelisted route",
			func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
				eval := is.New(t)

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
				eval := is.New(t)

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
			eval := is.New(t)

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
			eval.NoErr(err) // failed to create apiDefinition

			return ctx
		}).
		Assess("ApiDefinition must have blacklist field",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				eval := is.New(t)
				client := cfg.Client()

				testNS := ctx.Value(ctxNSKey).(string) //nolint:errcheck
				desiredApiDef := v1alpha1.ApiDefinition{
					ObjectMeta: metav1.ObjectMeta{Name: apiDefWithBlacklist, Namespace: testNS},
				}

				err := wait.For(conditions.New(client.Resources()).ResourceMatch(&desiredApiDef, func(object k8s.Object) bool {
					apiDef := object.(*v1alpha1.ApiDefinition) //nolint:errcheck
					return len(apiDef.Spec.VersionData.Versions[defaultVersion].ExtendedPaths.BlackList) == 1
				}), wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
				eval.NoErr(err)

				return ctx
			}).
		Assess("ApiDefniition should forbid traffic to blacklist route",
			func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
				eval := is.New(t)

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
				eval := is.New(t)

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
			eval := is.New(t)

			// Create ApiDefinition with whitelist + ignored extended path
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
			eval.NoErr(err) // failed to create apiDefinition

			return ctx
		}).
		Assess("ApiDefinition must have ignored field",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				eval := is.New(t)
				client := cfg.Client()

				testNS := ctx.Value(ctxNSKey).(string) //nolint:errcheck
				desiredApiDef := v1alpha1.ApiDefinition{
					ObjectMeta: metav1.ObjectMeta{Name: apiDefWithWhitelist, Namespace: testNS},
				}

				err := wait.For(conditions.New(client.Resources()).ResourceMatch(&desiredApiDef, func(object k8s.Object) bool {
					apiDef := object.(*v1alpha1.ApiDefinition) //nolint:errcheck
					return len(apiDef.Spec.VersionData.Versions[defaultVersion].ExtendedPaths.Ignored) == 1
				}), wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
				eval.NoErr(err)

				return ctx
			}).
		Assess("ApiDefniition should allow traffic to ignored route",
			func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
				eval := is.New(t)

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
				eval := is.New(t)

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
		apiDefPinning = "apidef-certificate-pinning"

		apiDefPinningViaSecret           = "apidef-with-secret"
		apiDefPinningViaSecretListenPath = "/secret"

		invalidApiDef           = "invalid-proxy-apidef"
		invalidApiDefListenPath = "/invalid"

		secretName = "secret"

		publicKeyID = "test-public-key-id"
		pubKeyPem   = []byte(`-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAyoiVyRRffvWcc/kQJj4h
tRPZeafPYAZzCgd9tQfQmtCJXWDAynL3slPfhAAuO0vWxyNHTTELbCHxD43nPgST
7HUwG/ZjsrY03+M4IaEiBvN53OsnJ5UekmH2G04HTZdsApoc9OSb+4aBGlkISsNx
n8SCRF3a8kn95tD27IToBozNXosKbyzKli/9g0rqmeQHqGHtLuuEsMZcWn9dXbKu
MB7n1c1e4XiBNgowQZGNWCU09fH56X/fN8QZ+OeP/1Fy5maOqjMReAAexkboS6yL
ZwOHzCPBGoxsay40cdI3pjz8UT8squMowlmZvhLNkOI1GkxucUXhFaaZoqmLrgh9
pwIDAQAB
-----END PUBLIC KEY-----`)
	)

	adCreate := features.New("Create ApiDefinition objects for Certificate Pinning").
		Setup(func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
			testNS := ctx.Value(ctxNSKey).(string) //nolint:errcheck
			client := envConf.Client()
			eval := is.New(t)

			// Create an ApiDefinition with Certificate Pinning using 'pinned_public_keys' field.
			// It contains a dummy public key which is not valid to be used.
			_, err := createTestAPIDef(ctx, testNS, func(apiDef *v1alpha1.ApiDefinition) {
				apiDef.Name = apiDefPinning
				apiDef.Spec.PinnedPublicKeys = map[string]string{"*": publicKeyID}
			}, envConf)
			eval.NoErr(err)

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
			eval.NoErr(err)

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
			eval.NoErr(err)

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
			eval.NoErr(err)

			return ctx
		}).
		Assess("ApiDefinition must have Certificate Pinning field defined",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				eval := is.New(t)
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
				eval.NoErr(err)

				return ctx
			}).
		Assess("Allow making requests based on the pinned public key defined via secret",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				eval := is.New(t)
				err := wait.For(func() (done bool, err error) {
					hc := &http.Client{}
					req, err := http.NewRequest(
						http.MethodGet,
						fmt.Sprintf("%s%s", gatewayLocalhost, apiDefPinningViaSecretListenPath),
						nil,
					)
					eval.NoErr(err)

					resp, err := hc.Do(req)
					eval.NoErr(err)

					if resp.StatusCode != 200 {
						t.Log("expected to access httpbin.org since it is pinned via public key.")
						return false, nil
					}

					return true, nil
				}, wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
				eval.NoErr(err)

				return ctx
			}).
		Assess("Prevent making requests to disallowed addresses based on the pinned public key defined via secret",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				eval := is.New(t)
				err := wait.For(func() (done bool, err error) {
					hc := &http.Client{}
					req, err := http.NewRequest(
						http.MethodGet,
						fmt.Sprintf("%s%s", gatewayLocalhost, invalidApiDefListenPath),
						nil,
					)
					eval.NoErr(err)

					resp, err := hc.Do(req)
					eval.NoErr(err)

					if resp.StatusCode == 200 {
						t.Log("unexpected access to invalid address")
						return false, nil
					}

					return true, nil
				}, wait.WithInterval(defaultWaitInterval), wait.WithTimeout(defaultWaitTimeout))
				eval.NoErr(err)

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
			eval := is.New(t)
			t.Log(testNS)
			_, err := createTestTlsSecret(ctx, testNS, nil, envConf)
			eval.NoErr(err)

			return ctx
		}).
		Setup(func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
			testNS := ctx.Value(ctxNSKey).(string) //nolint:errcheck
			eval := is.New(t)

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
			eval.NoErr(err) // failed to create apiDefinition

			return ctx
		}).
		Assess("ApiDefinition must have upstream field defined",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				eval := is.New(t)
				client := cfg.Client()
				testNS := ctx.Value(ctxNSKey).(string) //nolint:errcheck

				tlsSecret := v1.Secret{} //nolint:errcheck

				err := client.Resources(testNS).Get(ctx, certName, testNS, &tlsSecret)
				eval.NoErr(err)

				certPemBytes, ok := tlsSecret.Data["tls.crt"]
				if !ok {
					eval.Fail()
				}
				certFingerPrint, _ := cert.CalculateFingerPrint(certPemBytes)

				opConfSecret := v1.Secret{}
				err = client.Resources(opNs).Get(ctx, "tyk-operator-conf", opNs, &opConfSecret)
				eval.NoErr(err)

				tykAuth, ok := opConfSecret.Data["TYK_AUTH"]
				if !ok {
					eval.Fail()
				}

				tykOrg, ok := opConfSecret.Data["TYK_ORG"]
				if !ok {
					eval.Fail()
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

					eval.True(exists)

					return true, nil
				}, wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
				eval.NoErr(err)

				return ctx
			}).Feature()

	testenv.Test(t, adCreate)
}

func TestApiDefinitionBasicAuth(t *testing.T) {
	var (
		apiDefBasicAuth  = "apidef-basic-authentication"
		defaultVersion   = "Default"
		opNs             = "tyk-operator-system"
		tykConnectionURL = ""
		tykOrg           = ""
		tykAuth          = ""
	)

	mode := os.Getenv("TYK_MODE")

	switch mode {
	case "pro":
		tykConnectionURL = adminLocalhost
	case "ce":
		tykConnectionURL = gatewayLocalhost
	}

	testBasicAuth := features.New("Basic authentication").
		Setup(func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
			client := envConf.Client()
			eval := is.New(t)
			opConfSecret := v1.Secret{}

			err := client.Resources(opNs).Get(ctx, "tyk-operator-conf", opNs, &opConfSecret)
			eval.NoErr(err)

			data, ok := opConfSecret.Data["TYK_AUTH"]
			eval.True(ok)

			tykAuth = string(data)

			data, ok = opConfSecret.Data["TYK_ORG"]
			eval.True(ok)

			tykOrg = string(data)

			return ctx
		}).
		Setup(func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
			testNS := ctx.Value(ctxNSKey).(string) //nolint:errcheck
			eval := is.New(t)

			// Create ApiDefinition with Basic Authentication
			_, err := createTestAPIDef(ctx, testNS, func(apiDef *v1alpha1.ApiDefinition) {
				apiDef.Name = apiDefBasicAuth
				apiDef.Spec.UseBasicAuth = true
				apiDef.Spec.VersionData.DefaultVersion = defaultVersion
				apiDef.Spec.VersionData.NotVersioned = true
				apiDef.Spec.VersionData.Versions = map[string]model.VersionInfo{
					defaultVersion: {Name: defaultVersion},
				}
			}, envConf)
			eval.NoErr(err) // failed to create apiDefinition

			return ctx
		}).
		Assess("API must have basic authentication enabled",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				eval := is.New(t)
				client := cfg.Client()
				testNS := ctx.Value(ctxNSKey).(string) //nolint:errcheck

				var apiDef *model.APIDefinitionSpec

				err := wait.For(func() (done bool, err error) {
					env := environmet.Env{}
					env.Mode = v1alpha1.OperatorContextMode(mode)
					env.Org = tykOrg
					env.Auth = tykAuth
					env.URL = tykConnectionURL

					pkgContext := pkgclient.Context{
						Env: env,
						Log: log.NullLogger{},
					}

					reqContext := pkgclient.SetContext(context.Background(), pkgContext)

					// validate basic authentication field was set
					var apiDefCRD v1alpha1.ApiDefinition

					err = client.Resources().Get(ctx, apiDefBasicAuth, testNS, &apiDefCRD)
					if err != nil {
						return false, err
					}

					apiDef, err = klient.Universal.Api().Get(reqContext, apiDefCRD.Status.ApiID)
					if err != nil {
						return false, errors.New("API is not created yet")
					}

					eval.True(apiDef.UseBasicAuth)

					return true, nil
				}, wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
				eval.NoErr(err)

				eval.True(apiDef.UseBasicAuth)

				return ctx
			}).Feature()

	testenv.Test(t, testBasicAuth)
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
			eval := is.New(t)
			opConfSecret := v1.Secret{}

			err := client.Resources(opNs).Get(ctx, "tyk-operator-conf", opNs, &opConfSecret)
			eval.NoErr(err)

			data, ok := opConfSecret.Data["TYK_AUTH"]
			eval.True(ok)

			tykAuth = string(data)

			data, ok = opConfSecret.Data["TYK_ORG"]
			eval.True(ok)

			tykOrg = string(data)

			return ctx
		}).
		Setup(func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
			testNS := ctx.Value(ctxNSKey).(string) //nolint: errcheck
			eval := is.New(t)
			t.Log(testNS)

			_, err := createTestTlsSecret(ctx, testNS, nil, envConf)
			eval.NoErr(err)

			return ctx
		}).
		Setup(func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
			testNS := ctx.Value(ctxNSKey).(string) //nolint:errcheck
			eval := is.New(t)

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
			eval.NoErr(err) // failed to create apiDefinition

			return ctx
		}).Assess("Certificate from secret must be uploaded",
		func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			eval := is.New(t)
			client := cfg.Client()
			testNS := ctx.Value(ctxNSKey).(string) //nolint:errcheck

			tlsSecret := v1.Secret{} //nolint:errcheck

			err := client.Resources(testNS).Get(ctx, certName, testNS, &tlsSecret)
			eval.NoErr(err)

			certPemBytes, ok := tlsSecret.Data["tls.crt"]
			eval.True(ok)

			certFingerPrint, _ := cert.CalculateFingerPrint(certPemBytes)
			calculatedCertID := tykOrg + certFingerPrint

			err = wait.For(func() (done bool, err error) {
				env := environmet.Env{}
				env.Mode = v1alpha1.OperatorContextMode(mode)
				env.Org = tykOrg
				env.Auth = tykAuth
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
			eval.NoErr(err)

			ctx = context.WithValue(ctx, certIDCtxKey, calculatedCertID)

			return ctx
		}).
		Assess("API must have client certificate field defined",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				eval := is.New(t)
				client := cfg.Client()
				testNS := ctx.Value(ctxNSKey).(string) //nolint:errcheck

				certID := ctx.Value(certIDCtxKey)

				var apiDef *model.APIDefinitionSpec

				err := wait.For(func() (done bool, err error) {
					env := environmet.Env{}
					env.Mode = v1alpha1.OperatorContextMode(mode)
					env.Org = tykOrg
					env.Auth = tykAuth
					env.URL = tykConnectionURL

					pkgContext := pkgclient.Context{
						Env: env,
						Log: log.NullLogger{},
					}

					reqContext := pkgclient.SetContext(context.Background(), pkgContext)

					// validate client certificate field was set
					var apiDefCRD v1alpha1.ApiDefinition

					err = client.Resources().Get(ctx, apiDefClientMTLSWithCert, testNS, &apiDefCRD)
					if err != nil {
						return false, err
					}

					apiDef, err = klient.Universal.Api().Get(reqContext, apiDefCRD.Status.ApiID)
					if err != nil {
						return false, errors.New("API is not created yet")
					}

					eval.True(len(apiDef.ClientCertificates) == 1)
					eval.True(apiDef.ClientCertificates[0] == certID)

					return true, nil
				}, wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
				eval.NoErr(err)

				eval.True(len(apiDef.ClientCertificates) == 1)
				eval.True(apiDef.ClientCertificates[0] == certID)

				return ctx
			}).Feature()

	testWithoutCert := features.New("Client MTLS without certs").
		Setup(func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
			testNS := ctx.Value(ctxNSKey).(string) //nolint:errcheck
			eval := is.New(t)

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
			eval.NoErr(err) // failed to create apiDefinition

			return ctx
		}).
		Assess("API should be created even though certs doesn't exists",
			func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				eval := is.New(t)
				client := cfg.Client()
				testNS := ctx.Value(ctxNSKey).(string) //nolint:errcheck

				opConfSecret := v1.Secret{}
				err := client.Resources(opNs).Get(ctx, "tyk-operator-conf", opNs, &opConfSecret)
				eval.NoErr(err)

				tykAuth, ok := opConfSecret.Data["TYK_AUTH"]
				eval.True(ok)

				tykOrg, ok := opConfSecret.Data["TYK_ORG"]
				eval.True(ok)

				var apiDef *model.APIDefinitionSpec
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
					if err != nil {
						return false, err
					}

					apiDef, err = klient.Universal.Api().Get(reqContext, apiDefCRD.Status.ApiID)
					if err != nil {
						return false, err
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
	createAPI := features.New("Create GraphQL API").
		Assess("validate_executionMode", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			testNS := ctx.Value(ctxNSKey).(string) //nolint:errcheck
			eval := is.New(t)

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
					_, err := createTestAPIDef(ctx, testNS, func(ad *v1alpha1.ApiDefinition) {
						ad.Name = fmt.Sprintf("%s-%s", ad.Name, uuid.New().String())
						ad.Spec.Name = ad.Name
						ad.Spec.GraphQL = &model.GraphQLConfig{
							Enabled:       true,
							ExecutionMode: model.GraphQLExecutionMode(tc.ExecutionMode),
						}
					}, c)

					t.Log("Error=", err)
					eval.Equal(tc.ReturnErr, err != nil)
				})
			}

			return ctx
		}).Feature()

	testenv.Test(t, createAPI)
}

func TestApiDefinitionSubGraphExecutionMode(t *testing.T) {
	const (
		opNs = "tyk-operator-system"

		supportedMajorTykVersion = uint(4)
	)

	tykEnv := environmet.Env{}
	majorTykVersion := supportedMajorTykVersion
	r := &controllers.ApiDefinitionReconciler{}

	gqlSubGraph := features.New("GraphQL SubGraph Execution mode").
		Setup(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			eval := is.New(t)
			opConfSecret := v1.Secret{}

			err := c.Client().Resources(opNs).Get(ctx, "tyk-operator-conf", opNs, &opConfSecret)
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
			r = &controllers.ApiDefinitionReconciler{
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
					definition.Spec.GraphQL = &model.GraphQLConfig{
						GraphRef:      testSubGraphCRMetaName,
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

						return apiDefObj.Spec.GraphQL != nil &&
							apiDefObj.Spec.GraphQL.GraphRef == testSubGraphCRMetaName &&
							apiDefObj.Spec.GraphQL.Schema == testSubGraphSchema &&
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
					definition.ObjectMeta = metav1.ObjectMeta{
						Name: "another-api", Namespace: testNs,
					}
					definition.Spec.GraphQL = &model.GraphQLConfig{
						GraphRef:      testSubGraphCRMetaName,
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
					api.Spec.GraphQL.GraphRef = newSgName
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

						return apiDefObj.Spec.GraphQL != nil &&
							apiDefObj.Spec.GraphQL.GraphRef == newSgName &&
							apiDefObj.Spec.GraphQL.Schema == newSchema &&
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
					api.Spec.GraphQL.GraphRef = ""
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

						return apiDefObj.Spec.GraphQL != nil &&
							apiDefObj.Spec.GraphQL.GraphRef == "" &&
							apiDefObj.Spec.GraphQL.Schema == newSchema &&
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
