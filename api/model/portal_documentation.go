package model

// +kubebuilder:validation:Enum=swagger;swagger_custom_url;graphql
type DocumentationType string

type APIDocumentation struct {
	Id                string            `json:"id,omitempty"`
	Documentation     string            `json:"documentation"`
	DocumentationType DocumentationType `json:"doc_type"`
	// The policy_id of the APIDescription that this documentation is attached to.
	APIID string `json:"api_id"`
}
