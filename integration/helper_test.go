package integration

import (
	"context"

	"github.com/TykTechnologies/tyk-operator/api/model"
	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

func createTestAPIDef(ctx context.Context, namespace string, mutateFn func(*v1alpha1.ApiDefinition),
	envConf *envconf.Config,
) (*v1alpha1.ApiDefinition, error) {
	client := envConf.Client()
	var apiDef v1alpha1.ApiDefinition

	apiDef.Name = testApiDef
	apiDef.Spec.Name = testApiDef
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

func createTestOperatorContext(ctx context.Context, namespace string,
	envConf *envconf.Config) (*v1alpha1.OperatorContext, error) {
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

func createTestTlsSecret(ctx context.Context, namespace string, mutateFn func(*v1.Secret),
	envConf *envconf.Config,
) (*v1.Secret, error) {
	client := envConf.Client()
	var tlsSecret v1.Secret

	tlsSecret.Name = "test-tls-secret-name"
	tlsSecret.Namespace = namespace
	tlsSecret.Data = make(map[string][]byte)

	tlsSecret.Type = "kubernetes.io/tls"
	tlsSecret.Data["tls.key"] = []byte("-----BEGIN PRIVATE KEY-----\nMIIEvAIBADANBgkqhkiG9w0BAQEFAASCBKYwggSiAgEAAoIBAQCys8pxhaROYu6z\nqLAsrKyKDY+XgOhJfzctPWD6IsK4DeCiAOfWje3aXUZ6FvwlsMfW1vb6JdEQdsyo\n6YjI90HOZ+DcmH7Wc2oTV/pHRflx4IoWVr1lZmzmKCs4or0+Fk7TlwBjiAM0trya\nCxfMZGX2Vt1L5PP3yjlL3E/jyxalS3E9hLjhbv8nyf5ht8U2H54a1BmavXF0s2hc\njwSUc0TF+KcLI+k+loc+Y+cEun+0PDbAeq0RG/hSGnPz0qHItgyYBE4pHVRQXrNY\nq+3nPtN4BUpanV4w8TCRgwKdAlDbis/xCucFU3HczLQvbzot9uDpg+Ev1+CFi0A5\nfMsC152zAgMBAAECggEAaxkfcevDLgtSva+SbiPKgC5iaU0jabDpc55+eUrN4hrH\nDrB2QXrsGtud+lu+ICSTj+ljOUXixvg77duQU8kD0l0lQW/PTFz9LLykTYTdW2dT\nutGfTp8VEtbuGFJIEmayNVMhM4V3TmdaHwQY7jEZfopOtEZyBIZY0mMmKgIz/zl3\nLJ62trgvGmvGWR+MWDx2EP8rRhgY+FNS/S+rpKZujk2fiOEc71k+iTsAvSa2uWtM\nIEsXX1xOolJMXof2zrfwhnVX6XKPSDcbTGBOvfpUndQpvlJQe8qs0VRwUbC7ceQn\n2LDFu/5r5u4mx4jTHTGXt6gVLOCFwWm+ecTogG+ZoQKBgQDp67xeSpM6g3BUsZrL\nQGYoKJRBFDbh1aW3JEYhUh+brw5urXDLvLfpzmRJcrTTq7+MiIni1+5qeeJNOHYB\nNTk7gGA7LBItJReijaTcZVa3o48BTQwsRXKCZtby6uLBHbKbpH0XJURPrIupHVZg\nvtQMABMRwZ6CEJMlYfROcHjSEQKBgQDDkc0hFhs4XJwJAvY0lsY2z9IfjGBOvYJg\n6R13mjMM8a8ceRioTRFRWh1c7P6qiipIY4zBu/W6pNiuMU/8rMI+LacEjzPObI0J\nlnbLwIJ/qy+q7YMf02XAlFf73iaX5Cv/u+FwcxLlHu+XkhVWqs1P5RGKYZMzJytZ\nPXZxjEMvgwKBgA2+z2vPAAXBMXmYkhr9ZsNXVxbX5D2y+zDezcwpcjgIulVgla8z\nIK95dEUom12QywmsAEY3IAhbryOQfManZPyNF5qChXLnqhLgNd7JiaXy03VlHKEB\nV7A38MuHZ9mnMBabPMp+Yxw3bGF8mtXGgNlPq88wTGsiJDNfJSbyzvaxAoGACOhW\nKICiQsHtFXf+EM0hQBPdJTS2mj+FdbaIcg8i7h7/89MMLXY9KLBrD/V3b/sVC/EE\n0zolahfiCqUSWJbhzgU0Sz/egzNshRhGVudwyjHY3Pcudr+hLdFT5JPsvBRXcLF1\nBjMnlCoBjazIrgbfjRkI4H2rP7Q0BD+JaoiR8tMCgYBcpjRaY5z/mUBoCe6mf9Ts\nIeAMeaVfVlJZlr699Ix2CAnLzSeF0FfDibwrh2WapIYXpItTV6oEv+HTGqAHt6W5\nx9qqMl4RgV2L2k/ox+NyMZKx8DQ9Lv1jdEwBDjF/+0xTXurxW+g1ZUFYnD7Q9dif\nuNnays8krQv5B3h/8Bsbyw==\n-----END PRIVATE KEY-----\n") //nolint
	tlsSecret.Data["tls.crt"] = []byte("-----BEGIN CERTIFICATE-----\nMIICqDCCAZACCQDHVUhoyzm1tTANBgkqhkiG9w0BAQsFADAWMRQwEgYDVQQDDAtm\nb28uYmFyLmNvbTAeFw0yMjAzMTYwMDM4MzhaFw0yMzAzMTYwMDM4MzhaMBYxFDAS\nBgNVBAMMC2Zvby5iYXIuY29tMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKC\nAQEAsrPKcYWkTmLus6iwLKysig2Pl4DoSX83LT1g+iLCuA3gogDn1o3t2l1Gehb8\nJbDH1tb2+iXREHbMqOmIyPdBzmfg3Jh+1nNqE1f6R0X5ceCKFla9ZWZs5igrOKK9\nPhZO05cAY4gDNLa8mgsXzGRl9lbdS+Tz98o5S9xP48sWpUtxPYS44W7/J8n+YbfF\nNh+eGtQZmr1xdLNoXI8ElHNExfinCyPpPpaHPmPnBLp/tDw2wHqtERv4Uhpz89Kh\nyLYMmAROKR1UUF6zWKvt5z7TeAVKWp1eMPEwkYMCnQJQ24rP8QrnBVNx3My0L286\nLfbg6YPhL9fghYtAOXzLAtedswIDAQABMA0GCSqGSIb3DQEBCwUAA4IBAQCCUBsU\nAslwTYVCwPyFYG1qaB8ipxpRcsawRmah2BDiEjvd2UEYTk+LpFOEWLujdWxM9NHb\nW2WGYW5D4yVSLmdwR+ddJYAxWhKghg4hhO1Qpr7CdvJdRBz2SS9bc18gZ1ZCz/wl\nszKluhKmgBMwfpMSgwYmOggQgufAY4Q3llehA/6lWeyhxdpZ4xZ+m9U1h4JeFGTj\nIaryEbX2Fqm3MUeXyDgk65a9DNYRHFs9VMOYr4CZl7BMg/lFy7W8DcoxsIUaBbDu\n+HqNLh62N7i6Tg9p9euFPPkVu3oJkWulZNNEb+/g8u8dBGeyENXMD2+SBz3ZFZcF\ndvzZ+WvUvFyWa4XO\n-----END CERTIFICATE-----\n")                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               //nolint

	if mutateFn != nil {
		mutateFn(&tlsSecret)
	}

	err := client.Resources(namespace).Create(ctx, &tlsSecret)

	return &tlsSecret, err
}
