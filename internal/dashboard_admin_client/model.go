package dashboard_admin_client

import (
	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
)

type OrganizationsResponse struct {
	Organizations []v1alpha1.OrganizationSpec `json:"organisations"`
}

type CreateOrganizationResponse struct {
	Status  string `json:"Status"`
	Message string `json:"message"`
	Meta    string `json:"meta"`
}
