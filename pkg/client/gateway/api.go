package gateway

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	v1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/pkg/client"
)

var (
	apiCollisionError = errors.New("api id, listen path or slug collision")
)

type Api struct{}

func (a Api) All(ctx context.Context) ([]v1.APIDefinitionSpec, error) {
	res, err := client.Get(ctx, endpointAPIs, nil)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API Returned error: %d", res.StatusCode)
	}

	var list []v1.APIDefinitionSpec
	if err := client.JSON(res, &list); err != nil {
		return nil, err
	}
	return list, nil
}

func (a Api) Get(ctx context.Context, apiID string) (*v1.APIDefinitionSpec, error) {
	res, err := client.Get(ctx, client.Join(endpointAPIs, apiID), nil)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gateway API Returned error: %d", res.StatusCode)
	}
	var spec v1.APIDefinitionSpec
	if err := client.JSON(res, &spec); err != nil {
		return nil, err
	}
	return &spec, nil
}

func (a Api) Create(ctx context.Context, def *v1.APIDefinitionSpec) error {
	res, err := client.PostJSON(ctx, endpointAPIs, def)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	var resMsg ResponseMsg
	if err := client.JSON(res, &resMsg); err != nil {
		return err
	}

	if resMsg.Status != "ok" {
		return fmt.Errorf("API request completed, but with error: %s", resMsg.Message)
	}
	return nil
}

func (a Api) Update(ctx context.Context, def *v1.APIDefinitionSpec) error {
	res, err := client.PutJSON(ctx, client.Join(endpointAPIs, def.APIID), def)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	var resMsg ResponseMsg
	if err := client.JSON(res, &resMsg); err != nil {
		return err
	}
	if resMsg.Status != "ok" {
		return fmt.Errorf("API request completed, but with error: %s", resMsg.Message)
	}
	return nil
}

func (a Api) Delete(ctx context.Context, id string) error {
	res, err := client.Delete(ctx, client.Join(endpointAPIs, id), nil)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	var resMsg ResponseMsg
	if err := client.JSON(res, &resMsg); err != nil {
		return err
	}
	if resMsg.Status != "ok" {
		return fmt.Errorf("API request completed, but with error: %s", resMsg.Message)
	}
	return nil
}
