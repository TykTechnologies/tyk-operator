package dashboard_client

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/pkg/environmet"
	"github.com/TykTechnologies/tyk-operator/pkg/universal_client"
)

var env = environmet.Env{
	Mode: "pro",
	URL:  "http://localhost:3000",
	Auth: "secret",
	Org:  "testing",
}

const contentJSON = "application/json"

const ID = "5fd08ed769710900018bc196"

func TestAPI_requests(t *testing.T) {
	var e environmet.Env
	e.Parse()
	h := mockDash(t,
		&route{
			path:   "/api/apis",
			method: http.MethodPost,
			body:   "api.Create.body",
		},
		&route{
			path:   "/api/apis",
			method: http.MethodGet,
			body:   "api.All.body",
		},
		&route{
			path:   "/api/apis/5fd08ed769710900018bc196",
			method: http.MethodGet,
			body:   "api.Get.body",
		},
		&route{
			path:   "/api/apis/5fd08ed769710900018bc196",
			method: http.MethodPut,
			body:   "api.Update.body",
		},
		&route{
			path:   "/api/apis/5fd08ed769710900018bc196",
			method: http.MethodDelete,
			body:   "api.Delete.body",
		},
	)
	svr := httptest.NewServer(h)
	defer svr.Close()
	e.URL = svr.URL
	env = env.Merge(e)
	requestAPI(t, env, Kase{
		Name: "Create",
		Request: RequestKase{
			Path:   "/api/apis",
			Method: http.MethodPost,
			Headers: map[string]string{
				XAuthorization: env.Auth,
				XContentType:   contentJSON,
			},
		},
		Response: &ResponseKase{
			Body: ReadSample(t, "api.Create.body"),
		},
	})
	requestAPI(t, env, Kase{
		Name: "All",
		Request: RequestKase{
			Path:   "/api/apis",
			Method: http.MethodGet,
			Headers: map[string]string{
				XAuthorization: env.Auth,
				XContentType:   contentJSON,
			},
		},
		Response: &ResponseKase{
			Body: ReadSample(t, "api.All.body"),
		},
	})
	requestAPI(t, env, Kase{
		Name: "Get",
		Request: RequestKase{
			Path:   "/api/apis/5fd08ed769710900018bc196",
			Method: http.MethodGet,
			Headers: map[string]string{
				XAuthorization: env.Auth,
				XContentType:   contentJSON,
			},
		},
		Response: &ResponseKase{
			Body: ReadSample(t, "api.Get.body"),
		},
	})
	requestAPI(t, env, Kase{
		Name: "Update",
		Request: RequestKase{
			Path:   "/api/apis/5fd08ed769710900018bc196",
			Method: http.MethodPut,
			Headers: map[string]string{
				XAuthorization: env.Auth,
				XContentType:   contentJSON,
			},
		},
		Response: &ResponseKase{
			Body: ReadSample(t, "api.Update.body"),
		},
	})

	requestAPI(t, env, Kase{
		Name: "Delete",
		Request: RequestKase{
			Path:   "/api/apis/5fd08ed769710900018bc196",
			Method: http.MethodDelete,
			Headers: map[string]string{
				XAuthorization: env.Auth,
				XContentType:   contentJSON,
			},
		},
		Response: &ResponseKase{
			Body: ReadSample(t, "api.Delete.body"),
		},
	})
}

func requestAPI(t *testing.T, e environmet.Env, kase universal_client.Kase) {
	t.Helper()
	switch kase.Name {
	case "All":
		universal_client.RunRequestKase(t, e,
			func(c universal_client.Client) error {
				newKlient(c).Api().All()
				return nil
			},
			kase,
		)
	case "Get":
		universal_client.RunRequestKase(t, e,
			func(c universal_client.Client) error {
				newKlient(c).Api().Get(ID)
				return nil
			},
			kase,
		)
	case "Update":
		universal_client.RunRequestKase(t, e,
			func(c universal_client.Client) error {
				var s v1alpha1.APIDefinitionSpec
				Sample(t, "api."+kase.Name, &s)
				newKlient(c).Api().Update(&s)
				return nil
			},
			kase,
		)
	case "Create":
		universal_client.RunRequestKase(t, e,
			func(c universal_client.Client) error {
				var s v1alpha1.APIDefinitionSpec
				Sample(t, "api."+kase.Name, &s)
				newKlient(c).Api().Create(&s)
				return nil
			},
			kase,
		)
	case "Delete":
		universal_client.RunRequestKase(t, e,
			func(c universal_client.Client) error {
				newKlient(c).Api().Delete(ID)
				return nil
			},
			kase,
		)
	}
}
