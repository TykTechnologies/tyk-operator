package universal_client

import (
	v1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
)

type UniversalPortalAPI interface {
	Get(policyID string) (*v1.PortalAPISpec, error)
	All() ([]v1.PortalAPISpec, error)
	Create(spec *v1.PortalAPISpec) error
	Update(spec *v1.PortalAPISpec) error
	Delete(policyID string) error
}
