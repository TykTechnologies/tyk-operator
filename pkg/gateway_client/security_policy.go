package gateway_client

import (
	"context"

	v1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/pkg/universal_client"
)

// SecurityPolicy provides api for accessing policies on the tyk gateway
// NOTE: The gateway doesn't provide api for security policy so this is just a
// placeholder and does nothing except returning universal_client.ErrTODO on all methods.
type SecurityPolicy struct {
}

func (a SecurityPolicy) All(ctx context.Context) ([]v1.SecurityPolicySpec, error) {
	return nil, universal_client.ErrTODO
}

func (a SecurityPolicy) Get(ctx context.Context, namespacedName string) (*v1.SecurityPolicySpec, error) {
	return nil, universal_client.ErrTODO
}

func (a SecurityPolicy) Create(ctx context.Context, def *v1.SecurityPolicySpec) error {
	return universal_client.ErrTODO
}

func (a SecurityPolicy) Update(ctx context.Context, def *v1.SecurityPolicySpec) error {
	return universal_client.ErrTODO
}

func (a SecurityPolicy) Delete(ctx context.Context, namespacedName string) error {
	return universal_client.ErrTODO
}
