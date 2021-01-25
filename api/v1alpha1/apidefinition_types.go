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

// Method represents HTTP request method
// +kubebuilder:validation:Enum=GET;POST;PUT;PATCH;DELETE;OPTIONS;HEAD;CONNECT;TRACE
type HttpMethod string

// JWTSigningMethod algorithm used to sign jwt token
// +kubebuilder:validation:Enum=rsa;hmac;ecdsa
type JWTSigningMethod string

// ExecutionMode is the mode to define how an api behaves.
// +kubebuilder:validation:Enum=proxyOnly;executionEngine
type GraphQLExecutionMode string

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
	Method                 HttpMethod `json:"method"`
	Path                   string     `json:"path"`
	CacheKeyRegex          string     `json:"cache_key_regex"`
	CacheOnlyResponseCodes []int      `json:"cache_response_codes"`
}

type RequestInputType string

type TemplateData struct {
	Input          RequestInputType `json:"input_type"`
	Mode           TemplateMode     `json:"template_mode"`
	EnableSession  bool             `json:"enable_session"`
	TemplateSource string           `json:"template_source"`
	//FromDashboard  bool             `json:"from_dashboard"`
}

type TemplateMeta struct {
	TemplateData TemplateData `json:"template_data"`
	Path         string       `json:"path"`
	Method       HttpMethod   `json:"method"`
}

type TransformJQMeta struct {
	Filter string     `json:"filter"`
	Path   string     `json:"path"`
	Method HttpMethod `json:"method"`
}

type HeaderInjectionMeta struct {
	DeleteHeaders []string          `json:"delete_headers"`
	AddHeaders    map[string]string `json:"add_headers"`
	Path          string            `json:"path"`
	Method        HttpMethod        `json:"method"`
	ActOnResponse bool              `json:"act_on"`
}

type HardTimeoutMeta struct {
	Path    string     `json:"path"`
	Method  HttpMethod `json:"method"`
	TimeOut int        `json:"timeout"`
}

type TrackEndpointMeta struct {
	Path   string     `json:"path"`
	Method HttpMethod `json:"method"`
}

type InternalMeta struct {
	Path   string     `json:"path"`
	Method HttpMethod `json:"method"`
}

type RequestSizeMeta struct {
	Path      string     `json:"path"`
	Method    HttpMethod `json:"method"`
	SizeLimit int64      `json:"size_limit"`
}

type CircuitBreakerMeta struct {
	Path   string     `json:"path"`
	Method HttpMethod `json:"method"`

	// ThresholdPercent is the percentage of requests that fail before breaker is tripped
	ThresholdPercent string `json:"threshold_percent"`
	// Samples defines the number of requests to base the ThresholdPercent on
	Samples int64 `json:"samples"`
	// ReturnToServiceAfter represents the time in seconds to return back to the service
	ReturnToServiceAfter int  `json:"return_to_service_after"`
	DisableHalfOpenState bool `json:"disable_half_open_state,omitempty"`
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
	// Path represents the endpoint listen path
	Path   string     `json:"path"`
	Method HttpMethod `json:"method"`
	// MatchPattern is a regular expression pattern to match the path
	MatchPattern string `json:"match_pattern"`
	// RewriteTo is the target path on the upstream, or target URL we wish to rewrite to
	RewriteTo string           `json:"rewrite_to"`
	Triggers  []RoutingTrigger `json:"triggers"`
}

type VirtualMeta struct {
	ResponseFunctionName string     `json:"response_function_name"`
	FunctionSourceType   string     `json:"function_source_type"`
	FunctionSourceURI    string     `json:"function_source_uri"`
	Path                 string     `json:"path"`
	Method               HttpMethod `json:"method"`
	UseSession           bool       `json:"use_session"`
	ProxyOnError         bool       `json:"proxy_on_error"`
}

type MethodTransformMeta struct {
	Path     string     `json:"path"`
	Method   HttpMethod `json:"method"`
	ToMethod HttpMethod `json:"to_method"`
}

type ValidatePathMeta struct {
	// Allows override of default 422 Unprocessible Entity response code for validation errors.
	ErrorResponseCode int        `json:"error_response_code"`
	Path              string     `json:"path"`
	Method            HttpMethod `json:"method"`
	SchemaB64         string     `json:"schema_b64,omitempty"`

	//Schema    ExtraFields `json:"schema,omitempty"`
}

type ExtendedPathsSet struct {
	Ignored   []EndPointMeta `json:"ignored,omitempty"`
	WhiteList []EndPointMeta `json:"white_list,omitempty"`
	BlackList []EndPointMeta `json:"black_list,omitempty"`
	// List of paths which cache middleware should be enabled on
	Cached                  []string              `json:"cache,omitempty"`
	Transform               []TemplateMeta        `json:"transform,omitempty"`
	TransformResponse       []TemplateMeta        `json:"transform_response,omitempty"`
	TransformJQ             []TransformJQMeta     `json:"transform_jq,omitempty"`
	TransformJQResponse     []TransformJQMeta     `json:"transform_jq_response,omitempty"`
	TransformHeader         []HeaderInjectionMeta `json:"transform_headers,omitempty"`
	TransformResponseHeader []HeaderInjectionMeta `json:"transform_response_headers,omitempty"`
	AdvanceCacheConfig      []CacheMeta           `json:"advance_cache_config,omitempty"`
	HardTimeouts            []HardTimeoutMeta     `json:"hard_timeouts,omitempty"`
	CircuitBreaker          []CircuitBreakerMeta  `json:"circuit_breakers,omitempty"`
	URLRewrite              []URLRewriteMeta      `json:"url_rewrites,omitempty"`
	Virtual                 []VirtualMeta         `json:"virtual,omitempty"`
	SizeLimit               []RequestSizeMeta     `json:"size_limits,omitempty"`
	MethodTransforms        []MethodTransformMeta `json:"method_transforms,omitempty"`
	TrackEndpoints          []TrackEndpointMeta   `json:"track_endpoints,omitempty"`
	DoNotTrackEndpoints     []TrackEndpointMeta   `json:"do_not_track_endpoints,omitempty"`
	//ValidateJSON            []ValidatePathMeta    `json:"validate_json,omitempty"` //  Breaking integration test?
	Internal []InternalMeta `json:"internal,omitempty"`
}

type VersionInfo struct {
	Name                        string            `json:"name"`
	Expires                     string            `json:"expires,omitempty"`
	Paths                       VersionInfoPaths  `json:"paths,omitempty"`
	UseExtendedPaths            bool              `json:"use_extended_paths,omitempty"`
	ExtendedPaths               ExtendedPathsSet  `json:"extended_paths,omitempty"`
	GlobalHeaders               map[string]string `json:"global_headers,omitempty"`
	GlobalHeadersRemove         []string          `json:"global_headers_remove,omitempty"`
	GlobalResponseHeaders       map[string]string `json:"global_response_headers,omitempty"`
	GlobalResponseHeadersRemove []string          `json:"global_response_headers_remove,omitempty"`
	IgnoreEndpointCase          bool              `json:"ignore_endpoint_case,omitempty"`
	GlobalSizeLimit             int64             `json:"global_size_limit,omitempty"`
	OverrideTarget              string            `json:"override_target,omitempty"`
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
	RequireSession bool   `json:"require_session,omitempty"`
	RawBodyOnly    bool   `json:"raw_body_only,omitempty"`
}

type IdExtractorConfig struct {
	HeaderName      string `json:"header_name,omitempty"`
	FormParamName   string `json:"param_name,omitempty"`
	RegexExpression string `json:"regex_expression,omitempty"`
	RegexMatchIndex int    `json:"regex_match_index,omitempty"`
}

type MiddlewareIdExtractor struct {
	ExtractFrom     IdExtractorSource `json:"extract_from"`
	ExtractWith     IdExtractorType   `json:"extract_with"`
	ExtractorConfig IdExtractorConfig `json:"extractor_config"`
}

type MiddlewareSection struct {
	Pre         []MiddlewareDefinition `json:"pre,omitempty"`
	Post        []MiddlewareDefinition `json:"post,omitempty"`
	PostKeyAuth []MiddlewareDefinition `json:"post_key_auth,omitempty"`
	AuthCheck   MiddlewareDefinition   `json:"auth_check,omitempty"`
	Response    []MiddlewareDefinition `json:"response,omitempty"`
	Driver      MiddlewareDriver       `json:"driver"`
	IdExtractor MiddlewareIdExtractor  `json:"id_extractor,omitempty"`
}

type CacheOptions struct {
	// EnableCache turns global cache middleware on or off.
	// It is still possible to enable caching on a per-path basis by explicitly setting the endpoint cache middleware.
	// see `spec.version_data.versions.{VERSION}.extended_paths.cache[]`
	EnableCache bool `json:"enable_cache,omitempty"`
	// CacheTimeout is the TTL for a cached object in seconds
	CacheTimeout int64 `json:"cache_timeout"`
	// CacheAllSafeRequests caches responses to (GET, HEAD, OPTIONS) requests
	// overrides per-path cache settings in versions, applies across versions
	CacheAllSafeRequests bool `json:"cache_all_safe_requests,omitempty"`
	// CacheOnlyResponseCodes is an array of response codes which are safe to cache. e.g. 404
	CacheOnlyResponseCodes []int `json:"cache_response_codes,omitempty"`
	// EnableUpstreamCacheControl instructs Tyk Cache to respect upstream cache control headers
	EnableUpstreamCacheControl bool `json:"enable_upstream_cache_control,omitempty"`
	// CacheControlTTLHeader is the response header which tells Tyk how long it is safe to cache the response for
	CacheControlTTLHeader string `json:"cache_control_ttl_header,omitempty"`
	// CacheByHeaders allows header values to be used as part of the cache key
	CacheByHeaders []string `json:"cache_by_headers,omitempty"`
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
	Method              HttpMethod        `json:"method"`
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

	// For server use only, do not use
	ID string `json:"id,omitempty"`

	// Only set this field if you are referring
	// To an existing API in def.
	// The Operator will use this API ID to link the CRD with the API in Tyk
	// Note the values in the CRD will become the new source of truth, overriding the existing API Definition
	APIID string `json:"api_id,omitempty"`

	Name string `json:"name"`

	// OrgID is overwritten - no point setting this
	OrgID string `json:"org_id,omitempty"`

	// Active specifies if the api is enabled or not
	Active bool `json:"active,omitempty"`

	// Proxy
	Proxy Proxy `json:"proxy"`

	// +optional
	ListenPort int `json:"listen_port"`

	// +kubebuilder:validation:Enum=http;https;tcp;tls
	Protocol string `json:"protocol"`

	// Domain represents a custom host header that the gateway will listen on for this API
	Domain string `json:"domain,omitempty"`

	// DoNotTrack disables endpoint tracking for this API. Default is true, you need to explicitly set it to false
	DoNotTrack *bool `json:"do_not_track,omitempty"`

	// UseKeylessAccess will switch off all key checking. Some analytics will still be recorded, but rate-limiting,
	// quotas and security policies will not be possible (there is no session to attach requests to).
	UseKeylessAccess bool `json:"use_keyless,omitempty"`

	//EnableProxyProtocol bool          `json:"enable_proxy_protocol"`
	//UseOauth2           bool          `json:"use_oauth2"`
	//UseOpenID           bool          `json:"use_openid"`
	//OpenIDOptions       OpenIDOptions `json:"openid_options"`
	//Oauth2Meta          OAuth2Meta    `json:"oauth_meta"`

	// StripAuthData ensures that any security tokens used for accessing APIs are stripped and not leaked to the upstream
	StripAuthData bool `json:"strip_auth_data,omitempty"`

	Auth AuthConfig `json:"auth,omitempty"`

	// +optional
	AuthConfigs map[string]AuthConfig `json:"auth_configs,omitempty"`

	// UseStandardAuth enables simple bearer token authentication
	UseStandardAuth bool `json:"use_standard_auth,omitempty"`

	//UseBasicAuth               bool                  `json:"use_basic_auth"`
	//BasicAuth                  BasicAuthMeta         `json:"basic_auth"`
	//UseMutualTLSAuth           bool                  `json:"use_mutual_tls_auth"`
	//ClientCertificates         []string              `json:"client_certificates"`
	//UpstreamCertificates       map[string]string     `json:"upstream_certificates"`
	//PinnedPublicKeys           map[string]string     `json:"pinned_public_keys"`

	// EnableJWT enables JSON web token authentication
	EnableJWT bool `json:"enable_jwt,omitempty"`
	//UseGoPluginAuth            bool                  `json:"use_go_plugin_auth"`

	EnableCoProcessAuth        bool              `json:"enable_coprocess_auth,omitempty"`
	JWTSigningMethod           JWTSigningMethod  `json:"jwt_signing_method,omitempty"`
	JWTSource                  string            `json:"jwt_source,omitempty"`
	JWTIdentityBaseField       string            `json:"jwt_identity_base_field,omitempty"`
	JWTClientIDBaseField       string            `json:"jwt_client_base_field,omitempty"`
	JWTPolicyFieldName         string            `json:"jwt_policy_field_name,omitempty"`
	JWTDefaultPolicies         []string          `json:"jwt_default_policies,omitempty"`
	JWTIssuedAtValidationSkew  uint64            `json:"jwt_issued_at_validation_skew,omitempty"`
	JWTExpiresAtValidationSkew uint64            `json:"jwt_expires_at_validation_skew,omitempty"`
	JWTNotBeforeValidationSkew uint64            `json:"jwt_not_before_validation_skew,omitempty"`
	JWTSkipKid                 bool              `json:"jwt_skip_kid,omitempty"`
	JWTScopeToPolicyMapping    map[string]string `json:"jwt_scope_to_policy_mapping,omitempty"`
	JWTScopeClaimName          string            `json:"jwt_scope_claim_name,omitempty"`
	//NotificationsDetails       NotificationsManager  `json:"notifications"`
	//EnableSignatureChecking    bool                  `json:"enable_signature_checking"`
	//HmacAllowedClockSkew       json.Number           `json:"hmac_allowed_clock_skew"` // TODO: convert to float64
	//HmacAllowedAlgorithms      []string              `json:"hmac_allowed_algorithms"`
	//RequestSigning             RequestSigningMeta    `json:"request_signing"`
	//BaseIdentityProvidedBy     AuthTypeEnum          `json:"base_identity_provided_by"`

	VersionDefinition VersionDefinition `json:"definition,omitempty"`

	VersionData VersionData `json:"version_data,omitempty"`

	//UptimeTests                UptimeTests           `json:"uptime_tests"`
	//DisableRateLimit       bool                `json:"disable_rate_limit"`
	//DisableQuota           bool                `json:"disable_quota"`

	CustomMiddleware MiddlewareSection `json:"custom_middleware,omitempty"`
	//CustomMiddlewareBundle string              `json:"custom_middleware_bundle"`

	CacheOptions CacheOptions `json:"cache_options,omitempty"`
	//SessionLifetime        int64               `json:"session_lifetime"`
	//Internal               bool                `json:"internal"`
	//AuthProvider           AuthProviderMeta    `json:"auth_provider"`
	//SessionProvider        SessionProviderMeta `json:"session_provider"`
	////EventHandlers             EventHandlerMetaConfig `json:"event_handlers"`
	//EnableBatchRequestSupport bool `json:"enable_batch_request_support"`

	// EnableIPWhiteListing activates the ip whitelisting middleware.
	EnableIPWhiteListing bool `json:"enable_ip_whitelisting,omitempty"`

	// AllowedIPs is a list of IP address that are whitelisted.When this is
	// provided all IP address that is not on this list will be blocked and a 403 http
	// status will be returned. The IP address can be IPv4 or IPv6.IP in
	// CIDR notation is also supported.
	AllowedIPs []string `json:"allowed_ips,omitempty"`

	// EnableIPBlacklisting activates the ip blacklisting middleware.
	EnableIPBlacklisting bool `json:"enable_ip_blacklisting,omitempty"`

	// BlacklistedIPs is a list of IP address that will be blacklisted.This means if
	// origin IP matches any IP in this list a 403 http status code will be
	// returned. The IP address can be IPv4 or IPv6. IP in CIDR notation is also
	// supported.
	BlacklistedIPs []string `json:"blacklisted_ips,omitempty"`
	//DontSetQuotasOnCreate bool                `json:"dont_set_quota_on_create"`
	//ExpireAnalyticsAfter  int64               `json:"expire_analytics_after"` // must have an expireAt TTL index set (http://docs.mongodb.org/manual/tutorial/expire-data/)

	ResponseProcessors []ResponseProcessor `json:"response_processors,omitempty"`
	CORS               CORS                `json:"CORS,omitempty"`

	// Certificates is a list of Tyk Certificate IDs. e.g. orgid+fingerprint. Use CertificateSecretNames if using cert-manager
	Certificates []string `json:"certificates,omitempty"`

	// CertificateSecretNames represents the names of the secrets that the controller should look for in the in the current
	// namespace which contain the certificates.
	CertificateSecretNames []string `json:"certificate_secret_names,omitempty"`

	// Tags are named gateway nodes which tell gateway clusters whether to load an API or not.
	// for example, to load the API in an ARA gateway, you might want to include an `edge` tag.
	Tags []string `json:"tags,omitempty"`

	// EnableContextVars extracts request context variables from the start of the middleware chain.
	// Set this to true to make them available to your transforms.
	// Context Variables are available in the url rewriter, modify headers and body transforms.
	EnableContextVars bool `json:"enable_context_vars,omitempty"`

	//ConfigData MapStringInterface `json:"config_data"`

	//TagHeaders              []string        `json:"tag_headers"`
	//GlobalRateLimit         GlobalRateLimit `json:"global_rate_limit"`

	// EnableDetailedRecording instructs Tyk store the inbound request and outbound response data in HTTP Wire format
	// as part of the Analytics data
	EnableDetailedRecording bool `json:"enable_detailed_recording,omitempty"`

	GraphQL *GraphQLConfig `json:"graphql,omitempty"`
}

// Proxy outlines the API proxying functionality.
type Proxy struct {
	// If PreserveHostHeader is set to true then the host header in the outbound request is retained to be the
	// inbound hostname of the proxy.
	PreserveHostHeader bool `json:"preserve_host_header,omitempty"`

	// ListenPath represents the path to listen on. e.g. `/api` or `/` or `/httpbin`.
	// Any requests coming into the host, on the port that Tyk is configured to run on, that match this path will
	// have the rules defined in the API Definition applied. Versioning assumes that different versions of an API
	// will live on the same URL structure. If you are using URL-based versioning (e.g. /v1/function, /v2/function)
	// then it is recommended to set up a separate non-versioned definition for each version as they are essentially
	// separate APIs.
	ListenPath string `json:"listen_path,omitempty"`

	// TargetURL defines the target URL that the request should be proxied to.
	TargetURL string `json:"target_url"`

	// DisableStripSlash disables the stripping of the slash suffix from a URL.
	// when `true` a request to http://foo.bar/baz/ will be retained.
	// when `false` a request to http://foo.bar/baz/ will be matched to http://foo.bar/baz
	DisableStripSlash bool `json:"disable_strip_slash,omitempty"`

	// StripListenPath removes the inbound listen path in the outgoing request.
	// e.g. http://acme.com/httpbin/get where `httpbin` is the listen path. The `httpbin` listen path which is used
	// to identify the API loaded in Tyk is removed, and the outbound request would be http://httpbin.org/get
	StripListenPath bool `json:"strip_listen_path,omitempty"`

	// EnableLoadBalancing enables Tyk's round-robin loadbalancer. Tyk will ignore the TargetURL field, and rely on
	// the hosts in the Targets list
	EnableLoadBalancing bool `json:"enable_load_balancing,omitempty"`

	// Targets defines a list of upstream host targets. Tyk will then round-robin load balance between these targets.
	// EnableLoadBalancing must be set to true in order to take advantage of this feature.
	Targets []string `json:"target_list,omitempty"`

	// CheckHostAgainstUptimeTests will check the hostname of the outbound request against the downtime list generated
	// by the uptime test host checker. If the host is found, then it is skipped or removed from the load balancer.
	// This is only valid if uptime tests for the api are enabled.
	CheckHostAgainstUptimeTests bool `json:"check_host_against_uptime_tests,omitempty"`

	// Transport section exposes advanced transport level configurations such as minimum TLS version.
	Transport ProxyTransport `json:"transport,omitempty"`

	// TODO: Untested. Is there a use-case for SD inside a K8s environment?
	ServiceDiscovery ServiceDiscoveryConfiguration `json:"service_discovery,omitempty"`
}

type ProxyTransport struct {
	// SSLInsecureSkipVerify controls whether it is possible to use self-signed certificates when connecting to the
	// upstream. This is applied to `TykMakeHttpRequest` & `TykMakeBatchRequest` in virtual endpoint middleware.
	SSLInsecureSkipVerify bool `json:"ssl_insecure_skip_verify,omitempty"`
	// SSLCipherSuites is an array of acceptable cipher suites. A list of allowed cipher suites can be found in the
	// Go Crypto TLS package constants documentation https://golang.org/pkg/crypto/tls/#pkg-constants
	SSLCipherSuites []string `json:"ssl_ciphers,omitempty"`
	// SSLMinVersion defines the minimum TLS version the gateway will use to establish a connection to the upstream.
	// 1.0: 769; 1.1: 770; 1.2: 771; 1.3: 772.
	// +kubebuilder:validation:Enum=769;770;771;772
	SSLMinVersion uint16 `json:"ssl_min_version,omitempty"`
	// SSLForceCommonNameCheck forces hostname validation against the certificate Common Name
	SSLForceCommonNameCheck bool `json:"ssl_force_common_name_check,omitempty"`
	// ProxyURL specifies custom forward proxy & port. e.g. `http(s)://proxy.url:1234`
	ProxyURL string `json:"proxy_url,omitempty"`
}

// CORS cors settings
type CORS struct {
	// Enable when set to true it enables the cors middleware for the api
	Enable bool `json:"enable,omitempty"`
	// AllowedOrigins is a list of origin domains to allow access from.
	AllowedOrigins []string `json:"allowed_origins,omitempty"`
	// AllowedMethods is a list of methods to allow access via.
	AllowedMethods []HttpMethod `json:"allowed_methods,omitempty"`
	// AllowedHeaders are headers that are allowed within a request.
	AllowedHeaders []string `json:"allowed_headers,omitempty"`
	// ExposedHeaders is a list of headers that are exposed back in the response.
	ExposedHeaders []string `json:"exposed_headers,omitempty"`
	// AllowCredentials if true will allow cookies
	AllowCredentials bool `json:"allow_credentials,omitempty"`
	// MaxAge is the maximum age of credentials
	MaxAge int `json:"max_age,omitempty"`
	// OptionsPassthrough allow CORS OPTIONS preflight request to be proxied
	// directly to upstream, without authentication and rest of checks. This means
	// that pre-flight requests generated by web-clients such as SwaggerUI or the
	// Tyk Portal documentation system will be able to test the API using trial
	// keys. If your service handles CORS natively, then enable this option.
	OptionsPassthrough bool `json:"options_passthrough,omitempty"`
	// Debug if true, this option produces log files for the CORS middleware
	Debug bool `json:"debug,omitempty"`
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
	Versions       map[string]VersionInfo `json:"versions,omitempty"`
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
	UseParam          bool            `json:"use_param,omitempty"`
	ParamName         string          `json:"param_name,omitempty"`
	UseCookie         bool            `json:"use_cookie,omitempty"`
	CookieName        string          `json:"cookie_name,omitempty"`
	AuthHeaderName    string          `json:"auth_header_name"`
	UseCertificate    bool            `json:"use_certificate,omitempty"`
	ValidateSignature bool            `json:"validate_signature,omitempty"`
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

	// Enabled indicates if GraphQL proxy should be enabled.
	Enabled bool `json:"enabled"`

	ExecutionMode GraphQLExecutionMode `json:"execution_mode"`

	// Schema is the GraphQL Schema exposed by the GraphQL API/Upstream/Engine.
	Schema string `json:"schema"`

	// TypeFieldConfigurations is a rule set of data source and mapping of a schema field.
	TypeFieldConfigurations []TypeFieldConfiguration `json:"type_field_configurations,omitempty"`

	// GraphQLPlayground is the Playground specific configuration.
	GraphQLPlayground GraphQLPlayground `json:"playground,omitempty"`
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
	// +kubebuilder:validation:Enum=GraphQLDataSource;HTTPJSONDataSource
	Kind string `json:"kind"`
	// Config is the DataSource specific configuration object
	// Each Planner needs to make sure to parse their Config Object correctly
	Config DataSourceConfig `json:"data_source_config,omitempty"`
}

type DataSourceConfig struct {
	URL                        string                      `json:"url"`
	Method                     HttpMethod                  `json:"method"`
	Body                       string                      `json:"body,omitempty"`
	DefaultTypeName            string                      `json:"default_type_name,omitempty"`
	Headers                    []string                    `json:"headers,omitempty"`
	StatusCodeTypeNameMappings []StatusCodeTypeNameMapping `json:"status_code_type_name_mappings,omitempty"`
}

type StatusCodeTypeNameMapping struct {
	StatusCode int    `json:"status_code"`
	TypeName   string `json:"type_name,omitempty"`
}

type MappingConfiguration struct {
	Disabled bool   `json:"disabled"`
	Path     string `json:"path"`
}

// GraphQLPlayground represents the configuration for the public playground which will be hosted alongside the api.
type GraphQLPlayground struct {
	// Enabled indicates if the playground should be enabled.
	Enabled bool `json:"enabled"`
	// Path sets the path on which the playground will be hosted if enabled.
	Path string `json:"path"`
}

// ApiDefinitionStatus defines the observed state of ApiDefinition
type ApiDefinitionStatus struct {
	ApiID string `json:"api_id"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Domain",type=string,JSONPath=`.spec.domain`
// +kubebuilder:printcolumn:name="ListenPath",type=string,JSONPath=`.spec.proxy.listen_path`
// +kubebuilder:printcolumn:name="Proxy.TargetURL",type=string,JSONPath=`.spec.proxy.target_url`
// +kubebuilder:printcolumn:name="Enabled",type=boolean,JSONPath=`.spec.active`
// +kubebuilder:resource:shortName=tykapis
// ApiDefinition is the Schema for the apidefinitions API
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
