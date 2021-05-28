package dashboard

import (
	"context"
	"fmt"
	"net/http"

	tykv1alpha1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/pkg/client"
)

type Api struct{}

func (a Api) All(ctx context.Context) ([]tykv1alpha1.APIDefinitionSpec, error) {
	res, err := client.Get(ctx, endpointAPIs, nil,
		client.AddQuery(map[string]string{
			"p": "-2",
		}),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		return nil, client.ErrNotFound
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API Returned error: %d", res.StatusCode)
	}

	var apisResponse ApisResponse
	if err := client.JSON(res, &apisResponse); err != nil {
		return nil, err
	}

	var list []tykv1alpha1.APIDefinitionSpec
	for _, api := range apisResponse.Apis {
		list = append(list, api.ApiDefinition)
	}
	client.LInfo(ctx, "All api's", "Count", len(list))
	return list, nil
}

func (a Api) spec(def *tykv1alpha1.APIDefinitionSpec) tykv1alpha1.APIDefinitionSpec {
	o := def.DeepCopy()
	o.Context = nil
	return *o
}

func (a Api) Create(ctx context.Context, def *tykv1alpha1.APIDefinitionSpec) error {
	res, err := client.PostJSON(ctx, endpointAPIs,
		DashboardApi{
			ApiDefinition: a.spec(def),
		})
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return client.Error(res)
	}

	var resMsg ResponseMsg
	if err := client.JSON(res, &resMsg); err != nil {
		return err
	}
	if resMsg.Status != "OK" {
		return fmt.Errorf("API request completed, but with error: %s", resMsg.Message)
	}
	o, err := a.get(ctx, resMsg.Meta)
	if err != nil {
		return err
	}
	o.APIID = def.APIID
	return a.update(ctx, *o)
}

func (a Api) Get(ctx context.Context, id string) (*tykv1alpha1.APIDefinitionSpec, error) {
	all, err := a.All(ctx)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(all); i++ {
		if all[i].APIID == id {
			return &all[i], nil
		}
	}
	return nil, client.ErrNotFound
}

func (a Api) get(ctx context.Context, id string) (*tykv1alpha1.APIDefinitionSpec, error) {
	res, err := client.Get(ctx, client.Join(endpointAPIs, id), nil)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, client.Error(res)
	}
	var resMsg DashboardApi
	if err := client.JSON(res, &resMsg); err != nil {
		return nil, err
	}
	return &resMsg.ApiDefinition, nil
}

func (a Api) Update(ctx context.Context, def *tykv1alpha1.APIDefinitionSpec) error {
	x, err := a.Get(ctx, def.APIID)
	if err != nil {
		return err
	}
	o := a.spec(def)
	o.ID = x.ID
	return a.update(ctx, o)
}

func (a Api) update(ctx context.Context, o tykv1alpha1.APIDefinitionSpec) error {
	res, err := client.PutJSON(
		ctx, client.Join(endpointAPIs, o.ID), DashboardApi{
			ApiDefinition: o,
		},
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return client.Error(res)
	}

	var resMsg ResponseMsg
	if err := client.JSON(res, &resMsg); err != nil {
		return err
	}
	if resMsg.Status != "OK" {
		return fmt.Errorf("API request completed, but with error: %s", resMsg.Message)
	}
	return nil
}

func (a Api) Delete(ctx context.Context, id string) error {
	x, err := a.Get(ctx, id)
	if err != nil {
		return client.IgnoreNotFound(err)
	}
	res, err := client.Delete(ctx, client.Join(endpointAPIs, x.ID), nil)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	switch res.StatusCode {
	case http.StatusOK, http.StatusNotFound:
		return nil
	default:
		return client.Error(res)
	}
}
