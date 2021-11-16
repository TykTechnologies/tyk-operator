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
	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (

	// WatchNamespace is the constant for env variable WATCH_NAMESPACE
	// which specifies the Namespace to watch.
	// An empty value means the operator is running with cluster scope.
	WatchNamespace = "WATCH_NAMESPACE"

	// TykMode defines what environment the operator is running. The values are ce
	// for community edition and pro for pro version
	TykMode = "TYK_MODE"

	// TykURL holds the url to either tyk gateway or tyk dashboard
	TykURL = "TYK_URL"

	// TykAuth holds the authorization token used to make api calls to the
	// gateway/dashboard
	TykAuth = "TYK_AUTH"

	// TykORG holds the org id which perform api tasks with
	TykORG = "TYK_ORG"

	// SkipVerify the client will skip tls verification if this is true
	SkipVerify = "TYK_TLS_INSECURE_SKIP_VERIFY"

	// IngressClass overides the default class to watch for ingress
	IngressClass = "WATCH_INGRESS_CLASS"

	IngressTLSPort = "TYK_HTTPS_INGRESS_PORT"

	IngressHTTPPort = "TYK_HTTP_INGRESS_PORT"
)

// OperatorContextMode is the mode to which the admin api binding is done values are
// ce for community edition and pro for dashboard
// +kubebuilder:validation:Enum=ce;pro
type OperatorContextMode string

// OperatorContextSpec defines the desired state of OperatorContext
type OperatorContextSpec struct {
	// Reference to k8s secret resource that we load environment from.
	FromSecret *model.Target `json:"secretRef,omitempty"`
	// Env is the values of the admin api endpoint that the operator will use to
	// reconcile resources
	Env *Environment `json:"env,omitempty"`
}

type Environment struct {
	Mode               OperatorContextMode `json:"mode,omitempty"`
	URL                string              `json:"url,omitempty"`
	Auth               string              `json:"auth,omitempty"`
	Org                string              `json:"org,omitempty"`
	Ingress            Ingress             `json:"ingress,omitempty"`
	InsecureSkipVerify bool                `json:"insecureSkipVerify,omitempty"`
}

type Ingress struct {
	HTTPPort  int `json:"httpPort,omitempty"`
	HTTPSPort int `json:"httpsPort,omitempty"`
}

// OperatorContextStatus defines the observed state of OperatorContext
type OperatorContextStatus struct {
	LinkedApiDefinitions      []model.Target `json:"linked_api_definitions,omitempty"`
	LinkedApiDescriptions     []model.Target `json:"linked_api_descriptions,omitempty"`
	LinkedPortalAPICatalogues []model.Target `json:"linked_portal_catalogues,omitempty"`
	LinkedSecurityPolicies    []model.Target `json:"linked_security_policies,omitempty"`
	LinkedPortalConfigs       []model.Target `json:"linked_portal_configs,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// OperatorContext is the Schema for the operatorcontexts API
type OperatorContext struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OperatorContextSpec   `json:"spec,omitempty"`
	Status OperatorContextStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// OperatorContextList contains a list of OperatorContext
type OperatorContextList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OperatorContext `json:"items"`
}

// +kubebuilder:rbac:groups=tyk.tyk.io,resources=operatorcontexts,verbs=get;list

func init() {
	SchemeBuilder.Register(&OperatorContext{}, &OperatorContextList{})
}

func (opStatus *OperatorContextStatus) RemoveLinkedAPIDefinition(target model.Target, log logr.Logger) {
	for i, apiDef := range opStatus.LinkedApiDefinitions {
		if apiDef.Namespace == target.Namespace && apiDef.Name == target.Name {
			opStatus.LinkedApiDefinitions = append(opStatus.LinkedApiDefinitions[:i], opStatus.LinkedApiDefinitions[i+1:]...)

			return
		}
	}
}

func (opStatus *OperatorContextStatus) RemoveLinkedSecurityPolicies(target model.Target) {
	for i, sp := range opStatus.LinkedSecurityPolicies {
		if sp.Namespace == target.Namespace && sp.Name == target.Name {
			opStatus.LinkedSecurityPolicies = append(opStatus.LinkedSecurityPolicies[:i], opStatus.LinkedSecurityPolicies[i+1:]...)

			return
		}
	}
}

func (opStatus *OperatorContextStatus) RemoveLinkedApiDescriptions(target model.Target) {
	for i, apiDes := range opStatus.LinkedApiDescriptions {
		if apiDes.Namespace == target.Namespace && apiDes.Name == target.Name {
			opStatus.LinkedApiDescriptions = append(opStatus.LinkedApiDescriptions[:i], opStatus.LinkedApiDescriptions[i+1:]...)

			return
		}
	}
}

func (opStatus *OperatorContextStatus) RemoveLinkedPortalAPICatalogues(target model.Target) {
	for i, pcat := range opStatus.LinkedPortalAPICatalogues {
		if pcat.Namespace == target.Namespace && pcat.Name == target.Name {
			opStatus.LinkedPortalAPICatalogues = append(opStatus.LinkedPortalAPICatalogues[:i], opStatus.LinkedPortalAPICatalogues[i+1:]...)

			return
		}
	}
}

func (opStatus *OperatorContextStatus) RemoveLinkedPortalConfig(target model.Target) {
	for i, pconf := range opStatus.LinkedPortalAPICatalogues {
		if pconf.Namespace == target.Namespace && pconf.Name == target.Name {
			opStatus.LinkedPortalConfigs = append(opStatus.LinkedPortalConfigs[:i], opStatus.LinkedPortalConfigs[i+1:]...)

			return
		}
	}
}
