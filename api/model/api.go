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

// +kubebuilder:object:generate=true
package model

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
)

// MapStringInterfaceType represents a generic struct used as a map[string]interface{}. Since an arbitrary
// JSON fields defined as map[string]interface{} is not feasible to use as a Kubernetes CRD, unstructured.Unstructured
// type is used.
type MapStringInterfaceType struct {
	// +kubebuilder:pruning:PreserveUnknownFields
	unstructured.Unstructured `json:",inline"`
}

// ApiDefinitionSpec defines the desired state of ApiDefinition
type (
	AuthProviderCode     string
	SessionProviderCode  string
	StorageEngineCode    string
	TykEvent             string
	TykEventHandlerName  string
	EndpointMethodAction string
	TemplateMode         string
	MiddlewareDriver     string
	IdExtractorSource    string
	IdExtractorType      string
	AuthTypeEnum         string
	RoutingTriggerOnType string
)

// HttpMethod represents HTTP request method
// +kubebuilder:validation:Enum=GET;POST;PUT;PATCH;DELETE;OPTIONS;HEAD;CONNECT;TRACE
type HttpMethod string

// GraphQLExecutionMode is the mode to define how an api behaves.
// +kubebuilder:validation:Enum="";proxyOnly;executionEngine;supergraph;subgraph
type GraphQLExecutionMode string

const (
	SuperGraphExecutionMode GraphQLExecutionMode = "supergraph"
	SubGraphExecutionMode   GraphQLExecutionMode = "subgraph"
)

// APIProtocol is the network transport protocol supported by the gateway
// +kubebuilder:validation:Enum="";h2c;tcp;tls;http;https;
type APIProtocol string

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
	// FromDashboard  bool             `json:"from_dashboard"`
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
	ThresholdPercent Percent `json:"threshold_percent"`
	// Samples defines the number of requests to base the ThresholdPercent on
	Samples int64 `json:"samples"`
	// ReturnToServiceAfter represents the time in seconds to return back to the service
	ReturnToServiceAfter int   `json:"return_to_service_after"`
	DisableHalfOpenState *bool `json:"disable_half_open_state,omitempty"`
}

type StringRegexMap struct {
	MatchPattern string `json:"match_rx"`
	Reverse      *bool  `json:"reverse,omitempty"`
}

type RoutingTriggerOptions struct {
	HeaderMatches         map[string]StringRegexMap `json:"header_matches,omitempty"`
	QueryValMatches       map[string]StringRegexMap `json:"query_val_matches,omitempty"`
	PathPartMatches       map[string]StringRegexMap `json:"path_part_matches,omitempty"`
	SessionMetaMatches    map[string]StringRegexMap `json:"session_meta_matches,omitempty"`
	RequestContextMatches map[string]StringRegexMap `json:"request_context_matches,omitempty"`
	PayloadMatches        *StringRegexMap           `json:"payload_matches,omitempty"`
}

type RoutingTrigger struct {
	On                RoutingTriggerOnType  `json:"on"`
	Options           RoutingTriggerOptions `json:"options"`
	RewriteTo         *string               `json:"rewrite_to,omitempty"`
	RewriteToInternal *RewriteToInternal    `json:"rewrite_to_internal,omitempty"`
}

func (r *RoutingTrigger) collectLoopingTarget(fn func(Target)) {
	if r.RewriteToInternal != nil {
		x := r.RewriteToInternal.Target

		rewriteToInternal := r.RewriteToInternal.String()
		r.RewriteTo = &rewriteToInternal
		r.RewriteToInternal = nil

		fn(x)
	}
}

type URLRewriteMeta struct {
	// Path represents the endpoint listen path
	Path   string     `json:"path"`
	Method HttpMethod `json:"method"`
	// MatchPattern is a regular expression pattern to match the path
	MatchPattern string `json:"match_pattern"`
	// RewriteTo is the target path on the upstream, or target URL we wish to rewrite to
	RewriteTo *string `json:"rewrite_to,omitempty"`
	// RewriteToInternal serves as rewrite_to but used when rewriting to target
	// internal api's
	// When rewrite_to and rewrite_to_internal are both provided then
	// rewrite_to will take rewrite_to_internal
	RewriteToInternal *RewriteToInternal `json:"rewrite_to_internal,omitempty"`
	Triggers          []RoutingTrigger   `json:"triggers,omitempty"`
}

func (u *URLRewriteMeta) collectLoopingTarget(fn func(Target)) {
	if u.RewriteToInternal != nil {
		x := u.RewriteToInternal.Target

		if u.RewriteTo == nil {
			u.RewriteTo = new(string)
		}

		*u.RewriteTo = u.RewriteToInternal.String()
		u.RewriteToInternal = nil

		fn(x)
	}

	for i := 0; i < len(u.Triggers); i++ {
		u.Triggers[i].collectLoopingTarget(fn)
	}
}

// TargetInternal defines options that constructs a url that refers to an api that
// is loaded into the gateway.

type TargetInternal struct {
	// API a namespaced/name to the api definition resource that you are
	// targetting
	Target Target `json:"target,omitempty"`
	// Path path on target , this does not include query parameters.
	//	example /myendpoint
	Path *string `json:"path,omitempty"`

	// Query url query string to add to target
	//	example check_limits=true
	Query *string `json:"query,omitempty"`
}

func (i TargetInternal) String() string {
	host := i.Target.String()
	host = base64.RawURLEncoding.EncodeToString([]byte(host))

	path := ""
	query := ""

	if i.Path != nil {
		path = *i.Path
	}

	if i.Query != nil {
		query = *i.Query
	}

	u := url.URL{
		Scheme:   "tyk",
		Host:     host,
		RawPath:  path,
		RawQuery: query,
	}

	return u.String()
}

// RewriteToInternal defines options that constructs a url that refers to an api that
// is loaded into the gateway.
type RewriteToInternal struct {
	// API a namespaced/name to the api definition resource that you are
	// targetting
	Target Target `json:"target,omitempty"`
	// Path path on target , this does not include query parameters.
	//	example /myendpoint
	Path *string `json:"path,omitempty"`

	// Query url query string to add to target
	//	example check_limits=true
	Query *string `json:"query,omitempty"`
}

func (i RewriteToInternal) String() string {
	host := i.Target.String()
	host = base64.RawURLEncoding.EncodeToString([]byte(host))

	path := ""
	query := ""

	if i.Path != nil {
		path = *i.Path
	}

	if i.Query != nil {
		query = *i.Query
	}

	u := url.URL{
		Scheme:   "tyk",
		Host:     host,
		Path:     path,
		RawQuery: query,
	}

	return u.String()
}

type Target struct {
	// k8s resource name
	Name string `json:"name"`
	// The k8s namespace of the resource being targetted. When omitted this will be
	// set to the namespace of the object that is being reconciled.
	Namespace *string `json:"namespace,omitempty"`
}

func (t *Target) Parse(v string) {
	if strings.Contains(v, "/") {
		if p := strings.Split(v, "/"); len(p) == 2 {
			if t.Namespace == nil {
				t.Namespace = new(string)
			}

			*t.Namespace = p[0]
			t.Name = p[1]
		}
	}
}

func (t Target) String() string {
	return t.NS("").String()
}

// Equal returns true if t and o are equal
func (t Target) Equal(o Target) bool {
	namespaceMatches := false

	if t.Namespace == nil {
		if o.Namespace == nil {
			namespaceMatches = true
		} else if *o.Namespace == "" {
			namespaceMatches = true
		}
	} else {
		if *t.Namespace == "" && o.Namespace == nil {
			namespaceMatches = true
		} else if o.Namespace != nil {
			namespaceMatches = *t.Namespace == *o.Namespace
		}
	}

	return namespaceMatches && t.Name == o.Name
}

func (t Target) NS(defaultNS string) types.NamespacedName {
	if t.Namespace != nil && *t.Namespace != "" {
		defaultNS = *t.Namespace
	}

	return types.NamespacedName{Namespace: defaultNS, Name: t.Name}
}

func (t Target) NamespaceMatches(ns string) bool {
	if t.Namespace == nil && ns == "" {
		return true
	}

	if t.Namespace != nil && *t.Namespace == ns {
		return true
	}

	return false
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
	Disabled *bool `json:"disabled,omitempty"`
	// Allows override of default 422 Unprocessable Entity response code for validation errors.
	ErrorResponseCode int        `json:"error_response_code"`
	Path              string     `json:"path"`
	Method            HttpMethod `json:"method"`
	// Schema represents schema field that verifies user requests against a specified
	// JSON schema and check that the data sent to your API by a consumer is in the right format.
	Schema *MapStringInterfaceType `json:"schema"`
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
	ValidateJSON            []ValidatePathMeta    `json:"validate_json,omitempty"`
	Internal                []InternalMeta        `json:"internal,omitempty"`
}

func (e *ExtendedPathsSet) collectLoopingTarget(fn func(Target)) {
	if e == nil {
		return
	}

	for i := 0; i < len(e.URLRewrite); i++ {
		e.URLRewrite[i].collectLoopingTarget(fn)
	}
}

type VersionInfo struct {
	Name                        string            `json:"name"`
	Expires                     *string           `json:"expires,omitempty"`
	Paths                       *VersionInfoPaths `json:"paths,omitempty"`
	UseExtendedPaths            *bool             `json:"use_extended_paths,omitempty"`
	ExtendedPaths               *ExtendedPathsSet `json:"extended_paths,omitempty"`
	GlobalHeaders               map[string]string `json:"global_headers,omitempty"`
	GlobalHeadersRemove         []string          `json:"global_headers_remove,omitempty"`
	GlobalResponseHeaders       map[string]string `json:"global_response_headers,omitempty"`
	GlobalResponseHeadersRemove []string          `json:"global_response_headers_remove,omitempty"`
	IgnoreEndpointCase          *bool             `json:"ignore_endpoint_case,omitempty"`
	GlobalSizeLimit             int64             `json:"global_size_limit,omitempty"`
	OverrideTarget              *string           `json:"override_target,omitempty"`
}

func (v *VersionInfo) collectLoopingTarget(fn func(Target)) {
	v.ExtendedPaths.collectLoopingTarget(fn)
}

type VersionInfoPaths struct {
	Ignored   []string `json:"ignored,omitempty"`
	WhiteList []string `json:"white_list,omitempty"`
	BlackList []string `json:"black_list,omitempty"`
}

type AuthProviderMeta struct {
	Name          AuthProviderCode  `json:"name"`
	StorageEngine StorageEngineCode `json:"storage_engine"`
	// Meta          map[string]interface{} `json:"meta"`
}

type SessionProviderMeta struct {
	Name          SessionProviderCode `json:"name"`
	StorageEngine StorageEngineCode   `json:"storage_engine"`
	// Meta          map[string]interface{} `json:"meta"`
}

type EventHandlerTriggerConfig struct {
	Handler TykEventHandlerName `json:"handler_name"`
	// HandlerMeta map[string]interface{} `json:"handler_meta"`
}

type EventHandlerMetaConfig struct {
	Events map[TykEvent][]EventHandlerTriggerConfig `json:"events"`
}

type MiddlewareDefinition struct {
	Name           string `json:"name"`
	Path           string `json:"path"`
	RequireSession *bool  `json:"require_session,omitempty"`
	RawBodyOnly    *bool  `json:"raw_body_only,omitempty"`
}

type IdExtractorConfig struct {
	HeaderName      *string `json:"header_name,omitempty"`
	FormParamName   *string `json:"param_name,omitempty"`
	RegexExpression *string `json:"regex_expression,omitempty"`
	RegexMatchIndex int     `json:"regex_match_index,omitempty"`
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
	EnableCache *bool `json:"enable_cache,omitempty"`
	// CacheTimeout is the TTL for a cached object in seconds
	CacheTimeout int64 `json:"cache_timeout"`
	// CacheAllSafeRequests caches responses to (GET, HEAD, OPTIONS) requests
	// overrides per-path cache settings in versions, applies across versions
	CacheAllSafeRequests *bool `json:"cache_all_safe_requests,omitempty"`
	// CacheOnlyResponseCodes is an array of response codes which are safe to cache. e.g. 404
	CacheOnlyResponseCodes []int `json:"cache_response_codes,omitempty"`
	// EnableUpstreamCacheControl instructs Tyk Cache to respect upstream cache control headers
	EnableUpstreamCacheControl *bool `json:"enable_upstream_cache_control,omitempty"`
	// CacheControlTTLHeader is the response header which tells Tyk how long it is safe to cache the response for
	CacheControlTTLHeader *string `json:"cache_control_ttl_header,omitempty"`
	// CacheByHeaders allows header values to be used as part of the cache key
	CacheByHeaders []string `json:"cache_by_headers,omitempty"`
}

type ResponseProcessor struct {
	Name string `json:"name"`
	// Options interface{} `json:"options"`
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

// APIDefinitionSpec represents the configuration for a single proxied API and it's versions.
type APIDefinitionSpec struct {
	// For server use only, do not use
	ID *string `json:"id,omitempty" hash:"ignore"`

	// Only set this field if you are referring
	// to an existing API def.
	// The Operator will use this APIID to link the CR with the API in Tyk
	// Note: The values in the CR will become the new source of truth, overriding the existing API Definition
	APIID *string `json:"api_id,omitempty"`

	Name string `json:"name"`

	// OrgID is overwritten - no point setting this
	OrgID *string `json:"org_id,omitempty"`

	// Active specifies if the api is enabled or not
	Active *bool `json:"active,omitempty"`

	// Proxy
	Proxy Proxy `json:"proxy"`

	// +optional
	ListenPort int `json:"listen_port"`

	Protocol APIProtocol `json:"protocol"`

	EnableProxyProtocol *bool `json:"enable_proxy_protocol,omitempty"`

	// Domain represents a custom host header that the gateway will listen on for this API
	Domain *string `json:"domain,omitempty"`

	// DoNotTrack disables endpoint tracking for this API
	DoNotTrack *bool `json:"do_not_track,omitempty"`

	// UseKeylessAccess will switch off all key checking. Some analytics will still be recorded, but rate-limiting,
	// quotas and security policies will not be possible (there is no session to attach requests to).
	UseKeylessAccess *bool `json:"use_keyless,omitempty"`

	// UseOAuth2 enables oauth2 authorization
	UseOauth2 *bool `json:"use_oauth2,omitempty"`

	//+optional
	Oauth2Meta *OAuth2Meta `json:"oauth_meta,omitempty"`

	// UseOpenID           bool          `json:"use_openid"`
	// OpenIDOptions       OpenIDOptions `json:"openid_options"`

	// StripAuthData ensures that any security tokens used for accessing APIs are stripped and not leaked to the upstream
	StripAuthData *bool `json:"strip_auth_data,omitempty"`

	Auth AuthConfig `json:"auth,omitempty"`

	// +optional
	AuthConfigs map[string]AuthConfig `json:"auth_configs,omitempty"`

	// UseStandardAuth enables simple bearer token authentication
	UseStandardAuth *bool `json:"use_standard_auth,omitempty"`

	// UseBasicAuth enables basic authentication
	UseBasicAuth *bool `json:"use_basic_auth,omitempty"`
	// BasicAuth                  BasicAuthMeta         `json:"basic_auth"`

	// UseMutualTLSAuth enables mututal TLS authentication
	UseMutualTLSAuth      *bool    `json:"use_mutual_tls_auth,omitempty"`
	ClientCertificates    []string `json:"client_certificates,omitempty"`
	ClientCertificateRefs []string `json:"client_certificate_refs,omitempty"`

	// PinnedPublicKeys allows you to whitelist public keys used to generate certificates, so you will be protected in
	// case an upstream certificate is compromised. Please use PinnedPublicKeysRefs if using cert-manager.
	PinnedPublicKeys map[string]string `json:"pinned_public_keys,omitempty"`

	// PinnedPublicKeysRefs allows you to specify public keys using k8s secret.
	// It takes domain name as a key and secret name as a value.
	PinnedPublicKeysRefs map[string]string `json:"pinned_public_keys_refs,omitempty"`

	// UpstreamCertificates is a map of domains and certificate IDs that is used by the Tyk
	// Gateway to provide mTLS support for upstreams
	UpstreamCertificates map[string]string `json:"upstream_certificates,omitempty"`

	// UpstreamCertificateRefs is a map of domains and secret names that is used internally
	// to obtain certificates from secrets in order to establish mTLS support for upstreams
	UpstreamCertificateRefs map[string]string `json:"upstream_certificate_refs,omitempty"`

	// EnableJWT set JWT as the access method for this API.
	EnableJWT *bool `json:"enable_jwt,omitempty"`

	// Enable Go Plugin Auth. Needs to be combined with "use_keyless:false"
	UseGoPluginAuth *bool `json:"use_go_plugin_auth,omitempty"`

	EnableCoProcessAuth *bool `json:"enable_coprocess_auth,omitempty"`

	// JWTSigningMethod algorithm used to sign jwt token
	// +kubebuilder:validation:Enum="";rsa;hmac;ecdsa
	JWTSigningMethod *string `json:"jwt_signing_method,omitempty"`

	// JWTSource Must either be a base64 encoded valid RSA/HMAC key or a url to a
	// resource serving JWK, this key will then be used to validate inbound JWT and
	// throttle them according to the centralised JWT options and fields set in the
	// configuration.
	JWTSource *string `json:"jwt_source,omitempty"`

	// JWTIdentityBaseField Identifies the user or identity to be used in the
	// Claims of the JWT. This will fallback to sub if not found. This field forms
	// the basis of a new “virtual” token that gets used after validation. It means
	// policy attributes are carried forward through Tyk for attribution purposes.
	JWTIdentityBaseField *string `json:"jwt_identity_base_field,omitempty"`

	// JWTClientIDBaseField is the name of the field on JWT claim to use for client
	// id. This field is mutually exclusive to jwt_identity_base_field, meaning you
	// can only set/use one and jwt_identity_base_field takes precedence when both
	// are set.
	JWTClientIDBaseField *string `json:"jwt_client_base_field,omitempty"`

	// JWTPolicyFieldName The policy ID to apply to the virtual token generated for a JWT
	JWTPolicyFieldName *string `json:"jwt_policy_field_name,omitempty"`

	// JWTDefaultPolicies is a list of policies that will be used when base policy
	// can't be extracted from the JWT token. When this list is provided the first
	// element will be used as the base policy while the rest of elements will be applied.
	JWTDefaultPolicies []string `json:"jwt_default_policies,omitempty"`

	// JWTIssuedAtValidationSkew adds validation for  issued at JWT claim.
	// Given
	//	now = current unix time
	//	skew = jwt_issued_at_validation_skew
	//	iat = the issued at jwt claim
	// If iat > (now + skew) then validation will fail with "token used before issued"
	JWTIssuedAtValidationSkew uint64 `json:"jwt_issued_at_validation_skew,omitempty"`

	// JWTExpiresAtValidationSkew adds validation for  expired at JWT claim.
	// Given
	//	now = current unix time
	//	skew = jwt_expires_at_validation_skew
	//	exp = expired at
	// If exp > (now - skew) then validation will fail with "token has expired"
	JWTExpiresAtValidationSkew uint64 `json:"jwt_expires_at_validation_skew,omitempty"`

	// JWTNotBeforeValidationSkew adds validation for  not before  JWT claim.
	// Given
	//	now = current unix time
	//	skew = jwt_not_before_validation_skew
	//	nbf = the not before  jwt claim
	// If nbf > (now + skew) then validation will fail with "token is not valid yet"
	JWTNotBeforeValidationSkew uint64 `json:"jwt_not_before_validation_skew,omitempty"`

	// JWTSkipKid when true we ingore using kid as the identity for a JWT token and
	// instead use jwt_identity_base_field if it was set or fallback to sub JWT
	// claim.
	JWTSkipKid *bool `json:"jwt_skip_kid,omitempty"`

	// JWTScopeToPolicyMapping this is a mapping of scope value to policy id. If
	// this is set then a scope value found in this map will make the mappend
	// policy to be applied.
	JWTScopeToPolicyMapping map[string]string `json:"jwt_scope_to_policy_mapping,omitempty"`

	// JWTScopeClaimName overides the key used for scope values in the JWT claims.
	// By default the value is "scope"
	JWTScopeClaimName *string `json:"jwt_scope_claim_name,omitempty"`

	// NotificationsDetails       NotificationsManager  `json:"notifications"`
	// EnableSignatureChecking    bool                  `json:"enable_signature_checking"`
	// HmacAllowedClockSkew       json.Number           `json:"hmac_allowed_clock_skew"` // TODO: convert to float64
	// HmacAllowedAlgorithms      []string              `json:"hmac_allowed_algorithms"`
	// RequestSigning             RequestSigningMeta    `json:"request_signing"`

	// BaseIdentityProvidedBy sets Base Identity Provider for situation when multiple authentication mechanisms are used
	// +kubebuilder:validation:Enum=auth_token;hmac_key;basic_auth_user;jwt_claim;oidc_user;oauth_key
	BaseIdentityProvidedBy AuthTypeEnum `json:"base_identity_provided_by,omitempty"`

	VersionDefinition VersionDefinition `json:"definition,omitempty"`

	VersionData VersionData `json:"version_data,omitempty"`

	// UptimeTests                UptimeTests           `json:"uptime_tests"`

	// DisableRateLimit allows you to disable rate limits in a given API Definition.
	DisableRateLimit *bool `json:"disable_rate_limit,omitempty"`

	// DisableQuota allows you to disable quota middleware in a given API Definition.
	DisableQuota *bool `json:"disable_quota,omitempty"`

	// GlobalRateLimit is an API Level Global Rate Limit, which assesses all traffic coming into the API from all
	// sources and ensures that the overall rate limit is not exceeded.
	GlobalRateLimit GlobalRateLimit `json:"global_rate_limit,omitempty"`

	CustomMiddleware       MiddlewareSection `json:"custom_middleware,omitempty"`
	CustomMiddlewareBundle *string           `json:"custom_middleware_bundle,omitempty"`

	CacheOptions CacheOptions `json:"cache_options,omitempty"`

	// SessionLifetime this is duration in seconds before the session key expires
	// in redis.
	//
	// Example:
	// If you want the session keys to be alive only 24 hours you can set this
	// value to 86400 that we can break down to
	//	60 * 60 * 24 = Total seconds in a day
	//+optional
	SessionLifetime int64 `json:"session_lifetime,omitempty"`

	// Internal tells Tyk Gateway that this is a virtual API. It can only be routed to from other APIs.
	Internal *bool `json:"internal,omitempty"`
	//AuthProvider           AuthProviderMeta    `json:"auth_provider"`
	//SessionProvider        SessionProviderMeta `json:"session_provider"`
	////EventHandlers             EventHandlerMetaConfig `json:"event_handlers"`
	//EnableBatchRequestSupport bool `json:"enable_batch_request_support"`

	// EnableIPWhiteListing activates the ip whitelisting middleware.
	EnableIPWhiteListing *bool `json:"enable_ip_whitelisting,omitempty"`

	// AllowedIPs is a list of IP address that are whitelisted.When this is
	// provided all IP address that is not on this list will be blocked and a 403 http
	// status will be returned. The IP address can be IPv4 or IPv6.IP in
	// CIDR notation is also supported.
	AllowedIPs []string `json:"allowed_ips,omitempty"`

	// EnableIPBlacklisting activates the ip blacklisting middleware.
	EnableIPBlacklisting *bool `json:"enable_ip_blacklisting,omitempty"`

	// BlacklistedIPs is a list of IP address that will be blacklisted.This means if
	// origin IP matches any IP in this list a 403 http status code will be
	// returned. The IP address can be IPv4 or IPv6. IP in CIDR notation is also
	// supported.
	BlacklistedIPs []string `json:"blacklisted_ips,omitempty"`
	// DontSetQuotasOnCreate bool                `json:"dont_set_quota_on_create"`

	// must have an expireAt TTL index set (http://docs.mongodb.org/manual/tutorial/expire-data/)
	// ExpireAnalyticsAfter  int64               `json:"expire_analytics_after"`

	ResponseProcessors []ResponseProcessor `json:"response_processors,omitempty"`
	CORS               CORS                `json:"CORS,omitempty"`

	// Certificates is a list of Tyk Certificate IDs. e.g. orgid+fingerprint. Use CertificateSecretNames if using cert-manager
	Certificates []string `json:"certificates,omitempty"`

	// CertificateSecretNames represents the names of the secrets that the controller should look for in the current
	// namespace which contain the certificates.
	CertificateSecretNames []string `json:"certificate_secret_names,omitempty"`

	// Tags are named gateway nodes which tell gateway clusters whether to load an API or not.
	// for example, to load the API in an ARA gateway, you might want to include an `edge` tag.
	Tags []string `json:"tags,omitempty"`

	// EnableContextVars extracts request context variables from the start of the middleware chain.
	// Set this to true to make them available to your transforms.
	// Context Variables are available in the url rewriter, modify headers and body transforms.
	EnableContextVars *bool `json:"enable_context_vars,omitempty"`

	// ConfigData can be used to pass custom attributes (a JSON object) into your middleware, such
	// as a virtual endpoint or header transform.
	// +optional
	// +nullable
	ConfigData *MapStringInterfaceType `json:"config_data"`

	TagHeaders []string `json:"tag_headers,omitempty"`

	// EnableDetailedRecording instructs Tyk store the inbound request and outbound response data in HTTP Wire format
	// as part of the Analytics data
	EnableDetailedRecording *bool `json:"enable_detailed_recording,omitempty"`

	GraphQL *GraphQLConfig `json:"graphql,omitempty"`

	// +optional
	// +nullable
	DetailedTracing *bool `json:"detailed_tracing,omitempty"`
}

func (a *APIDefinitionSpec) CollectLoopingTarget() (targets []Target) {
	fn := func(t Target) {
		targets = append(targets, t)
	}

	a.Proxy.collectLoopingTarget(fn)
	a.VersionData.collectLoopingTarget(fn)

	return
}

// Proxy outlines the API proxying functionality.
type Proxy struct {
	// If PreserveHostHeader is set to true then the host header in the outbound request is retained to be the
	// inbound hostname of the proxy.
	PreserveHostHeader *bool `json:"preserve_host_header,omitempty"`

	// ListenPath represents the path to listen on. e.g. `/api` or `/` or `/httpbin`.
	// Any requests coming into the host, on the port that Tyk is configured to run on, that match this path will
	// have the rules defined in the API Definition applied. Versioning assumes that different versions of an API
	// will live on the same URL structure. If you are using URL-based versioning (e.g. /v1/function, /v2/function)
	// then it is recommended to set up a separate non-versioned definition for each version as they are essentially
	// separate APIs.
	ListenPath *string `json:"listen_path,omitempty"`

	// TargetURL defines the target URL that the request should be proxied to.
	TargetURL      string          `json:"target_url"`
	TargetInternal *TargetInternal `json:"target_internal,omitempty"`

	// DisableStripSlash disables the stripping of the slash suffix from a URL.
	// when `true` a request to http://foo.bar/baz/ will be retained.
	// when `false` a request to http://foo.bar/baz/ will be matched to http://foo.bar/baz
	DisableStripSlash *bool `json:"disable_strip_slash,omitempty"`

	// StripListenPath removes the inbound listen path in the outgoing request.
	// e.g. http://acme.com/httpbin/get where `httpbin` is the listen path. The `httpbin` listen path which is used
	// to identify the API loaded in Tyk is removed, and the outbound request would be http://httpbin.org/get
	StripListenPath *bool `json:"strip_listen_path,omitempty"`

	// EnableLoadBalancing enables Tyk's round-robin loadbalancer. Tyk will ignore the TargetURL field, and rely on
	// the hosts in the Targets list
	EnableLoadBalancing *bool `json:"enable_load_balancing,omitempty"`

	// Targets defines a list of upstream host targets. Tyk will then round-robin load balance between these targets.
	// EnableLoadBalancing must be set to true in order to take advantage of this feature.
	Targets []string `json:"target_list,omitempty"`

	// CheckHostAgainstUptimeTests will check the hostname of the outbound request against the downtime list generated
	// by the uptime test host checker. If the host is found, then it is skipped or removed from the load balancer.
	// This is only valid if uptime tests for the api are enabled.
	CheckHostAgainstUptimeTests *bool `json:"check_host_against_uptime_tests,omitempty"`

	// Transport section exposes advanced transport level configurations such as minimum TLS version.
	Transport ProxyTransport `json:"transport,omitempty"`

	// TODO: Untested. Is there a use-case for SD inside a K8s environment?
	ServiceDiscovery ServiceDiscoveryConfiguration `json:"service_discovery,omitempty"`
}

func (p *Proxy) collectLoopingTarget(fn func(Target)) {
	if p.TargetInternal != nil {
		x := p.TargetInternal.Target
		p.TargetURL = p.TargetInternal.String()
		p.TargetInternal = nil

		fn(x)
	}
}

type ProxyTransport struct {
	// SSLInsecureSkipVerify controls whether it is possible to use self-signed certificates when connecting to the
	// upstream. This is applied to `TykMakeHttpRequest` & `TykMakeBatchRequest` in virtual endpoint middleware.
	SSLInsecureSkipVerify *bool `json:"ssl_insecure_skip_verify,omitempty"`

	// SSLCipherSuites is an array of acceptable cipher suites. A list of allowed cipher suites can be found in the
	// Go Crypto TLS package constants documentation https://golang.org/pkg/crypto/tls/#pkg-constants
	SSLCipherSuites []string `json:"ssl_ciphers,omitempty"`

	// SSLMinVersion defines the minimum TLS version the gateway will use to establish a connection to the upstream.
	// 1.0: 769; 1.1: 770; 1.2: 771; 1.3: 772.
	// +kubebuilder:validation:Enum=769;770;771;772
	SSLMinVersion uint16 `json:"ssl_min_version,omitempty"`

	// SSLForceCommonNameCheck forces hostname validation against the certificate Common Name
	SSLForceCommonNameCheck *bool `json:"ssl_force_common_name_check,omitempty"`

	// ProxyURL specifies custom forward proxy & port. e.g. `http(s)://proxy.url:1234`
	ProxyURL *string `json:"proxy_url,omitempty"`
}

// CORS cors settings

type CORS struct {
	// Enable when set to true it enables the cors middleware for the api
	Enable *bool `json:"enable,omitempty"`

	// AllowedOrigins is a list of origin domains to allow access from.
	AllowedOrigins []string `json:"allowed_origins,omitempty"`

	// AllowedMethods is a list of methods to allow access via.
	AllowedMethods []HttpMethod `json:"allowed_methods,omitempty"`

	// AllowedHeaders are headers that are allowed within a request.
	AllowedHeaders []string `json:"allowed_headers,omitempty"`

	// ExposedHeaders is a list of headers that are exposed back in the response.
	ExposedHeaders []string `json:"exposed_headers,omitempty"`

	// AllowCredentials if true will allow cookies
	AllowCredentials *bool `json:"allow_credentials,omitempty"`

	// MaxAge is the maximum age of credentials
	MaxAge int `json:"max_age,omitempty"`

	// OptionsPassthrough allow CORS OPTIONS preflight request to be proxied
	// directly to upstream, without authentication and rest of checks. This means
	// that pre-flight requests generated by web-clients such as SwaggerUI or the
	// Tyk Portal documentation system will be able to test the API using trial
	// keys. If your service handles CORS natively, then enable this option.
	OptionsPassthrough *bool `json:"options_passthrough,omitempty"`

	// Debug if true, this option produces log files for the CORS middleware
	Debug *bool `json:"debug,omitempty"`
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

func (v *VersionData) collectLoopingTarget(fn func(Target)) {
	for k, value := range v.Versions {
		value.collectLoopingTarget(fn)
		v.Versions[k] = value
	}
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
	// AllowedAccessTypes are an array of allowable access types.
	AllowedAccessTypes []AccessTypeEnum `json:"allowed_access_types"` // osin.AccessRequestType

	// AllowedAuthorizeTypes is an array of allowable `response_type` parameters `token` or authorization code `code`.
	// Choose token for client_credentials or implicit grant types.
	AllowedAuthorizeTypes []AuthorizeTypeEnum `json:"allowed_authorize_types"` // osin.AuthorizeRequestType

	// Login form to handle user login.
	AuthLoginRedirect *string `json:"auth_login_redirect,omitempty"`
}

// +kubebuilder:validation:Enum=authorization_code;refresh_token;password;client_credentials
type AccessTypeEnum string

// +kubebuilder:validation:Enum=code;token
type AuthorizeTypeEnum string

type AuthConfig struct {
	UseParam          *bool           `json:"use_param,omitempty"`
	ParamName         *string         `json:"param_name,omitempty"`
	UseCookie         *bool           `json:"use_cookie,omitempty"`
	CookieName        *string         `json:"cookie_name,omitempty"`
	AuthHeaderName    string          `json:"auth_header_name"`
	UseCertificate    *bool           `json:"use_certificate,omitempty"`
	ValidateSignature *bool           `json:"validate_signature,omitempty"`
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
	// Rate represents the number of requests allowed within a specified time window (Per)
	Rate int `json:"rate"`

	// Per represents a time window in seconds
	Per int `json:"per"`
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

type GraphQLFieldConfig struct {
	TypeName              string   `json:"type_name"`
	FieldName             string   `json:"field_name"`
	DisableDefaultMapping bool     `json:"disable_default_mapping"`
	Path                  []string `json:"path"`
}

// +kubebuilder:validation:Enum=REST;GraphQL;Kafka
type GraphQLEngineDataSourceKind string

type GraphQLEngineDataSource struct {
	Kind       GraphQLEngineDataSourceKind `json:"kind"`
	Name       string                      `json:"name"`
	Internal   bool                        `json:"internal"`
	RootFields []GraphQLTypeFields         `json:"root_fields"`
	Config     MapStringInterfaceType      `json:"config"`
}

type GraphQLEngineGlobalHeader struct {
	// Key is the name of the request header
	Key string `json:"key"`
	// Value holds the value of the request header
	Value string `json:"value"`
}

type GraphQLTypeFields struct {
	Type   string   `json:"type"`
	Fields []string `json:"fields"`
}

type GraphQLEngineConfig struct {
	// +nullable
	FieldConfigs []GraphQLFieldConfig `json:"field_configs"`
	// +nullable
	DataSources []GraphQLEngineDataSource `json:"data_sources"`
	// GlobalHeaders for managing headers for UDG and all associated data sources
	// +nullable
	GlobalHeaders []GraphQLEngineGlobalHeader `json:"global_headers"`
}

type GraphQLSubgraphConfig struct {
	SDL string `json:"sdl"`
}

type GraphQLSubgraphEntity struct {
	// UUID v4 string (!not the same as _id of APIDefinition)
	APIID string `json:"api_id"`
	Name  string `json:"name"`
	// The internal URL of the subgraph
	URL string `json:"url"`
	// the schema definition language of the subgraph
	SDL string `json:"sdl"`

	// +nullable
	Headers map[string]string `json:"headers"`
}

type GraphQLSupergraphConfig struct {
	// UpdatedAt contains the date and time of the last update of a supergraph API.
	UpdatedAt            *metav1.Time            `json:"updated_at,omitempty"`
	Subgraphs            []GraphQLSubgraphEntity `json:"subgraphs,omitempty"`
	MergedSDL            *string                 `json:"merged_sdl,omitempty"`
	GlobalHeaders        map[string]string       `json:"global_headers,omitempty"`
	DisableQueryBatching *bool                   `json:"disable_query_batching,omitempty"`
}

// +kubebuilder:validation:Enum="1";"2"
type GraphQLConfigVersion string

// GraphQLConfig is the root config object for a GraphQL API.
type GraphQLConfig struct {
	// Enabled indicates if GraphQL proxy should be enabled.
	Enabled bool `json:"enabled"`

	ExecutionMode GraphQLExecutionMode `json:"execution_mode"`

	// Version defines the version of the GraphQL config and engine to be used.
	Version GraphQLConfigVersion `json:"version,omitempty"`

	// Schema is the GraphQL Schema exposed by the GraphQL API/Upstream/Engine.
	Schema *string `json:"schema,omitempty"`

	// LastSchemaUpdate contains the date and time of the last triggered schema update to the upstream.
	LastSchemaUpdate *metav1.Time `json:"last_schema_update,omitempty"`

	// TypeFieldConfigurations is a rule set of data source and mapping of a schema field.
	TypeFieldConfigurations []TypeFieldConfiguration `json:"type_field_configurations,omitempty"`

	// GraphQLPlayground is the Playground specific configuration.
	GraphQLPlayground GraphQLPlayground `json:"playground,omitempty"`

	// Engine holds the configuration for engine v2 and upwards.
	Engine GraphQLEngineConfig `json:"engine,omitempty"`

	// Proxy holds the configuration for a proxy only api.
	Proxy GraphQLProxyConfig `json:"proxy,omitempty"`

	// Subgraph holds the configuration for a GraphQL federation subgraph.
	Subgraph GraphQLSubgraphConfig `json:"subgraph,omitempty"`

	GraphRef *string `json:"graph_ref,omitempty"`

	// Supergraph holds the configuration for a GraphQL federation supergraph.
	Supergraph GraphQLSupergraphConfig `json:"supergraph,omitempty"`

	// Introspection holds the configuration for GraphQL Introspection
	Introspection GraphQLIntrospectionConfig `json:"introspection,omitempty"`
}

type GraphQLIntrospectionConfig struct {
	Disabled bool `json:"disabled,omitempty"`
}

type GraphQLProxyConfig struct {
	// +nullable
	AuthHeaders map[string]string `json:"auth_headers"`
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
	Body                       *string                     `json:"body,omitempty"`
	DefaultTypeName            *string                     `json:"default_type_name,omitempty"`
	Headers                    []string                    `json:"headers,omitempty"`
	StatusCodeTypeNameMappings []StatusCodeTypeNameMapping `json:"status_code_type_name_mappings,omitempty"`
}

type StatusCodeTypeNameMapping struct {
	StatusCode int     `json:"status_code"`
	TypeName   *string `json:"type_name,omitempty"`
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

// APIDefinitionSpecList is a list of api definitions
type APIDefinitionSpecList struct {
	Apis []*APIDefinitionSpec `json:"apis"`
}

// ListAPIOptions options passed as url query when getting a list of api's
type ListAPIOptions struct {
	Compressed *bool   `json:"compressed,omitempty"`
	Query      *string `json:"q,omitempty"`
	Pages      int     `json:"p,omitempty"`
	Sort       *string `json:"sort,omitempty"`
	Category   *string `json:"category,omitempty"`
	AuthType   *string `json:"auth_type,omitempty"`
	Graph      *bool   `json:"graph,omitempty"`
}

// Params returns url.Values that matches what the admin api expects from ls.
func (ls ListAPIOptions) Params() url.Values {
	o := make(map[string]interface{})

	b, _ := json.Marshal(ls)
	json.Unmarshal(b, &o)

	u := make(url.Values)

	for k, v := range o {
		u[k] = []string{fmt.Sprint(v)}
	}

	return u
}

func (in *MapStringInterfaceType) DeepCopyInto(out *MapStringInterfaceType) {
	// controller-gen cannot handle the interface{} type of an aliased Unstructured,
	// thus we write our own DeepCopyInto function.
	if out != nil {
		casted := in.Unstructured
		deepCopy := casted.DeepCopy()
		out.Object = deepCopy.Object
	}
}

func (in *MapStringInterfaceType) UnmarshalJSON(data []byte) error {
	m := make(map[string]interface{})
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}

	in.Object = m

	return nil
}

func (in *MapStringInterfaceType) MarshalJSON() (data []byte, err error) {
	return json.Marshal(in.Object)
}
