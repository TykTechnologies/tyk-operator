package dashboard_client

import (
	"testing"

	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/pkg/environmet"
	"github.com/TykTechnologies/tyk-operator/pkg/universal_client"
	ctrl "sigs.k8s.io/controller-runtime"
)

func getClient() universal_client.UniversalClient {
	return newKlient(universal_client.Client{
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
}

func TestPortalApi_All(t *testing.T) {
	t.SkipNow()
	c := getClient()
	apiSpec, err := c.PortalCatalogue().All()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%#v", apiSpec)
}

func TestPortalApi_Get(t *testing.T) {
	t.SkipNow()
	c := getClient()
	apiSpec, err := c.PortalCatalogue().Get("607f48e9d3626e691e800102")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%#v", apiSpec)
}

func TestPortalApi_Create(t *testing.T) {
	t.SkipNow()
	c := getClient()
	err := c.PortalCatalogue().Create(&v1alpha1.PortalAPISpec{
		Name:             "Foo",
		ShortDescription: "Bar",
		LongDescription:  "Baz",
		Show:             true,
		PolicyID:         "abcde",
		Version:          "v2",
		IsKeyless:        false,
		Config:           v1alpha1.PortalAPIConfig{},
		Fields:           nil,
		AuthType:         "authToken",
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestPortalApi_Update(t *testing.T) {
	t.SkipNow()
	c := getClient()
	err := c.PortalCatalogue().Update(&v1alpha1.PortalAPISpec{
		Name:             "Bar",
		ShortDescription: "Baz",
		LongDescription:  "Foo",
		Show:             true,
		PolicyID:         "abcde",
		Version:          "v2",
		IsKeyless:        false,
		Config:           v1alpha1.PortalAPIConfig{},
		Fields:           nil,
		AuthType:         "authToken",
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestPortalApi_Delete(t *testing.T) {
	t.SkipNow()
	c := getClient()
	err := c.PortalCatalogue().Delete("abcde")
	if err != nil {
		t.Fatal(err)
	}
}
