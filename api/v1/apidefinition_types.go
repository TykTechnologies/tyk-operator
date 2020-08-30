/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	"encoding/json"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ApiDefinitionSpec defines the desired state of ApiDefinition
type AuthProviderCode string
type SessionProviderCode string
type StorageEngineCode string
type TykEvent string
type TykEventHandlerName string
type EndpointMethodAction string
type TemplateMode string
type MiddlewareDriver string
type IdExtractorSource string
type IdExtractorType string
type AuthTypeEnum string
type RoutingTriggerOnType string

type NotificationsManager struct {
	SharedSecret      string `json:"shared_secret"`
	OAuthKeyChangeURL string `json:"oauth_on_keychange_url"`
}

type EndpointMethodMeta struct {
	Action  EndpointMethodAction `json:"action"`
	Code    int                  `json:"code"`
	Data    string               `json:"data"`
	Headers map[string]string    `json:"headers"`
}

type EndPointMeta struct {
	Path          string                        `json:"path"`
	IgnoreCase    bool                          `json:"ignore_case"`
	MethodActions map[string]EndpointMethodMeta `json:"method_actions"`
}

type CacheMeta struct {
	Method                 string `json:"method"`
	Path                   string `json:"path"`
	CacheKeyRegex          string `json:"cache_key_regex"`
	CacheOnlyResponseCodes []int  `json:"cache_response_codes"`
}

type RequestInputType string

type TemplateData struct {
	Input          RequestInputType `json:"input_type"`
	Mode           TemplateMode     `json:"template_mode"`
	EnableSession  bool             `json:"enable_session"`
	TemplateSource string           `json:"template_source"`
}

type TemplateMeta struct {
	TemplateData TemplateData `json:"template_data"`
	Path         string       `json:"path"`
	Method       string       `json:"method"`
}

type TransformJQMeta struct {
	Filter string `json:"filter"`
	Path   string `json:"path"`
	Method string `json:"method"`
}

type HeaderInjectionMeta struct {
	DeleteHeaders []string          `json:"delete_headers"`
	AddHeaders    map[string]string `json:"add_headers"`
	Path          string            `json:"path"`
	Method        string            `json:"method"`
	ActOnResponse bool              `json:"act_on"`
}

type HardTimeoutMeta struct {
	Path    string `json:"path"`
	Method  string `json:"method"`
	TimeOut int    `json:"timeout"`
}

type TrackEndpointMeta struct {
	Path   string `json:"path"`
	Method string `json:"method"`
}

type InternalMeta struct {
	Path   string `json:"path"`
	Method string `json:"method"`
}

type RequestSizeMeta struct {
	Path      string `json:"path"`
	Method    string `json:"method"`
	SizeLimit int64  `json:"size_limit"`
}

type CircuitBreakerMeta struct {
	Path                 string `json:"path"`
	Method               string `json:"method"`
	ThresholdPercent     int    `json:"threshold_percent"`
	Samples              int64  `json:"samples"`
	ReturnToServiceAfter int    `json:"return_to_service_after"`
}

type StringRegexMap struct {
	MatchPattern string `json:"match_rx"`
	Reverse      bool   `json:"reverse"`
	//matchRegex   *regexp.Regexp
}

type RoutingTriggerOptions struct {
	HeaderMatches         map[string]StringRegexMap `json:"header_matches"`
	QueryValMatches       map[string]StringRegexMap `json:"query_val_matches"`
	PathPartMatches       map[string]StringRegexMap `json:"path_part_matches"`
	SessionMetaMatches    map[string]StringRegexMap `json:"session_meta_matches"`
	RequestContextMatches map[string]StringRegexMap `json:"request_context_matches"`
	PayloadMatches        StringRegexMap            `json:"payload_matches"`
}

type RoutingTrigger struct {
	On        RoutingTriggerOnType  `json:"on"`
	Options   RoutingTriggerOptions `json:"options"`
	RewriteTo string                `json:"rewrite_to"`
}

type URLRewriteMeta struct {
	Path         string           `json:"path"`
	Method       string           `json:"method"`
	MatchPattern string           `json:"match_pattern"`
	RewriteTo    string           `json:"rewrite_to"`
	Triggers     []RoutingTrigger `json:"triggers"`
	//MatchRegexp  *regexp.Regexp   `json:"-"`
}

type VirtualMeta struct {
	ResponseFunctionName string `json:"response_function_name"`
	FunctionSourceType   string `json:"function_source_type"`
	FunctionSourceURI    string `json:"function_source_uri"`
	Path                 string `json:"path"`
	Method               string `json:"method"`
	UseSession           bool   `json:"use_session"`
	ProxyOnError         bool   `json:"proxy_on_error"`
}

type MethodTransformMeta struct {
	Path     string `json:"path"`
	Method   string `json:"method"`
	ToMethod string `json:"to_method"`
}

type ValidatePathMeta struct {
	Path   string `json:"path"`
	Method string `json:"method"`
	//Schema    map[string]interface{} `json:"schema"`
	SchemaB64 string `json:"schema_b64,omitempty"`
	// Allows override of default 422 Unprocessible Entity response code for validation errors.
	ErrorResponseCode int `json:"error_response_code"`
}

type ExtendedPathsSet struct {
	Ignored                 []EndPointMeta        `json:"ignored,omitempty"`
	WhiteList               []EndPointMeta        `json:"white_list,omitempty"`
	BlackList               []EndPointMeta        `json:"black_list,omitempty"`
	Cached                  []string              `json:"cache,omitempty"`
	AdvanceCacheConfig      []CacheMeta           `json:"advance_cache_config,omitempty"`
	Transform               []TemplateMeta        `json:"transform,omitempty"`
	TransformResponse       []TemplateMeta        `json:"transform_response,omitempty"`
	TransformJQ             []TransformJQMeta     `json:"transform_jq,omitempty"`
	TransformJQResponse     []TransformJQMeta     `json:"transform_jq_response,omitempty"`
	TransformHeader         []HeaderInjectionMeta `json:"transform_headers,omitempty"`
	TransformResponseHeader []HeaderInjectionMeta `json:"transform_response_headers,omitempty"`
	HardTimeouts            []HardTimeoutMeta     `json:"hard_timeouts,omitempty"`
	CircuitBreaker          []CircuitBreakerMeta  `json:"circuit_breakers,omitempty"`
	URLRewrite              []URLRewriteMeta      `json:"url_rewrites,omitempty"`
	Virtual                 []VirtualMeta         `json:"virtual,omitempty"`
	SizeLimit               []RequestSizeMeta     `json:"size_limits,omitempty"`
	MethodTransforms        []MethodTransformMeta `json:"method_transforms,omitempty"`
	TrackEndpoints          []TrackEndpointMeta   `json:"track_endpoints,omitempty"`
	DoNotTrackEndpoints     []TrackEndpointMeta   `json:"do_not_track_endpoints,omitempty"`
	ValidateJSON            []ValidatePathMeta    `json:"validate_json,omitempty"`
	Internal                []InternalMeta        `json:"internal,omitempty"`
}

type VersionInfo struct {
	Name                        string            `json:"name"`
	Expires                     string            `json:"expires"`
	Paths                       VersionInfoPaths  `json:"paths"`
	UseExtendedPaths            bool              `json:"use_extended_paths"`
	ExtendedPaths               ExtendedPathsSet  `json:"extended_paths"`
	GlobalHeaders               map[string]string `json:"global_headers"`
	GlobalHeadersRemove         []string          `json:"global_headers_remove"`
	GlobalResponseHeaders       map[string]string `json:"global_response_headers"`
	GlobalResponseHeadersRemove []string          `json:"global_response_headers_remove"`
	IgnoreEndpointCase          bool              `json:"ignore_endpoint_case"`
	GlobalSizeLimit             int64             `json:"global_size_limit"`
	OverrideTarget              string            `json:"override_target"`
}

type VersionInfoPaths struct {
	Ignored   []string `json:"ignored"`
	WhiteList []string `json:"white_list"`
	BlackList []string `json:"black_list"`
}

type AuthProviderMeta struct {
	Name          AuthProviderCode  `json:"name"`
	StorageEngine StorageEngineCode `json:"storage_engine"`
	//Meta          map[string]interface{} `json:"meta"`
}

type SessionProviderMeta struct {
	Name          SessionProviderCode `json:"name"`
	StorageEngine StorageEngineCode   `json:"storage_engine"`
	//Meta          map[string]interface{} `json:"meta"`
}

type EventHandlerTriggerConfig struct {
	Handler TykEventHandlerName `json:"handler_name"`
	//HandlerMeta map[string]interface{} `json:"handler_meta"`
}

type EventHandlerMetaConfig struct {
	Events map[TykEvent][]EventHandlerTriggerConfig `json:"events"`
}

type MiddlewareDefinition struct {
	Name           string `json:"name"`
	Path           string `json:"path"`
	RequireSession bool   `json:"require_session"`
	RawBodyOnly    bool   `json:"raw_body_only"`
}

type MiddlewareIdExtractor struct {
	ExtractFrom IdExtractorSource `json:"extract_from"`
	ExtractWith IdExtractorType   `json:"extract_with"`
	//ExtractorConfig map[string]interface{} `json:"extractor_config"`
}

type MiddlewareSection struct {
	Pre         []MiddlewareDefinition `json:"pre"`
	Post        []MiddlewareDefinition `json:"post"`
	PostKeyAuth []MiddlewareDefinition `json:"post_key_auth"`
	AuthCheck   MiddlewareDefinition   `json:"auth_check"`
	Response    []MiddlewareDefinition `json:"response"`
	Driver      MiddlewareDriver       `json:"driver"`
	IdExtractor MiddlewareIdExtractor  `json:"id_extractor"`
}

type CacheOptions struct {
	CacheTimeout               int64    `json:"cache_timeout"`
	EnableCache                bool     `json:"enable_cache"`
	CacheAllSafeRequests       bool     `json:"cache_all_safe_requests"`
	CacheOnlyResponseCodes     []int    `json:"cache_response_codes"`
	EnableUpstreamCacheControl bool     `json:"enable_upstream_cache_control"`
	CacheControlTTLHeader      string   `json:"cache_control_ttl_header"`
	CacheByHeaders             []string `json:"cache_by_headers"`
}

type ResponseProcessor struct {
	Name string `json:"name"`
	//Options interface{} `json:"options"`
}

type HostCheckObject struct {
	CheckURL            string            `json:"url"`
	Protocol            string            `json:"protocol"`
	Timeout             time.Duration     `json:"timeout"`
	EnableProxyProtocol bool              `json:"enable_proxy_protocol"`
	Commands            []CheckCommand    `json:"commands"`
	Method              string            `json:"method"`
	Headers             map[string]string `json:"headers"`
	Body                string            `json:"body"`
}

type CheckCommand struct {
	Name    string `json:"name"`
	Message string `json:"message"`
}

type ServiceDiscoveryConfiguration struct {
	UseDiscoveryService bool   `json:"use_discovery_service"`
	QueryEndpoint       string `json:"query_endpoint"`
	UseNestedQuery      bool   `json:"use_nested_query"`
	ParentDataPath      string `json:"parent_data_path"`
	DataPath            string `json:"data_path"`
	PortDataPath        string `json:"port_data_path"`
	TargetPath          string `json:"target_path"`
	UseTargetList       bool   `json:"use_target_list"`
	CacheTimeout        int64  `json:"cache_timeout"`
	EndpointReturnsList bool   `json:"endpoint_returns_list"`
}

type OIDProviderConfig struct {
	Issuer    string            `json:"issuer"`
	ClientIDs map[string]string `json:"client_ids"`
}

type OpenIDOptions struct {
	Providers         []OIDProviderConfig `json:"providers"`
	SegregateByClient bool                `json:"segregate_by_client"`
}

// APIDefinition represents the configuration for a single proxied API and it's versions.
// +kubebuilder:object:generate=true
type APIDefinitionSpec struct {
	APIID string `json:"api_id"`
	OrgID string `json:"org_id"`
	//Id                         string                `json:"id,omitempty"`
	Name                       string                `json:"name"`
	Slug                       string                `json:"slug"`
	Proxy                      Proxy                 `json:"proxy"`
	ListenPort                 int                   `json:"listen_port"`
	Protocol                   string                `json:"protocol"`
	EnableProxyProtocol        bool                  `json:"enable_proxy_protocol"`
	UseKeylessAccess           bool                  `json:"use_keyless"`
	UseOauth2                  bool                  `json:"use_oauth2"`
	UseOpenID                  bool                  `json:"use_openid"`
	OpenIDOptions              OpenIDOptions         `json:"openid_options"`
	Oauth2Meta                 OAuth2Meta            `json:"oauth_meta"`
	Auth                       AuthConfig            `json:"auth"` // Deprecated: Use AuthConfigs instead.
	AuthConfigs                map[string]AuthConfig `json:"auth_configs"`
	UseBasicAuth               bool                  `json:"use_basic_auth"`
	BasicAuth                  BasicAuthMeta         `json:"basic_auth"`
	UseMutualTLSAuth           bool                  `json:"use_mutual_tls_auth"`
	ClientCertificates         []string              `json:"client_certificates"`
	UpstreamCertificates       map[string]string     `json:"upstream_certificates"`
	PinnedPublicKeys           map[string]string     `json:"pinned_public_keys"`
	EnableJWT                  bool                  `json:"enable_jwt"`
	UseStandardAuth            bool                  `json:"use_standard_auth"`
	UseGoPluginAuth            bool                  `json:"use_go_plugin_auth"`
	EnableCoProcessAuth        bool                  `json:"enable_coprocess_auth"`
	JWTSigningMethod           string                `json:"jwt_signing_method"`
	JWTSource                  string                `json:"jwt_source"`
	JWTIdentityBaseField       string                `json:"jwt_identity_base_field"`
	JWTClientIDBaseField       string                `json:"jwt_client_base_field"`
	JWTPolicyFieldName         string                `json:"jwt_policy_field_name"`
	JWTDefaultPolicies         []string              `json:"jwt_default_policies"`
	JWTIssuedAtValidationSkew  uint64                `json:"jwt_issued_at_validation_skew"`
	JWTExpiresAtValidationSkew uint64                `json:"jwt_expires_at_validation_skew"`
	JWTNotBeforeValidationSkew uint64                `json:"jwt_not_before_validation_skew"`
	JWTSkipKid                 bool                  `json:"jwt_skip_kid"`
	JWTScopeToPolicyMapping    map[string]string     `json:"jwt_scope_to_policy_mapping"`
	JWTScopeClaimName          string                `json:"jwt_scope_claim_name"`
	NotificationsDetails       NotificationsManager  `json:"notifications"`
	EnableSignatureChecking    bool                  `json:"enable_signature_checking"`
	HmacAllowedClockSkew       json.Number           `json:"hmac_allowed_clock_skew"` // TODO: convert to float64
	HmacAllowedAlgorithms      []string              `json:"hmac_allowed_algorithms"`
	RequestSigning             RequestSigningMeta    `json:"request_signing"`
	BaseIdentityProvidedBy     AuthTypeEnum          `json:"base_identity_provided_by"`
	VersionDefinition          VersionDefinition     `json:"definition"`
	VersionData                VersionData           `json:"version_data"`
	//UptimeTests                UptimeTests           `json:"uptime_tests"`

	DisableRateLimit       bool                `json:"disable_rate_limit"`
	DisableQuota           bool                `json:"disable_quota"`
	CustomMiddleware       MiddlewareSection   `json:"custom_middleware"`
	CustomMiddlewareBundle string              `json:"custom_middleware_bundle"`
	CacheOptions           CacheOptions        `json:"cache_options"`
	SessionLifetime        int64               `json:"session_lifetime"`
	Active                 bool                `json:"active"`
	Internal               bool                `json:"internal"`
	AuthProvider           AuthProviderMeta    `json:"auth_provider"`
	SessionProvider        SessionProviderMeta `json:"session_provider"`
	//EventHandlers             EventHandlerMetaConfig `json:"event_handlers"`
	EnableBatchRequestSupport bool `json:"enable_batch_request_support"`
	EnableIpWhiteListing      bool `json:"enable_ip_whitelisting"`
	// +optional
	AllowedIPs            []string            `json:"allowed_ips"`
	EnableIpBlacklisting  bool                `json:"enable_ip_blacklisting"`
	BlacklistedIPs        []string            `json:"blacklisted_ips"`
	DontSetQuotasOnCreate bool                `json:"dont_set_quota_on_create"`
	ExpireAnalyticsAfter  int64               `json:"expire_analytics_after"` // must have an expireAt TTL index set (http://docs.mongodb.org/manual/tutorial/expire-data/)
	ResponseProcessors    []ResponseProcessor `json:"response_processors"`
	// +optional
	CORS              CORS     `json:"CORS"`
	Domain            string   `json:"domain"`
	Certificates      []string `json:"certificates"`
	DoNotTrack        bool     `json:"do_not_track"`
	Tags              []string `json:"tags"`
	EnableContextVars bool     `json:"enable_context_vars"`
	//ConfigData              map[string]interface{} `json:"config_data"`
	TagHeaders              []string        `json:"tag_headers"`
	GlobalRateLimit         GlobalRateLimit `json:"global_rate_limit"`
	StripAuthData           bool            `json:"strip_auth_data"`
	EnableDetailedRecording bool            `json:"enable_detailed_recording"`
	GraphQL                 GraphQLConfig   `json:"graphql"`
}

type Proxy struct {
	PreserveHostHeader          bool                          `json:"preserve_host_header"`
	ListenPath                  string                        `json:"listen_path"`
	TargetURL                   string                        `json:"target_url"`
	DisableStripSlash           bool                          `json:"disable_strip_slash"`
	StripListenPath             bool                          `json:"strip_listen_path"`
	EnableLoadBalancing         bool                          `json:"enable_load_balancing"`
	Targets                     []string                      `json:"target_list"`
	CheckHostAgainstUptimeTests bool                          `json:"check_host_against_uptime_tests"`
	ServiceDiscovery            ServiceDiscoveryConfiguration `json:"service_discovery"`
	Transport                   ProxyTransport                `json:"transport"`
}

type ProxyTransport struct {
	SSLInsecureSkipVerify   bool     `json:"ssl_insecure_skip_verify"`
	SSLCipherSuites         []string `json:"ssl_ciphers"`
	SSLMinVersion           uint16   `json:"ssl_min_version"`
	SSLForceCommonNameCheck bool     `json:"ssl_force_common_name_check"`
	ProxyURL                string   `json:"proxy_url"`
}

type CORS struct {
	Enable             bool     `json:"enable"`
	AllowedOrigins     []string `json:"allowed_origins"`
	AllowedMethods     []string `json:"allowed_methods"`
	AllowedHeaders     []string `json:"allowed_headers"`
	ExposedHeaders     []string `json:"exposed_headers"`
	AllowCredentials   bool     `json:"allow_credentials"`
	MaxAge             int      `json:"max_age"`
	OptionsPassthrough bool     `json:"options_passthrough"`
	Debug              bool     `json:"debug"`
}

type UptimeTests struct {
	CheckList []HostCheckObject `json:"check_list"`
	Config    UptimeTestConfig  `json:"config"`
}

type UptimeTestConfig struct {
	ExpireUptimeAnalyticsAfter int64                         `json:"expire_utime_after"` // must have an expireAt TTL index set (http://docs.mongodb.org/manual/tutorial/expire-data/)
	ServiceDiscovery           ServiceDiscoveryConfiguration `json:"service_discovery"`
	RecheckWait                int                           `json:"recheck_wait"`
}

type VersionData struct {
	NotVersioned   bool                   `json:"not_versioned"`
	DefaultVersion string                 `json:"default_version"`
	Versions       map[string]VersionInfo `json:"versions"`
}

type VersionDefinition struct {
	Location  string `json:"location"`
	Key       string `json:"key"`
	StripPath bool   `json:"strip_path"`
}

type BasicAuthMeta struct {
	DisableCaching     bool   `json:"disable_caching"`
	CacheTTL           int    `json:"cache_ttl"`
	ExtractFromBody    bool   `json:"extract_from_body"`
	BodyUserRegexp     string `json:"body_user_regexp"`
	BodyPasswordRegexp string `json:"body_password_regexp"`
}

type OAuth2Meta struct {
	AllowedAccessTypes     []string `json:"allowed_access_types"`    // osin.AccessRequestType
	AllowedAuthorizeTypes  []string `json:"allowed_authorize_types"` // osin.AuthorizeRequestType
	AuthorizeLoginRedirect string   `json:"auth_login_redirect"`
}

type AuthConfig struct {
	UseParam          bool            `json:"use_param"`
	ParamName         string          `json:"param_name"`
	UseCookie         bool            `json:"use_cookie"`
	CookieName        string          `json:"cookie_name"`
	AuthHeaderName    string          `json:"auth_header_name"`
	UseCertificate    bool            `json:"use_certificate"`
	ValidateSignature bool            `json:"validate_signature"`
	Signature         SignatureConfig `json:"signature,omitempty"`
}

type SignatureConfig struct {
	Algorithm        string `json:"algorithm"`
	Header           string `json:"header"`
	Secret           string `json:"secret"`
	AllowedClockSkew int64  `json:"allowed_clock_skew"`
	ErrorCode        int    `json:"error_code"`
	ErrorMessage     string `json:"error_message"`
}

type GlobalRateLimit struct {
	Rate json.Number `json:"rate"`
	Per  json.Number `json:"per"`
}

type BundleManifest struct {
	FileList         []string          `json:"file_list"`
	CustomMiddleware MiddlewareSection `json:"custom_middleware"`
	Checksum         string            `json:"checksum"`
	Signature        string            `json:"signature"`
}

type RequestSigningMeta struct {
	IsEnabled       bool     `json:"is_enabled"`
	Secret          string   `json:"secret"`
	KeyId           string   `json:"key_id"`
	Algorithm       string   `json:"algorithm"`
	HeaderList      []string `json:"header_list"`
	CertificateId   string   `json:"certificate_id"`
	SignatureHeader string   `json:"signature_header"`
}

// GraphQLConfig is the root config object for a GraphQL API.
type GraphQLConfig struct {
	// Enabled indicates if GraphQL should be enabled.
	Enabled bool `json:"enabled"`
	// ExecutionMode is the mode to define how an api behaves.
	ExecutionMode GraphQLExecutionMode `json:"execution_mode"`
	// Schema is the GraphQL Schema exposed by the GraphQL API/Upstream/Engine.
	Schema string `json:"schema"`
	// LastSchemaUpdate contains the date and time of the last triggered schema update to the upstream
	//LastSchemaUpdate *time.Time `json:"last_schema_update,omitempty"`
	// TypeFieldConfigurations is a rule set of data source and mapping of a schema field.
	TypeFieldConfigurations []TypeFieldConfiguration `json:"type_field_configurations"`
	// GraphQLPlayground is the Playground specific configuration.
	GraphQLPlayground GraphQLPlayground `json:"playground"`
}

type TypeFieldConfiguration struct {
	TypeName   string                `json:"type_name"`
	FieldName  string                `json:"field_name"`
	Mapping    *MappingConfiguration `json:"mapping"`
	DataSource SourceConfig          `json:"data_source"`
}

type SourceConfig struct {
	// Kind defines the unique identifier of the DataSource
	// Kind needs to match to the Planner "DataSourceName" name
	Name string `json:"kind"`
	// Config is the DataSource specific configuration object
	// Each Planner needs to make sure to parse their Config Object correctly
	Config json.RawMessage `json:"data_source_config"`
}

type MappingConfiguration struct {
	Disabled bool   `json:"disabled"`
	Path     string `json:"path"`
}

// GraphQLExecutionMode is the mode in which the GraphQL Middleware should operate.
type GraphQLExecutionMode string

// GraphQLPlayground represents the configuration for the public playground which will be hosted alongside the api.
type GraphQLPlayground struct {
	// Enabled indicates if the playground should be enabled.
	Enabled bool `json:"enabled"`
	// Path sets the path on which the playground will be hosted if enabled.
	Path string `json:"path"`
}

// ApiDefinitionStatus defines the observed state of ApiDefinition
type ApiDefinitionStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Hashes []string `json:"hashes"`
	Ids    []string `json:"paths"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// ApiDefinition is the Schema for the apidefinitions API
type ApiDefinition struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   APIDefinitionSpec   `json:"spec,omitempty"`
	Status ApiDefinitionStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ApiDefinitionList contains a list of ApiDefinition
type ApiDefinitionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []APIDefinitionSpec `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ApiDefinition{}, &ApiDefinitionList{})
}
