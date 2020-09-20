/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SecurityPolicySpec defines the desired state of SecurityPolicy
type SecurityPolicySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	MID   string `json:"_id,omitempty"`
	ID    string `json:"id,omitempty"`
	Name  string `json:"name"`
	OrgID string `json:"org_id,omitempty"`
	// State can be active, draft or deny
	State                         string                      `json:"state"`
	Active                        bool                        `json:"active"`
	IsInactive                    bool                        `json:"is_inactive"`
	AccessRightsArray             []AccessDefinition          `json:"access_rights_array"`
	AccessRights                  map[string]AccessDefinition `json:"access_rights,omitempty"`
	Rate                          int64                       `json:"rate"`
	Per                           int64                       `json:"per"`
	QuotaMax                      int64                       `json:"quota_max"`
	QuotaRenewalRate              int64                       `json:"quota_renewal_rate"`
	ThrottleInterval              int64                       `json:"throttle_interval"`
	ThrottleRetryLimit            int                         `json:"throttle_retry_limit"`
	MaxQueryDepth                 int                         `json:"max_query_depth"`
	HMACEnabled                   bool                        `json:"hmac_enabled,omitempty"`
	EnableHTTPSignatureValidation bool                        `json:"enable_http_signature_validation,omitempty"`
	Tags                          []string                    `json:"tags,omitempty"`
	// KeyExpiresIn is the number of seconds till key expiry. For 1 hour is 3600. Default never expire or 0
	KeyExpiresIn int64            `json:"key_expires_in"`
	Partitions   PolicyPartitions `json:"partitions,omitempty"`
}

// from tyk/session.go
// AccessDefinition defines which versions of an API a key has access to
type AccessDefinition struct {
	APIName  string   `json:"api_name"`
	APIID    string   `json:"api_id"`
	Versions []string `json:"versions"`
	//RestrictedTypes []graphql.Type `json:"restricted_types"`
	Limit          APILimit     `json:"limit,omitempty"`
	AllowanceScope string       `json:"allowance_scope,omitempty"`
	AllowedURLs    []AccessSpec `json:"allowed_urls,omitempty"` // mapped string MUST be a valid regex
}

// from tyk/session.go
// APILimit stores quota and rate limit on ACL level (per API)
type APILimit struct {
	Rate               int64  `json:"rate"`
	Per                int64  `json:"per"`
	ThrottleInterval   int64  `json:"throttle_interval"`
	ThrottleRetryLimit int    `json:"throttle_retry_limit"`
	MaxQueryDepth      int    `json:"max_query_depth"`
	QuotaMax           int64  `json:"quota_max"`
	QuotaRenews        int64  `json:"quota_renews"`
	QuotaRemaining     int64  `json:"quota_remaining"`
	QuotaRenewalRate   int64  `json:"quota_renewal_rate"`
	SetBy              string `json:"-"`
}

// from tyk/session.go
// AccessSpecs define what URLS a user has access to an what methods are enabled
type AccessSpec struct {
	URL     string   `json:"url"`
	Methods []string `json:"methods"`
}

type PolicyPartitions struct {
	Quota      bool `json:"quota"`
	RateLimit  bool `json:"rate_limit"`
	Complexity bool `json:"complexity"`
	Acl        bool `json:"acl"`
	PerAPI     bool `json:"per_api"`
}

// SecurityPolicyStatus defines the observed state of SecurityPolicy
type SecurityPolicyStatus struct {
	// TODO: add ID here which references the policy_id that was created in Tyk
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=tykpolicies
// SecurityPolicy is the Schema for the securitypolicies API
type SecurityPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SecurityPolicySpec   `json:"spec,omitempty"`
	Status SecurityPolicyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SecurityPolicyList contains a list of SecurityPolicy
type SecurityPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SecurityPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SecurityPolicy{}, &SecurityPolicyList{})
}
