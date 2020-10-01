package mock_client

import (
	v1 "github.com/TykTechnologies/tyk-operator/api/v1"
)

type Api struct {
	*Client
}

func (a Api) All() ([]v1.APIDefinitionSpec, error) {
	return nil, nil
}

func (a Api) Get(apiID string) (*v1.APIDefinitionSpec, error) {
	return nil, nil
}

func (a Api) Create(def *v1.APIDefinitionSpec) (string, error) {
	return "", nil
}

func (a Api) Update(_ string, def *v1.APIDefinitionSpec) error {
	return nil
}

func (a Api) Delete(id string) error {
	return nil
}
