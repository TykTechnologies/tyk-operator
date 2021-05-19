package dashboard

import (
	"context"
	"fmt"
	"net/http"

	tykv1alpha1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/pkg/client/universal"
)

type Api struct {
	*Client
}

func (a Api) All(ctx context.Context) ([]tykv1alpha1.APIDefinitionSpec, error) {
	res, err := a.Client.Get(a.Env.JoinURL(endpointAPIs), nil,
		universal.AddQuery(map[string]string{
			"p": "-2",
		}),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		return nil, universal.ErrNotFound
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API Returned error: %d", res.StatusCode)
	}

	var apisResponse ApisResponse
	if err := universal.JSON(res, &apisResponse); err != nil {
		return nil, err
	}

	var list []tykv1alpha1.APIDefinitionSpec
	for _, api := range apisResponse.Apis {
		list = append(list, api.ApiDefinition)
	}
	a.Log.Info("All api's", "Count", len(list))
	return list, nil
}

func (a Api) Create(ctx context.Context, def *tykv1alpha1.APIDefinitionSpec) error {
	res, err := a.Client.PostJSON(a.Env.JoinURL(endpointAPIs),
		DashboardApi{
			ApiDefinition: *def,
		})
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return universal.Error(res)
	}

	var resMsg ResponseMsg
	if err := universal.JSON(res, &resMsg); err != nil {
		return err
	}
	if resMsg.Status != "OK" {
		return fmt.Errorf("API request completed, but with error: %s", resMsg.Message)
	}
	o, err := a.get(resMsg.Meta)
	if err != nil {
		return err
	}
	o.APIID = def.APIID
	return a.update(*o)
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
	return nil, universal.ErrNotFound
}

func (a Api) get(id string) (*tykv1alpha1.APIDefinitionSpec, error) {
	res, err := a.Client.Get(a.Env.JoinURL(endpointAPIs, id), nil)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, universal.Error(res)
	}
	var resMsg DashboardApi
	if err := universal.JSON(res, &resMsg); err != nil {
		return nil, err
	}
	return &resMsg.ApiDefinition, nil
}

func (a Api) Update(ctx context.Context, def *tykv1alpha1.APIDefinitionSpec) error {
	x, err := a.Get(ctx, def.APIID)
	if err != nil {
		return err
	}
	o := *def
	o.ID = x.ID
	return a.update(o)
}

func (a Api) update(o tykv1alpha1.APIDefinitionSpec) error {
	res, err := a.Client.PutJSON(
		a.Env.JoinURL(endpointAPIs, o.ID), DashboardApi{
			ApiDefinition: o,
		},
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return universal.Error(res)
	}

	var resMsg ResponseMsg
	if err := universal.JSON(res, &resMsg); err != nil {
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
		return universal.IgnoreNotFound(err)
	}
	res, err := a.Client.Delete(a.Env.JoinURL(endpointAPIs, x.ID), nil)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	switch res.StatusCode {
	case http.StatusOK, http.StatusNotFound:
		return nil
	default:
		return universal.Error(res)
	}
}
