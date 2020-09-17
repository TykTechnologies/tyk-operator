package universal_client

import v1 "github.com/TykTechnologies/tyk-operator/api/v1"

type UniversalSecurityPolicy interface {
	All() ([]v1.SecurityPolicySpec, error)
	Get(polId string) (*v1.SecurityPolicySpec, error)
	Create(def *v1.SecurityPolicySpec) (string, error)
	Update(def *v1.SecurityPolicySpec) error
	Delete(id string) error
}
