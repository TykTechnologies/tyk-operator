package mock_client

import (
	v1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
)

type SecurityPolicy struct {
	*Client
}

func (p SecurityPolicy) All() ([]v1.SecurityPolicySpec, error) {
	return nil, nil
}

func (p SecurityPolicy) Get(polId string) (*v1.SecurityPolicySpec, error) {
	return nil, nil
}

func (p SecurityPolicy) Create(def *v1.SecurityPolicySpec) (string, error) {
	return "", nil
}

func (p SecurityPolicy) Update(def *v1.SecurityPolicySpec) error {
	return nil
}

func (p SecurityPolicy) Delete(id string) error {
	return nil
}
