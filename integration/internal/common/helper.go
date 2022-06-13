package common

import (
	"context"

	"github.com/TykTechnologies/tyk-operator/api/model"
	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

func CreateTestAPIDef(ctx context.Context, namespace string, mutateFn func(*v1alpha1.ApiDefinition),
	envConf *envconf.Config,
) (*v1alpha1.ApiDefinition, error) {
	client := envConf.Client()
	var apiDef v1alpha1.ApiDefinition

	apiDef.Name = TestApiDef
	apiDef.Spec.Name = TestApiDef
	apiDef.Namespace = namespace
	apiDef.Spec.Protocol = "http"
	apiDef.Spec.UseKeylessAccess = true
	apiDef.Spec.Active = true
	apiDef.Spec.VersionData = model.VersionData{
		DefaultVersion: "Default",
		NotVersioned:   true,
		Versions:       map[string]model.VersionInfo{"Default": {Name: "Default"}},
	}
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

func CreateTestOperatorContext(ctx context.Context, namespace string,
	envConf *envconf.Config,
) (*v1alpha1.OperatorContext, error) {
	var operatorCtx v1alpha1.OperatorContext

	client := envConf.Client()

	operatorCtx.Name = TestOperatorCtx
	operatorCtx.Namespace = namespace
	operatorCtx.Spec.FromSecret = &model.Target{
		Name:      "tyk-operator-conf",
		Namespace: OperatorNamespace,
	}

	err := client.Resources(namespace).Create(ctx, &operatorCtx)

	return &operatorCtx, err
}

func CreateTestTlsSecret(ctx context.Context, namespace string, mutateFn func(*v1.Secret),
	envConf *envconf.Config,
) (*v1.Secret, error) {
	client := envConf.Client()
	var tlsSecret v1.Secret

	tlsSecret.Name = "test-tls-secret-name"
	tlsSecret.Namespace = namespace
	tlsSecret.Data = make(map[string][]byte)

	tlsSecret.Type = v1.SecretTypeTLS
	tlsSecret.Data["tls.key"] = HttpbinTLSKey
	tlsSecret.Data["tls.crt"] = HttpbinTLSCert

	if mutateFn != nil {
		mutateFn(&tlsSecret)
	}

	err := client.Resources(namespace).Create(ctx, &tlsSecret)

	return &tlsSecret, err
}
