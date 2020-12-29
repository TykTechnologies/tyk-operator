package dashboard_client

import (
	"fmt"
	"net/http"

	tykv1alpha1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/pkg/universal_client"
)

type Api struct {
	*Client
}

func (a Api) All() ([]tykv1alpha1.APIDefinitionSpec, error) {
	res, err := a.Client.Get(a.Env.JoinURL(endpointAPIs), nil,
		universal_client.AddQuery(map[string]string{
			"p": "-2",
		}),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API Returned error: %d", res.StatusCode)
	}

	var apisResponse ApisResponse
	if err := universal_client.JSON(res, &apisResponse); err != nil {
		return nil, err
	}

	var list []tykv1alpha1.APIDefinitionSpec
	for _, api := range apisResponse.Apis {
		list = append(list, api.ApiDefinition)
	}
	return list, nil
}

func (a Api) Create(_ string, def *tykv1alpha1.APIDefinitionSpec) error {
	def.OrgID = a.Env.Org
	dashboardAPIRequest := DashboardApi{
		ApiDefinition: *def,
	}
	res, err := a.Client.PostJSON(a.Env.JoinURL(endpointAPIs), dashboardAPIRequest)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return universal_client.Error(res)
	}

	var resMsg ResponseMsg
	if err := universal_client.JSON(res, &resMsg); err != nil {
		return err
	}
	if resMsg.Status != "OK" {
		return fmt.Errorf("API request completed, but with error: %s", resMsg.Message)
	}
	def.APIID = resMsg.Meta
	return nil
}

func (a Api) Get(id string) (*tykv1alpha1.APIDefinitionSpec, error) {
	res, err := a.Client.Get(a.Env.JoinURL(endpointAPIs, id), nil)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, universal_client.Error(res)
	}
	var resMsg DashboardApi
	if err := universal_client.JSON(res, &resMsg); err != nil {
		return nil, err
	}
	return &resMsg.ApiDefinition, nil
}

func (a Api) Update(def *tykv1alpha1.APIDefinitionSpec) error {
	id := def.ID
	if id == "" {
		id = def.APIID
	}
	res, err := a.Client.PutJSON(
		a.Env.JoinURL(endpointAPIs, id), DashboardApi{
			ApiDefinition: *def,
		},
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return universal_client.Error(res)
	}

	var resMsg ResponseMsg
	if err := universal_client.JSON(res, &resMsg); err != nil {
		return err
	}
	if resMsg.Status != "OK" {
		return fmt.Errorf("API request completed, but with error: %s", resMsg.Message)
	}
	return nil
}

func (a Api) Delete(id string) error {
	res, err := a.Client.Delete(a.Env.JoinURL(endpointAPIs, id), nil)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	switch res.StatusCode {
	case http.StatusOK, http.StatusNotFound:
		return nil
	default:
		return universal_client.Error(res)
	}
}
