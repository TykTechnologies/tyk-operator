package gateway_client

import (
	v1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
)

// SecurityPolicy provides api for accessing policies on the tyk gateway
// NOTE: The gateway doesn't provide api for security policy so this is just a
// placeholder and does nothing except returning ErrTODO on all methods.
type SecurityPolicy struct {
	*Client
}

func (a SecurityPolicy) All() ([]v1.SecurityPolicySpec, error) {
	return nil, ErrTODO
}

func (a SecurityPolicy) Get(namespacedName string) (*v1.SecurityPolicySpec, error) {
	return nil, ErrTODO
}

func (a SecurityPolicy) Create(def *v1.SecurityPolicySpec) error {
	return ErrTODO
}

func (a SecurityPolicy) Update(def *v1.SecurityPolicySpec) error {
	return ErrTODO
}

func (a SecurityPolicy) Delete(namespacedName string) error {
	return ErrTODO
}
