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

// PortalAPICatalogueSpec defines the desired state of PortalAPICatalogue
type PortalAPICatalogueSpec struct {
	OrgID              string         `json:"org_id,omitempty"`
	Email              string         `json:"email,omitempty"`
	APIDescriptionList []model.Target `json:"apis,omitempty"`
	Context            *model.Target  `json:"contextRef,omitempty"`
}

// PortalAPICatalogueStatus defines the observed state of PortalAPICatalogue
type PortalAPICatalogueStatus struct {
	ID string `json:"id,omitempty"`
	// maps model.Target.String() => Documentation ID
	Documentation map[string]string `json:"apis,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// PortalAPICatalogue is the Schema for the portalapicatalogues API
type PortalAPICatalogue struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PortalAPICatalogueSpec   `json:"spec,omitempty"`
	Status PortalAPICatalogueStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PortalAPICatalogueList contains a list of PortalAPICatalogue
type PortalAPICatalogueList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PortalAPICatalogue `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PortalAPICatalogue{}, &PortalAPICatalogueList{})
}
