package dashboard

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/TykTechnologies/tyk-operator/pkg/client"
	"github.com/TykTechnologies/tyk-operator/pkg/environmet"
)

const testCertID = "5fd08e0f69710900018bc19568492b39a512286d3e71c4c673faa7f094ffef324d12bf3b485c295221e97150"

func TestCert(t *testing.T) {
	var e environmet.Env
	e.Parse()
	h := mockDash(t,
		&route{
			path:   "/api/certs",
			method: http.MethodGet,
			body:   "cert.All.body",
		},
		&route{
			path:   "/api/certs/5fd08e0f69710900018bc19568492b39a512286d3e71c4c673faa7f094ffef324d12bf3b485c295221e97150",
			method: http.MethodGet,
			body:   "cert.Exist.body",
		},
		&route{
			path:   "/api/certs",
			method: http.MethodPost,
			body:   "cert.Upload.body",
		},
	)
	svr := httptest.NewServer(h)
	defer svr.Close()
	e.URL = svr.URL
	e = env().Merge(e)
	requestCert(t, e, Kase{
		Name: "All",
		Request: RequestKase{
			Path:   "/api/certs",
			Method: http.MethodGet,
			Headers: map[string]string{
				XAuthorization: e.Auth,
				XContentType:   contentJSON,
			},
		},
		Response: &ResponseKase{
			Body: ReadSample(t, "cert.All.body"),
		},
	})
	requestCert(t, e, Kase{
		Name: "Exist",
		Request: RequestKase{
			Path:   "/api/certs/5fd08e0f69710900018bc19568492b39a512286d3e71c4c673faa7f094ffef324d12bf3b485c295221e97150",
			Method: http.MethodGet,
			Headers: map[string]string{
				XAuthorization: e.Auth,
				XContentType:   contentJSON,
			},
		},
		Response: &ResponseKase{
			Body: ReadSample(t, "cert.Exist.body"),
		},
	})
	requestCert(t, e, Kase{
		Name: "Upload",
		Request: RequestKase{
			Path:   "/api/certs",
			Method: http.MethodPost,
			Headers: map[string]string{
				XAuthorization: e.Auth,
			},
			Callback: func(t *testing.T, h *http.Request) {
				e := "multipart/form-data;"
				g := h.Header.Get(XContentType)
				if !strings.HasPrefix(g, e) {
					t.Errorf("expected %s header to have %q prefix got %q", XContentType, e, g)
				}
			},
		},
		Response: &ResponseKase{
			Body: ReadSample(t, "cert.Upload.body"),
		},
	})
}

func requestCert(t *testing.T, e environmet.Env, kase client.Kase) {
	t.Helper()
	switch kase.Name {
	case "All":
		client.RunRequestKase(t, e,
			func(ctx context.Context) error {
				newKlient().Certificate().All(ctx)
				return nil
			},
			kase,
		)
	case "Upload":
		client.RunRequestKase(t, e,
			func(ctx context.Context) error {
				key := ReadSampleFile(t, "cert.Key.pem")
				cert := ReadSampleFile(t, "cert.Cert.pem")
				newKlient().Certificate().Upload(ctx, []byte(key), []byte(cert))
				return nil
			},
			kase,
		)
	case "Exist":
		client.RunRequestKase(t, e,
			func(ctx context.Context) error {
				newKlient().Certificate().Exists(ctx, testCertID)
				return nil
			},
			kase,
		)
	}
}
