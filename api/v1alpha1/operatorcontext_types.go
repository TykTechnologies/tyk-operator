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
	// FromSecret when this is specified, the given secret will be used to load
	// values for the context environment
	FromSecret *Target `json:"fromSecret,omitempty"`
	// Env is the values of the admin api endpoint that the operator will use to
	// reconcile resources
	Env *Environment `json:"env,omitempty"`
}

type Environment struct {
	Mode               OperatorContextMode `json:"mode,omitempty"`
	InsecureSkipVerify bool                `json:"insecureSkipVerify,omitempty"`
	URL                string              `json:"url,omitempty"`
	Auth               string              `json:"auth,omitempty"`
	Org                string              `json:"org,omitempty"`
	Ingress            Ingress             `json:"ingress,omitempty"`
}

type Ingress struct {
	HTTPPort  int `json:"httpPort,omitempty"`
	HTTPSPort int `json:"httpsPort,omitempty"`
}

// OperatorContextStatus defines the observed state of OperatorContext
type OperatorContextStatus struct{}

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

func init() {
	SchemeBuilder.Register(&OperatorContext{}, &OperatorContextList{})
}
