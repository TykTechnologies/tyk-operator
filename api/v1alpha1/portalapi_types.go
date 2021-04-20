package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PortalAPIConfig struct {
	OrgId string `json:"org_id,omitempty"`
	//SignupFields         []interface{} `json:"signup_fields"`
	//KeyRequestFields     []interface{} `json:"key_request_fields"`
	RequireKeyApproval   bool   `json:"require_key_approval,omitempty"`
	SecureKeyApproval    bool   `json:"secure_key_approval,omitempty"`
	RedirectOnKeyRequest bool   `json:"redirect_on_key_request,omitempty"`
	RedirectTo           string `json:"redirect_to,omitempty"`

	//EnableMultiSelection bool   `json:"enable_multi_selection"`

	DisableLogin       bool   `json:"disable_login"`
	DisableSignup      bool   `json:"disable_signup"`
	DisableAutoLogin   bool   `json:"disable_auto_login"`
	CatalogueLoginOnly bool   `json:"catalogue_login_only"`
	OauthUsageLimit    int    `json:"oauth_usage_limit"`
	Email              string `json:"email,omitempty"`
	//MailOptions          struct {
	//	MailFromName  string `json:"mail_from_name"`
	//	MailFromEmail string `json:"mail_from_email"`
	//	EmailCopy     struct {
	//		WelcomeEmail struct {
	//			Enabled       bool   `json:"enabled"`
	//			Subject       string `json:"subject"`
	//			Body          string `json:"body"`
	//			SignOff       string `json:"sign_off"`
	//			HideTokenData bool   `json:"hide_token_data"`
	//		} `json:"welcome_email"`
	//		KeyEmail struct {
	//			Enabled       bool   `json:"enabled"`
	//			Subject       string `json:"subject"`
	//			Body          string `json:"body"`
	//			SignOff       string `json:"sign_off"`
	//			HideTokenData bool   `json:"hide_token_data"`
	//		} `json:"key_email"`
	//		ResetPasswordEmail struct {
	//			Enabled       bool   `json:"enabled"`
	//			Subject       string `json:"subject"`
	//			Body          string `json:"body"`
	//			SignOff       string `json:"sign_off"`
	//			HideTokenData bool   `json:"hide_token_data"`
	//		} `json:"reset_password_email"`
	//	} `json:"email_copy"`
	//} `json:"mail_options"`
	Override bool `json:"override"`
	Hashkeys bool `json:"hashkeys"`
}

type NamespaceName struct {
	// Namespace of the resource to target
	Namespace string `json:"namespace"`
	// Name of the resource to target
	Name string `json:"name"`
}

// PortalAPISpec defines the desired state of PortalAPI
type PortalAPISpec struct {
	Name             string `json:"name"`
	ShortDescription string `json:"short_description"`
	LongDescription  string `json:"long_description"`
	Show             bool   `json:"show"`

	// Explicitly set the PolicyID or populate the SecurityPolicy field with the namespace/name
	PolicyID       string        `json:"policy_id,omitempty"`
	SecurityPolicy NamespaceName `json:"security_policy,omitempty"`

	Documentation string `json:"documentation,omitempty"`

	// Version should always be v2 - so we will leave it as optional & set it later on
	Version string `json:"version,omitempty"`

	// IsKeyless is like a flag to determine whether somebody can click the get key button
	IsKeyless bool `json:"is_keyless,omitempty"`

	Config PortalAPIConfig   `json:"config,omitempty"`
	Fields map[string]string `json:"fields,omitempty"`

	// AuthType TODO should be an enum. maybe we can derive it from the underlying api types? leave empty for keyless
	// +kubebuilder:validation:Enum=authToken
	AuthType string `json:"auth_type,omitempty"`
}

// PortalAPIStatus defines the observed state of PortalAPI
type PortalAPIStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// PortalAPI is the Schema for the portalapis API
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=tykpapis
type PortalAPI struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PortalAPISpec   `json:"spec,omitempty"`
	Status PortalAPIStatus `json:"status,omitempty"`
}

// PortalAPIList contains a list of PortalAPI
// +kubebuilder:object:root=true
type PortalAPIList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PortalAPI `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PortalAPI{}, &PortalAPIList{})
}
