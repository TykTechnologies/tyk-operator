package universal

import (
	"context"

	"github.com/TykTechnologies/tyk-operator/api/model"
	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
)

type Policy interface {
	All(ctx context.Context) ([]v1alpha1.SecurityPolicySpec, error)
	// Get retruns the policy with the given id.
	Get(ctx context.Context, id string) (*v1alpha1.SecurityPolicySpec, error)
	// Create creates a new def and updates id and other fields. It is up to the
	// caller to update any fields that will be set after the policy has been
	// created for instance _id
	Create(ctx context.Context, def *v1alpha1.SecurityPolicySpec) error
	// Update this will update an existing policy
	Update(ctx context.Context, def *v1alpha1.SecurityPolicySpec) error
	//Delete deletes policy id id
	Delete(ctx context.Context, id string) error
}

type Portal interface {
	Policy() Policy
	Documentation() Documentation
	Catalogue() Catalogue
}

type Documentation interface {
	Upload(
		ctx context.Context, o *model.APIDocumentation,
	) (*model.Result, error)
	Delete(ctx context.Context, id string) (*model.Result, error)
}

type Catalogue interface {
	Get(ctx context.Context) (*model.APICatalogue, error)
	Create(ctx context.Context, o *model.APICatalogue) (*model.Result, error)
	Update(ctx context.Context, o *model.APICatalogue) (*model.Result, error)
}
