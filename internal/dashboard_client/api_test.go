package dashboard_client

import (
	"testing"

	v1 "github.com/TykTechnologies/tyk-operator/api/v1"
)

func TestApi_All(t *testing.T) {
	c := getClient()
	apis, err := c.Api().All()
	if err != nil {
		t.Fatal(err.Error())
	}

	for _, a := range apis {
		t.Log("api:", a.APIID, a.Slug, a.Proxy.ListenPath, a.OrgID)
	}
}

func TestApi_Create(t *testing.T) {
	c := getClient()

	t.Log("creating API")

	dashboardAPIRequest := DashboardApi{
		ApiDefinition: v1.APIDefinitionSpec{
			APIID:  "friendlyID",
			Name:   "sedky #k8s #ns:default #foo #bar #baz",
			Active: true,
			Proxy: v1.Proxy{
				ListenPath:      "/httpbin",
				TargetURL:       "https://httpbin.org",
				StripListenPath: true,
			},
			UseKeylessAccess: true,
			VersionData: v1.VersionData{
				DefaultVersion: "Default",
				NotVersioned:   true,
				Versions: map[string]v1.VersionInfo{
					"Default": {
						Name:             "Default",
						UseExtendedPaths: true,
					},
				},
			},
		},
	}

	apiId, err := c.Api().Create(&dashboardAPIRequest.ApiDefinition)
	if err != nil {
		t.Fatal(err)
	}

	inserted, err := c.Api().Get(apiId)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("orgID: %s, apiId: %s", inserted.OrgID, inserted.APIID)

	t.Log("cleanup")
	err = c.Api().Delete(apiId)
	if err != nil {
		t.Fatal(err)
	}
}

func TestApi_Update(t *testing.T) {

	c := getClient()

	t.Log("creating api for update")

	dashboardAPIRequest := DashboardApi{
		ApiDefinition: v1.APIDefinitionSpec{
			Name:   "api to update",
			Active: false,
			Proxy: v1.Proxy{
				ListenPath: "/api_to_update",
			},
			UseKeylessAccess: true,
			Auth: v1.AuthConfig{
				AuthHeaderName: "Authorization",
			},
			VersionData: v1.VersionData{
				Versions: map[string]v1.VersionInfo{
					"Default": {
						Name:             "Default",
						UseExtendedPaths: true,
					},
				},
			},
		},
	}

	apiId, err := c.Api().Create(&dashboardAPIRequest.ApiDefinition)
	if err != nil {
		t.Fatal(err)
	}

	inserted, err := c.Api().Get(apiId)
	if err != nil {
		t.Fatal(err)
	}

	// try to update the api
	inserted.Name = "updated api"
	inserted.Active = true
	if err := c.Api().Update(apiId, inserted); err != nil {
		t.Fatal(err)
	}

	updated, err := c.Api().Get(apiId)
	if err != nil {
		t.Fatal("well that sucks!")
	}

	if updated.Name != inserted.Name {
		t.Fatal("api name not updated")
	}

	if updated.Active != inserted.Active {
		t.Fatal("unable to activate api")
	}

	t.Log("cleanup")
	err = c.Api().Delete(apiId)
	if err != nil {
		t.Fatal(err)
	}
}
