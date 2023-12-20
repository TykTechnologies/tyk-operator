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

// PortalConfigSpec defines the desired state of PortalConfig
type PortalConfigSpec struct {
	model.PortalModelPortalConfig `json:",inline"`
	Context                       *model.Target `json:"contextRef,omitempty"`
}

// PortalConfigStatus defines the observed state of PortalConfig
type PortalConfigStatus struct {
	ID string `json:"id,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// PortalConfig is the Schema for the portalconfigs API
// +kubebuilder:resource:categories=tyk
type PortalConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PortalConfigSpec   `json:"spec,omitempty"`
	Status PortalConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PortalConfigList contains a list of PortalConfig
type PortalConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PortalConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PortalConfig{}, &PortalConfigList{})
}
