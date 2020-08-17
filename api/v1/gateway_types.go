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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// GatewaySpec defines the desired state of Gateway
type GatewaySpec struct {
	// +kubebuilder:validation:Minimum=0
	// Size is the size of the gateway deployment
	Size   int32  `json:"size"`
	Config Config `json:"config"`
}

type Config struct {
	// +kubebuilder:validation:Minimum=0
	ListenPort int32  `json:"listen_port"`
	Secret     string `json:"secret"`
	NodeSecret string `json:"node_secret"`
	//AllowInsecureConfigs      bool                    `json:"allow_insecure_configs"`
	//PublicKeyPath             string                  `json:"public_key_path"`
	//AllowRemoteConfig         bool                    `bson:"allow_remote_config" json:"allow_remote_config"`
	//Security                  SecurityConfig          `json:"security"`
	//HttpServerOptions         HttpServerOptionsConfig `json:"http_server_options"`
	//ReloadWaitTime            int                     `bson:"reload_wait_time" json:"reload_wait_time"`
	//VersionHeader             string                  `json:"version_header"`
	//UseAsyncSessionWrite      bool                    `json:"optimisations_use_async_session_write"`
	//SuppressRedisSignalReload bool                    `json:"suppress_redis_signal_reload"`
	//// Gateway Security Policies
	//HashKeys                bool           `json:"hash_keys"`
	//HashKeyFunction         string         `json:"hash_key_function"`
	//EnableHashedKeysListing bool           `json:"enable_hashed_keys_listing"`
	//MinTokenLength          int            `json:"min_token_length"`
	//EnableAPISegregation    bool           `json:"enable_api_segregation"`
	//TemplatePath            string         `json:"template_path"`
	//Policies                PoliciesConfig `json:"policies"`
	//DisablePortWhiteList    bool           `json:"disable_ports_whitelist"`
	//// Defines the ports that will be available for the api services to bind to.
	//// This is a map of protocol to PortWhiteList. This allows per protocol
	//// configurations.
	//PortWhiteList map[string]PortWhiteList `json:"ports_whitelist"`
	//
	//// CE Configurations
	//AppPath string `json:"app_path"`
	//
	//// Dashboard Configurations
	//UseDBAppConfigs          bool                   `json:"use_db_app_configs"`
	//DBAppConfOptions         DBAppConfOptionsConfig `json:"db_app_conf_options"`
	//Storage                  StorageOptionsConf     `json:"storage"`
	//DisableDashboardZeroConf bool                   `json:"disable_dashboard_zeroconf"`
	//
	//// Slave Configurations
	//SlaveOptions   SlaveOptionsConfig `json:"slave_options"`
	//ManagementNode bool               `json:"management_node"`
	////AuthOverride   AuthOverrideConf   `json:"auth_override"`
	//
	//// Rate Limiting Strategy
	//EnableNonTransactionalRateLimiter bool    `json:"enable_non_transactional_rate_limiter"`
	//EnableSentinelRateLimiter         bool    `json:"enable_sentinel_rate_limiter"`
	//EnableRedisRollingLimiter         bool    `json:"enable_redis_rolling_limiter"`
	//DRLNotificationFrequency          int     `json:"drl_notification_frequency"`
	//DRLThreshold                      float64 `json:"drl_threshold"`
	//
	//// Organization configurations
	//EnforceOrgDataAge               bool          `json:"enforce_org_data_age"`
	//EnforceOrgDataDetailLogging     bool          `json:"enforce_org_data_detail_logging"`
	//EnforceOrgQuotas                bool          `json:"enforce_org_quotas"`
	//ExperimentalProcessOrgOffThread bool          `json:"experimental_process_org_off_thread"`
	//Monitor                         MonitorConfig `json:"monitor"`
	//
	//// Client-Gateway Configuration
	//MaxIdleConns         int   `bson:"max_idle_connections" json:"max_idle_connections"`
	//MaxIdleConnsPerHost  int   `bson:"max_idle_connections_per_host" json:"max_idle_connections_per_host"`
	//MaxConnTime          int64 `json:"max_conn_time"`
	//CloseIdleConnections bool  `json:"close_idle_connections"`
	//CloseConnections     bool  `json:"close_connections"`
	//EnableCustomDomains  bool  `json:"enable_custom_domains"`
	//// If AllowMasterKeys is set to true, session objects (key definitions) that do not have explicit access rights set
	//// will be allowed by Tyk. This means that keys that are created have access to ALL APIs, which in many cases is
	//// unwanted behaviour unless you are sure about what you are doing.
	//AllowMasterKeys bool `json:"allow_master_keys"`
	//
	//// Gateway-Service Configuration
	//ServiceDiscovery              ServiceDiscoveryConf `json:"service_discovery"`
	//ProxySSLInsecureSkipVerify    bool                 `json:"proxy_ssl_insecure_skip_verify"`
	//ProxyEnableHttp2              bool                 `json:"proxy_enable_http2"`
	//ProxySSLMinVersion            uint16               `json:"proxy_ssl_min_version"`
	//ProxySSLCipherSuites          []string             `json:"proxy_ssl_ciphers"`
	//ProxyDefaultTimeout           float64              `json:"proxy_default_timeout"`
	//ProxySSLDisableRenegotiation  bool                 `json:"proxy_ssl_disable_renegotiation"`
	//ProxyCloseConnections         bool                 `json:"proxy_close_connections"`
	//UptimeTests                   UptimeTestsConfig    `json:"uptime_tests"`
	//HealthCheck                   HealthCheckConfig    `json:"health_check"`
	//OauthRefreshExpire            int64                `json:"oauth_refresh_token_expire"`
	//OauthTokenExpire              int32                `json:"oauth_token_expire"`
	//OauthTokenExpiredRetainPeriod int32                `json:"oauth_token_expired_retain_period"`
	//OauthRedirectUriSeparator     string               `json:"oauth_redirect_uri_separator"`
	//OauthErrorStatusCode          int                  `json:"oauth_error_status_code"`
	//EnableKeyLogging              bool                 `json:"enable_key_logging"`
	//SSLForceCommonNameCheck       bool                 `json:"ssl_force_common_name_check"`
	//
	//// Proxy analytics configuration
	//EnableAnalytics bool                  `json:"enable_analytics"`
	//AnalyticsConfig AnalyticsConfigConfig `json:"analytics_config"`
	//
	//LivenessCheck LivenessCheckConfig `json:"liveness_check"`
	//// Cache
	//DnsCache                 DnsCacheConfig        `json:"dns_cache"`
	//DisableRegexpCache       bool                  `json:"disable_regexp_cache"`
	//RegexpCacheExpire        int32                 `json:"regexp_cache_expire"`
	//LocalSessionCache        LocalSessionCacheConf `json:"local_session_cache"`
	//EnableSeperateCacheStore bool                  `json:"enable_separate_cache_store"`
	//CacheStorage             StorageOptionsConf    `json:"cache_storage"`
	//
	//// Middleware/Plugin Configuration
	//EnableBundleDownloader   bool            `bson:"enable_bundle_downloader" json:"enable_bundle_downloader"`
	//BundleBaseURL            string          `bson:"bundle_base_url" json:"bundle_base_url"`
	//BundleInsecureSkipVerify bool            `bson:"bundle_insecure_skip_verify" json:"bundle_insecure_skip_verify"`
	//EnableJSVM               bool            `json:"enable_jsvm"`
	//JSVMTimeout              int             `json:"jsvm_timeout"`
	//DisableVirtualPathBlobs  bool            `json:"disable_virtual_path_blobs"`
	//TykJSPath                string          `json:"tyk_js_path"`
	//MiddlewarePath           string          `json:"middleware_path"`
	//CoProcessOptions         CoProcessConfig `json:"coprocess_options"`
	//IgnoreEndpointCase       bool            `json:"ignore_endpoint_case"`
	//
	//// Monitoring, Logging & Profiling
	//LogLevel                string         `json:"log_level"`
	//HealthCheckEndpointName string         `json:"health_check_endpoint_name"`
	//NewRelic                NewRelicConfig `json:"newrelic"`
	//HTTPProfile             bool           `json:"enable_http_profiler"`
	//UseRedisLog             bool           `json:"use_redis_log"`
	//SentryCode              string         `json:"sentry_code"`
	//SentryLogLevel          string         `json:"sentry_log_level"`
	//UseSentry               bool           `json:"use_sentry"`
	//UseSyslog               bool           `json:"use_syslog"`
	//UseGraylog              bool           `json:"use_graylog"`
	//UseLogstash             bool           `json:"use_logstash"`
	//Track404Logs            bool           `json:"track_404_logs"`
	//GraylogNetworkAddr      string         `json:"graylog_network_addr"`
	//LogstashNetworkAddr     string         `json:"logstash_network_addr"`
	//SyslogTransport         string         `json:"syslog_transport"`
	//LogstashTransport       string         `json:"logstash_transport"`
	//SyslogNetworkAddr       string         `json:"syslog_network_addr"`
	//StatsdConnectionString  string         `json:"statsd_connection_string"`
	//StatsdPrefix            string         `json:"statsd_prefix"`
	//
	//// Event System
	////EventHandlers        apidef.EventHandlerMetaConfig         `json:"event_handlers"`
	////EventTriggers        map[apidef.TykEvent][]TykEventHandler `json:"event_trigers_defunct"`  // Deprecated: Config.GetEventTriggers instead.
	////EventTriggersDefunct map[apidef.TykEvent][]TykEventHandler `json:"event_triggers_defunct"` // Deprecated: Config.GetEventTriggers instead.
	//
	//// TODO: These config options are not documented - What do they do?
	//SessionUpdatePoolSize          int   `json:"session_update_pool_size"`
	//SessionUpdateBufferSize        int   `json:"session_update_buffer_size"`
	//SupressDefaultOrgStore         bool  `json:"suppress_default_org_store"`
	//LegacyEnableAllowanceCountdown bool  `bson:"legacy_enable_allowance_countdown" json:"legacy_enable_allowance_countdown"`
	//GlobalSessionLifetime          int64 `bson:"global_session_lifetime" json:"global_session_lifetime"`
	//ForceGlobalSessionLifetime     bool  `bson:"force_global_session_lifetime" json:"force_global_session_lifetime"`
	//HideGeneratorHeader            bool  `json:"hide_generator_header"`
	//KV                             struct {
	//	Consul ConsulConfig `json:"consul"`
	//	Vault  VaultConfig  `json:"vault"`
	//} `json:"kv"`
	//
	//// Secrets are key-value pairs that can be accessed in the dashboard via "secrets://"
	//Secrets map[string]string `json:"secrets"`
	//
	//// OverrideMessages is used to override returned API error codes and messages.
	//OverrideMessages map[string]TykError `bson:"override_messages" json:"override_messages"`
	//
	//// Cloud flag shows that gateway runs in Tyk-cloud.
	//Cloud bool `json:"cloud"`
}

//type PoliciesConfig struct {
//	PolicySource           string `json:"policy_source"`
//	PolicyConnectionString string `json:"policy_connection_string"`
//	PolicyRecordName       string `json:"policy_record_name"`
//	AllowExplicitPolicyID  bool   `json:"allow_explicit_policy_id"`
//}
//
//type DBAppConfOptionsConfig struct {
//	ConnectionString string   `json:"connection_string"`
//	NodeIsSegmented  bool     `json:"node_is_segmented"`
//	Tags             []string `json:"tags"`
//}
//
//type StorageOptionsConf struct {
//	Type                  string            `json:"type"`
//	Host                  string            `json:"host"`
//	Port                  int               `json:"port"`
//	Hosts                 map[string]string `json:"hosts"` // Deprecated: Addrs instead.
//	Addrs                 []string          `json:"addrs"`
//	MasterName            string            `json:"master_name"`
//	Username              string            `json:"username"`
//	Password              string            `json:"password"`
//	Database              int               `json:"database"`
//	MaxIdle               int               `json:"optimisation_max_idle"`
//	MaxActive             int               `json:"optimisation_max_active"`
//	Timeout               int               `json:"timeout"`
//	EnableCluster         bool              `json:"enable_cluster"`
//	UseSSL                bool              `json:"use_ssl"`
//	SSLInsecureSkipVerify bool              `json:"ssl_insecure_skip_verify"`
//}
//
//type NormalisedURLConfig struct {
//	Enabled            bool                 `json:"enabled"`
//	NormaliseUUIDs     bool                 `json:"normalise_uuids"`
//	NormaliseNumbers   bool                 `json:"normalise_numbers"`
//	Custom             []string             `json:"custom_patterns"`
//	//CompiledPatternSet NormaliseURLPatterns `json:"-"` // see analytics.go
//}
//
////type NormaliseURLPatterns struct {
////	UUIDs  *regexp.Regexp   `json:"uuids"`
////	IDs    *regexp.Regexp   `json:"ids"`
////	Custom []*regexp.Regexp `json:"custom"`
////}
//
//type AnalyticsConfigConfig struct {
//	Type                    string              `json:"type"`
//	IgnoredIPs              []string            `json:"ignored_ips"`
//	EnableDetailedRecording bool                `json:"enable_detailed_recording"`
//	EnableGeoIP             bool                `json:"enable_geo_ip"`
//	GeoIPDBLocation         string              `json:"geo_ip_db_path"`
//	NormaliseUrls           NormalisedURLConfig `json:"normalise_urls"`
//	PoolSize                int                 `json:"pool_size"`
//	RecordsBufferSize       uint64              `json:"records_buffer_size"`
//	StorageExpirationTime   int                 `json:"storage_expiration_time"`
//	ignoredIPsCompiled      map[string]bool
//}
//
//type HealthCheckConfig struct {
//	EnableHealthChecks      bool  `json:"enable_health_checks"`
//	HealthCheckValueTimeout int64 `json:"health_check_value_timeouts"`
//}
//
//type LivenessCheckConfig struct {
//	CheckDuration time.Duration `json:"check_duration"`
//}
//
//type DnsCacheConfig struct {
//	Enabled                   bool              `json:"enabled"`
//	TTL                       int64             `json:"ttl"`
//	CheckInterval             int64             `json:"-" ignored:"true"` //controls cache cleanup interval. By convention shouldn't be exposed to config or env_variable_setup
//	MultipleIPsHandleStrategy IPsHandleStrategy `json:"multiple_ips_handle_strategy"`
//}
//
//type IPsHandleStrategy string
//
//type MonitorConfig struct {
//	EnableTriggerMonitors bool               `json:"enable_trigger_monitors"`
//	Config                WebHookHandlerConf `json:"configuration"`
//	GlobalTriggerLimit    float64            `json:"global_trigger_limit"`
//	MonitorUserKeys       bool               `json:"monitor_user_keys"`
//	MonitorOrgKeys        bool               `json:"monitor_org_keys"`
//}
//
//type WebHookHandlerConf struct {
//	Method       string            `bson:"method" json:"method"`
//	TargetPath   string            `bson:"target_path" json:"target_path"`
//	TemplatePath string            `bson:"template_path" json:"template_path"`
//	HeaderList   map[string]string `bson:"header_map" json:"header_map"`
//	EventTimeout int64             `bson:"event_timeout" json:"event_timeout"`
//}
//
//type SlaveOptionsConfig struct {
//	UseRPC                          bool    `json:"use_rpc"`
//	UseSSL                          bool    `json:"use_ssl"`
//	SSLInsecureSkipVerify           bool    `json:"ssl_insecure_skip_verify"`
//	ConnectionString                string  `json:"connection_string"`
//	RPCKey                          string  `json:"rpc_key"`
//	APIKey                          string  `json:"api_key"`
//	EnableRPCCache                  bool    `json:"enable_rpc_cache"`
//	BindToSlugsInsteadOfListenPaths bool    `json:"bind_to_slugs"`
//	DisableKeySpaceSync             bool    `json:"disable_keyspace_sync"`
//	GroupID                         string  `json:"group_id"`
//	CallTimeout                     int     `json:"call_timeout"`
//	PingTimeout                     int     `json:"ping_timeout"`
//	RPCPoolSize                     int     `json:"rpc_pool_size"`
//	KeySpaceSyncInterval            float32 `json:"key_space_sync_interval"`
//}
//
//type LocalSessionCacheConf struct {
//	DisableCacheSessionState bool `json:"disable_cached_session_state"`
//	CachedSessionTimeout     int  `json:"cached_session_timeout"`
//	CacheSessionEviction     int  `json:"cached_session_eviction"`
//}
//
//type HttpServerOptionsConfig struct {
//	OverrideDefaults       bool       `json:"override_defaults"`
//	ReadTimeout            int        `json:"read_timeout"`
//	WriteTimeout           int        `json:"write_timeout"`
//	UseSSL                 bool       `json:"use_ssl"`
//	UseLE_SSL              bool       `json:"use_ssl_le"`
//	EnableHttp2            bool       `json:"enable_http2"`
//	SSLInsecureSkipVerify  bool       `json:"ssl_insecure_skip_verify"`
//	EnableWebSockets       bool       `json:"enable_websockets"`
//	Certificates           []CertData `json:"certificates"`
//	SSLCertificates        []string   `json:"ssl_certificates"`
//	ServerName             string     `json:"server_name"`
//	MinVersion             uint16     `json:"min_version"`
//	FlushInterval          int        `json:"flush_interval"`
//	SkipURLCleaning        bool       `json:"skip_url_cleaning"`
//	SkipTargetPathEscaping bool       `json:"skip_target_path_escaping"`
//	Ciphers                []string   `json:"ssl_ciphers"`
//}
//
////type AuthOverrideConf struct {
////	ForceAuthProvider    bool                       `json:"force_auth_provider"`
////	//AuthProvider         apidef.AuthProviderMeta    `json:"auth_provider"`
////	ForceSessionProvider bool                       `json:"force_session_provider"`
////	//SessionProvider      apidef.SessionProviderMeta `json:"session_provider"`
////}
//
//type UptimeTestsConfigDetail struct {
//	FailureTriggerSampleSize int  `json:"failure_trigger_sample_size"`
//	TimeWait                 int  `json:"time_wait"`
//	CheckerPoolSize          int  `json:"checker_pool_size"`
//	EnableUptimeAnalytics    bool `json:"enable_uptime_analytics"`
//}
//
//type UptimeTestsConfig struct {
//	Disable bool                    `json:"disable"`
//	Config  UptimeTestsConfigDetail `json:"config"`
//}
//
//type ServiceDiscoveryConf struct {
//	DefaultCacheTimeout int `json:"default_cache_timeout"`
//}
//
//type CoProcessConfig struct {
//	EnableCoProcess     bool   `json:"enable_coprocess"`
//	CoProcessGRPCServer string `json:"coprocess_grpc_server"`
//	GRPCRecvMaxSize     int    `json:"grpc_recv_max_size"`
//	GRPCSendMaxSize     int    `json:"grpc_send_max_size"`
//	PythonPathPrefix    string `json:"python_path_prefix"`
//	PythonVersion       string `json:"python_version"`
//}
//
//type CertificatesConfig struct {
//	API        []string          `json:"apis"`
//	Upstream   map[string]string `json:"upstream"`
//	ControlAPI []string          `json:"control_api"`
//	Dashboard  []string          `json:"dashboard_api"`
//	MDCB       []string          `json:"mdcb_api"`
//}
//
//type SecurityConfig struct {
//	PrivateCertificateEncodingSecret string             `json:"private_certificate_encoding_secret"`
//	ControlAPIUseMutualTLS           bool               `json:"control_api_use_mutual_tls"`
//	PinnedPublicKeys                 map[string]string  `json:"pinned_public_keys"`
//	Certificates                     CertificatesConfig `json:"certificates"`
//}
//
//type NewRelicConfig struct {
//	AppName    string `json:"app_name"`
//	LicenseKey string `json:"license_key"`
//}
//
//// ServicePort defines a protocol and port on which a service can bind to
//type ServicePort struct {
//	Protocol string `json:"protocol"`
//	Port     int    `json:"port"`
//}
//
//// PortWhiteList defines ports that will be allowed by the gateway.
//type PortWhiteList struct {
//	Ranges []PortRange `json:"ranges,omitempty"`
//	Ports  []int       `json:"ports,omitempty"`
//}
//
//// PortRange defines a range of ports inclusively.
//type PortRange struct {
//	From int `json:"from"`
//	To   int `json:"to"`
//}
//
//type TykError struct {
//	Message string `json:"message"`
//	Code    int    `json:"code"`
//}
//
//// VaultConfig is used to configure the creation of a client
//// This is a stripped down version of the Config struct in vault's API client
//type VaultConfig struct {
//	// Address is the address of the Vault server. This should be a complete
//	// URL such as "http://vault.example.com".
//	Address string `json:"address"`
//
//	// AgentAddress is the address of the local Vault agent. This should be a
//	// complete URL such as "http://vault.example.com".
//	AgentAddress string `json:"agent_address"`
//
//	// MaxRetries controls the maximum number of times to retry when a vault
//	// serer occurs
//	MaxRetries int `json:"max_retries"`
//
//	Timeout time.Duration `json:"timeout"`
//
//	// Token is the vault root token
//	Token string `json:"token"`
//
//	// KVVersion is the version number of Vault. Usually defaults to 2
//	KVVersion int `json:"kv_version"`
//}
//
//// ConsulConfig is used to configure the creation of a client
//// This is a stripped down version of the Config struct in consul's API client
//type ConsulConfig struct {
//	// Address is the address of the Consul server
//	Address string `json:"address"`
//
//	// Scheme is the URI scheme for the Consul server
//	Scheme string `json:"scheme"`
//
//	// Datacenter to use. If not provided, the default agent datacenter is used.
//	Datacenter string `json:"datacenter"`
//
//	// HttpAuth is the auth info to use for http access.
//	HttpAuth struct {
//		// Username to use for HTTP Basic Authentication
//		Username string `json:"username"`
//
//		// Password to use for HTTP Basic Authentication
//		Password string `json:"password"`
//	} `json:"http_auth"`
//
//	// WaitTime limits how long a Watch will block. If not provided,
//	// the agent default values will be used.
//	WaitTime time.Duration `json:"wait_time"`
//
//	// Token is used to provide a per-request ACL token
//	// which overrides the agent's default token.
//	Token string `json:"token"`
//
//	TLSConfig struct {
//		Address string `json:"address"`
//
//		CAFile string `json:"ca_file"`
//
//		CAPath string `json:"ca_path"`
//
//		CertFile string `json:"cert_file"`
//
//		KeyFile string `json:"key_file"`
//
//		InsecureSkipVerify bool `json:"insecure_skip_verify"`
//	} `json:"tls_config"`
//}
//
//type CertData struct {
//	Name     string `json:"domain_name"`
//	CertFile string `json:"cert_file"`
//	KeyFile  string `json:"key_file"`
//}

// GatewayStatus defines the observed state of Gateway
type GatewayStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// The names of the Tyk Gateway pods
	Nodes []string `json:"nodes"`
}

// +kubebuilder:object:root=true

// Gateway is the Schema for the gateways API
// +kubebuilder:subresource:status
type Gateway struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GatewaySpec   `json:"spec,omitempty"`
	Status GatewayStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// GatewayList contains a list of Gateway
type GatewayList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Gateway `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Gateway{}, &GatewayList{})
}
