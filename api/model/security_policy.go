package model

import (
	"github.com/TykTechnologies/graphql-go-tools/pkg/graphql"
)

type SecurityPolicySpec struct {
	// MID ("_id") is generated by Tyk once the resource is created.
	// Do NOT fill this in.
	MID string `json:"_id,omitempty" oss:"ignore"`
	// If you are linking an existing Policy ID to a new YAML CRD, then
	// fill in this ID field with the "_id" field.
	// See policies.md readme for more info
	ID string `json:"id,omitempty"`

	// Name represents the name of the security policy as displayed in the Dashboard
	Name string `json:"name"`

	// OrgID is overwritten - no point setting this
	OrgID string `json:"org_id,omitempty"`

	// +kubebuilder:validation:Enum=active;draft;deny
	// State can be active, draft or deny
	// active: All keys are active and new keys can be created.
	// draft: All keys are active but no new keys can be created.
	// deny: All keys are deactivated and no keys can be created.
	State string `json:"state" oss:"ignore"`

	// Active must be set to `true` for Tyk to load the security policy into memory.
	Active bool `json:"active"`

	// IsInactive applies to the key itself. Allows enabling or disabling the policy without deleting it.
	IsInactive        bool                        `json:"is_inactive,omitempty"`
	AccessRightsArray []*AccessDefinition         `json:"access_rights_array,omitempty" oss:"ignore"`
	AccessRights      map[string]AccessDefinition `json:"access_rights,omitempty"`

	// Rate limit per X seconds (x="Per"), omit or "-1" for unlimited
	Rate int64 `json:"rate,omitempty"`

	// To be used in conjunction with "Rate".  Per seconds. 1 minute=60.  1 hour=3600
	// omit or "-1" for unlimited
	Per int64 `json:"per,omitempty"`

	// Value of Quota allowed, omit or "-1" for unlimited
	QuotaMax int64 `json:"quota_max,omitempty"`

	// Value reset length, in seconds, omit or "-1" for unlimited
	QuotaRenewalRate int64 `json:"quota_renewal_rate,omitempty"`

	// If rate limited, how many seconds to retry a request for.  omit or "-1" for unlimited
	ThrottleInterval int64 `json:"throttle_interval,omitempty"`

	// Number of retries before returning error.   omit or "-1" for unlimited
	ThrottleRetryLimit int `json:"throttle_retry_limit,omitempty"`

	// Max depth of a GraphQL query
	MaxQueryDepth                 int  `json:"max_query_depth,omitempty"`
	HMACEnabled                   bool `json:"hmac_enabled,omitempty"`
	EnableHTTPSignatureValidation bool `json:"enable_http_signature_validation,omitempty"`

	// Custom tags to apply to the key, get transfered to the analytics
	Tags []string `json:"tags,omitempty"`

	// KeyExpiresIn is the number of seconds till key expiry. For 1 hour is 3600. Default never expire or 0
	KeyExpiresIn int64             `json:"key_expires_in,omitempty"`
	Partitions   *PolicyPartitions `json:"partitions,omitempty"`

	// LastUpdated                   string                           `json:"last_updated"`
	MetaData map[string]string `json:"meta_data,omitempty"`
	// GraphQL                       map[string]GraphAccessDefinition `json:"graphql_access_rights"`
}

// GraphQLType represents a GraphQL Type for Tyk.
type GraphQLType graphql.Type

// GraphQLTypeList represents a list of GraphQLType.
type GraphQLTypeList []GraphQLType

// AccessDefinition defines which versions of an API a key has access to
type AccessDefinition struct {
	// Namespace of the ApiDefinition resource to target
	Namespace string `json:"namespace" oss:"ignore" pro:"ignore"`
	// Name of the ApiDefinition resource to target
	Name string `json:"name" oss:"ignore" pro:"ignore"`

	// TODO: APIName should not really be needed, as is auto-set from the APIDefnition Resource
	APIName string `json:"api_name,omitempty"`
	// TODO: APIID should not really be needed, as is auto-set from the APIDefnition Resource
	APIID    string   `json:"api_id,omitempty"`
	Versions []string `json:"versions,omitempty"`

	// Field access of GraphQL APIs can be restricted by setting up an allowed types list in a policy
	// or directly on a key.
	AllowedTypes GraphQLTypeList `json:"allowed_types,omitempty"`

	// Field access of GraphQL APIs can be restricted by setting up an allowed types list in a policy
	// or directly on a key.
	RestrictedTypes GraphQLTypeList `json:"restricted_types,omitempty"`

	// DisableIntrospection disables GraphQL introspection if it is set to True.
	DisableIntrospection bool `json:"disable_introspection,omitempty"`

	// FieldAccessRights is array of depth limit settings per GraphQL APIs.
	FieldAccessRights []FieldAccessDefinition `json:"field_access_rights,omitempty"`

	// Limit          APILimit     `json:"limit,omitempty"`

	AllowanceScope string       `json:"allowance_scope,omitempty"`
	AllowedURLs    []AccessSpec `json:"allowed_urls,omitempty"` // mapped string MUST be a valid regex
}

// FieldAccessDefinition represent a struct for depth limit settings per API.
type FieldAccessDefinition struct {
	// TypeName points to a type on which depth limit is set.
	// It can be either Query (most common case) or Mutation
	TypeName string `json:"type_name,omitempty"`
	// FieldName represents the name of the Query or Mutation which the limit applies to.
	FieldName string `json:"field_name,omitempty"`
	// Limit specifies the numerical value of the limit.
	Limits FieldLimits `json:"limits,omitempty"`
}

// FieldLimits represents a struct for the numerical value of the depth limit for a GraphQL query.
type FieldLimits struct {
	// MaxQueryDepth represents the numerical value of the limit.
	MaxQueryDepth int64 `json:"max_query_depth"`
}

// APILimit stores quota and rate limit on ACL level (per API)
type APILimit struct {
	Rate               int64 `json:"rate"`
	Per                int64 `json:"per"`
	ThrottleInterval   int64 `json:"throttle_interval"`
	ThrottleRetryLimit int   `json:"throttle_retry_limit"`
	MaxQueryDepth      int   `json:"max_query_depth"`
	QuotaMax           int64 `json:"quota_max"`
	QuotaRenews        int64 `json:"quota_renews"`
	QuotaRemaining     int64 `json:"quota_remaining"`
	QuotaRenewalRate   int64 `json:"quota_renewal_rate"`
}

// AccessSpec defines what URLS a user has access to and what methods are enabled
type AccessSpec struct {
	URL     string   `json:"url"`
	Methods []string `json:"methods"`
}

type PolicyPartitions struct {
	Quota      bool `json:"quota,omitempty"`
	RateLimit  bool `json:"rate_limit,omitempty"`
	Complexity bool `json:"complexity,omitempty"`
	Acl        bool `json:"acl,omitempty"`
	PerAPI     bool `json:"per_api,omitempty"`
}
