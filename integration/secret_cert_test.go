package integration

import (
	"context"
	"testing"

	"golang.org/x/mod/semver"
	"k8s.io/apimachinery/pkg/types"

	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/controllers"
	tykClient "github.com/TykTechnologies/tyk-operator/pkg/client"
	"github.com/TykTechnologies/tyk-operator/pkg/client/klient"
	"github.com/TykTechnologies/tyk-operator/pkg/environment"
	"github.com/buger/jsonparser"

	"github.com/matryer/is"
	v1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestCertificateUpload(t *testing.T) {
	var (
		tlsSecretName = "test-cert-upload-secret"
		apiDefName    = "test-cert-upload-api"
		testNS        string

		apiDef *v1alpha1.ApiDefinition

		tykEnv environment.Env
		tykCtx context.Context
	)

	f := features.New("test certificate upload").
		Setup(func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
			eval := is.New(t)
			var err error

			testNS, _ = ctx.Value(ctxNSKey).(string) // nolint: errcheck

			// Obtain Environment configuration to be able to connect Tyk.
			tykEnv, err = generateEnvConfig(ctx, envConf)
			eval.NoErr(err)

			tykCtx = tykClient.SetContext(context.Background(), tykClient.Context{
				Env: tykEnv,
				Log: log.NullLogger{},
			})

			// Create TLS secret
			_, err = createTestTlsSecret(ctx, testNS, envConf, func(s *v1.Secret) {
				s.Name = tlsSecretName
			})
			eval.NoErr(err)

			// Create test api definition
			apiDef, err = createTestAPIDef(ctx, envConf, testNS, func(api *v1alpha1.ApiDefinition) {
				api.Name = apiDefName
				api.Spec.UpstreamCertificateRefs = map[string]string{"*": tlsSecretName}
			})
			eval.NoErr(err)

			err = waitForTykResourceCreation(envConf, apiDef)
			eval.NoErr(err)

			return ctx
		}).Assess("validate certificate is uploaded",
		func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
			eval := is.New(t)

			// Validate the certificate is uploaded
			err := wait.For(func() (done bool, err error) {
				tykApi, err := klient.Universal.Api().Get(tykCtx, apiDef.Status.ApiID)
				if err != nil {
					t.Log("Failed to fetch Tyk API", "id", apiDef.Status.ApiID, "error", err)

					return false, err
				}

				if tykApi.UpstreamCertificates == nil || len(tykApi.UpstreamCertificates) == 0 {
					t.Log("Upstream certificate is not uploaded yet. It is still empty/nil")

					return false, nil
				}

				return true, nil
			}, wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))

			eval.NoErr(err)

			return ctx
		}).Assess("secret is successfully reconciled on next run",
		func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
			eval := is.New(t)

			cl, err := createTestClient(envConf.Client())
			eval.NoErr(err)

			r := controllers.SecretCertReconciler{
				Client: cl,
				Log:    log.NullLogger{},
				Scheme: cl.Scheme(),
				Env:    tykEnv,
			}

			err = wait.For(func() (done bool, err error) {
				_, err = r.Reconcile(tykCtx, ctrl.Request{
					NamespacedName: types.NamespacedName{
						Name:      tlsSecretName,
						Namespace: testNS,
					},
				})

				if err != nil {
					t.Log("Failed to reconcile secret", "name", tlsSecretName, "namespace", testNS, "error", err)
					return false, nil
				}

				return true, nil
			})
			eval.NoErr(err)

			return ctx
		}).Feature()

	testenv.Test(t, f)
}

func TestCertificateUploadForOAS(t *testing.T) {
	var (
		tlsSecretName = "test-oas-cert-upload-secret"

		testNS string
		tykOas *v1alpha1.TykOasApiDefinition
		tykEnv environment.Env
		tykCtx context.Context
	)
	f := features.New("test certificate upload").
		Setup(func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
			eval := is.New(t)
			var err error

			testNS, _ = ctx.Value(ctxNSKey).(string) // nolint: errcheck

			// Obtain Environment configuration to be able to connect Tyk.
			tykEnv, err = generateEnvConfig(ctx, envConf)
			eval.NoErr(err)

			res := semver.Compare(tykEnv.TykVersion, "v5.3")
			if res < 0 {
				t.Skip("OAS support is added in Tyk in v5.3")
			}

			tykCtx = tykClient.SetContext(context.Background(), tykClient.Context{
				Env: tykEnv,
				Log: log.NullLogger{},
			})

			// Create TLS secret
			_, err = createTestTlsSecret(ctx, testNS, envConf, func(s *v1.Secret) {
				s.Name = tlsSecretName
			})
			eval.NoErr(err)

			tykOASDoc := `{
				"info": {
				  "title": "Petstore",
				  "version": "1.0.0"
				},
				"openapi": "3.0.3",
				"components": {},
				"paths": {},
				"x-tyk-api-gateway": {
				  "info": {
					"name": "Petstore",
					"state": {
					  "active": true
					}
				  },
				  "upstream": {
					"url": "https://petstore.swagger.io/v2"
				  },
				  "server": {
					"listenPath": {
					  "value": "/petstore/",
					  "strip": true
					},
					"clientCertificates":{
					  "enabled": false
					}
				  }
				}
			  }`

			// Create test TykOasApiDefinition
			tykOas, _, err = createTestOASApi(ctx, testNS, testOASCmName, envConf, tykOASDoc, nil,
				func(oas *v1alpha1.TykOasApiDefinition) *v1alpha1.TykOasApiDefinition {
					oas.Spec.ClientCertificate.Enabled = true
					oas.Spec.ClientCertificate.Allowlist = []string{tlsSecretName}

					return oas
				})
			eval.NoErr(err)

			err = waitForTykResourceCreation(envConf, tykOas)
			eval.NoErr(err)

			return ctx
		}).Assess("validate certificate is uploaded",
		func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
			eval := is.New(t)

			// Validate the certificate is uploaded
			err := wait.For(func() (done bool, err error) {
				tykOASApi, err := klient.Universal.TykOAS().Get(tykCtx, tykOas.Status.ID)
				if err != nil {
					t.Log("Failed to fetch Tyk OAS API", "id", tykOas.Status.ID, "error", err)

					return false, nil
				}

				val, _, _, err := jsonparser.Get([]byte(tykOASApi), controllers.OASClientCertAllowlistPath...)
				if err != nil {
					return false, err
				}

				allowList := []string{}
				_, err = jsonparser.ArrayEach(val, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
					val, pErr := jsonparser.ParseString(value)
					if pErr != nil {
						return
					}

					allowList = append(allowList, val)
				})
				if err != nil {
					return false, err
				}

				if len(allowList) == 0 {
					t.Log("Client certificate is not uploaded yet. It is still empty/nil")

					return false, nil
				}

				return true, nil
			}, wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))

			eval.NoErr(err)

			return ctx
		}).Feature()

	testenv.Test(t, f)
}
