package dashboard_client

import (
	v1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
)

type ApisResponse struct {
	Apis  []DashboardApi `json:"apis"`
	Pages int            `json:"pages"`
}

type DashboardApi struct {
	CreatedAt     string               `json:"created_at,omitempty"`
	ApiDefinition v1.APIDefinitionSpec `json:"api_definition"`
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
