package integration

import (
	"context"
	"testing"

	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/controllers"
	tykClient "github.com/TykTechnologies/tyk-operator/pkg/client"
	"github.com/TykTechnologies/tyk-operator/pkg/client/klient"
	"github.com/TykTechnologies/tyk-operator/pkg/environment"
	"github.com/go-logr/logr"
	"github.com/matryer/is"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
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
				Log: logr.Discard(),
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
				Log:    logr.Discard(),
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
