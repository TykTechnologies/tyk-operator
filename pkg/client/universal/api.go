package universal

import (
	"context"

	"github.com/TykTechnologies/tyk-operator/api/model"
)

type Api interface {
	Create(ctx context.Context, def *model.APIDefinitionSpec) (*model.Result, error)
	Get(ctx context.Context, id string) (*model.APIDefinitionSpec, error)
	Update(ctx context.Context, spec *model.APIDefinitionSpec) (*model.Result, error)
	Delete(ctx context.Context, id string) (*model.Result, error)
	List(ctx context.Context, options ...model.ListAPIOptions) (*model.APIDefinitionSpecList, error)
}
