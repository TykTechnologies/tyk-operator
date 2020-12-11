package gateway_client

import (
	v1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/pkg/universal_client"
)

// SecurityPolicy provides api for accessing policies on the tyk gateway
// NOTE: The gateway doesn't provide api for security policy so this is just a
// placeholder and does nothing except returning universal_client.ErrTODO on all methods.
type SecurityPolicy struct {
}

func (a SecurityPolicy) All() ([]v1.SecurityPolicySpec, error) {
	return nil, universal_client.ErrTODO
}

func (a SecurityPolicy) Get(namespacedName string) (*v1.SecurityPolicySpec, error) {
	return nil, universal_client.ErrTODO
}

func (a SecurityPolicy) Create(def *v1.SecurityPolicySpec) error {
	return universal_client.ErrTODO
}

func (a SecurityPolicy) Update(def *v1.SecurityPolicySpec) error {
	return universal_client.ErrTODO
}

func (a SecurityPolicy) Delete(namespacedName string) error {
	return universal_client.ErrTODO
}
