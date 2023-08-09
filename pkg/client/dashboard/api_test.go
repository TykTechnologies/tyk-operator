package dashboard

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TykTechnologies/tyk-operator/api/model"
	"github.com/TykTechnologies/tyk-operator/pkg/client"
	"github.com/TykTechnologies/tyk-operator/pkg/environment"
)

const contentJSON = "application/json"

const testAPIID = "ZGVmYXVsdC9odHRwYmlu"

func TestAPI(t *testing.T) {
	var e environment.Env

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
			path:   "/api/apis/ZGVmYXVsdC9odHRwYmlu",
			method: http.MethodGet,
			body:   "api.Get.body",
		},
		&route{
			path:   "/api/apis/5fd08ed769710900018bc196",
			method: http.MethodPut,
			body:   "api.Update.body",
		},
		&route{
			path:   "/api/apis/ZGVmYXVsdC9odHRwYmlu",
			method: http.MethodPut,
			body:   "api.Update.body",
		},
		&route{
			path:   "/api/apis/ZGVmYXVsdC9odHRwYmlu",
			method: http.MethodDelete,
			body:   "api.Delete.body",
		},
	)
	svr := httptest.NewServer(h)

	defer svr.Close()

	e.URL = svr.URL
	e = env().Merge(e)
	requestAPI(t, &e, "Create",
		// TODO:(gernest) This only covers the case when an api already exists.
		// Add case of creating fresh new api
		Kase{
			Name: "Get",
			Request: RequestKase{
				Path:   "/api/apis/ZGVmYXVsdC9odHRwYmlu",
				Method: http.MethodGet,
				Headers: map[string]string{
					XAuthorization: e.Auth,
					XContentType:   contentJSON,
				},
			},
			Response: &ResponseKase{
				Body: ReadSample(t, "api.Get.body"),
			},
		},
		Kase{
			Name: "Update",
			Request: RequestKase{
				Path:   "/api/apis/ZGVmYXVsdC9odHRwYmlu",
				Method: http.MethodPut,
				Headers: map[string]string{
					XAuthorization: e.Auth,
					XContentType:   contentJSON,
				},
			},
			Response: &ResponseKase{
				Body: ReadSample(t, "api.Update.body"),
			},
		},
	)

	requestAPI(t, &e, "All", Kase{
		Name: "All",
		Request: RequestKase{
			Path:   "/api/apis",
			Method: http.MethodGet,
			Headers: map[string]string{
				XAuthorization: e.Auth,
				XContentType:   contentJSON,
			},
		},
		Response: &ResponseKase{
			Body: ReadSample(t, "api.All.body"),
		},
	})

	requestAPI(t, &e, "Get",
		Kase{
			Name: "All",
			Request: RequestKase{
				Path:   "/api/apis/ZGVmYXVsdC9odHRwYmlu",
				Method: http.MethodGet,
				Headers: map[string]string{
					XAuthorization: e.Auth,
					XContentType:   contentJSON,
				},
			},
			Response: &ResponseKase{
				Body: ReadSample(t, "api.Get.body"),
			},
		},
	)

	requestAPI(t, &e, "Update",
		Kase{
			Name: "Update",
			Request: RequestKase{
				Path:   "/api/apis/ZGVmYXVsdC9odHRwYmlu",
				Method: http.MethodPut,
				Headers: map[string]string{
					XAuthorization: e.Auth,
					XContentType:   contentJSON,
				},
			},
			Response: &ResponseKase{
				Body: ReadSample(t, "api.Update.body"),
			},
		})

	requestAPI(t, &e, "Delete",
		Kase{
			Name: "Delete",
			Request: RequestKase{
				Path:   "/api/apis/ZGVmYXVsdC9odHRwYmlu",
				Method: http.MethodDelete,
				Headers: map[string]string{
					XAuthorization: e.Auth,
					XContentType:   contentJSON,
				},
			},
			Response: &ResponseKase{
				Body: ReadSample(t, "api.Delete.body"),
			},
		})
}

func requestAPI(t *testing.T, e *environment.Env, name string, kase ...client.Kase) {
	t.Helper()
	t.Run(name, func(t *testing.T) {
		switch name {
		case "All":
			client.RunRequestKase(t, e,
				func(ctx context.Context) error {
					newKlient().Api().List(ctx)
					return nil
				},
				kase...,
			)
		case "Get":
			client.RunRequestKase(t, e,
				func(ctx context.Context) error {
					newKlient().Api().Get(ctx, testAPIID)
					return nil
				},
				kase...,
			)
		case "Update":
			client.RunRequestKase(t, e,
				func(ctx context.Context) error {
					var s model.APIDefinitionSpec
					Sample(t, "api."+name, &s)
					newKlient().Api().Update(ctx, &s)
					return nil
				},
				kase...,
			)
		case "Create":
			client.RunRequestKase(t, e,
				func(ctx context.Context) error {
					var s model.APIDefinitionSpec
					Sample(t, "api."+name, &s)
					newKlient().Api().Create(ctx, &s)
					return nil
				},
				kase...,
			)
		case "Delete":
			client.RunRequestKase(t, e,
				func(ctx context.Context) error {
					newKlient().Api().Delete(ctx, testAPIID)
					return nil
				},
				kase...,
			)
		}
	})
}
