package integration

import (
	"context"
	"errors"
	"os"

	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"

	"github.com/TykTechnologies/tyk-operator/api/model"
	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/pkg/environmet"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	cr "sigs.k8s.io/controller-runtime/pkg/client"
	e2eKlient "sigs.k8s.io/e2e-framework/klient"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

const (
	opNs                   = "tyk-operator-system"
	testSubGraphCRMetaName = "test-subgraph"
	testSubGraphSchema     = "test-schema"
	testSubGraphSDL        = "test-SDL"
	testApiDef             = "test-api-def"
	testOperatorCtx        = "mycontext"
	testSecurityPolicy     = "test-security-policy"
	gatewayLocalhost       = "http://localhost:7000"
	operatorSecret         = "tyk-operator-conf"
)

// createTestClient creates controller-runtime client by wrapping given e2e test client. It can be used to create
// Reconciler for CRs such as ApiDefinitionReconciler.
func createTestClient(k e2eKlient.Client) (cr.Client, error) {
	scheme := runtime.NewScheme()

	cl, err := cr.New(k.RESTConfig(), cr.Options{Scheme: scheme})
	if err != nil {
		return nil, err
	}

	return cl, v1alpha1.AddToScheme(scheme)
}

// generateApiDef generates a sample ApiDefinition CR. It won't create the newly created CR on your k8s.
func generateApiDef(ns string, mutateFn func(*v1alpha1.ApiDefinition)) *v1alpha1.ApiDefinition {
	var apiDef v1alpha1.ApiDefinition

	apiDef.Name = testApiDef
	apiDef.Namespace = ns
	apiDef.Spec.Name = testApiDef
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

	return &apiDef
}

// generateSubGraphCR generates a sample SubGraph CR. It won't create the newly created CR on your k8s.
func generateSubGraphCR(namespace string, mutateFn func(graph *v1alpha1.SubGraph)) *v1alpha1.SubGraph {
	sg := &v1alpha1.SubGraph{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testSubGraphCRMetaName,
			Namespace: namespace,
		},
		Spec: v1alpha1.SubGraphSpec{
			SubGraphSpec: model.SubGraphSpec{
				Schema: testSubGraphSchema,
				SDL:    testSubGraphSDL,
			},
		},
	}

	if mutateFn != nil {
		mutateFn(sg)
	}

	return sg
}

// createTestAPIDef generates ApiDefinition CR and creates it on your cluster. You can modify new ApiDefinition
// CR via mutateFn.
func createTestAPIDef(ctx context.Context, c *envconf.Config, ns string, mutateFn func(*v1alpha1.ApiDefinition),
) (*v1alpha1.ApiDefinition, error) {
	apiDef := generateApiDef(ns, mutateFn)

	err := c.Client().Resources(ns).Create(ctx, apiDef)
	if err != nil {
		return nil, err
	}

	return apiDef, err
}

// waitForTykResourceCreation waits the creation of resource on Tyk GW or Dashboard. In order to understand whether
// the resource is created successfully or not, it checks status of CRs.
func waitForTykResourceCreation(envConf *envconf.Config, obj k8s.Object) error {
	err := wait.For(conditions.New(envConf.Client().Resources()).ResourceMatch(obj, func(obj k8s.Object) bool {
		switch val := obj.(type) {
		case *v1alpha1.ApiDefinition:
			if val.Status.ApiID != "" {
				return true
			}
		case *v1alpha1.SecurityPolicy:
			if val.Status.PolID != "" {
				return true
			}
		}

		return false
	}), wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))

	return err
}

// createTestOperatorContext creates a sample OperatorContext resource on k8s.
func createTestOperatorContext(ctx context.Context, ns string, c *envconf.Config) (*v1alpha1.OperatorContext, error) {
	var operatorCtx v1alpha1.OperatorContext

	operatorCtx.Name = testOperatorCtx
	operatorCtx.Namespace = ns
	operatorCtx.Spec.FromSecret = &model.Target{
		Name:      operatorSecret,
		Namespace: opNs,
	}

	return &operatorCtx, c.Client().Resources(ns).Create(ctx, &operatorCtx)
}

// createTestPolicy creates a test policy on k8s.
func createTestPolicy(ctx context.Context, c *envconf.Config, namespace string, mutateFn func(*v1alpha1.SecurityPolicy),
) (*v1alpha1.SecurityPolicy, error) {
	var policy v1alpha1.SecurityPolicy

	policy.Name = testSecurityPolicy
	policy.Namespace = namespace
	policy.Spec = v1alpha1.SecurityPolicySpec{
		Name:   testSecurityPolicy,
		Active: true,
		State:  "active",
	}

	if mutateFn != nil {
		mutateFn(&policy)
	}

	err := c.Client().Resources(namespace).Create(ctx, &policy)
	if err != nil {
		return nil, err
	}

	return &policy, err
}

// createTestTlsSecret creates a TLS type of Secret object on your k8s.
func createTestTlsSecret(ctx context.Context, ns string, fn func(*v1.Secret), c *envconf.Config) (*v1.Secret, error) {
	var tlsSecret v1.Secret

	tlsSecret.Name = "test-tls-secret-name"
	tlsSecret.Namespace = ns
	tlsSecret.Data = make(map[string][]byte)

	tlsSecret.Type = "kubernetes.io/tls"
	tlsSecret.Data["tls.key"] = []byte("-----BEGIN PRIVATE KEY-----\nMIIEvAIBADANBgkqhkiG9w0BAQEFAASCBKYwggSiAgEAAoIBAQCys8pxhaROYu6z\nqLAsrKyKDY+XgOhJfzctPWD6IsK4DeCiAOfWje3aXUZ6FvwlsMfW1vb6JdEQdsyo\n6YjI90HOZ+DcmH7Wc2oTV/pHRflx4IoWVr1lZmzmKCs4or0+Fk7TlwBjiAM0trya\nCxfMZGX2Vt1L5PP3yjlL3E/jyxalS3E9hLjhbv8nyf5ht8U2H54a1BmavXF0s2hc\njwSUc0TF+KcLI+k+loc+Y+cEun+0PDbAeq0RG/hSGnPz0qHItgyYBE4pHVRQXrNY\nq+3nPtN4BUpanV4w8TCRgwKdAlDbis/xCucFU3HczLQvbzot9uDpg+Ev1+CFi0A5\nfMsC152zAgMBAAECggEAaxkfcevDLgtSva+SbiPKgC5iaU0jabDpc55+eUrN4hrH\nDrB2QXrsGtud+lu+ICSTj+ljOUXixvg77duQU8kD0l0lQW/PTFz9LLykTYTdW2dT\nutGfTp8VEtbuGFJIEmayNVMhM4V3TmdaHwQY7jEZfopOtEZyBIZY0mMmKgIz/zl3\nLJ62trgvGmvGWR+MWDx2EP8rRhgY+FNS/S+rpKZujk2fiOEc71k+iTsAvSa2uWtM\nIEsXX1xOolJMXof2zrfwhnVX6XKPSDcbTGBOvfpUndQpvlJQe8qs0VRwUbC7ceQn\n2LDFu/5r5u4mx4jTHTGXt6gVLOCFwWm+ecTogG+ZoQKBgQDp67xeSpM6g3BUsZrL\nQGYoKJRBFDbh1aW3JEYhUh+brw5urXDLvLfpzmRJcrTTq7+MiIni1+5qeeJNOHYB\nNTk7gGA7LBItJReijaTcZVa3o48BTQwsRXKCZtby6uLBHbKbpH0XJURPrIupHVZg\nvtQMABMRwZ6CEJMlYfROcHjSEQKBgQDDkc0hFhs4XJwJAvY0lsY2z9IfjGBOvYJg\n6R13mjMM8a8ceRioTRFRWh1c7P6qiipIY4zBu/W6pNiuMU/8rMI+LacEjzPObI0J\nlnbLwIJ/qy+q7YMf02XAlFf73iaX5Cv/u+FwcxLlHu+XkhVWqs1P5RGKYZMzJytZ\nPXZxjEMvgwKBgA2+z2vPAAXBMXmYkhr9ZsNXVxbX5D2y+zDezcwpcjgIulVgla8z\nIK95dEUom12QywmsAEY3IAhbryOQfManZPyNF5qChXLnqhLgNd7JiaXy03VlHKEB\nV7A38MuHZ9mnMBabPMp+Yxw3bGF8mtXGgNlPq88wTGsiJDNfJSbyzvaxAoGACOhW\nKICiQsHtFXf+EM0hQBPdJTS2mj+FdbaIcg8i7h7/89MMLXY9KLBrD/V3b/sVC/EE\n0zolahfiCqUSWJbhzgU0Sz/egzNshRhGVudwyjHY3Pcudr+hLdFT5JPsvBRXcLF1\nBjMnlCoBjazIrgbfjRkI4H2rP7Q0BD+JaoiR8tMCgYBcpjRaY5z/mUBoCe6mf9Ts\nIeAMeaVfVlJZlr699Ix2CAnLzSeF0FfDibwrh2WapIYXpItTV6oEv+HTGqAHt6W5\nx9qqMl4RgV2L2k/ox+NyMZKx8DQ9Lv1jdEwBDjF/+0xTXurxW+g1ZUFYnD7Q9dif\nuNnays8krQv5B3h/8Bsbyw==\n-----END PRIVATE KEY-----\n") //nolint
	tlsSecret.Data["tls.crt"] = []byte("-----BEGIN CERTIFICATE-----\nMIICqDCCAZACCQDHVUhoyzm1tTANBgkqhkiG9w0BAQsFADAWMRQwEgYDVQQDDAtm\nb28uYmFyLmNvbTAeFw0yMjAzMTYwMDM4MzhaFw0yMzAzMTYwMDM4MzhaMBYxFDAS\nBgNVBAMMC2Zvby5iYXIuY29tMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKC\nAQEAsrPKcYWkTmLus6iwLKysig2Pl4DoSX83LT1g+iLCuA3gogDn1o3t2l1Gehb8\nJbDH1tb2+iXREHbMqOmIyPdBzmfg3Jh+1nNqE1f6R0X5ceCKFla9ZWZs5igrOKK9\nPhZO05cAY4gDNLa8mgsXzGRl9lbdS+Tz98o5S9xP48sWpUtxPYS44W7/J8n+YbfF\nNh+eGtQZmr1xdLNoXI8ElHNExfinCyPpPpaHPmPnBLp/tDw2wHqtERv4Uhpz89Kh\nyLYMmAROKR1UUF6zWKvt5z7TeAVKWp1eMPEwkYMCnQJQ24rP8QrnBVNx3My0L286\nLfbg6YPhL9fghYtAOXzLAtedswIDAQABMA0GCSqGSIb3DQEBCwUAA4IBAQCCUBsU\nAslwTYVCwPyFYG1qaB8ipxpRcsawRmah2BDiEjvd2UEYTk+LpFOEWLujdWxM9NHb\nW2WGYW5D4yVSLmdwR+ddJYAxWhKghg4hhO1Qpr7CdvJdRBz2SS9bc18gZ1ZCz/wl\nszKluhKmgBMwfpMSgwYmOggQgufAY4Q3llehA/6lWeyhxdpZ4xZ+m9U1h4JeFGTj\nIaryEbX2Fqm3MUeXyDgk65a9DNYRHFs9VMOYr4CZl7BMg/lFy7W8DcoxsIUaBbDu\n+HqNLh62N7i6Tg9p9euFPPkVu3oJkWulZNNEb+/g8u8dBGeyENXMD2+SBz3ZFZcF\ndvzZ+WvUvFyWa4XO\n-----END CERTIFICATE-----\n")                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               //nolint

	if fn != nil {
		fn(&tlsSecret)
	}

	return &tlsSecret, c.Client().Resources(ns).Create(ctx, &tlsSecret)
}

// generateEnvConfig creates a config structure to connect your Tyk installation. It parses k8s secret object
// and reads required connection credentials from there.
func generateEnvConfig(operatorConfSecret *v1.Secret) (environmet.Env, error) {
	data, ok := operatorConfSecret.Data["TYK_AUTH"]
	if !ok {
		return environmet.Env{}, errors.New("failed to parse TYK_AUTH from operator secret")
	}

	tykAuth := string(data)

	data, ok = operatorConfSecret.Data["TYK_ORG"]
	if !ok {
		return environmet.Env{}, errors.New("failed to parse TYK_ORG from operator secret")
	}

	tykOrg := string(data)
	tykVersion := "v4.2"

	data, ok = operatorConfSecret.Data["TYK_VERSION"]
	if ok && len(data) != 0 {
		tykVersion = string(data)
	}

	mode := os.Getenv("TYK_MODE")
	var tykConnectionURL string

	switch mode {
	case "pro":
		tykConnectionURL = adminLocalhost
	case "ce":
		tykConnectionURL = gatewayLocalhost
	}

	return environmet.Env{
		Environment: v1alpha1.Environment{
			Auth: tykAuth,
			Org:  tykOrg,
			Mode: v1alpha1.OperatorContextMode(mode),
			URL:  tykConnectionURL,
		},
		TykVersion: tykVersion,
	}, nil
}
