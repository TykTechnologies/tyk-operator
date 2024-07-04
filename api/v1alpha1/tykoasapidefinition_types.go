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

	// Versioning provides versioning information about this OAS API
	Versioning *TykOASVersioning `json:"versioning,omitempty"`

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

// LocationType defines the type for the location enum
type LocationType string

// Define the allowed values for LocationType
const (
	LocationHeader   LocationType = "header"
	LocationURLParam LocationType = "url-param"
	LocationURL      LocationType = "url"
)

// TykOASVersioning contains verisoning information for an TykOASAPIDefinition.
type TykOASVersioning struct {
	// Default contains the default version name if a request is issued without a version.
	Default string `json:"default"`

	// Enabled is a boolean flag, if set to true it will enable versioning of the API.
	Enabled bool `json:"enabled"`

	// Key contains the name of the key to check for versioning information.
	Key string `json:"key"`

	// Location contains versioning location information. It can be one of the following:
	// header, url-param, url.
	// +kubebuilder:validation:Enum=header;url-param;url
	Location *LocationType `json:"location"`

	// Name contains the name of the version.
	Name string `json:"name"`

	// Versions contains a list of versions that map to individual API IDs.
	Versions []TykOASVersion `json:"versions,omitempty"`
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
	// Name is the name of the OAS API within Tyk.
	Name string `json:"name,omitempty"`
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
	// VersioningStatus shows the status of a Versioned TykOasAPIDefinition.
	VersioningStatus *VersioningStatus `json:"versioningStatus,omitempty"`
	// LinkedByPolicies is a list policies that references this OAS API Definition.
	//+optional
	LinkedByPolicies []model.Target `json:"linkedByPolicies,omitempty"`
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

	// LatestConfigMapHash stores the hash of ConfigMap that is being used by TykOasApiDefinition.
	LatestConfigMapHash string `json:"latestConfigMapHash,omitempty"`
}

// TykOASVersion represents each OAS API Definition used as a version.
type TykOASVersion struct {
	// Name contains the name of the refrenced TykOasApiDefinition.
	Name string `json:"name"`

	// TykOasApiDefinitionRef references a TykOasApiDefinition.
	TykOasApiDefinitionRef string `json:"tykOasApiDefinitionRef"`

	// Namespace contains the namespace where the version was installed.
	Namespace string `json:"namespace,omitempty"`
}

// VersioningStatus contains the status of a versioned TykOasAPI.
type VersioningStatus struct {
	// IsVersionedAPI indicates if the API is versioned.
	IsVersionedAPI bool `json:"isVersionedAPI,omitempty"`
	// BaseAPIVersionContextRef specifies the namespace and name of the
	// Base API a versioned API is linked to.
	BaseAPIVersionContextRef *model.Target `json:"baseAPIVersionContextRef,omitempty"`
	// IsDefaultVersion specifies if the OAS API is the default  Version.
	IsDefaultVersion bool `json:"isDefaultVersion,omitempty"`
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

func (t *TykOasApiDefinition) RemoveOASVersionStatus() {
	t.Status.VersioningStatus = nil
}

func (status *TykOasApiDefinitionStatus) GetIsVersionedAPI() bool {
	if status.VersioningStatus == nil {
		return false
	}

	return status.VersioningStatus.IsVersionedAPI
}

func (status *TykOasApiDefinitionStatus) GetIsDefaultVersion() bool {
	if status.VersioningStatus == nil {
		return false
	}

	return status.VersioningStatus.IsDefaultVersion
}

func (status *TykOasApiDefinitionStatus) GetBaseVersionName() string {
	if status.VersioningStatus == nil {
		return ""
	}

	return status.VersioningStatus.BaseAPIVersionContextRef.Name
}

func (status *TykOasApiDefinitionStatus) GetBaseVersionNamespace() string {
	if status.VersioningStatus == nil {
		return ""
	}

	return *status.VersioningStatus.BaseAPIVersionContextRef.Namespace
}

func (status *TykOasApiDefinitionStatus) SetBaseVersionName(name string) {
	status.VersioningStatus.BaseAPIVersionContextRef.Name = name
}

func (status *TykOasApiDefinitionStatus) SetBaseVersionNamespace(name *string) {
	status.VersioningStatus.BaseAPIVersionContextRef.Namespace = name
}

func (status *TykOasApiDefinitionStatus) SetIsVersionedAPI(versioned bool) {
	status.VersioningStatus.IsVersionedAPI = versioned
}

func (status *TykOasApiDefinitionStatus) NewVersioningStatus() {
	versioningStatus := VersioningStatus{
		BaseAPIVersionContextRef: &model.Target{},
	}

	status.VersioningStatus = &versioningStatus
}

func (in *TykOasApiDefinition) GetLinkedPolicies() []model.Target {
	return in.Status.LinkedByPolicies
}

func (in *TykOasApiDefinition) SetLinkedPolicies(result []model.Target) {
	in.Status.LinkedByPolicies = result
}

func (in *TykOasApiDefinition) ApiName() string {
	return in.Status.Name
}

func (in *TykOasApiDefinition) StatusApiID() string {
	return in.Status.ID
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
