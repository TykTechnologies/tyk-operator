package universal

import (
	"context"

	"github.com/TykTechnologies/tyk-operator/api/model"
)

type OAS interface {
	Create(ctx context.Context, data string) (*model.Result, error)
	Update(ctx context.Context, id string, data string) error
	Delete(ctx context.Context, id string) error
}
