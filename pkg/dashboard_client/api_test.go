package dashboard_client

import (
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

var testAPI = v1alpha1.APIDefinitionSpec{
	ID: "test_api_id",
}

func TestAPI_requests(t *testing.T) {
	requestAPI(t, env, Kase{
		Name: "All",
		Path: "/api/apis",
		Headers: map[string]string{
			XAuthorization: env.Auth,
			XContentType:   contentJSON,
		},
	})
}

func requestAPI(t *testing.T, e environmet.Env, kase universal_client.Kase) {
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
				newKlient(c).Api().Get(testAPI.ID)
				return nil
			},
			kase,
		)
	}
}
