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

// APIDescriptionSpec defines the desired state of APIDescription
type APIDescriptionSpec struct {
	model.APIDescription `json:",inline"`
	Context              *model.Target    `json:"contextRef,omitempty"`
	APIDocumentation     APIDocumentation `json:"docs"`
}

type APIDocumentation struct {
	DocumentationType model.DocumentationType `json:"doc_type"`
	Documentation     string                  `json:"documentation,omitempty"`
}

// APIDescriptionStatus defines the observed state of APIDescription
type APIDescriptionStatus struct{}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// APIDescription is the Schema for the apidescriptions API
type APIDescription struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   APIDescriptionSpec   `json:"spec,omitempty"`
	Status APIDescriptionStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// APIDescriptionList contains a list of APIDescription
type APIDescriptionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []APIDescription `json:"items"`
}

func init() {
	SchemeBuilder.Register(&APIDescription{}, &APIDescriptionList{})
}
