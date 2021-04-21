package dashboard_client

import (
	"fmt"
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

	var portalCatalogueResponse PortalCatalogueResponse
	if err := universal_client.JSON(res, &portalCatalogueResponse); err != nil {
		return nil, err
	}

	a.Log.Info("All api's", "count", len(portalCatalogueResponse.Apis))
	return portalCatalogueResponse.Apis, nil
}

func (a PortalApi) Get(policyID string) (*tykv1alpha1.PortalAPISpec, error) {
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

	var portalCatalogueResponse PortalCatalogueResponse
	if err := universal_client.JSON(res, &portalCatalogueResponse); err != nil {
		return nil, err
	}

	for _, catalogueEntry := range portalCatalogueResponse.Apis {
		if catalogueEntry.PolicyID == policyID {
			return &catalogueEntry, nil
		}
	}

	return nil, fmt.Errorf("policy with id %s not found", policyID)
}

func (a PortalApi) Create(def *tykv1alpha1.PortalAPISpec) error {
	res, err := a.Client.Get(a.Env.JoinURL(endpointPortalCatalogue), nil)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		return universal_client.ErrNotFound
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("API Returned error: %d", res.StatusCode)
	}

	var portalCatalogueResponse PortalCatalogueResponse
	if err := universal_client.JSON(res, &portalCatalogueResponse); err != nil {
		return err
	}

	for _, catalogueEntry := range portalCatalogueResponse.Apis {
		if catalogueEntry.PolicyID == def.PolicyID {
			return fmt.Errorf("unable to create catalogue api as policy with id %s already exists", def.PolicyID)
		}
	}
	// insert this catalogue entry into the catalogue
	portalCatalogueResponse.Apis = append(portalCatalogueResponse.Apis, *def)

	// put it back
	putRes, err := a.Client.PutJSON(a.Env.JoinURL(endpointPortalCatalogue), portalCatalogueResponse)
	if err != nil {
		return err
	}
	defer putRes.Body.Close()

	if putRes.StatusCode != http.StatusOK {
		return fmt.Errorf("API Returned error: %d", res.StatusCode)
	}

	return nil
}

func (a PortalApi) Update(spec *tykv1alpha1.PortalAPISpec) error {
	res, err := a.Client.Get(a.Env.JoinURL(endpointPortalCatalogue), nil)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		return universal_client.ErrNotFound
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("API Returned error: %d", res.StatusCode)
	}

	var portalCatalogueResponse PortalCatalogueResponse
	if err := universal_client.JSON(res, &portalCatalogueResponse); err != nil {
		return err
	}

	foundIndex := 0
	for i, catalogueEntry := range portalCatalogueResponse.Apis {
		if catalogueEntry.PolicyID == spec.PolicyID {
			foundIndex = i
		}
	}

	if foundIndex == 0 {
		return fmt.Errorf("unable to update catalogue with policy %s as it does not exist", spec.PolicyID)
	}

	portalCatalogueResponse.Apis[foundIndex] = *spec

	putRes, err := a.Client.PutJSON(a.Env.JoinURL(endpointPortalCatalogue), portalCatalogueResponse)
	if err != nil {
		return err
	}
	defer putRes.Body.Close()

	if putRes.StatusCode != http.StatusOK {
		return fmt.Errorf("API Returned error: %d", res.StatusCode)
	}

	return nil
}

func (a PortalApi) Delete(policyID string) error {
	res, err := a.Client.Get(a.Env.JoinURL(endpointPortalCatalogue), nil)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		return universal_client.ErrNotFound
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("API Returned error: %d", res.StatusCode)
	}

	var portalCatalogueResponse PortalCatalogueResponse
	if err := universal_client.JSON(res, &portalCatalogueResponse); err != nil {
		return err
	}

	foundIndex := 0
	for i, catalogueEntry := range portalCatalogueResponse.Apis {
		if catalogueEntry.PolicyID == policyID {
			foundIndex = i
		}
	}

	if foundIndex == 0 {
		return fmt.Errorf("unable to delete catalogue with policy %s as it does not exist", policyID)
	}

	portalCatalogueResponse.Apis = removeFromSliceByIndex(portalCatalogueResponse.Apis, foundIndex)

	putRes, err := a.Client.PutJSON(a.Env.JoinURL(endpointPortalCatalogue), portalCatalogueResponse)
	if err != nil {
		return err
	}
	defer putRes.Body.Close()

	if putRes.StatusCode != http.StatusOK {
		return fmt.Errorf("API Returned error: %d", res.StatusCode)
	}

	return nil
}

// order not important
func removeFromSliceByIndex(s []tykv1alpha1.PortalAPISpec, i int) []tykv1alpha1.PortalAPISpec {
	s[len(s)-1], s[i] = s[i], s[len(s)-1]
	return s[:len(s)-1]
}
