package dashboard_client

import (
	"fmt"
	"io/ioutil"
	"net/http"

	tykv1alpha1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/pkg/universal_client"
)

type PortalApi struct {
	*Client
}

func (a PortalApi) All() ([]tykv1alpha1.PortalAPISpec, error) {
	res, err := a.Client.Get(a.Env.JoinURL(endpointPortalCatalogue), nil)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		return nil, universal_client.ErrNotFound
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API Returned error: %d", res.StatusCode)
	}

	copyRes, err := ioutil.ReadAll(res.Body)
	ioutil.WriteFile("get.json", copyRes, 0600)

	var portalCatalogueResponse PortalCatalogueResponse
	if err := universal_client.JSON(res, &portalCatalogueResponse); err != nil {
		return nil, err
	}

	a.Log.Info("All api's", "count", len(portalCatalogueResponse.Apis))
	return portalCatalogueResponse.Apis, nil
}

func (a PortalApi) Create(def *tykv1alpha1.PortalAPISpec) error {
	panic("implement me")
}

func (a PortalApi) Get(id string) (*tykv1alpha1.PortalAPISpec, error) {
	panic("implement me")
}

func (a PortalApi) Update(def *tykv1alpha1.PortalAPISpec) error {
	panic("implement me")
}

func (a PortalApi) Delete(id string) error {
	panic("implement me")
}
