package dashboard_client

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/pkg/environmet"
	"github.com/TykTechnologies/tyk-operator/pkg/universal_client"
	ctrl "sigs.k8s.io/controller-runtime"
)

const portalCataloguePolicyID = "607f48e9d3626e691e800102"

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

func TestPortalApi_impl(t *testing.T) {
	e := environmet.Env{
		//Namespace:          "",
		Mode:               "pro",
		InsecureSkipVerify: true,
		URL:                "http://localhost:3000",
		Auth:               "0f8c2d38d42a46434f597a3ecab48d08",
		Org:                "5d67b96d767e02015ea84a6f",
		IngressClass:       "tyk",
	}
	e = env().Merge(e)
	h := mockDash(t,
		&route{
			path:   "/api/portal/catalogue",
			method: http.MethodPost,
			body:   "portalCatalogue.Create.body",
		},
		&route{
			path:   "/api/portal/catalogue",
			method: http.MethodGet,
			body:   "portalCatalogue.All.body",
		},
		&route{
			path:   "/api/portal/catalogue/5fd202b669710900018bc19c",
			method: http.MethodGet,
			body:   "portalCatalogue.Get.body",
		},
		&route{
			path:   "/api/portal/catalogue/5fd202b669710900018bc19c",
			method: http.MethodPut,
			body:   "portalCatalogue.Update.body",
		},
		&route{
			path:   "/api/portal/catalogue/5fd202b669710900018bc19c",
			method: http.MethodDelete,
			body:   "portalCatalogue.Delete.body",
		},
	)
	svr := httptest.NewServer(h)
	defer svr.Close()
	e.URL = svr.URL
	requestPortal(t, e, Kase{
		Name: "All",
		Request: RequestKase{
			Path:   endpointPortalCatalogue,
			Method: http.MethodGet,
			Headers: map[string]string{
				XAuthorization: e.Auth,
				XContentType:   contentJSON,
			},
		},
		Response: &ResponseKase{
			Body: ReadSample(t, "portalCatalogue.All.body"),
		},
	})
}

func requestPortal(t *testing.T, e environmet.Env, kase universal_client.Kase) {
	t.Helper()
	switch kase.Name {
	case "All":
		universal_client.RunRequestKase(t, e,
			func(c universal_client.Client) error {
				newKlient(c).PortalCatalogue().All()
				return nil
			},
			kase,
		)
	case "Get":
		universal_client.RunRequestKase(t, e,
			func(c universal_client.Client) error {
				newKlient(c).PortalCatalogue().Get(portalCataloguePolicyID)
				return nil
			},
			kase,
		)
	case "Update":
		universal_client.RunRequestKase(t, e,
			func(c universal_client.Client) error {
				var s v1alpha1.PortalAPISpec
				Sample(t, "portalCatalogue."+kase.Name, &s)
				newKlient(c).PortalCatalogue().Update(&s)
				return nil
			},
			kase,
		)
	case "Create":
		universal_client.RunRequestKase(t, e,
			func(c universal_client.Client) error {
				var s v1alpha1.PortalAPISpec
				Sample(t, "portalCatalogue."+kase.Name, &s)
				newKlient(c).PortalCatalogue().Create(&s)
				return nil
			},
			kase,
		)
	case "Delete":
		universal_client.RunRequestKase(t, e,
			func(c universal_client.Client) error {
				newKlient(c).PortalCatalogue().Delete(portalCataloguePolicyID)
				return nil
			},
			kase,
		)
	}
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
