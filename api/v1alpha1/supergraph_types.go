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

type SubGraphReference struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// SuperGraphSpec defines the desired state of SuperGraph
type SuperGraphSpec struct {
	SubgraphsRefs []SubGraphReference `json:"subgraphs_refs"`
	MergedSDL     string              `json:"merged_sdl,omitempty"`
	Schema        string              `json:"schema,omitempty"`
}

// SuperGraphStatus defines the observed state of SuperGraph
type SuperGraphStatus struct{}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// SuperGraph is the Schema for the supergraphs API
type SuperGraph struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SuperGraphSpec   `json:"spec,omitempty"`
	Status SuperGraphStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SuperGraphList contains a list of SuperGraph
type SuperGraphList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SuperGraph `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SuperGraph{}, &SuperGraphList{})
}
