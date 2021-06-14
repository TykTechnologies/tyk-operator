package model

type APICatalogue struct {
	Id    string           `json:"id"`
	OrgId string           `json:"org_id"`
	APIS  []APIDescription `json:"apis"`
	Email string           `json:"email"`
}

type CatalogueVersion string

type APIDescription struct {
	Name             string                   `json:"name,omitempty"`
	ShortDescription string                   `json:"short_description,omitempty"`
	LongDescription  string                   `json:"long_description,omitempty"`
	Show             bool                     `json:"show,omitempty"`
	PolicyID         string                   `json:"policy_id,omitempty"`
	Documentation    string                   `json:"documentation,omitempty"`
	Version          CatalogueVersion         `json:"version,omitempty,omitempty"`
	IsKeyless        bool                     `json:"is_keyless,omitempty"`
	Config           *PortalModelPortalConfig `json:"config,omitempty"`
	Fields           map[string]string        `json:"fields,omitempty"`
	AuthType         string                   `json:"auth_type,omitempty"`
}

type PortalModelPortalConfig struct {
	Id                   string       `json:"id,omitempty"`
	OrgID                string       `json:"org_id,omitempty"`
	SignUpFields         []string     `json:"signup_fields,omitempty"`
	KeyRequestFields     []string     `json:"key_request_fields,omitempty"`
	RequireKeyApproval   bool         `json:"require_key_approval,omitempty"`
	SecureKeyApproval    bool         `json:"secure_key_approval,omitempty"`
	RedirectOnKeyRequest bool         `json:"redirect_on_key_request,omitempty"`
	RedirectTo           string       `json:"redirect_to,omitempty"`
	EnableMultiSelection bool         `json:"enable_multi_selection,omitempty"`
	DisableLogin         bool         `json:"disable_login,omitempty"`
	DisableSignup        bool         `json:"disable_signup,omitempty"`
	DisableAutoLogin     bool         `json:"disable_auto_login,omitempty"`
	CatalogueLoginOnly   bool         `json:"catalogue_login_only,omitempty"`
	OAuthUsageLimit      int          `json:"oauth_usage_limit,omitempty"`
	Email                string       `json:"email,omitempty"`
	MailOptions          *MailOptions `json:"mail_options,omitempty"`
	DCROptions           *DCROptions  `json:"dcr_options,omitempty"`
	EnableDCR            bool         `json:"enable_dcr,omitempty"`
	Override             bool         `json:"override,omitempty"`
}

type DCROptions struct {
	IDPHost                 string   `json:"idp_host"`
	AccessToken             string   `json:"access_token"`
	RegistrationEndpoint    string   `json:"registration_endpoint"`
	Provider                string   `json:"provider"`
	GrantTypes              []string `json:"grant_types"`
	ResponseTypes           []string `json:"response_types"`
	TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method"`
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
