package dashboard

import (
	"github.com/TykTechnologies/tyk-operator/api/model"
	v1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
)

type ApisResponse struct {
	Apis  []DashboardApi `json:"apis"`
	Pages int            `json:"pages"`
}

type DashboardApi struct {
	CreatedAt       string                  `json:"created_at,omitempty"`
	ApiDefinition   model.APIDefinitionSpec `json:"api_definition"`
	UserOwners      []string                `json:"user_owners"`
	UserGroupOwners []string                `json:"user_group_owners"`
}

type PoliciesResponse struct {
	Policies []v1.SecurityPolicySpec `json:"data"`
	Pages    int                     `json:"pages"`
}

type CertErrorResponse struct {
	Status  string `json:"Status"`
	Message string `json:"Message"`
}

type CertResponse struct {
	Id      string `json:"id"`
	Message string `json:"message"`
	Status  string `json:"status"`
}
