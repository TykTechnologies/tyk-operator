package common

import (
	"time"

	"github.com/TykTechnologies/tyk-operator/pkg/environment"
)

var Env environment.Env

type ctxKey string

var (
	HttpbinTLSKey  = []byte("-----BEGIN PRIVATE KEY-----\nMIIEvAIBADANBgkqhkiG9w0BAQEFAASCBKYwggSiAgEAAoIBAQCys8pxhaROYu6z\nqLAsrKyKDY+XgOhJfzctPWD6IsK4DeCiAOfWje3aXUZ6FvwlsMfW1vb6JdEQdsyo\n6YjI90HOZ+DcmH7Wc2oTV/pHRflx4IoWVr1lZmzmKCs4or0+Fk7TlwBjiAM0trya\nCxfMZGX2Vt1L5PP3yjlL3E/jyxalS3E9hLjhbv8nyf5ht8U2H54a1BmavXF0s2hc\njwSUc0TF+KcLI+k+loc+Y+cEun+0PDbAeq0RG/hSGnPz0qHItgyYBE4pHVRQXrNY\nq+3nPtN4BUpanV4w8TCRgwKdAlDbis/xCucFU3HczLQvbzot9uDpg+Ev1+CFi0A5\nfMsC152zAgMBAAECggEAaxkfcevDLgtSva+SbiPKgC5iaU0jabDpc55+eUrN4hrH\nDrB2QXrsGtud+lu+ICSTj+ljOUXixvg77duQU8kD0l0lQW/PTFz9LLykTYTdW2dT\nutGfTp8VEtbuGFJIEmayNVMhM4V3TmdaHwQY7jEZfopOtEZyBIZY0mMmKgIz/zl3\nLJ62trgvGmvGWR+MWDx2EP8rRhgY+FNS/S+rpKZujk2fiOEc71k+iTsAvSa2uWtM\nIEsXX1xOolJMXof2zrfwhnVX6XKPSDcbTGBOvfpUndQpvlJQe8qs0VRwUbC7ceQn\n2LDFu/5r5u4mx4jTHTGXt6gVLOCFwWm+ecTogG+ZoQKBgQDp67xeSpM6g3BUsZrL\nQGYoKJRBFDbh1aW3JEYhUh+brw5urXDLvLfpzmRJcrTTq7+MiIni1+5qeeJNOHYB\nNTk7gGA7LBItJReijaTcZVa3o48BTQwsRXKCZtby6uLBHbKbpH0XJURPrIupHVZg\nvtQMABMRwZ6CEJMlYfROcHjSEQKBgQDDkc0hFhs4XJwJAvY0lsY2z9IfjGBOvYJg\n6R13mjMM8a8ceRioTRFRWh1c7P6qiipIY4zBu/W6pNiuMU/8rMI+LacEjzPObI0J\nlnbLwIJ/qy+q7YMf02XAlFf73iaX5Cv/u+FwcxLlHu+XkhVWqs1P5RGKYZMzJytZ\nPXZxjEMvgwKBgA2+z2vPAAXBMXmYkhr9ZsNXVxbX5D2y+zDezcwpcjgIulVgla8z\nIK95dEUom12QywmsAEY3IAhbryOQfManZPyNF5qChXLnqhLgNd7JiaXy03VlHKEB\nV7A38MuHZ9mnMBabPMp+Yxw3bGF8mtXGgNlPq88wTGsiJDNfJSbyzvaxAoGACOhW\nKICiQsHtFXf+EM0hQBPdJTS2mj+FdbaIcg8i7h7/89MMLXY9KLBrD/V3b/sVC/EE\n0zolahfiCqUSWJbhzgU0Sz/egzNshRhGVudwyjHY3Pcudr+hLdFT5JPsvBRXcLF1\nBjMnlCoBjazIrgbfjRkI4H2rP7Q0BD+JaoiR8tMCgYBcpjRaY5z/mUBoCe6mf9Ts\nIeAMeaVfVlJZlr699Ix2CAnLzSeF0FfDibwrh2WapIYXpItTV6oEv+HTGqAHt6W5\nx9qqMl4RgV2L2k/ox+NyMZKx8DQ9Lv1jdEwBDjF/+0xTXurxW+g1ZUFYnD7Q9dif\nuNnays8krQv5B3h/8Bsbyw==\n-----END PRIVATE KEY-----\n") //nolint
	HttpbinTLSCert = []byte("-----BEGIN CERTIFICATE-----\nMIICqDCCAZACCQDHVUhoyzm1tTANBgkqhkiG9w0BAQsFADAWMRQwEgYDVQQDDAtm\nb28uYmFyLmNvbTAeFw0yMjAzMTYwMDM4MzhaFw0yMzAzMTYwMDM4MzhaMBYxFDAS\nBgNVBAMMC2Zvby5iYXIuY29tMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKC\nAQEAsrPKcYWkTmLus6iwLKysig2Pl4DoSX83LT1g+iLCuA3gogDn1o3t2l1Gehb8\nJbDH1tb2+iXREHbMqOmIyPdBzmfg3Jh+1nNqE1f6R0X5ceCKFla9ZWZs5igrOKK9\nPhZO05cAY4gDNLa8mgsXzGRl9lbdS+Tz98o5S9xP48sWpUtxPYS44W7/J8n+YbfF\nNh+eGtQZmr1xdLNoXI8ElHNExfinCyPpPpaHPmPnBLp/tDw2wHqtERv4Uhpz89Kh\nyLYMmAROKR1UUF6zWKvt5z7TeAVKWp1eMPEwkYMCnQJQ24rP8QrnBVNx3My0L286\nLfbg6YPhL9fghYtAOXzLAtedswIDAQABMA0GCSqGSIb3DQEBCwUAA4IBAQCCUBsU\nAslwTYVCwPyFYG1qaB8ipxpRcsawRmah2BDiEjvd2UEYTk+LpFOEWLujdWxM9NHb\nW2WGYW5D4yVSLmdwR+ddJYAxWhKghg4hhO1Qpr7CdvJdRBz2SS9bc18gZ1ZCz/wl\nszKluhKmgBMwfpMSgwYmOggQgufAY4Q3llehA/6lWeyhxdpZ4xZ+m9U1h4JeFGTj\nIaryEbX2Fqm3MUeXyDgk65a9DNYRHFs9VMOYr4CZl7BMg/lFy7W8DcoxsIUaBbDu\n+HqNLh62N7i6Tg9p9euFPPkVu3oJkWulZNNEb+/g8u8dBGeyENXMD2+SBz3ZFZcF\ndvzZ+WvUvFyWa4XO\n-----END CERTIFICATE-----\n")                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               //nolint
)

const (
	CtxNSKey     ctxKey = "namespaceName"
	CtxApiName   ctxKey = "apiName"
	CtxOpCtxName ctxKey = "opCtxName"

	DefaultWaitTimeout  = 30 * time.Second
	DefaultWaitInterval = 1 * time.Second

	OperatorNamespace = "tyk-operator-system"
	GatewayLocalhost  = "http://localhost:7000"

	TestApiDef      = "test-http"
	TestOperatorCtx = "mycontext"
)
