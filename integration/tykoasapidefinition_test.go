package integration

import (
	"context"
	"fmt"
	"testing"

	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/controllers"
	tykClient "github.com/TykTechnologies/tyk-operator/pkg/client"
	"github.com/TykTechnologies/tyk-operator/pkg/client/klient"
	"github.com/TykTechnologies/tyk-operator/pkg/environment"
	retry "github.com/avast/retry-go"
	"github.com/buger/jsonparser"
	"github.com/matryer/is"
	"golang.org/x/mod/semver"
	v1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	cr "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestOASCreate(t *testing.T) {
	var (
		testNS string
		tykOAS *v1alpha1.TykOasApiDefinition
		eval   = is.New(t)
		tykCtx context.Context
	)

	f := features.New("Create Tyk OAS API").
		Setup(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			var err error

			// Obtain Environment configuration to be able to connect Tyk.
			tykEnv, err := generateEnvConfig(ctx, c)
			eval.NoErr(err)

			res := semver.Compare(tykEnv.TykVersion, "v4.1")
			if res < 0 {
				t.Skip("OAS support is added in Tyk in v4.1")
			}

			tykCtx = tykClient.SetContext(context.Background(), tykClient.Context{
				Env: tykEnv,
				Log: log.NullLogger{},
			})

			var ok bool

			testNS, ok = ctx.Value(ctxNSKey).(string)
			eval.True(ok)

			tykOAS, _, err = createTestOASApi(ctx, testNS, c, "", nil)
			eval.NoErr(err)

			err = waitForTykResourceCreation(c, tykOAS)
			eval.NoErr(err)

			return ctx
		}).Assess("Ensure OAS definition is created on Tyk",
		func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			err := wait.For(
				conditions.New(c.Client().Resources(testNS)).ResourceMatch(tykOAS, func(object k8s.Object) bool {
					currTykOas, ok := object.(*v1alpha1.TykOasApiDefinition)
					eval.True(ok)

					t.Logf("looking for %+v", currTykOas.Status)
					return klient.Universal.TykOAS().Exists(tykCtx, currTykOas.Status.ID) == true
				}))
			eval.NoErr(err)

			return ctx
		}).Feature()

	testenv.Test(t, f)
}

func TestInvalidTykOAS(t *testing.T) {
	var (
		testNS string
		eval   = is.New(t)
	)

	f := features.New("Create invalid Tyk OAS API").
		Setup(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			// Obtain Environment configuration to be able to connect Tyk.
			tykEnv, err := generateEnvConfig(ctx, c)
			eval.NoErr(err)

			res := semver.Compare(tykEnv.TykVersion, "v4.1")
			if res < 0 {
				t.Skip("OAS support is added in Tyk in v4.1")
			}

			var ok bool

			testNS, ok = ctx.Value(ctxNSKey).(string)
			eval.True(ok)

			return ctx
		}).Assess("Ensure Tyk OAS CR creation fails",
		func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			invalidOASDoc := "{}"

			tykOAS, _, err := createTestOASApi(ctx, testNS, c, invalidOASDoc, nil)
			eval.NoErr(err)

			err = wait.For(conditions.New(c.Client().Resources(testNS)).ResourceMatch(tykOAS, func(object k8s.Object) bool {
				var ok bool
				tykOAS, ok := object.(*v1alpha1.TykOasApiDefinition)
				eval.True(ok)

				return tykOAS.Status.LatestTransaction.Status == v1alpha1.Failed
			}))
			eval.NoErr(err)

			return ctx
		}).Feature()

	testenv.Test(t, f)
}

func TestOASDelete(t *testing.T) {
	var (
		testNS string
		tykOAS *v1alpha1.TykOasApiDefinition
		eval   = is.New(t)
		tykCtx context.Context
	)

	f := features.New("Test OAS Delete").
		Setup(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			// Obtain Environment configuration to be able to connect Tyk.
			tykEnv, err := generateEnvConfig(ctx, c)
			eval.NoErr(err)

			res := semver.Compare(tykEnv.TykVersion, "v4.1")
			if res < 0 {
				t.Skip("OAS support is added in Tyk in v4.1")
			}

			tykCtx = tykClient.SetContext(context.Background(), tykClient.Context{
				Env: tykEnv,
				Log: log.NullLogger{},
			})

			var ok bool

			testNS, ok = ctx.Value(ctxNSKey).(string)
			eval.True(ok)

			tykOAS, _, err = createTestOASApi(ctx, testNS, c, "", nil)
			eval.NoErr(err)

			err = waitForTykResourceCreation(c, tykOAS)
			eval.NoErr(err)

			return ctx
		}).Assess("Deleting Tyk OAS CRD deletes Tyk OAS API from Tyk",
		func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			apiID := tykOAS.Status.ID

			err := c.Client().Resources(testNS).Delete(ctx, tykOAS)
			eval.NoErr(err)

			err = wait.For(conditions.New(c.Client().Resources(testNS)).ResourceDeleted(tykOAS),
				wait.WithInterval(defaultWaitInterval), wait.WithTimeout(defaultWaitTimeout))
			eval.NoErr(err)

			exists := klient.Universal.TykOAS().Exists(tykCtx, apiID)
			eval.True(!exists)

			return ctx
		}).Feature()

	testenv.Test(t, f)
}

func TestOASUpdate(t *testing.T) {
	var (
		testNS string
		tykOAS *v1alpha1.TykOasApiDefinition
		tykCtx context.Context
		cm     *v1.ConfigMap
		tykEnv environment.Env
	)

	eval := is.New(t)

	f := features.New("Test OAS Update").
		Setup(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			var err error

			// Obtain Environment configuration to be able to connect Tyk.
			tykEnv, err = generateEnvConfig(ctx, c)
			eval.NoErr(err)

			res := semver.Compare(tykEnv.TykVersion, "v4.1")
			if res < 0 {
				t.Skip("OAS support is added in Tyk in v4.1")
			}

			tykCtx = tykClient.SetContext(context.Background(), tykClient.Context{
				Env: tykEnv,
				Log: log.NullLogger{},
			})

			var ok bool

			testNS, ok = ctx.Value(ctxNSKey).(string)
			eval.True(ok)

			tykOAS, cm, err = createTestOASApi(ctx, testNS, c, "", nil)
			eval.NoErr(err)

			err = waitForTykResourceCreation(c, tykOAS)
			eval.NoErr(err)

			return ctx
		}).Assess("Updating Configmap updates Tyk OAS API",
		func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			newListenPath := "/petstore-update"
			quotedListenPath := fmt.Sprintf("\"%s\"", newListenPath)
			oasDoc := cm.Data[testOASConfKeyName]

			result, err := jsonparser.Set([]byte(oasDoc), []byte(quotedListenPath),
				controllers.TykOASExtenstionStr, "server", "listenPath", "value")
			eval.NoErr(err)

			cm.Data[testOASConfKeyName] = string(result)

			err = c.Client().Resources(testNS).Update(ctx, cm)
			eval.NoErr(err)

			// Create ApiDefinition Reconciler.
			cl, err := createTestClient(c.Client())
			eval.NoErr(err)

			r := controllers.TykOasApiDefinitionReconciler{
				Client: cl,
				Log:    log.NullLogger{},
				Scheme: cl.Scheme(),
				Env:    tykEnv,
			}

			err = wait.For(conditions.New(c.Client().Resources(testNS)).ResourceMatch(tykOAS, func(object k8s.Object) bool {
				_, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: cr.ObjectKeyFromObject(tykOAS)})
				return err == nil
			}))
			eval.NoErr(err)

			var tykOASApi string

			err = retry.Do(func() error {
				tykOASApi, err = klient.Universal.TykOAS().Get(tykCtx, tykOAS.Status.ID)
				return err
			})
			eval.NoErr(err)

			var targetURL string
			targetURL, err = jsonparser.GetString([]byte(tykOASApi),
				controllers.TykOASExtenstionStr, "server", "listenPath", "value")

			eval.NoErr(err)
			eval.Equal(targetURL, newListenPath)

			return ctx
		}).Feature()

	testenv.Test(t, f)
}
