package dashboard

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/pkg/client"
	"github.com/TykTechnologies/tyk-operator/pkg/environmet"
)

const testSecurityPolicyID = "5fd202b669710900018bc19c"

func TestSecurityPolicy(t *testing.T) {
	var e environmet.Env
	e.Parse()
	e = env().Merge(e)
	h := mockDash(t,
		&route{
			path:   "/api/portal/policies",
			method: http.MethodPost,
			body:   "policy.Create.body",
		},
		&route{
			path:   "/api/portal/policies",
			method: http.MethodGet,
			body:   "policy.All.body",
		},
		&route{
			path:   "/api/portal/policies/5fd202b669710900018bc19c",
			method: http.MethodGet,
			body:   "policy.Get.body",
		},
		&route{
			path:   "/api/portal/policies/5fd202b669710900018bc19c",
			method: http.MethodPut,
			body:   "policy.Update.body",
		},
		&route{
			path:   "/api/portal/policies/5fd202b669710900018bc19c",
			method: http.MethodDelete,
			body:   "policy.Delete.body",
		},
	)
	svr := httptest.NewServer(h)
	defer svr.Close()
	e.URL = svr.URL

	requestSecurityPolicy(t, e, Kase{
		Name: "Create",
		Request: RequestKase{
			Path:   "/api/portal/policies",
			Method: http.MethodPost,
			Headers: map[string]string{
				XAuthorization: e.Auth,
				XContentType:   contentJSON,
			},
		},
		Response: &ResponseKase{
			Body: ReadSample(t, "policy.Create.body"),
		},
	})
	requestSecurityPolicy(t, e, Kase{
		Name: "All",
		Request: RequestKase{
			Path:   "/api/portal/policies",
			Method: http.MethodGet,
			Headers: map[string]string{
				XAuthorization: e.Auth,
				XContentType:   contentJSON,
			},
		},
		Response: &ResponseKase{
			Body: ReadSample(t, "policy.All.body"),
		},
	})

	requestSecurityPolicy(t, e, Kase{
		Name: "Get",
		Request: RequestKase{
			Path:   "/api/portal/policies/5fd202b669710900018bc19c",
			Method: http.MethodGet,
			Headers: map[string]string{
				XAuthorization: e.Auth,
				XContentType:   contentJSON,
			},
		},
		Response: &ResponseKase{
			Body: ReadSample(t, "policy.Get.body"),
		},
	})
	requestSecurityPolicy(t, e, Kase{
		Name: "Update",
		Request: RequestKase{
			Path:   "/api/portal/policies/5fd202b669710900018bc19c",
			Method: http.MethodPut,
			Headers: map[string]string{
				XAuthorization: e.Auth,
				XContentType:   contentJSON,
			},
		},
		Response: &ResponseKase{
			Body: ReadSample(t, "policy.Update.body"),
		},
	})
	requestSecurityPolicy(t, e, Kase{
		Name: "Delete",
		Request: RequestKase{
			Path:   "/api/portal/policies/5fd202b669710900018bc19c",
			Method: http.MethodDelete,
			Headers: map[string]string{
				XAuthorization: e.Auth,
				XContentType:   contentJSON,
			},
		},
		Response: &ResponseKase{
			Body: ReadSample(t, "policy.Delete.body"),
		},
	})

}

func requestSecurityPolicy(t *testing.T, e environmet.Env, kase client.Kase) {
	t.Helper()
	switch kase.Name {
	case "All":
		client.RunRequestKase(t, e,
			func(ctx context.Context) error {
				newKlient().Portal().Policy().All(ctx)
				return nil
			},
			kase,
		)
	case "Get":
		client.RunRequestKase(t, e,
			func(ctx context.Context) error {
				newKlient().Portal().Policy().Get(ctx, testSecurityPolicyID)
				return nil
			},
			kase,
		)
	case "Update":
		client.RunRequestKase(t, e,
			func(ctx context.Context) error {
				var s v1alpha1.SecurityPolicySpec
				Sample(t, "policy."+kase.Name, &s)
				newKlient().Portal().Policy().Update(ctx, &s)
				return nil
			},
			kase,
		)
	case "Create":
		client.RunRequestKase(t, e,
			func(ctx context.Context) error {
				var s v1alpha1.SecurityPolicySpec
				Sample(t, "policy."+kase.Name, &s)
				newKlient().Portal().Policy().Create(ctx, &s)
				return nil
			},
			kase,
		)
	case "Delete":
		client.RunRequestKase(t, e,
			func(ctx context.Context) error {
				newKlient().Portal().Policy().Delete(ctx, testSecurityPolicyID)
				return nil
			},
			kase,
		)
	}
}
