package dashboard

import (
	"context"

	"github.com/TykTechnologies/tyk-operator/api/model"
	"github.com/TykTechnologies/tyk-operator/pkg/client"
	"github.com/TykTechnologies/tyk-operator/pkg/client/universal"
)

const endpointCatalogue = "/api/portal/catalogue/"

var _ universal.Catalogue = Catalogue{}

type Catalogue struct{}

func (Catalogue) Get(ctx context.Context) (*model.APICatalogue, error) {
	var o model.APICatalogue
	err := client.Data(&o)(
		client.Get(ctx, endpointCatalogue, nil),
	)
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (Catalogue) Create(ctx context.Context, o *model.APICatalogue) (*model.Result, error) {
	return client.Result(client.PostJSON(
		ctx, endpointCatalogue, o,
	))
}

func (Catalogue) Update(ctx context.Context, o *model.APICatalogue) (*model.Result, error) {
	return client.Result(client.PutJSON(
		ctx, endpointCatalogue, o,
	))
}
