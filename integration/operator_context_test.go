package integration

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/TykTechnologies/tyk-operator/api/model"
	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/matryer/is"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestOperatorContextCreate(t *testing.T) {
	listenPath := "/test-opctx"

	opCreate := features.New("Operator Context").
		Setup(func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
			testNS := ctx.Value(ctxNSKey).(string)
			eval := is.New(t)

			// create operator context
			opCtx, err := createTestOperatorContext(ctx, testNS, envConf, nil)
			eval.NoErr(err) // failed to create operatorcontext

			// create api definition
			_, err = createTestAPIDef(ctx, envConf, testNS, func(apiDef *v1alpha1.ApiDefinition) {
				apiDef.Spec.Context = &model.Target{
					Name:      opCtx.Name,
					Namespace: opCtx.Namespace,
				}
				apiDef.Spec.Proxy.ListenPath = listenPath
			})
			eval.NoErr(err) // failed to create apiDefinition

			// create api definition with empty namespace for contextRef
			_, err = createTestAPIDef(ctx, envConf, testNS, func(apiDef *v1alpha1.ApiDefinition) {
				apiDef.Name = "empty-ns"
				apiDef.Spec.Context = &model.Target{Name: opCtx.Name}
				apiDef.Spec.Proxy.ListenPath = "/empty-ns"
			})

			eval.NoErr(err) // failed to create apiDefinition

			return ctx
		}).
		Assess("context status is updated",
			func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
				testNS := ctx.Value(ctxNSKey).(string) //nolint:errcheck
				client := envConf.Client()
				eval := is.New(t)

				opCtx := v1alpha1.OperatorContext{ObjectMeta: metav1.ObjectMeta{Name: testOperatorCtx, Namespace: testNS}}

				err := wait.For(conditions.New(client.Resources()).ResourceMatch(&opCtx, func(object k8s.Object) bool {
					operatCtx := object.(*v1alpha1.OperatorContext) //nolint:errcheck

					// only one apidef will get linked
					// other one has empty namespace
					if len(operatCtx.Status.LinkedApiDefinitions) != 1 {
						return false
					}

					if operatCtx.Status.LinkedApiDefinitions[0].Namespace != testNS ||
						operatCtx.Status.LinkedApiDefinitions[0].Name != testApiDef {
						return false
					}

					return true
				}), wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
				eval.NoErr(err)

				return ctx
			}).Assess("apidef was created in dashboard",
		func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
			eval := is.New(t)

			err := wait.For(func() (done bool, err error) {
				resp, getErr := http.Get(fmt.Sprintf("%s/%s/get", gatewayLocalhost, listenPath))
				if getErr != nil {
					t.Log(getErr)
					return false, nil
				}

				if resp.StatusCode != http.StatusOK {
					t.Logf("API is not created yet. Got response code %d", resp.StatusCode)
					return false, nil
				}

				return true, nil
			}, wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
			eval.NoErr(err)

			return ctx
		}).Assess("apidef with empty namespace in contextRef was not created in dashboard",
		func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
			eval := is.New(t)

			err := wait.For(func() (done bool, err error) {
				resp, getErr := http.Get(fmt.Sprintf("%s/empty-ns/get", gatewayLocalhost))
				if getErr != nil {
					t.Log(getErr)
					return false, nil
				}

				if resp.StatusCode != 404 {
					t.Log("API definition should not be created on dashboard")
					return false, nil
				}

				return true, nil
			}, wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))

			eval.NoErr(err)

			return ctx
		}).Feature()

	testenv.Test(t, opCreate)
}

func TestOperatorContextDelete(t *testing.T) {
	delApiDef := features.New("Api Definition Delete").
		Setup(func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
			testNS := ctx.Value(ctxNSKey).(string)
			eval := is.New(t)
			client := envConf.Client()

			operatorCtx, err := createTestOperatorContext(ctx, testNS, envConf, nil)
			eval.NoErr(err) // failed to create operatorcontext

			ctx = context.WithValue(ctx, ctxOpCtxName, operatorCtx.Name)

			// create api definition
			def, err := createTestAPIDef(ctx, envConf, testNS, func(apiDef *v1alpha1.ApiDefinition) {
				apiDef.Spec.Context = &model.Target{
					Name:      operatorCtx.Name,
					Namespace: operatorCtx.Namespace,
				}
			})
			eval.NoErr(err) // failed to create apiDefinition

			ctx = context.WithValue(ctx, ctxApiName, def.Name)

			err = wait.For(conditions.New(client.Resources()).ResourceMatch(operatorCtx, func(object k8s.Object) bool {
				opCtx := object.(*v1alpha1.OperatorContext) //nolint:errcheck

				if len(opCtx.Status.LinkedApiDefinitions) == 0 {
					t.Log(opCtx)
					return false
				}

				return true
			}), wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))

			eval.NoErr(err)

			return ctx
		}).
		Assess("context ref should not get deleted while it is still been refered",
			func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
				testNS := ctx.Value(ctxNSKey).(string)        //nolint:errcheck
				opCtxName := ctx.Value(ctxOpCtxName).(string) //nolint:errcheck

				client := envConf.Client()
				eval := is.New(t)
				opCtx := v1alpha1.OperatorContext{ObjectMeta: metav1.ObjectMeta{Name: opCtxName, Namespace: testNS}}

				err := client.Resources().Delete(ctx, &opCtx)
				eval.NoErr(err)

				err = client.Resources(testNS).Get(ctx, opCtxName, testNS, &opCtx)
				eval.NoErr(err)

				return ctx
			}).
		Assess("delete api def should delete operator context",
			func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
				eval := is.New(t)
				testNS := ctx.Value(ctxNSKey).(string)        //nolint:errcheck
				opCtxName := ctx.Value(ctxOpCtxName).(string) //nolint:errcheck
				apiDefName := ctx.Value(ctxApiName).(string)  //nolint:errcheck

				apiDef := v1alpha1.ApiDefinition{ObjectMeta: metav1.ObjectMeta{Name: apiDefName, Namespace: testNS}}

				err := wait.For(
					conditions.New(c.Client().Resources()).ResourceMatch(&apiDef, func(object k8s.Object) bool {
						apiObj, ok := object.(*v1alpha1.ApiDefinition)
						eval.True(ok)

						time.Sleep(1 * time.Second)
						err := c.Client().Resources().Delete(ctx, apiObj)
						if err != nil {
							t.Logf("failed to delete ApiDefinition, err: %v", err)
							return false
						}

						return true
					}),
					wait.WithTimeout(defaultWaitTimeout),
					wait.WithInterval(defaultWaitInterval),
				)
				eval.NoErr(err)

				//err = envConf.Client().Resources(testNS).Delete(ctx, &apiDef)
				//eval.NoErr(err)

				opCtx := v1alpha1.OperatorContext{ObjectMeta: metav1.ObjectMeta{Name: opCtxName, Namespace: testNS}}

				err = wait.For(conditions.New(c.Client().Resources(testNS)).ResourceDeleted(&opCtx),
					wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
				eval.NoErr(err)

				return ctx
			}).
		Feature()

	//updateApiDef := features.New("Remove context from Api Definition").
	//	Setup(func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
	//		testNS := ctx.Value(ctxNSKey).(string)
	//		eval := is.New(t)
	//		client := envConf.Client()
	//
	//		operatorCtx, err := createTestOperatorContext(ctx, testNS, envConf, nil)
	//		eval.NoErr(err) // failed to create operatorcontext
	//
	//		ctx = context.WithValue(ctx, ctxOpCtxName, operatorCtx.Name)
	//
	//		// create api definition
	//		def, err := createTestAPIDef(ctx, envConf, testNS, func(apiDef *v1alpha1.ApiDefinition) {
	//			apiDef.Spec.Context = &model.Target{
	//				Name:      operatorCtx.Name,
	//				Namespace: operatorCtx.Namespace,
	//			}
	//		})
	//		eval.NoErr(err) // failed to create apiDefinition
	//
	//		ctx = context.WithValue(ctx, ctxApiName, def.Name)
	//
	//		err = wait.For(conditions.New(client.Resources()).ResourceMatch(operatorCtx, func(object k8s.Object) bool {
	//			opCtx := object.(*v1alpha1.OperatorContext) //nolint:errcheck
	//
	//			if len(opCtx.Status.LinkedApiDefinitions) == 0 {
	//				return false
	//			}
	//
	//			return true
	//		}), wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
	//
	//		eval.NoErr(err)
	//
	//		return ctx
	//	}).Assess("context ref should not get deleted while it is still been refered",
	//	func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
	//		testNS := ctx.Value(ctxNSKey).(string)        //nolint:errcheck
	//		opCtxName := ctx.Value(ctxOpCtxName).(string) //nolint:errcheck
	//
	//		client := envConf.Client()
	//		eval := is.New(t)
	//
	//		opCtx := v1alpha1.OperatorContext{ObjectMeta: metav1.ObjectMeta{Name: opCtxName, Namespace: testNS}}
	//
	//		err := client.Resources().Delete(ctx, &opCtx)
	//		eval.NoErr(err)
	//
	//		err = client.Resources(testNS).Get(ctx, testOperatorCtx, testNS, &opCtx)
	//		eval.NoErr(err)
	//
	//		return ctx
	//	}).
	//	Assess("removing reference from apiDefinition should delete operator context",
	//		func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
	//			testNS := ctx.Value(ctxNSKey).(string)        //nolint:errcheck
	//			opCtxName := ctx.Value(ctxOpCtxName).(string) //nolint:errcheck
	//			apiDefName := ctx.Value(ctxApiName).(string)  //nolint:errcheck
	//
	//			client := envConf.Client()
	//			eval := is.New(t)
	//
	//			var apiDef v1alpha1.ApiDefinition
	//			opCtx := v1alpha1.OperatorContext{ObjectMeta: metav1.ObjectMeta{Name: opCtxName, Namespace: testNS}}
	//
	//			err := client.Resources(testNS).Get(ctx, apiDefName, testNS, &apiDef)
	//			eval.NoErr(err)
	//
	//			apiDef.Spec.Context = nil
	//
	//			err = client.Resources(testNS).Update(ctx, &apiDef)
	//			eval.NoErr(err)
	//
	//			err = wait.For(conditions.New(client.Resources()).ResourceDeleted(&opCtx),
	//				wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))
	//
	//			eval.NoErr(err)
	//
	//			return ctx
	//		}).
	//	Feature()

	testenv.Test(t, delApiDef)
	//testenv.Test(t, updateApiDef)
}
