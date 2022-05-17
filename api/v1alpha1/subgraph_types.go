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

type SubgraphConfig struct {
	SDL    string `json:"sdl"`
	Schema string `json:"schema"`
}

// SubGraphSpec defines the desired state of SubGraph
type SubGraphSpec struct {
	// Subgraph holds the configuration for a GraphQL federation subgraph.
	Subgraph SubgraphConfig `json:"subgraph,omitempty"`
}

// SubGraphStatus defines the observed state of SubGraph
type SubGraphStatus struct {
	APIID string `json:"APIID"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// SubGraph is the Schema for the subgraphs API
type SubGraph struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SubGraphSpec   `json:"spec,omitempty"`
	Status SubGraphStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SubGraphList contains a list of SubGraph
type SubGraphList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SubGraph `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SubGraph{}, &SubGraphList{})
}
