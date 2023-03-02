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
	"github.com/TykTechnologies/tyk-operator/api/model"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TykOASApiDefinitionSpec defines the desired state of TykOASApiDefinition
type TykOASApiDefinitionSpec struct {
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of TykOASApiDefinition. Edit tykoasapidefinition_types.go to remove/update
	OASRef *OASReference `json:"oas_ref"`
}

type OASReference struct {
	model.Target `json:",inline"`
	KeyName      string `json:"key_name"`
}

// TykOASApiDefinitionStatus defines the observed state of TykOASApiDefinition
type TykOASApiDefinitionStatus struct {
	ApiID string `json:"api_id"`

	// LinkedByPolicies is a list policies that references this api definition
	//+optional
	LinkedByPolicies []model.Target `json:"linked_by_policies,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// TykOASApiDefinition is the Schema for the tykoasapidefinitions API
type TykOASApiDefinition struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TykOASApiDefinitionSpec   `json:"spec,omitempty"`
	Status TykOASApiDefinitionStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TykOASApiDefinitionList contains a list of TykOASApiDefinition
type TykOASApiDefinitionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TykOASApiDefinition `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TykOASApiDefinition{}, &TykOASApiDefinitionList{})
}