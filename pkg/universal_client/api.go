package universal_client

import (
	"context"

	v1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
)

type UniversalApi interface {
	Get(ctx context.Context, apiID string) (*v1.APIDefinitionSpec, error)
	All(ctx context.Context) ([]v1.APIDefinitionSpec, error)
	Create(ctx context.Context, spec *v1.APIDefinitionSpec) error
	Update(ctx context.Context, def *v1.APIDefinitionSpec) error
	Delete(ctx context.Context, id string) error
}
