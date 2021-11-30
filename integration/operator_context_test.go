package integration

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/TykTechnologies/tyk-operator/api/model"
	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/matryer/is"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestOperatorContextCreate(t *testing.T) {
	opCreate := features.New("Operator Context").
		Assess("Create", func(ctx context.Context, t *testing.T, envConf *envconf.Config) context.Context {
			testNS := ctx.Value(ctxNSKey).(string)
			is := is.New(t)

			// create api definition
			_, err := createTestAPIDef(ctx, testNS, envConf)
			is.NoErr(err) // failed to create apiDefinition

			_, err = createTestOperatorContext(ctx, testNS, envConf)
			is.NoErr(err) // failed to create operatorcontext

			time.Sleep(reconcileDelay)

			resp, err := http.Get("http://localhost:7000/httpbin/get")
			is.NoErr(err)
			is.Equal(resp.StatusCode, 200)

			return ctx
		}).Feature()

	testenv.Test(t, opCreate)
}

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
			apiDef, err := createTestAPIDef(ctx, testNS, envConf)
			is.NoErr(err) // failed to create apiDefinition

			time.Sleep(reconcileDelay)

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

			time.Sleep(reconcileDelay)

			// once api defintions referring to it get deleted, operator context gets deleted
			err = client.Resources().Get(ctx, operatorCtx.Name, testNS, &result)
			is.True(err != nil)

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
			apidef, err := createTestAPIDef(ctx, testNS, envConf)
			is.NoErr(err) // failed to create apiDefinition

			time.Sleep(reconcileDelay)

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

			time.Sleep(reconcileDelay)

			err = client.Resources().Get(ctx, operatorCtx.Name, testNS, &result)
			is.True(err != nil)

			return ctx
		}).Feature()

	testenv.Test(t, delApiDef)
	testenv.Test(t, updateApiDef)
}

func createTestAPIDef(ctx context.Context, namespace string, envConf *envconf.Config) (*v1alpha1.ApiDefinition, error) {
	var apiDef v1alpha1.ApiDefinition

	client := envConf.Client()

	apiDef.Name = "test-http"
	apiDef.Spec.Name = "test-http"
	apiDef.Namespace = namespace
	apiDef.Spec.Protocol = "http"
	apiDef.Spec.Context = &model.Target{
		Namespace: namespace,
		Name:      "mycontext",
	}
	apiDef.Spec.UseKeylessAccess = true
	apiDef.Spec.Active = true
	apiDef.Spec.Proxy = model.Proxy{
		ListenPath:      "/httpbin",
		TargetURL:       "http://httpbin.default.svc:8000",
		StripListenPath: true,
	}

	err := client.Resources(namespace).Create(ctx, &apiDef)

	return &apiDef, err
}

func createTestOperatorContext(ctx context.Context, namespace string, envConf *envconf.Config) (*v1alpha1.OperatorContext, error) {
	var operatorCtx v1alpha1.OperatorContext

	client := envConf.Client()

	operatorCtx.Name = "mycontext"
	operatorCtx.Namespace = namespace
	operatorCtx.Spec.FromSecret = &model.Target{
		Name:      "tyk-operator-conf",
		Namespace: operatorNamespace,
	}

	err := client.Resources(namespace).Create(ctx, &operatorCtx)

	return &operatorCtx, err
}
