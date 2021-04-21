package dashboard_client

import (
	"testing"

	"github.com/TykTechnologies/tyk-operator/pkg/environmet"
	"github.com/TykTechnologies/tyk-operator/pkg/universal_client"
	ctrl "sigs.k8s.io/controller-runtime"
)

func TestPortalApi_All(t *testing.T) {
	c := newKlient(universal_client.Client{
		Env: environmet.Env{
			//Namespace:          "",
			Mode:               "pro",
			InsecureSkipVerify: true,
			URL:                "http://localhost:3000",
			Auth:               "0f8c2d38d42a46434f597a3ecab48d08",
			Org:                "5d67b96d767e02015ea84a6f",
			IngressClass:       "tyk",
		},
		Log:           ctrl.Log.WithName("controllers").WithName("ApiDefinition"),
		BeforeRequest: nil,
		Do:            nil,
	})
	apiSpec, err := c.PortalCatalogue().All()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%#v", apiSpec)
}
