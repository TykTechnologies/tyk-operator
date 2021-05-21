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

type Documentation struct{}

func (Documentation) Upload(
	ctx context.Context, o *model.APIDocumentation,
) (*model.Result, error) {
	return nil, client.ErrTODO
}

func (Documentation) Delete(ctx context.Context, id string) (*model.Result, error) {
	return nil, client.ErrTODO
}
