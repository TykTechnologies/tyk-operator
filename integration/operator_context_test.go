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
					Namespace: opCtx.Namespace}
			}, envConf)
			is.NoErr(err) // failed to create apiDefinition

			_, err = createTestAPIDef(ctx, testNS, func(apiDef *v1alpha1.ApiDefinition) {
				apiDef.Name = "empty-ns"
				apiDef.Spec.Context = &model.Target{Name: opCtx.Name}
				apiDef.Spec.Proxy.ListenPath = "/empty-ns"
			}, envConf)

			is.NoErr(err) // failed to create apiDefinition

			return ctx
		}).
		Assess("context status is updated", func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
			testNS := ctx.Value(ctxNSKey).(string)
			client := envConf.Client()

			opCtx := v1alpha1.OperatorContext{ObjectMeta: metav1.ObjectMeta{Name: testOperatorCtx, Namespace: testNS}}

			wait.For(conditions.New(client.Resources()).ResourceMatch(&opCtx, func(object k8s.Object) bool {
				operatCtx := object.(*v1alpha1.OperatorContext)

				if len(operatCtx.Status.LinkedApiDefinitions) != 1 {
					return false
				}

				if operatCtx.Status.LinkedApiDefinitions[0].Namespace != testNS || operatCtx.Status.LinkedApiDefinitions[0].Name != testApiDef {
					return false
				}

				return true
			}))

			return ctx
		}).Assess("apidef was created in dashboard", func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
		wait.For(func() (done bool, err error) {
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
		}, wait.WithTimeout(retryOperationTimeout))

		return ctx

	}).Assess("apidef was created in dashboard", func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
		wait.For(func() (done bool, err error) {
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
		}, wait.WithTimeout(retryOperationTimeout))

		return ctx
	}).Feature()

	testenv.Test(t, opCreate)
}

/*

func TestOperatorContextDelete(t *testing.T) {
	delApiDef := features.New("Operator Context Delete").
		Assess("Delete Api Defintion", func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
			testNS := ctx.Value(ctxNSKey).(string)
			is := is.New(t)
			client := envConf.Client()

			// create operator context
			operatorCtx, err := createTestOperatorContext(ctx, testNS, envConf)
			is.NoErr(err) // failed to create operatorcontext

			// create api definition
			apiDef, err := createTestAPIDef(ctx, testNS, func(apiDef *v1alpha1.ApiDefinition) {
				apiDef.Spec.Context = &model.Target{
					Name:      operatorCtx.Name,
					Namespace: operatorCtx.Namespace}
			}, envConf)
			is.NoErr(err) // failed to create apiDefinition

			err = retryOperation(retryOperationTimeout, reconcileDelay, func() error {
				var opCtx v1alpha1.OperatorContext

				// shouldn't get deleted
				if errGet := client.Resources().Get(ctx, operatorCtx.Name, testNS, &opCtx); errGet != nil {
					t.Log(errGet)
					return errGet
				}

				if len(opCtx.Status.LinkedApiDefinitions) == 0 {
					t.Log("operator context status is not updated yet")
					return errors.New("operator context status is not updated yet")
				}

				t.Log("Operation completed successfully")

				return nil
			})
			is.NoErr(err)

			// try to delete operator context
			err = client.Resources().Delete(ctx, operatorCtx)
			is.NoErr(err)

			time.Sleep(reconcileDelay)

			var result v1alpha1.OperatorContext
			// shouldn't get deleted
			err = client.Resources().Get(ctx, operatorCtx.Name, testNS, &result)
			is.NoErr(err)

			// delete apidef
			err = client.Resources().Delete(ctx, apiDef)
			is.NoErr(err)

			err = retryOperation(retryOperationTimeout, reconcileDelay, func() error {
				var result v1alpha1.OperatorContext

				// should get deleted
				if errGet := client.Resources().Get(ctx, operatorCtx.Name, testNS, &result); errGet != nil {
					return nil
				}

				return errors.New("Should get deleted")
			})
			is.NoErr(err)

			return ctx
		}).Feature()

	updateApiDef := features.New("Operator Context Delete").
		Assess("Remove contextRef from Api Defintion", func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
			testNS := ctx.Value(ctxNSKey).(string)
			is := is.New(t)
			client := envConf.Client()

			// create operator context
			operatorCtx, err := createTestOperatorContext(ctx, testNS, envConf)
			is.NoErr(err) // failed to create operatorcontext

			// create api definition
			apidef, err := createTestAPIDef(ctx, testNS, func(apiDef *v1alpha1.ApiDefinition) {
				apiDef.Spec.Context = &model.Target{
					Name:      operatorCtx.Name,
					Namespace: operatorCtx.Namespace}
			}, envConf)
			is.NoErr(err) // failed to create apiDefinition

			err = retryOperation(retryOperationTimeout, reconcileDelay, func() error {
				var opCtx v1alpha1.OperatorContext
				// shouldn't get deleted
				if errGet := client.Resources().Get(ctx, operatorCtx.Name, testNS, &opCtx); errGet != nil {
					return errGet
				}

				if len(opCtx.Status.LinkedApiDefinitions) == 0 {
					return errors.New("operator context status is not updated yet")
				}

				return nil
			})
			is.NoErr(err)

			// try to delete operator context
			err = client.Resources().Delete(ctx, operatorCtx)
			is.NoErr(err)

			time.Sleep(reconcileDelay)

			var result v1alpha1.OperatorContext
			// shouldn't get deleted
			err = client.Resources().Get(ctx, operatorCtx.Name, testNS, &result)
			is.NoErr(err)

			err = client.Resources().Get(ctx, apidef.Name, apidef.Namespace, apidef)
			is.NoErr(err)

			apidef.Spec.Context = nil

			err = client.Resources().Update(ctx, apidef)
			is.NoErr(err)

			err = retryOperation(retryOperationTimeout, reconcileDelay, func() error {
				var result v1alpha1.OperatorContext

				// should get deleted
				if err = client.Resources().Get(ctx, operatorCtx.Name, testNS, &result); err != nil {
					return nil
				}

				return errors.New("Should get deleted")
			})
			is.NoErr(err)

			return ctx
		}).Feature()

	testenv.Test(t, delApiDef)
	testenv.Test(t, updateApiDef)
}
*/

func createTestAPIDef(ctx context.Context, namespace string, mutateFn func(*v1alpha1.ApiDefinition), envConf *envconf.Config) (*v1alpha1.ApiDefinition, error) {
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
