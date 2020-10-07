/*


Licensed under the Mozilla Public License (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.mozilla.org/en-US/MPL/2.0/

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//type OrganizationAPI struct {
//	APIHumanName string `json:"api_human_name"`
//	APIID        string `json:"api_id"`
//}
//
//type OrganizationEvent struct {
//	Webhook string `json:"webhook"`
//	Email   string `json:"email"`
//	Redis   bool   `json:"redis"`
//}
//
//type OrganizationUI struct {
//	Languages   json.RawMessage `json:"languages"`
//	HideHelp    bool            `json:"hide_help,omitempty"`
//	DefaultLang string          `json:"json:default_lang"`
//}

// OrganizationSpec defines the desired state of Organization
type OrganizationSpec struct {
	ID           string `json:"id,omitempty"`
	OwnerName    string `json:"owner_name"`
	OwnerSlug    string `json:"owner_slug,omitempty"`
	CNAMEEnabled bool   `json:"cname_enabled"`
	CNAME        string `json:"cname"`

	//APIs           []OrganizationAPI            `json:"apis"`
	//SSOEnabled     bool                         `json:"sso_enabled,omitempty"`
	//DeveloperQuota int                          `json:"developer_quota,omitempty"`
	//DeveloperCount int                          `json:"developer_count"`
	//EventOptions   map[string]OrganizationEvent `json:"event_options,omitempty"`
	//HybridEnabled  bool                         `json:"hybrid_enabled,omitempty"`
	//UI             OrganizationUI               `json:"ui,omitempty"`
	//// Foo is an example field of Organization. Edit Organization_types.go to remove/update
	//Foo string `json:"foo,omitempty"`
}

// OrganizationStatus defines the observed state of Organization
type OrganizationStatus struct {
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=tykorgs
// Organization is the Schema for the organizations API
type Organization struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OrganizationSpec   `json:"spec,omitempty"`
	Status OrganizationStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// OrganizationList contains a list of Organization
type OrganizationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Organization `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Organization{}, &OrganizationList{})
}
