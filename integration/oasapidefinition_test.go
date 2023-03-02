package integration

import (
	"context"
	"testing"

	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/controllers"
	tykClient "github.com/TykTechnologies/tyk-operator/pkg/client"
	"github.com/TykTechnologies/tyk-operator/pkg/client/klient"
	"github.com/TykTechnologies/tyk-operator/pkg/environmet"
	"github.com/matryer/is"
	v1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	cr "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

// TestDeletingNonexistentOasApi tests if deleting nonexistent resources cause an error or not on k8s level.
// Assume that the user deleted OasApiDefinition resource from Tyk instead of deleting it via kubectl.
// This will create a drift between Tyk and K8s. Deleting the same resource from k8s shouldn't cause any
// external API errors such as 404.
func TestDeletingNonexistentOasApi(t *testing.T) {
	var (
		eval   = is.New(t)
		tykCtx context.Context
		tykEnv environmet.Env
	)

	testDeletingNonexistentAPIs := features.New("Deleting Nonexistent OAS APIs from k8s").
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
		Assess("Delete nonexistent OasApiDefinition from k8s successfully",
			func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
				// To begin with, delete the OasApiDefinition from Tyk, which is the wrong thing to do because it'll
				// cause a drift between k8s and Tyk. Now, deleting ApiDefinition CR from k8s,
				// `kubectl delete tykapis <resource_name>`, must be handled gracefully.
				testNs, ok := ctx.Value(ctxNSKey).(string)
				eval.True(ok)

				// First, create the OasApiDefinition
				oasApiDefCR, err := createTestOASApiDef(ctx, c, testNs, func(apiDef *v1alpha1.TykOASApiDefinition) {})
				eval.NoErr(err)

				err = waitForTykResourceCreation(c, oasApiDefCR)
				eval.NoErr(err)

				err = deleteApiDefinitionFromTyk(tykCtx, oasApiDefCR.Status.ApiID)
				eval.NoErr(err)

				err = c.Client().Resources(testNs).Delete(ctx, oasApiDefCR)
				eval.NoErr(err)

				err = wait.For(
					conditions.New(c.Client().Resources()).ResourceDeleted(oasApiDefCR),
					wait.WithTimeout(defaultWaitTimeout),
					wait.WithInterval(defaultWaitInterval),
				)
				eval.NoErr(err)

				return ctx
			}).Feature()

	testenv.Test(t, testDeletingNonexistentAPIs)
}

// TestReconcileNonexistentOasApi tests whether reconciliation finishes successfully if OasApiDefinition
// does not exist on Tyk. Reconciliation logic must handle API calls to Tyk Gateway / Dashboard for
// nonexistent OasApiDefinitions and create it if needed. So that, the k8s remains as source of truth.
func TestReconcileNonexistentOasApi(t *testing.T) {
	var (
		eval   = is.New(t)
		tykCtx context.Context
		tykEnv environmet.Env
		r      controllers.TykOASApiDefinitionReconciler
	)

	testReconcilingNonexistentAPIs := features.New("Reconciling Nonexistent OasApiDefinition CRs").
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

			r = controllers.TykOASApiDefinitionReconciler{
				Client: cl,
				Log:    log.NullLogger{},
				Scheme: cl.Scheme(),
				Env:    tykEnv,
			}

			return ctx
		}).
		Assess("Create a drift between Tyk and k8s by deleting an OasApiDefinition from Tyk",
			func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
				// To begin with, we should create a drift between Tyk and K8s. In order to do that
				// first create an ApiDefinition, then delete it from Tyk via Tyk API calls. The next
				// reconciliation request must understand nonexistent entity and create it from scratch.
				testNs, ok := ctx.Value(ctxNSKey).(string)
				eval.True(ok)

				// First, create the ApiDefinition.
				apiDefCR, err := createTestOASApiDef(ctx, c, testNs, nil)
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
				// must understand the change between Tyk and K8s and create nonexistent OasApiDefinition again based
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
