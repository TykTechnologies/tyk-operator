package integration

import (
	"context"
	"net/http"
	"testing"

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

const (
	testApiDef      = "test-http"
	testOperatorCtx = "mycontext"
)

func TestOperatorContextCreate(t *testing.T) {
	opCreate := features.New("Operator Context").
		Setup(func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
			testNS := ctx.Value(ctxNSKey).(string)
			is := is.New(t)

			// create operator context
			opCtx, err := createTestOperatorContext(ctx, testNS, envConf)
			is.NoErr(err) // failed to create operatorcontext

			// create api definition
			_, err = createTestAPIDef(ctx, testNS, func(apiDef *v1alpha1.ApiDefinition) {
				apiDef.Spec.Context = &model.Target{
					Name:      opCtx.Name,
					Namespace: opCtx.Namespace,
				}
			}, envConf)
			is.NoErr(err) // failed to create apiDefinition

			// create api defintion with empty namespace for contextRef
			_, err = createTestAPIDef(ctx, testNS, func(apiDef *v1alpha1.ApiDefinition) {
				apiDef.Name = "empty-ns"
				apiDef.Spec.Context = &model.Target{Name: opCtx.Name}
				apiDef.Spec.Proxy.ListenPath = "/empty-ns"
			}, envConf)

			is.NoErr(err) // failed to create apiDefinition

			return ctx
		}).
		Assess("context status is updated",
			func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
				testNS := ctx.Value(ctxNSKey).(string) //nolint:errcheck
				client := envConf.Client()
				is := is.New(t)

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
				}))
				is.NoErr(err)

				return ctx
			}).Assess("apidef was created in dashboard",
		func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
			is := is.New(t)

			err := wait.For(func() (done bool, err error) {
				resp, getErr := http.Get("http://localhost:7000/httpbin/get")
				if getErr != nil {
					t.Log(getErr)
					return false, nil
				}

				if resp.StatusCode != 200 {
					t.Log("API is not created yet")
					return false, nil
				}

				return true, nil
			})
			is.NoErr(err)

			return ctx
		}).Assess("apidef with empty namespace in contextRef was not created in dashboard",
		func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
			is := is.New(t)

			err := wait.For(func() (done bool, err error) {
				resp, getErr := http.Get("http://localhost:7000/empty-ns/get")
				if getErr != nil {
					t.Log(getErr)
					return false, nil
				}

				if resp.StatusCode != 404 {
					t.Log("API definition should not be created on dashbaord")
					return false, nil
				}

				return true, nil
			})

			is.NoErr(err)

			return ctx
		}).Feature()

	testenv.Test(t, opCreate)
}

func TestOperatorContextDelete(t *testing.T) {
	delApiDef := features.New("Api Definition Delete").
		Setup(func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
			testNS := ctx.Value(ctxNSKey).(string)
			is := is.New(t)
			client := envConf.Client()

			operatorCtx, err := createTestOperatorContext(ctx, testNS, envConf)
			is.NoErr(err) // failed to create operatorcontext

			ctx = context.WithValue(ctx, ctxOpCtxName, operatorCtx.Name)

			// create api definition
			def, err := createTestAPIDef(ctx, testNS, func(apiDef *v1alpha1.ApiDefinition) {
				apiDef.Spec.Context = &model.Target{
					Name:      operatorCtx.Name,
					Namespace: operatorCtx.Namespace,
				}
			}, envConf)
			is.NoErr(err) // failed to create apiDefinition

			ctx = context.WithValue(ctx, ctxApiName, def.Name)

			err = wait.For(conditions.New(client.Resources()).ResourceMatch(operatorCtx, func(object k8s.Object) bool {
				opCtx := object.(*v1alpha1.OperatorContext) //nolint:errcheck

				if len(opCtx.Status.LinkedApiDefinitions) == 0 {
					t.Log(opCtx)
					return false
				}

				return true
			}))

			is.NoErr(err)

			return ctx
		}).Assess("context ref should not get deleted while it is still been refered",
		func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
			testNS := ctx.Value(ctxNSKey).(string)        //nolint:errcheck
			opCtxName := ctx.Value(ctxOpCtxName).(string) //nolint:errcheck

			client := envConf.Client()
			is := is.New(t)

			opCtx := v1alpha1.OperatorContext{ObjectMeta: metav1.ObjectMeta{Name: opCtxName, Namespace: testNS}}

			err := client.Resources().Delete(ctx, &opCtx)
			is.NoErr(err)

			err = client.Resources(testNS).Get(ctx, opCtxName, testNS, &opCtx)
			is.NoErr(err)

			return ctx
		}).
		Assess("delete api def should delete operator context",
			func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
				testNS := ctx.Value(ctxNSKey).(string)        //nolint:errcheck
				opCtxName := ctx.Value(ctxOpCtxName).(string) //nolint:errcheck
				apiDefName := ctx.Value(ctxApiName).(string)  //nolint:errcheck

				client := envConf.Client()
				is := is.New(t)

				apiDef := v1alpha1.ApiDefinition{ObjectMeta: metav1.ObjectMeta{Name: apiDefName, Namespace: testNS}}
				opCtx := v1alpha1.OperatorContext{ObjectMeta: metav1.ObjectMeta{Name: opCtxName, Namespace: testNS}}

				err := client.Resources(testNS).Delete(ctx, &apiDef)
				is.NoErr(err)

				err = wait.For(conditions.New(client.Resources()).ResourceDeleted(&opCtx))
				is.NoErr(err)

				return ctx
			}).
		Feature()

	updateApiDef := features.New("Remove context from Api Definition").
		Setup(func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
			testNS := ctx.Value(ctxNSKey).(string)
			is := is.New(t)
			client := envConf.Client()

			operatorCtx, err := createTestOperatorContext(ctx, testNS, envConf)
			is.NoErr(err) // failed to create operatorcontext

			ctx = context.WithValue(ctx, ctxOpCtxName, operatorCtx.Name)

			// create api definition
			def, err := createTestAPIDef(ctx, testNS, func(apiDef *v1alpha1.ApiDefinition) {
				apiDef.Spec.Context = &model.Target{
					Name:      operatorCtx.Name,
					Namespace: operatorCtx.Namespace,
				}
			}, envConf)
			is.NoErr(err) // failed to create apiDefinition

			ctx = context.WithValue(ctx, ctxApiName, def.Name)

			err = wait.For(conditions.New(client.Resources()).ResourceMatch(operatorCtx, func(object k8s.Object) bool {
				opCtx := object.(*v1alpha1.OperatorContext) //nolint:errcheck

				if len(opCtx.Status.LinkedApiDefinitions) == 0 {
					return false
				}

				return true
			}))

			is.NoErr(err)

			return ctx
		}).Assess("context ref should not get deleted while it is still been refered",
		func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
			testNS := ctx.Value(ctxNSKey).(string)        //nolint:errcheck
			opCtxName := ctx.Value(ctxOpCtxName).(string) //nolint:errcheck

			client := envConf.Client()
			is := is.New(t)

			opCtx := v1alpha1.OperatorContext{ObjectMeta: metav1.ObjectMeta{Name: opCtxName, Namespace: testNS}}

			err := client.Resources().Delete(ctx, &opCtx)
			is.NoErr(err)

			err = client.Resources(testNS).Get(ctx, testOperatorCtx, testNS, &opCtx)
			is.NoErr(err)

			return ctx
		}).
		Assess("removing reference from apiDefinition should delete operator context",
			func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
				testNS := ctx.Value(ctxNSKey).(string)        //nolint:errcheck
				opCtxName := ctx.Value(ctxOpCtxName).(string) //nolint:errcheck
				apiDefName := ctx.Value(ctxApiName).(string)  //nolint:errcheck

				client := envConf.Client()
				is := is.New(t)

				var apiDef v1alpha1.ApiDefinition
				opCtx := v1alpha1.OperatorContext{ObjectMeta: metav1.ObjectMeta{Name: opCtxName, Namespace: testNS}}

				err := client.Resources(testNS).Get(ctx, apiDefName, testNS, &apiDef)
				is.NoErr(err)

				apiDef.Spec.Context = nil

				err = client.Resources(testNS).Update(ctx, &apiDef)
				is.NoErr(err)

				err = wait.For(conditions.New(client.Resources()).ResourceDeleted(&opCtx))

				is.NoErr(err)

				return ctx
			}).
		Feature()

	testenv.Test(t, delApiDef)
	testenv.Test(t, updateApiDef)
}

func createTestAPIDef(ctx context.Context, namespace string, mutateFn func(*v1alpha1.ApiDefinition),
	envConf *envconf.Config) (*v1alpha1.ApiDefinition, error) {
	client := envConf.Client()
	var apiDef v1alpha1.ApiDefinition

	apiDef.Name = testApiDef
	apiDef.Spec.Name = testApiDef
	apiDef.Namespace = namespace
	apiDef.Spec.Protocol = "http"
	apiDef.Spec.UseKeylessAccess = true
	apiDef.Spec.Active = true
	apiDef.Spec.Proxy = model.Proxy{
		ListenPath:      "/httpbin",
		TargetURL:       "http://httpbin.default.svc:8000",
		StripListenPath: true,
	}

	if mutateFn != nil {
		mutateFn(&apiDef)
	}

	err := client.Resources(namespace).Create(ctx, &apiDef)

	return &apiDef, err
}

func createTestOperatorContext(ctx context.Context, namespace string, envConf *envconf.Config) (*v1alpha1.OperatorContext, error) {
	var operatorCtx v1alpha1.OperatorContext

	client := envConf.Client()

	operatorCtx.Name = testOperatorCtx
	operatorCtx.Namespace = namespace
	operatorCtx.Spec.FromSecret = &model.Target{
		Name:      "tyk-operator-conf",
		Namespace: operatorNamespace,
	}

	err := client.Resources(namespace).Create(ctx, &operatorCtx)

	return &operatorCtx, err
}
