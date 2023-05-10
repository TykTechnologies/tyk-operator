package mockgw

import v1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"

type CertResponse struct {
	Id      string `json:"id"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

type PoliciesResponse struct {
	Policies []v1.SecurityPolicySpec `json:"data"`
	Pages    int                     `json:"pages"`
}
