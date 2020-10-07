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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// WebhookSpec defines the desired state of Webhook
type WebhookSpec struct {
	//ApiModel     `bson:"api_model,omitempty" json:"api_model,omitempty"`
	// System generated resource. represents the ID to use in API  calls
	ID string `json:"id,omitempty"`
	// The org ID to which this webhook belongs to
	OrgID string `bson:"org_id" json:"org_id,omitempty"`
	// System generated resource, will be set to the {namespace/name} of the CRD
	Name         string            `bson:"name" json:"name"`
	Method       string            `bson:"method" json:"method"`
	TargetPath   string            `bson:"target_path" json:"target_path"`
	TemplatePath string            `bson:"template_path" json:"template_path,omitempty"`
	HeaderList   map[string]string `bson:"header_map" json:"header_map,omitempty"`
	EventTimeout int64             `bson:"event_timeout" json:"event_timeout,omitempty"`
}

// WebhookStatus defines the observed state of Webhook
type WebhookStatus struct {
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Webhook is the Schema for the webhooks API
type Webhook struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WebhookSpec   `json:"spec,omitempty"`
	Status WebhookStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// WebhookList contains a list of Webhook
type WebhookList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Webhook `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Webhook{}, &WebhookList{})
}
