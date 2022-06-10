package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	common2 "github.com/TykTechnologies/tyk-operator/integration/internal/common"

	"github.com/TykTechnologies/tyk-operator/api/model"
	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/pkg/cert"
	"github.com/matryer/is"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestApiDefinitionCertificatePinning(t *testing.T) {
	if strings.TrimSpace(os.Getenv("TYK_MODE")) == "ce" {
		t.Log("CE is not feasible to test at the moment.")
		return
	}

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
			testNS := ctx.Value(common2.CtxNSKey).(string) //nolint:errcheck
			client := envConf.Client()
			is := is.New(t)

			// Create an ApiDefinition with Certificate Pinning using 'pinned_public_keys' field.
			// It contains a dummy public key which is not valid to be used.
			_, err := common2.CreateTestAPIDef(ctx, testNS, func(apiDef *v1alpha1.ApiDefinition) {
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
			_, err = common2.CreateTestAPIDef(ctx, testNS, func(apiDef *v1alpha1.ApiDefinition) {
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
			_, err = common2.CreateTestAPIDef(ctx, testNS, func(apiDef *v1alpha1.ApiDefinition) {
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

				testNS := ctx.Value(common2.CtxNSKey).(string) //nolint:errcheck
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
				}), wait.WithTimeout(common2.DefaultWaitTimeout), wait.WithInterval(common2.DefaultWaitInterval))
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
						fmt.Sprintf("%s%s", common2.GatewayLocalhost, apiDefPinningViaSecretListenPath),
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
				}, wait.WithTimeout(common2.DefaultWaitTimeout), wait.WithInterval(common2.DefaultWaitInterval))
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
						fmt.Sprintf("%s%s", common2.GatewayLocalhost, invalidApiDefListenPath),
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
				}, wait.WithInterval(common2.DefaultWaitInterval), wait.WithTimeout(common2.DefaultWaitTimeout))
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
		dashboardLocalHost  = "http://localhost:7200"
	)

	if strings.TrimSpace(os.Getenv("TYK_MODE")) == "ce" {
		t.Log("CE is not feasible to test at the moment.")
		return
	}

	adCreate := features.New("Create an ApiDefinition for Upstream TLS").
		Setup(func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
			testNS := ctx.Value(common2.CtxNSKey).(string) //nolint: errcheck
			is := is.New(t)
			t.Log(testNS)
			_, err := common2.CreateTestTlsSecret(ctx, testNS, nil, envConf)
			is.NoErr(err)

			return ctx
		}).
		Setup(func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
			testNS := ctx.Value(common2.CtxNSKey).(string) //nolint:errcheck
			is := is.New(t)

			// Create ApiDefinition with Upstream certificate
			_, err := common2.CreateTestAPIDef(ctx, testNS, func(apiDef *v1alpha1.ApiDefinition) {
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
				testNS := ctx.Value(common2.CtxNSKey).(string) //nolint:errcheck

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

					if len(certResponse.Certs) < 1 {
						return false, nil
					}
					return true, nil
				}, wait.WithTimeout(common2.DefaultWaitTimeout), wait.WithInterval(common2.DefaultWaitInterval))
				is.NoErr(err)

				return ctx
			}).Feature()

	testenv.Test(t, adCreate)
}
