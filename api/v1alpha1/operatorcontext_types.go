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

// +kubebuilder:validation:Enum=ce;pro
type OperatorMode string

// OperatorContextSpec defines the desired state of OperatorContext
type OperatorContextSpec struct {
	Env        *Env            `json:"env,omitempty"`
	FromSecret *OperatorSecret `json:"from_secret,omitempty"`
}

type OperatorSecret struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
}

type Env struct {
	Namespace          []string     `json:"namespace,omitempty"`
	Mode               OperatorMode `json:"mode"`
	InsecureSkipVerify bool         `json:"ssl_insecure_skip_verify,omitempty"`
	URL                string       `json:"url"`
	ORG                string       `json:"org"`
	Auth               string       `json:"auth"`
}

// OperatorContextStatus defines the observed state of OperatorContext
type OperatorContextStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=tykctx

// OperatorContext is the Schema for the operatorcontexts API
type OperatorContext struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OperatorContextSpec   `json:"spec,omitempty"`
	Status OperatorContextStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// OperatorContextList contains a list of OperatorContext
type OperatorContextList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OperatorContext `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OperatorContext{}, &OperatorContextList{})
}
