package dashboard

import (
	"context"

	"github.com/TykTechnologies/tyk-operator/api/model"
	"github.com/TykTechnologies/tyk-operator/pkg/client"
	"github.com/TykTechnologies/tyk-operator/pkg/client/universal"
)

const endpointConfiguration = "/api/portal/configuration"

var _ universal.Configuration = Configuration{}

type Configuration struct{}

func (Configuration) Get(ctx context.Context) (*model.PortalModelPortalConfig, error) {
	var o model.PortalModelPortalConfig

	err := client.Data(&o)(
		client.Get(ctx, endpointConfiguration, nil),
	)
	if err != nil {
		return nil, err
	}

	return &o, nil
}

func (Configuration) Create(
	ctx context.Context, o *model.PortalModelPortalConfig,
) (*model.Result, error) {
	return client.Result(client.PostJSON(
		ctx, endpointConfiguration, o,
	))
}

func (Configuration) Update(
	ctx context.Context, o *model.PortalModelPortalConfig,
) (*model.Result, error) {
	return client.Result(client.PutJSON(
		ctx, endpointConfiguration, o,
	))
}
