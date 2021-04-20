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

// PortalAPISpec defines the desired state of PortalAPI
type PortalAPISpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of PortalAPI. Edit PortalAPI_types.go to remove/update
	Foo string `json:"foo,omitempty"`
}

// PortalAPIStatus defines the observed state of PortalAPI
type PortalAPIStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// PortalAPI is the Schema for the portalapis API
type PortalAPI struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PortalAPISpec   `json:"spec,omitempty"`
	Status PortalAPIStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PortalAPIList contains a list of PortalAPI
type PortalAPIList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PortalAPI `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PortalAPI{}, &PortalAPIList{})
}
