package gateway

import (
	"context"

	"github.com/TykTechnologies/tyk-operator/api/model"
	"github.com/TykTechnologies/tyk-operator/pkg/client"
	"github.com/TykTechnologies/tyk-operator/pkg/client/universal"
)

type Portal struct{}

func (Portal) Policy() universal.Policy {
	return SecurityPolicy{}
}

func (Portal) Documentation() universal.Documentation {
	return Documentation{}
}

func (Portal) Catalogue() universal.Catalogue {
	return Catalogue{}
}

type Documentation struct{}

func (Documentation) Upload(
	ctx context.Context, o *model.APIDocumentation,
) (*model.Result, error) {
	return nil, client.ErrTODO
}

func (Documentation) Delete(ctx context.Context, id string) (*model.Result, error) {
	return nil, client.ErrTODO
}

var _ universal.Catalogue = Catalogue{}

type Catalogue struct{}

func (Catalogue) Get(ctx context.Context) (*model.APICatalogue, error) {
	return nil, client.ErrTODO
}

func (Catalogue) Create(ctx context.Context, o *model.APICatalogue) (*model.Result, error) {
	return nil, client.ErrTODO
}

func (Catalogue) Update(ctx context.Context, o *model.APICatalogue) (*model.Result, error) {
	return nil, client.ErrTODO
}
