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

// APICatalogueSpec defines the desired state of APICatalogue
type APICatalogueSpec struct {
	APIDescriptionList []*model.Target `json:"apis,omitempty"`
	Context            *model.Target   `json:"contextRef,omitempty"`
}

// APICatalogueStatus defines the observed state of APICatalogue
type APICatalogueStatus struct {
	ID string `json:"id,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// APICatalogue is the Schema for the apicatalogues API
type APICatalogue struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   APICatalogueSpec   `json:"spec,omitempty"`
	Status APICatalogueStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// APICatalogueList contains a list of APICatalogue
type APICatalogueList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []APICatalogue `json:"items"`
}

func init() {
	SchemeBuilder.Register(&APICatalogue{}, &APICatalogueList{})
}
