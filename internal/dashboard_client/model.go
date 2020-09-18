package dashboard_client

import (
	v1 "github.com/TykTechnologies/tyk-operator/api/v1"
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
