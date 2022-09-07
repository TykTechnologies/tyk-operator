package dashboard

import (
	"context"

	"github.com/TykTechnologies/tyk-operator/api/model"
	"github.com/TykTechnologies/tyk-operator/pkg/client"
)

const endpointAPIs = "/api/apis"

type Api struct{}

func (a Api) Create(ctx context.Context, def *model.APIDefinitionSpec) (*model.Result, error) {
	_, err := a.Get(ctx, def.APIID)
	if err == nil {
		return a.update(ctx, &model.Result{Meta: def.APIID}, def)
	}

	var o model.Result

	octx := client.GetContext(ctx)

	octx.Log.Info("create request", "body", DashboardApi{
		ApiDefinition:   *def,
		UserOwners:      octx.Env.UserOwners,
		UserGroupOwners: octx.Env.UserGroupOwners,
	})

	err = client.Data(&o)(client.PostJSON(ctx, endpointAPIs,
		DashboardApi{
			ApiDefinition:   *def,
			UserOwners:      octx.Env.UserOwners,
			UserGroupOwners: octx.Env.UserGroupOwners,
		}))
	if err != nil {
		return nil, err
	}

	api, err := a.Get(ctx, o.Meta)
	if err != nil {
		return nil, err
	}

	api.APIID = def.APIID

	return a.update(ctx, &o, api)
}

func (a Api) Get(ctx context.Context, id string) (*model.APIDefinitionSpec, error) {
	var o DashboardApi

	err := client.Data(&o)(client.Get(
		ctx, client.Join(endpointAPIs, id), nil,
	))
	if err != nil {
		return nil, err
	}

	return &o.ApiDefinition, nil
}

func (a Api) Update(ctx context.Context, spec *model.APIDefinitionSpec) (*model.Result, error) {
	var o model.Result

	octx := client.GetContext(ctx)

	err := client.Data(&o)(client.PutJSON(
		ctx, client.Join(endpointAPIs, spec.APIID), DashboardApi{
			ApiDefinition:   *spec,
			UserOwners:      octx.Env.UserOwners,
			UserGroupOwners: octx.Env.UserGroupOwners,
		},
	))
	if err != nil {
		return nil, err
	}

	return &o, nil
}

func (a Api) update(ctx context.Context, result *model.Result, spec *model.APIDefinitionSpec) (*model.Result, error) {
	var o model.Result

	err := client.Data(&o)(client.PutJSON(
		ctx, client.Join(endpointAPIs, result.Meta), DashboardApi{
			ApiDefinition: *spec,
		},
	))
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (a Api) Delete(ctx context.Context, id string) (*model.Result, error) {
	var o model.Result

	err := client.Data(&o)(
		client.Delete(ctx, client.Join(endpointAPIs, id), nil),
	)
	if err != nil {
		return nil, err
	}

	return &o, nil
}

// List list all apis in the dashboard. options controls filtering and sorting
func (Api) List(
	ctx context.Context,
	options ...model.ListAPIOptions,
) (*model.APIDefinitionSpecList, error) {
	opts := model.ListAPIOptions{Pages: -2}

	if len(options) > 0 {
		opts = options[0]
	}

	var o ApisResponse

	err := client.Data(&o)(client.Get(ctx, endpointAPIs, nil,
		client.AddQuery(opts.Params()),
	))
	if err != nil {
		return nil, err
	}

	var a model.APIDefinitionSpecList

	for _, v := range o.Apis {
		v := v
		a.Apis = append(a.Apis, &v.ApiDefinition)
	}

	return &a, nil
}
