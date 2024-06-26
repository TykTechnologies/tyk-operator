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

// TykOasApiDefinitionSpec defines the desired state of TykOasApiDefinition
type TykOasApiDefinitionSpec struct {
	// Context specify namespace/name of the OperatorContext object used for
	// reconciling this APIDefinition
	Context *model.Target `json:"contextRef,omitempty"`
	// TykOAS provides storage information about Tyk OAS
	TykOAS TykOASReference `json:"tykOAS"`

	// ClientCertificate is used to configure client certificates settings needed
	// for MTLS connection between Tyk and client.
	// It is used to set `server.clientCertificate` field of Tyk OAS API
	ClientCertificate ClientCertificateConfig `json:"clientCertificate,omitempty"`
}

// ClientCertificateConfig specifies configuration for client certificates
type ClientCertificateConfig struct {
	// Enabled activates mTLS for the API.
	Enabled bool `json:"enabled,omitempty"`
	// Allowlist stores list of k8s secret names storing client certificates
	Allowlist []string `json:"allowlist,omitempty"`
}

type TykOASReference struct {
	// ConfigmapRef provides information of configmap in which Tyk OAS is stored
	ConfigmapRef ConfigMapReference `json:"configmapRef"`
}

type ConfigMapReference struct {
	// Name is the name of configmap
	Name string `json:"name"`
	// Namespace is the namespace of configmap.
	// If Namespace is not provided, we assume that the ConfigMap is in the same
	// namespace as TykOasApiDefinition resource.
	Namespace string `json:"namespace,omitempty"`
	// KeyName is the key of configmap in which Tyk OAS doc is stored
	KeyName string `json:"keyName"`
}

// TykOasApiDefinitionStatus defines the observed state of TykOasApiDefinition
type TykOasApiDefinitionStatus struct {
	// ID is the unique identifier of the API within Tyk.
	ID string `json:"id,omitempty"`
	// Domain is the custom domain used by the API
	Domain string `json:"domain,omitempty"`
	// ListenPath is the base path on Tyk to which requests for this API will be sent.
	ListenPath string `json:"listenPath,omitempty"`
	// TargetURL is the upstream address to which requests will be proxied.
	TargetURL string `json:"targetURL,omitempty"`
	// Enabled represents if API is enabled or not
	Enabled bool `json:"enabled,omitempty"`
	// LatestTransaction provides status information about the last reconciliation.
	LatestTransaction TransactionInfo `json:"latestTransaction,omitempty"`
	// IngressTemplate shows whether this CR is used as Ingress Template or not.
	IngressTemplate bool `json:"ingressTemplate,omitempty"`

	// LatestTykSpecHash stores the hash of OAS API Definition created on Tyk. This information is updated after
	// creating or updating the TykOasApiDefinition. It is useful for Operator to understand the need for
	// running update operation or not. If there is a change in latestTykSpecHash as well as latestCRDSpecHash,
	// Operator runs update logic and updates resources on Tyk Gateway or Tyk Dashboard.
	LatestTykSpecHash string `json:"latestTykSpecHash,omitempty"`

	// LatestCRDSpecHash stores the hash of TykOasApiDefinition CR created on K8s. This information is updated after
	// creating or updating the TykOasApiDefinition. It is useful for Operator to understand the need for
	// running update operation or not. If there is a change in latestCRDSpecHash as well as latestTykSpecHash,
	// Operator runs update logic and updates resources on Tyk Gateway or Tyk Dashboard.
	LatestCRDSpecHash string `json:"latestCRDSpecHash,omitempty"`
}

// TykOasApiDefinition is the Schema for the tykoasapidefinitions API
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Domain",type=string,JSONPath=`.status.domain`
// +kubebuilder:printcolumn:name="ListenPath",type=string,JSONPath=`.status.listenPath`
// +kubebuilder:printcolumn:name="Proxy.TargetURL",type=string,JSONPath=`.status.targetURL`
// +kubebuilder:printcolumn:name="Enabled",type=boolean,JSONPath=`.status.enabled`
// +kubebuilder:printcolumn:name="SyncStatus",type=string,JSONPath=`.status.latestTransaction.status`
// +kubebuilder:printcolumn:name="IngressTemplate",type=boolean,JSONPath=`.status.ingressTemplate`
// +kubebuilder:resource:categories="tyk",shortName="tykoas"
type TykOasApiDefinition struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TykOasApiDefinitionSpec   `json:"spec,omitempty"`
	Status TykOasApiDefinitionStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TykOasApiDefinitionList contains a list of TykOasApiDefinition
type TykOasApiDefinitionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TykOasApiDefinition `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TykOasApiDefinition{}, &TykOasApiDefinitionList{})
}
