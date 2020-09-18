package dashboard_client

import (
	"errors"
	"fmt"
	"net/http"

	v1 "github.com/TykTechnologies/tyk-operator/api/v1"
	"github.com/levigross/grequests"
)

var (
	apiCollisionError = errors.New("api id, listen path or slug collision")
)

type Api struct {
	*Client
}

func (a Api) All() ([]v1.APIDefinitionSpec, error) {
	fullPath := JoinUrl(a.url, endpointAPIs)

	// -2 means get all pages
	queryStruct := struct {
		Pages int `url:"p"`
	}{
		Pages: -2,
	}
	opts := a.opts
	opts.QueryStruct = queryStruct

	res, err := grequests.Get(fullPath, a.opts)
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

	var list []v1.APIDefinitionSpec
	for _, api := range apisResponse.Apis {
		list = append(list, api.ApiDefinition)
	}

	return list, nil
}

// TODO: logic to prevent create without an "api_id"
func (a Api) Create(def *v1.APIDefinitionSpec) (string, error) {
	// Create
	opts := a.opts

	dashboardAPIRequest := DashboardApi{
		ApiDefinition: *def,
	}

	opts.JSON = dashboardAPIRequest
	fullPath := JoinUrl(a.url, endpointAPIs)

	res, err := grequests.Post(fullPath, opts)
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

	inserted, err := a.Get(resMsg.Meta)
	if err != nil {
		return "", fmt.Errorf("API request (lookup) failed: %s", resMsg.Message)
	}

	def.OrgID = inserted.OrgID
	err = a.Update(resMsg.Meta, def)
	if resMsg.Status != "OK" {
		return "", fmt.Errorf("API request (update name) failed: %s", resMsg.Message)
	}

	return resMsg.Meta, nil
}

func (a Api) Get(apiID string) (*v1.APIDefinitionSpec, error) {
	// Create
	opts := a.opts
	fullPath := JoinUrl(a.url, endpointAPIs, apiID)

	res, err := grequests.Get(fullPath, opts)
	if err != nil {
		return nil, err
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

func (a Api) Update(apiID string, def *v1.APIDefinitionSpec) error {
	// Update
	dashboardAPIRequest := DashboardApi{
		ApiDefinition: *def,
	}

	opts := a.opts
	opts.JSON = dashboardAPIRequest
	fullPath := JoinUrl(a.url, endpointAPIs, apiID)

	res, err := grequests.Put(fullPath, opts)
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
	delPath := JoinUrl(a.url, endpointAPIs, id)

	res, err := grequests.Delete(delPath, a.opts)
	if err != nil {
		return err
	}

	if res.StatusCode == http.StatusOK {
		return nil
	}

	if res.StatusCode == http.StatusInternalServerError {
		// Tyk returns internal server error if api is already deleted
		return nil
	}

	return fmt.Errorf("API Returned error: %s", res.String())
}
