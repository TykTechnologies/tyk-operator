package v1alpha1

type APICatalogueSpec struct {
	Id    string           `json:"id"`
	OrgId string           `json:"org_id"`
	APIS  []APIDescription `json:"apis"`
	Email string           `json:"email"`
}

type CatalogueVersion string

type APIDescription struct {
	Name             string                   `json:"name"`
	ShortDescription string                   `json:"short_description"`
	LongDescription  string                   `json:"long_description"`
	Show             bool                     `json:"show"`
	APIID            string                   `json:"api_id"`
	PolicyID         string                   `json:"policy_id"`
	Documentation    string                   `json:"documentation"`
	Version          *CatalogueVersion        `json:"version,omitempty"`
	IsKeyless        bool                     `json:"is_keyless,omitempty"`
	Config           *PortalModelPortalConfig `son:"config,omitempty"`
	Fields           map[string]string        `json:"fields,omitempty"`
	AuthType         string                   `json:"auth_type,omitempty"`
}

type PortalModelPortalConfig struct {
	Id                   string       `json:"id"`
	OrgID                string       `json:"org_id"`
	SignUpFields         []string     `json:"signup_fields"`
	KeyRequestFields     []string     `json:"key_request_fields"`
	RequireKeyApproval   bool         `json:"require_key_approval"`
	SecureKeyApproval    bool         `json:"secure_key_approval"`
	RedirectOnKeyRequest bool         `json:"redirect_on_key_request"`
	RedirectTo           string       `json:"redirect_to"`
	EnableMultiSelection bool         `json:"enable_multi_selection"`
	DisableLogin         bool         `json:"disable_login"`
	DisableSignup        bool         `json:"disable_signup"`
	DisableAutoLogin     bool         `json:"disable_auto_login"`
	CatalogueLoginOnly   bool         `son:"catalogue_login_only"`
	OAuthUsageLimit      int          `json:"oauth_usage_limit"`
	Email                string       `json:"email"`
	MailOptions          *MailOptions `json:"mail_options"`
	DCROptions           DCROptions   `json:"dcr_options"`
	EnableDCR            bool         `json:"enable_dcr"`
	Override             bool         `json:"override"`
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
