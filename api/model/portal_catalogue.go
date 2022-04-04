package model

type APICatalogue struct {
	Id    string           `json:"id"`
	OrgId string           `json:"org_id"`
	APIS  []APIDescription `json:"apis"`
	Email string           `json:"email"`
}

// All auth_type from dashboard validation:Enum=multiAuth;keyless;basic;hmac;jwt;oauth;openid;mutualTLS;authToken;custom;other
// +kubebuilder:validation:Enum=keyless;jwt;oauth;authToken
type AuthType string

type APIDescription struct {
	// Name is the title of the API that you wish to be published to the catalogue
	Name string `json:"name,omitempty"`

	// AuthType displays as a badge next to the name of the API
	AuthType AuthType `json:"auth_type,omitempty"`

	// Show toggles visibility of the API in the portal catalogue
	Show bool `json:"show,omitempty"`

	// TODO: I don't think this is exposed to the default portal templates.
	ShortDescription string `json:"short_description,omitempty"`

	// LongDescription can be markdown. It allows you to describe the capabilities of the API and is displayed just
	// below the name and AuthType in the catalogue listing page.
	LongDescription string `json:"long_description,omitempty"`

	// IsKeyless toggles visibility of the `Request an API Key button`. Use this when AuthType is keyless, jwt or oauth.
	IsKeyless bool `json:"is_keyless,omitempty"`

	// Version should always be v2
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum=v2
	// +kubebuilder:default=v2
	Version string `json:"version,omitempty"`

	// Config allows you to optionally override various fields in the PortalConfig.
	// TODO: This is an advanced capability which has not been fully tested with Tyk Operator as yet.
	Config *PortalModelPortalConfig `json:"config,omitempty"`

	// Fields is a generic map of key:value pairs.
	// You may wish to use this to tag a catalogue as type:internal or type:public
	// Then apply logic at the template layer to dynamically display catalogue apis to different user types.
	Fields map[string]string `json:"fields,omitempty"`

	// PolicyID explicitly sets the policy_id to be published. We do not recommend that this value is set directly.
	// Rather, use `policyRef` instead.
	PolicyID string `json:"policy_id,omitempty"`

	// Do not set Documentation. Use `docs` instead.
	Documentation string `json:"documentation,omitempty"`
}

type PortalModelPortalConfig struct {
	// Set by the server. DO NOT set this field it is read only.
	Id string `json:"id,omitempty"`

	// OrgID is the organization ID
	OrgID string `json:"org_id,omitempty"`

	// SignUpFields is a slice of fields which are asked of the portal developer when they register for an account
	SignUpFields []string `json:"signup_fields,omitempty"`

	// KeyRequestFields is a slice of fields which are asked of the portal developer when requesting an api key
	KeyRequestFields []string `json:"key_request_fields,omitempty"`

	// RequireKeyApproval requires reviewing of all key requests before approving
	// them. By default developers will auto-enroll into an API and be given an API
	// key. If you wish to review key requests before giving developers access to
	// the API, enable this option and you will manually need to approve them in
	// the 'Key requests' section of the dashboard.
	RequireKeyApproval bool `json:"require_key_approval,omitempty"`

	// SecureKeyApproval enables Secure key approval.By default, API keys when
	// approved are sent in an email to the Developer. By enabling this option, the
	// Developer will instead receive a link where they can go to generate the API
	// key securely
	SecureKeyApproval bool `json:"secure_key_approval,omitempty"`

	// RedirectOnKeyRequest redirects key requests. WHen set to true it will
	// redirect key requests to the url specified in RedirectTo field
	RedirectOnKeyRequest bool `json:"redirect_on_key_request,omitempty"`

	// RedirectTo is a url used to redirect key requests
	RedirectTo string `json:"redirect_to,omitempty"`

	// EnableMultiSelection enables subscribing to multiple APIs with single
	// key.User will be able subscript any combination of exposed catalogues of the
	// same authentication type
	EnableMultiSelection bool `json:"enable_multi_selection,omitempty"`

	// DisableLogin disables login on developer portal.If you do not want
	// developers to be able to login, set this value to true. You can use this
	// configuration option in the portal templates to display or hide the login
	// form as well as disabling the actual login method.
	DisableLogin bool `json:"disable_login,omitempty"`

	// DisableSignup stop developer sign-up to the portal.This will stop developers
	// from being able to signup, they will need to be created manually, or
	// assigned keys via the key management API.
	DisableSignup bool `json:"disable_signup,omitempty"`

	DisableAutoLogin bool `json:"disable_auto_login,omitempty"`

	// CatalogueLoginOnly limits access to catalogues for login users only.
	CatalogueLoginOnly bool `json:"catalogue_login_only,omitempty"`

	// OAuthUsageLimit is the maximum permitted number of OAuth clients
	OAuthUsageLimit int `json:"oauth_usage_limit,omitempty"`

	Email string `json:"email,omitempty"`

	MailOptions *MailOptions `json:"mail_options,omitempty"`

	// EnableDCR activates dynamic client registration.
	EnableDCR bool `json:"enable_dcr,omitempty"`

	// DCROptions dynamic client registration options.
	DCROptions *DCROptions `json:"dcr_options,omitempty"`

	// Override overides global settings. These Catalogue settings are currently
	// being overwritten by the Global Catalogue settings. Toggle the checkbox
	// below to override them for this Catalogue.
	Override bool `json:"override,omitempty"`
}

// +kubebuilder:validation:Enum=client_credentials;authorization_code;refresh_token
type GrantTypeEnum string

// +kubebuilder:validation:Enum=code;token
type ResponseTypeEnum string

// DCROptions are the configuration metadata for dynamic client registration. To enable DCR, ensure EnableDCR is true.
type DCROptions struct {
	// IDPHost is the fully qualified hostname of the Identity Provider.
	// e.g. https://mysubdomain.eu.auth0.com
	IDPHost string `json:"idp_host"`

	// RegistrationEndpoint is the registration_endpoint as presented in the /.well-known/openid-configuration document.
	RegistrationEndpoint string `json:"registration_endpoint"`

	// AccessToken represents an optional bearer token to authenticate with against the registration endpoint
	AccessToken string `json:"access_token,omitempty"`

	// Provider is an optional enum of the provider which allows Tyk to register clients outside the standard DCR spec
	// and perform provider specific logic.
	// If your provider is not in this list, please omit. Upon failure, submit a support ticket so that we may extend
	// support for your provider.
	// +kubebuilder:validation:Enum=gluu;keycloak;okta
	Provider string `json:"provider,omitempty"`

	// GrantTypes is an array of OAuth 2.0 grant type strings that the client can use at
	// the token endpoint.
	GrantTypes []GrantTypeEnum `json:"grant_types"`

	// ResponseTypes is an array of OAuth 2.0 response type strings that the client can
	// use at the authorization endpoint.
	ResponseTypes []ResponseTypeEnum `json:"response_types,omitempty"`

	// TokenEndpointAuthMethod is an indicator of the requested authentication method for the token endpoint.
	// "none": The client is a public client and does not have a client secret.
	// "client_secret_post": The client uses the HTTP POST parameters
	// "client_secret_basic": The client uses HTTP Basic authentication
	// +kubebuilder:validation:Enum=client_secret_basic;client_secret_post;client_secret_jwt;private_key_jwt;none
	TokenEndpointAuthMethod string `json:"token_endpoint_auth_method"`
}

type MailOptions struct {
	MailFromName  string           `json:"mail_from_name"`
	MailFromEmail string           `json:"mail_from_email"`
	EmailCopy     EmailCopyOptions `json:"email_copy"`
}

type EmailCopyOptions struct {
	WelcomeEmail       EmailConfigMeta `json:"welcome_email"`
	APIKeyEmail        EmailConfigMeta `json:"key_email"`
	ResetPasswordEmail EmailConfigMeta `json:"reset_password_email"`
}

type EmailConfigMeta struct {
	Enabled       bool   `bson:"enabled" json:"enabled"`
	EmailSubject  string `bson:"subject" json:"subject"`
	EmailBody     string `bson:"body" json:"body"`
	EmailSignoff  string `bson:"sign_off" json:"sign_off"`
	HideTokenData bool   `bson:"hide_token_data" json:"hide_token_data"`
}
