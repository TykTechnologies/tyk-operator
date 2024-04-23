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

// APIDefinitionSpec represents the configuration for a single proxied API and it's versions.
// +kubebuilder:object:generate=true
type APIDefinitionSpec struct {
	model.APIDefinitionSpec `json:",inline"`
	// Context specify namespace/name of the OperatorContext object used for
	// reconciling this APIDefinition
	Context *model.Target `json:"contextRef,omitempty"`
}

// ApiDefinitionStatus defines the observed state of ApiDefinition
type ApiDefinitionStatus struct {
	ApiID string `json:"api_id"`

	// OrgID corresponds to the Organization ID that this API belongs to.
	OrgID string `json:"orgId,omitempty"`

	// LinkedByPolicies is a list policies that references this api definition
	//+optional
	LinkedByPolicies []model.Target `json:"linked_by_policies,omitempty"`

	// LinkedByAPIs is a list of ApiDefinition namespaced/name that links to this
	// resource
	LinkedByAPIs []model.Target `json:"linked_by_apis,omitempty"`

	// LinkedToAPIs is a list of ApiDefinition namespaced/name that this resource
	// links to.
	LinkedToAPIs []model.Target `json:"linked_to_apis,omitempty"`

	// LinkedToSubgraph corresponds to the name of the Subgraph CR if the ApiDefinition is GraphQL Federation Subgraph.
	// ApiDefinition CR can only be linked to Subgraph CRs that are created in the same namespace as ApiDefinition CR.
	LinkedToSubgraph string `json:"linked_to_subgraph,omitempty"`

	// LatestTykSpecHash stores the hash of ApiDefinition created on Tyk. This information is updated after
	// creating or updating the ApiDefinition. It is useful for Operator to understand running update
	// operation or not. If there is a change in latestTykSpecHash as well as latestCRDSpecHash, Operator
	// runs update logic and updates resources on Tyk Gateway or Tyk Dashboard.
	LatestTykSpecHash string `json:"latestTykSpecHash,omitempty"`

	// LatestCRDSpecHash stores the hash of ApiDefinition CRD created on K8s. This information is updated after
	// creating or updating the ApiDefinition. It is useful for Operator to understand running update
	// operation or not. If there is a change in latestCRDSpecHash as well as latestTykSpecHash, Operator
	// runs update logic and updates resources on Tyk Gateway or Tyk Dashboard.
	LatestCRDSpecHash string `json:"latestCRDSpecHash,omitempty"`

	LatestTransaction TransactionInfo `json:"latestTransaction,omitempty"`
}

// ApiDefinition is the Schema for the apidefinitions API
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Domain",type=string,JSONPath=`.spec.domain`
// +kubebuilder:printcolumn:name="ListenPath",type=string,JSONPath=`.spec.proxy.listen_path`
// +kubebuilder:printcolumn:name="Proxy.TargetURL",type=string,JSONPath=`.spec.proxy.target_url`
// +kubebuilder:printcolumn:name="Enabled",type=boolean,JSONPath=`.spec.active`
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.latestTransaction.status`
// +kubebuilder:resource:categories="tyk",shortName="tykapis"
type ApiDefinition struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   APIDefinitionSpec   `json:"spec,omitempty"`
	Status ApiDefinitionStatus `json:"status,omitempty"`
}

// ApiDefinitionList contains a list of ApiDefinition
// +kubebuilder:object:root=true
type ApiDefinitionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ApiDefinition `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ApiDefinition{}, &ApiDefinitionList{})
}
