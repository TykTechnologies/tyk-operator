package dashboard_client

import (
	v1 "github.com/TykTechnologies/tyk-operator/api/v1"
)

type SecurityPolicy struct {
	*Client
}

func (s *SecurityPolicy) All() ([]v1.SecurityPolicySpec, error) {
	panic("implement me")
}

func (s *SecurityPolicy) Get(polId string) (*v1.SecurityPolicySpec, error) {
	panic("implement me")
}

func (s *SecurityPolicy) Create(def *v1.SecurityPolicySpec) (string, error) {
	panic("implement me")
}

func (s *SecurityPolicy) Update(def *v1.SecurityPolicySpec) error {
	panic("implement me")
}

func (s *SecurityPolicy) Delete(id string) error {
	panic("implement me")
}
