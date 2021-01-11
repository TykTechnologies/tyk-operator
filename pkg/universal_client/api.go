package universal_client

import (
	v1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
)

type UniversalApi interface {
	Get(apiID string) (*v1.APIDefinitionSpec, error)
	All() ([]v1.APIDefinitionSpec, error)
	Create(spec *v1.APIDefinitionSpec) error
	Update(def *v1.APIDefinitionSpec) error
	Delete(id string) error
}
