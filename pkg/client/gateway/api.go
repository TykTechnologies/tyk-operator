package gateway

import (
	"context"
	"errors"

	"github.com/TykTechnologies/tyk-operator/api/model"
	"github.com/TykTechnologies/tyk-operator/pkg/client"
)

var (
	apiCollisionError = errors.New("api id, listen path or slug collision")
)

type Api struct{}

func (a Api) List(ctx context.Context, options ...model.ListAPIOptions) (*model.APIDefinitionSpecList, error) {
	var o model.APIDefinitionSpecList
	err := client.Data(&o.Apis)(client.Get(ctx, endpointAPIs, nil))
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (a Api) Get(ctx context.Context, apiID string) (*model.APIDefinitionSpec, error) {
	var spec model.APIDefinitionSpec
	err := client.Data(&spec)(client.Get(ctx, client.Join(endpointAPIs, apiID), nil))
	if err != nil {
		return nil, err
	}
	return &spec, nil
}

func (a Api) Create(ctx context.Context, def *model.APIDefinitionSpec) (*model.Result, error) {
	return a.createOrUpdate(ctx, def)
}

func (a Api) createOrUpdate(ctx context.Context, def *model.APIDefinitionSpec) (*model.Result, error) {
	var o model.Result
	err := client.Data(&o)(client.PostJSON(ctx, endpointAPIs, def))
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (a Api) Update(ctx context.Context, def *model.APIDefinitionSpec) (*model.Result, error) {
	return a.createOrUpdate(ctx, def)
}

func (a Api) Delete(ctx context.Context, id string) (*model.Result, error) {
	var o model.Result
	err := client.Data(&o)(client.Delete(ctx, client.Join(endpointAPIs, id), nil))
	if err != nil {
		return nil, err
	}
	return &o, nil
}
