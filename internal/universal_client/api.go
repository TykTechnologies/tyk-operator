package universal_client

import (
	v1 "github.com/TykTechnologies/tyk-operator/api/v1"
)

type UniversalApi interface {
	Get(apiID string) (*v1.APIDefinitionSpec, error)
	All() ([]v1.APIDefinitionSpec, error)
	Create(def *v1.APIDefinitionSpec) (apiID string, err error)
	Update(apiID string, def *v1.APIDefinitionSpec) error
	Delete(id string) error
}
