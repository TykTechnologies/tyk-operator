package gateway_client

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	v1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/pkg/universal_client"
)

var (
	apiCollisionError = errors.New("api id, listen path or slug collision")
)

type Api struct {
	*Client
}

func (a Api) All(ctx context.Context) ([]v1.APIDefinitionSpec, error) {
	res, err := a.Client.Get(a.Env.JoinURL(endpointAPIs), nil)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API Returned error: %d", res.StatusCode)
	}

	var list []v1.APIDefinitionSpec
	if err := universal_client.JSON(res, &list); err != nil {
		return nil, err
	}
	return list, nil
}

func (a Api) Get(ctx context.Context, apiID string) (*v1.APIDefinitionSpec, error) {
	res, err := a.Client.Get(a.Env.JoinURL(endpointAPIs, apiID), nil)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gateway API Returned error: %d", res.StatusCode)
	}
	var spec v1.APIDefinitionSpec
	if err := universal_client.JSON(res, &spec); err != nil {
		return nil, err
	}
	return &spec, nil
}

func (a Api) Create(ctx context.Context, def *v1.APIDefinitionSpec) error {
	res, err := a.PostJSON(a.Env.JoinURL(endpointAPIs), def)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	var resMsg ResponseMsg
	if err := universal_client.JSON(res, &resMsg); err != nil {
		return err
	}

	if resMsg.Status != "ok" {
		return fmt.Errorf("API request completed, but with error: %s", resMsg.Message)
	}
	return nil
}

func (a Api) Update(ctx context.Context, def *v1.APIDefinitionSpec) error {
	res, err := a.PutJSON(a.Env.JoinURL(endpointAPIs, def.APIID), def)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	var resMsg ResponseMsg
	if err := universal_client.JSON(res, &resMsg); err != nil {
		return err
	}
	if resMsg.Status != "ok" {
		return fmt.Errorf("API request completed, but with error: %s", resMsg.Message)
	}
	return nil
}

func (a Api) Delete(ctx context.Context, id string) error {
	res, err := a.Client.Delete(a.Env.JoinURL(endpointAPIs, id), nil)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	var resMsg ResponseMsg
	if err := universal_client.JSON(res, &resMsg); err != nil {
		return err
	}
	if resMsg.Status != "ok" {
		return fmt.Errorf("API request completed, but with error: %s", resMsg.Message)
	}
	return nil
}
