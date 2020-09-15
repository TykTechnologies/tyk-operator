package dashboard_client

import (
	"testing"

	v1 "github.com/TykTechnologies/tyk-operator/api/v1"
)

func TestApi_All(t *testing.T) {
	c := getClient()
	apis, err := c.Api.All()
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

	apiId, err := c.Api.Create(&dashboardAPIRequest)
	if err != nil {
		t.Fatal(err)
	}

	inserted, err := c.Api.Get(apiId)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("orgID: %s, apiId: %s", inserted.ApiDefinition.OrgID, inserted.ApiDefinition.APIID)
}

func TestApi_Create_API_ID_Override(t *testing.T) {
	//t.Skip("PoC hook up with policies to see if overriding the API_ID sense.")

	c := getClient()

	newID := "NEED_A_SENSIBLE_CUSTOM_NAME_HERE"
	t.Log("creating API with ID", newID)

	dashboardAPIRequest := DashboardApi{
		ApiDefinition: v1.APIDefinitionSpec{
			APIID: newID,
			Name:  "override api id",
			//OrgID:            "5f5d48438e18ef0001fda615",
			Active: true,
			Proxy: v1.Proxy{
				ListenPath: "/override_api_id",
			},
			ListenPort:       0,
			Protocol:         "",
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

	apiId, err := c.Api.Create(&dashboardAPIRequest)
	if err != nil {
		t.Fatal(err)
	}

	inserted, err := c.Api.Get(apiId)
	if err != nil {
		t.Fatal(err)
	}

	// try to update the ID
	inserted.ApiDefinition.APIID = newID
	if err := c.Api.Update(apiId, inserted); err != nil {
		t.Fatal(err)
	}

	updated, err := c.Api.Get(newID)
	if err != nil {
		t.Fatal("new api id doesn't appear to have been set properly")
	}

	if updated.ApiDefinition.APIID != newID {
		t.Fatalf("expected api_id: %s, got: %s", newID, updated.ApiDefinition.APIID)
	}
}

func TestApi_Update(t *testing.T) {
	//t.Fatal("no test")
}

func TestApi_Delete(t *testing.T) {
	//t.Fatal("no test")
}
