package dashboard_client

import (
	"fmt"
	"net/http"

	tykv1alpha1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/levigross/grequests"
)

type Api struct {
	*Client
}

func (a Api) All() ([]tykv1alpha1.APIDefinitionSpec, error) {
	sess := grequests.NewSession(a.opts())

	fullPath := a.env.JoinURL(endpointAPIs)

	// -2 means get all pages
	queryStruct := struct {
		Pages int `url:"p"`
	}{
		Pages: -2,
	}

	sess.RequestOptions.QueryStruct = queryStruct

	res, err := sess.Get(fullPath, nil)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API Returned error: %d", res.StatusCode)
	}

	var apisResponse ApisResponse
	if err := res.JSON(&apisResponse); err != nil {
		return nil, err
	}

	var list []tykv1alpha1.APIDefinitionSpec
	for _, api := range apisResponse.Apis {
		list = append(list, api.ApiDefinition)
	}

	return list, nil
}

func (a Api) Create(def *tykv1alpha1.APIDefinitionSpec) (string, error) {
	// Create
	sess := grequests.NewSession(a.opts())

	def.OrgID = a.env.Org
	dashboardAPIRequest := DashboardApi{
		ApiDefinition: *def,
	}

	fullPath := a.env.JoinURL(endpointAPIs)

	res, err := sess.Post(fullPath, &grequests.RequestOptions{JSON: dashboardAPIRequest})
	if err != nil {
		return "", err
	}

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API Returned error: %v (code: %v)", res.String(), res.StatusCode)
	}

	var resMsg ResponseMsg
	if err := res.JSON(&resMsg); err != nil {
		return "", err
	}

	if resMsg.Status != "OK" {
		return "", fmt.Errorf("API request completed, but with error: %s", resMsg.Message)
	}

	return resMsg.Meta, nil
}

func (a Api) Get(apiID string) (*tykv1alpha1.APIDefinitionSpec, error) {
	if apiID == "" {
		return nil, nil
	}

	// Create
	sess := grequests.NewSession(a.opts())
	fullPath := a.env.JoinURL(endpointAPIs, apiID)

	res, err := sess.Get(fullPath, nil)
	if err != nil {
		return nil, err
	}

	// Todo, hacky because we dont know best way to show API not found
	if res.StatusCode == http.StatusBadRequest {
		return nil, nil
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API Returned error: %v (code: %v)", res.String(), res.StatusCode)
	}

	var resMsg DashboardApi
	if err := res.JSON(&resMsg); err != nil {
		return nil, err
	}

	return &resMsg.ApiDefinition, nil
}

func (a Api) Update(apiID string, def *tykv1alpha1.APIDefinitionSpec) error {
	// Update
	dashboardAPIRequest := DashboardApi{
		ApiDefinition: *def,
	}

	sess := grequests.NewSession(a.opts())

	fullPath := a.env.JoinURL(endpointAPIs, apiID)

	res, err := sess.Put(fullPath, &grequests.RequestOptions{JSON: dashboardAPIRequest})
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("API Returned error: %v (code: %v)", res.String(), res.StatusCode)
	}

	var resMsg ResponseMsg
	if err := res.JSON(&resMsg); err != nil {
		return err
	}
	if resMsg.Status != "OK" {
		return fmt.Errorf("API request completed, but with error: %s", resMsg.Message)
	}
	return nil
}

func (a Api) Delete(id string) error {
	delPath := a.env.JoinURL(endpointAPIs, id)
	sess := grequests.NewSession(a.opts())
	res, err := sess.Delete(delPath, nil)
	if err != nil {
		return err
	}

	if res.StatusCode == http.StatusOK {
		return nil
	}

	if res.StatusCode == http.StatusNotFound {
		// Tyk returns 404 if api is already deleted
		return nil
	}
	return fmt.Errorf("API Returned error: %s", res.String())
}
