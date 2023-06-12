package integration

import (
	"context"
	"testing"

	"k8s.io/apimachinery/pkg/types"

	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/controllers"
	tykClient "github.com/TykTechnologies/tyk-operator/pkg/client"
	"github.com/TykTechnologies/tyk-operator/pkg/client/klient"
	"github.com/TykTechnologies/tyk-operator/pkg/environmet"

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

		tykEnv environmet.Env
		tykCtx context.Context
	)

	f := features.New("test certificate upload").
		Setup(func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
			eval := is.New(t)
			var err error

			// Obtain Environment configuration to be able to connect Tyk.
			tykEnv, err = generateEnvConfig(ctx, envConf)
			eval.NoErr(err)

			tykCtx = tykClient.SetContext(context.Background(), tykClient.Context{
				Env: tykEnv,
				Log: log.NullLogger{},
			})

			// Create a secret
			testNS, _ = ctx.Value(ctxNSKey).(string)

			_, err = createTestTlsSecret(ctx, testNS, envConf, func(s *v1.Secret) {
				s.Name = tlsSecretName
			})
			eval.NoErr(err)

			apiDef, err = createTestAPIDef(ctx, envConf, testNS, func(api *v1alpha1.ApiDefinition) {
				api.Name = apiDefName
				api.Spec.UpstreamCertificateRefs = map[string]string{"*": tlsSecretName}
			})
			eval.NoErr(err)

			waitForTykResourceCreation(envConf, apiDef)

			return ctx
		}).Assess("validate certificate is uploaded", func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
		eval := is.New(t)

		// Validate the certificate is uploaded
		err := wait.For(func() (done bool, err error) {
			tykApi, err := klient.Universal.Api().Get(tykCtx, apiDef.Status.ApiID)
			if err != nil {
				return false, err
			}

			if tykApi.UpstreamCertificates == nil {
				return false, nil
			}

			if len(tykApi.UpstreamCertificates) == 0 {
				return false, nil
			}

			return true, nil
		}, wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))

		eval.NoErr(err)

		return ctx
	}).Assess("validate certificate is successfully reconciled on restart of operator",
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

			_, err = r.Reconcile(tykCtx, ctrl.Request{NamespacedName: types.NamespacedName{Name: tlsSecretName, Namespace: testNS}})
			eval.NoErr(err)

			return ctx
		}).Feature()

	testenv.Test(t, f)
}
