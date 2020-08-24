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

// ApiDefinitionSpec defines the desired state of ApiDefinition
type ApiDefinitionSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of ApiDefinition. Edit ApiDefinition_types.go to remove/update
	ListenPath string `json:"listen_path"`

	UseKeyless bool `json:"use_keyless"`

	//{
	//	"created_at": "2020-08-21T00:50:39+01:00",
	//	"api_model": {},
	//	"api_definition": {
	//	"api_id": "a2cb34658c7e41505b5e20c8fd4be357",
	//	"jwt_issued_at_validation_skew": 0,
	//	"upstream_certificates": {},
	//	"use_keyless": true,
	//	"enable_coprocess_auth": false,
	//	"base_identity_provided_by": "",
	//	"custom_middleware": {
	//	"pre": [],
	//	"post": [],
	//	"post_key_auth": [],
	//	"auth_check": {
	//	"name": "",
	//	"path": "",
	//	"require_session": false,
	//	"raw_body_only": false
	//},
	//	"response": [],
	//	"driver": "",
	//	"id_extractor": {
	//	"extract_from": "",
	//	"extract_with": "",
	//	"extractor_config": {}
	//}
	//},
	//	"disable_quota": false,
	//	"custom_middleware_bundle": "",
	//	"cache_options": {
	//	"cache_timeout": 60,
	//	"enable_cache": true,
	//	"cache_all_safe_requests": false,
	//	"cache_response_codes": [],
	//	"enable_upstream_cache_control": false,
	//	"cache_control_ttl_header": "",
	//	"cache_by_headers": []
	//},
	//	"enable_ip_blacklisting": false,
	//	"tag_headers": [],
	//	"jwt_scope_to_policy_mapping": {},
	//	"pinned_public_keys": {},
	//	"expire_analytics_after": 0,
	//	"domain": "",
	//	"openid_options": {
	//	"providers": [],
	//	"segregate_by_client": false
	//},
	//	"jwt_policy_field_name": "",
	//	"enable_proxy_protocol": false,
	//	"jwt_default_policies": [],
	//	"active": true,
	//	"jwt_expires_at_validation_skew": 0,
	//	"config_data": {},
	//	"notifications": {
	//	"shared_secret": "",
	//	"oauth_on_keychange_url": ""
	//},
	//	"jwt_client_base_field": "",
	//	"auth": {
	//	"use_param": false,
	//	"param_name": "",
	//	"use_cookie": false,
	//	"cookie_name": "",
	//	"auth_header_name": "Authorization",
	//	"use_certificate": false,
	//	"validate_signature": false,
	//	"signature": {
	//	"algorithm": "",
	//	"header": "",
	//	"secret": "",
	//	"allowed_clock_skew": 0,
	//	"error_code": 0,
	//	"error_message": ""
	//}
	//},
	//	"check_host_against_uptime_tests": false,
	//	"auth_provider": {
	//	"name": "",
	//	"storage_engine": "",
	//	"meta": {}
	//},
	//	"blacklisted_ips": [],
	//	"graphql": {
	//	"enabled": false,
	//	"execution_mode": "proxyOnly",
	//	"schema": "",
	//	"type_field_configurations": [],
	//	"playground": {
	//	"enabled": false,
	//	"path": ""
	//}
	//},
	//	"hmac_allowed_clock_skew": -1,
	//	"dont_set_quota_on_create": false,
	//	"uptime_tests": {
	//	"check_list": [],
	//	"config": {
	//	"expire_utime_after": 0,
	//	"service_discovery": {
	//	"use_discovery_service": false,
	//	"query_endpoint": "",
	//	"use_nested_query": false,
	//	"parent_data_path": "",
	//	"data_path": "",
	//	"cache_timeout": 60
	//},
	//	"recheck_wait": 0
	//}
	//},
	//	"enable_jwt": false,
	//	"do_not_track": false,
	//	"name": "httpbin",
	//	"slug": "httpbin",
	//	"oauth_meta": {
	//	"allowed_access_types": [],
	//	"allowed_authorize_types": [],
	//	"auth_login_redirect": ""
	//},
	//	"CORS": {
	//	"enable": false,
	//	"max_age": 24,
	//	"allow_credentials": false,
	//	"exposed_headers": [],
	//	"allowed_headers": [],
	//	"options_passthrough": false,
	//	"debug": false,
	//	"allowed_origins": [],
	//	"allowed_methods": []
	//},
	//	"event_handlers": {
	//	"events": {}
	//},
	//	"proxy": {
	//	"target_url": "http://httpbin.org",
	//	"service_discovery": {
	//	"endpoint_returns_list": false,
	//	"cache_timeout": 0,
	//	"parent_data_path": "",
	//	"query_endpoint": "",
	//	"use_discovery_service": false,
	//	"_sd_show_port_path": false,
	//	"target_path": "",
	//	"use_target_list": false,
	//	"use_nested_query": false,
	//	"data_path": "",
	//	"port_data_path": ""
	//},
	//	"check_host_against_uptime_tests": false,
	//	"transport": {
	//	"ssl_insecure_skip_verify": false,
	//	"ssl_ciphers": [],
	//	"ssl_min_version": 0,
	//	"ssl_force_common_name_check": false,
	//	"proxy_url": ""
	//},
	//	"target_list": [],
	//	"preserve_host_header": false,
	//	"strip_listen_path": true,
	//	"enable_load_balancing": false,
	//	"listen_path": "/httpbin/",
	//	"disable_strip_slash": false
	//},
	//	"client_certificates": [],
	//	"use_basic_auth": false,
	//	"version_data": {
	//	"not_versioned": true,
	//	"default_version": "",
	//	"versions": {
	//	"Default": {
	//	"name": "Default",
	//	"expires": "",
	//	"paths": {
	//	"ignored": [],
	//	"white_list": [],
	//	"black_list": []
	//},
	//	"use_extended_paths": true,
	//	"extended_paths": {
	//	"ignored": [],
	//	"white_list": [],
	//	"black_list": [],
	//	"transform": [],
	//	"transform_response": [],
	//	"transform_jq": [],
	//	"transform_jq_response": [],
	//	"transform_headers": [],
	//	"transform_response_headers": [],
	//	"hard_timeouts": [],
	//	"circuit_breakers": [],
	//	"url_rewrites": [],
	//	"virtual": [],
	//	"size_limits": [],
	//	"method_transforms": [],
	//	"track_endpoints": [],
	//	"do_not_track_endpoints": [],
	//	"validate_json": [],
	//	"internal": []
	//},
	//	"global_headers": {},
	//	"global_headers_remove": [],
	//	"global_response_headers": {},
	//	"global_response_headers_remove": [],
	//	"ignore_endpoint_case": false,
	//	"global_size_limit": 0,
	//	"override_target": ""
	//}
	//}
	//},
	//	"jwt_scope_claim_name": "",
	//	"use_standard_auth": false,
	//	"session_lifetime": 0,
	//	"hmac_allowed_algorithms": [],
	//	"disable_rate_limit": false,
	//	"definition": {
	//	"location": "header",
	//	"key": "x-api-version",
	//	"strip_path": false
	//},
	//	"use_oauth2": false,
	//	"jwt_source": "",
	//	"jwt_signing_method": "",
	//	"jwt_not_before_validation_skew": 0,
	//	"use_go_plugin_auth": false,
	//	"jwt_identity_base_field": "",
	//	"allowed_ips": [],
	//	"request_signing": {
	//	"is_enabled": false,
	//	"secret": "",
	//	"key_id": "",
	//	"algorithm": "",
	//	"header_list": [],
	//	"certificate_id": "",
	//	"signature_header": ""
	//},
	//	"org_id": "5d67b96d767e02015ea84a6f",
	//	"enable_ip_whitelisting": false,
	//	"global_rate_limit": {
	//	"rate": 0,
	//	"per": 0
	//},
	//	"protocol": "",
	//	"enable_context_vars": false,
	//	"tags": [],
	//	"basic_auth": {
	//	"disable_caching": false,
	//	"cache_ttl": 0,
	//	"extract_from_body": false,
	//	"body_user_regexp": "",
	//	"body_password_regexp": ""
	//},
	//	"listen_port": 0,
	//	"session_provider": {
	//	"name": "",
	//	"storage_engine": "",
	//	"meta": {}
	//},
	//	"auth_configs": {
	//	"authToken": {
	//	"use_param": false,
	//	"param_name": "",
	//	"use_cookie": false,
	//	"cookie_name": "",
	//	"auth_header_name": "Authorization",
	//	"use_certificate": false,
	//	"validate_signature": false,
	//	"signature": {
	//	"algorithm": "",
	//	"header": "",
	//	"secret": "",
	//	"allowed_clock_skew": 0,
	//	"error_code": 0,
	//	"error_message": ""
	//}
	//},
	//	"basic": {
	//	"use_param": false,
	//	"param_name": "",
	//	"use_cookie": false,
	//	"cookie_name": "",
	//	"auth_header_name": "Authorization",
	//	"use_certificate": false,
	//	"validate_signature": false,
	//	"signature": {
	//	"algorithm": "",
	//	"header": "",
	//	"secret": "",
	//	"allowed_clock_skew": 0,
	//	"error_code": 0,
	//	"error_message": ""
	//}
	//},
	//	"coprocess": {
	//	"use_param": false,
	//	"param_name": "",
	//	"use_cookie": false,
	//	"cookie_name": "",
	//	"auth_header_name": "Authorization",
	//	"use_certificate": false,
	//	"validate_signature": false,
	//	"signature": {
	//	"algorithm": "",
	//	"header": "",
	//	"secret": "",
	//	"allowed_clock_skew": 0,
	//	"error_code": 0,
	//	"error_message": ""
	//}
	//},
	//	"hmac": {
	//	"use_param": false,
	//	"param_name": "",
	//	"use_cookie": false,
	//	"cookie_name": "",
	//	"auth_header_name": "Authorization",
	//	"use_certificate": false,
	//	"validate_signature": false,
	//	"signature": {
	//	"algorithm": "",
	//	"header": "",
	//	"secret": "",
	//	"allowed_clock_skew": 0,
	//	"error_code": 0,
	//	"error_message": ""
	//}
	//},
	//	"jwt": {
	//	"use_param": false,
	//	"param_name": "",
	//	"use_cookie": false,
	//	"cookie_name": "",
	//	"auth_header_name": "Authorization",
	//	"use_certificate": false,
	//	"validate_signature": false,
	//	"signature": {
	//	"algorithm": "",
	//	"header": "",
	//	"secret": "",
	//	"allowed_clock_skew": 0,
	//	"error_code": 0,
	//	"error_message": ""
	//}
	//},
	//	"oauth": {
	//	"use_param": false,
	//	"param_name": "",
	//	"use_cookie": false,
	//	"cookie_name": "",
	//	"auth_header_name": "Authorization",
	//	"use_certificate": false,
	//	"validate_signature": false,
	//	"signature": {
	//	"algorithm": "",
	//	"header": "",
	//	"secret": "",
	//	"allowed_clock_skew": 0,
	//	"error_code": 0,
	//	"error_message": ""
	//}
	//},
	//	"oidc": {
	//	"use_param": false,
	//	"param_name": "",
	//	"use_cookie": false,
	//	"cookie_name": "",
	//	"auth_header_name": "Authorization",
	//	"use_certificate": false,
	//	"validate_signature": false,
	//	"signature": {
	//	"algorithm": "",
	//	"header": "",
	//	"secret": "",
	//	"allowed_clock_skew": 0,
	//	"error_code": 0,
	//	"error_message": ""
	//}
	//}
	//},
	//	"strip_auth_data": false,
	//	"id": "5f3f0c4fd3626eac273e46c1",
	//	"certificates": [],
	//	"enable_signature_checking": false,
	//	"use_openid": false,
	//	"internal": false,
	//	"jwt_skip_kid": false,
	//	"enable_batch_request_support": false,
	//	"enable_detailed_recording": false,
	//	"response_processors": [],
	//	"use_mutual_tls_auth": false
	//},
	//	"hook_references": [],
	//	"is_site": false,
	//	"sort_by": 0,
	//	"user_group_owners": [],
	//	"user_owners": []
	//}

}

// ApiDefinitionStatus defines the observed state of ApiDefinition
type ApiDefinitionStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// ApiDefinition is the Schema for the apidefinitions API
type ApiDefinition struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApiDefinitionSpec   `json:"spec,omitempty"`
	Status ApiDefinitionStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ApiDefinitionList contains a list of ApiDefinition
type ApiDefinitionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ApiDefinition `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ApiDefinition{}, &ApiDefinitionList{})
}
