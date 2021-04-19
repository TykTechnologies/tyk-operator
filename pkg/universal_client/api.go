package universal_client

import (
	"context"

	v1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
)

type UniversalApi interface {
	Get(ctx context.Context, apiID string) (*v1.APIDefinitionSpec, error)
	All(context.Context) ([]v1.APIDefinitionSpec, error)
	Create(context.Context, *v1.APIDefinitionSpec) error
	Update(context.Context, *v1.APIDefinitionSpec) error
	Delete(ctx context.Context, id string) error
}
