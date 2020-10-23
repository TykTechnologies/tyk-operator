# API Definitions

## Support 
An API Definition describes the configuration of an API. It instructs Tyk Gateway how to configure the API.

| Symbol | Description |
| --------- | --------- |
| ✅ | Fully supported |
| ⚠️ | Untested / Requires Documentation |
| ❌️ | Not currently supported |

## APIDefinition

### APIDefinition - Authentication

| Type | Supported | Comments |
| ----------- | --------- | --------- |
| [Keyless (open)](./../config/samples/httpbin.yaml) | ✅ | - |
| [Static Bearer Token](./../config/samples/httpbin_protected.yaml) | ✅ | - |
| JWT | ❌️ | Not implemented |
| OpenID Connect | ❌ | Not implemented |
| OAuth2 | ❌ | Not implemented |

API Definition Features

| Type | Supported | Comments |
| ----------- | --------- | --------- |
| Cross Origin Resource Sharing (CORS) | ❌ | Not implemented |
| [Custom Plugins - Go](./api_definitions/custom_plugin.md) | ⚠️ | Untested |
| [Custom Plugins - gRPC](./api_definitions/custom_plugin.md) | ⚠️ | Untested |
| [Custom Plugins - Javascript](./api_definitions/custom_plugin.md) | ✅ |
| [Custom Plugins - Lua](./api_definitions/custom_plugin.md) | ⚠️ | Untested |
| [Custom Plugins - Python](./api_definitions/custom_plugin.md) | ⚠️ | Untested |
| [Webhooks](./api_definitions/webhooks.md) | ❌ | [See issue](https://github.com/TykTechnologies/tyk-operator/issues/62) |

## APIDefinition - Endpoint Middleware

| Endpoint Middleware  | Supported | Comments |
| ----------- | --------- | --------- |
| Analytics - Endpoint Tracking | ⚠️ | Untested |
| [Availability - Circuit Breaker](./../config/samples/httpbin_timeout.yaml) | ❌ | Incompatible types string vs float64 |
| [Availability - Enforced Timeouts](./../config/samples/httpbin_timeout.yaml) | ✅ | - |
| [Headers - Global Request Add](../config/samples/httpbin_global-headers.yaml) | ✅ | - |
| [Headers - Global Request Remove](../config/samples/httpbin_global-headers.yaml) | ✅ | - |
| [Headers - Global Response Add](../config/samples/httpbin_global-headers.yaml) | ✅ | - |
| [Headers - Global Response Remove](../config/samples/httpbin_global-headers.yaml) | ✅ | - |
| [Performance - Cache](./../config/samples/httpbin_cache.yaml) | ✅ | - |
| [Plugin - Virtual Endpoint](./api_definitions/custom_plugin.md) | ✅ | - |
| [Security - Allow list](#) | ⚠️ | #92 |
| [Security - Block list](#) | ⚠️ | #92 |
| [Security - Ignore list](#) | ⚠️ | #92 |
| Transform - Internal | ⚠️ | #93 |
| Transform - Method | ⚠️ | #93 |
| Transform - Mock | ⚠️ | #93 |
| [Transform - Request Body](../config/samples/httpbin_transform.yaml) | ✅ | - |
| [Transform - Response Body](../config/samples/httpbin_transform.yaml) | ✅ | - |
| Transform - Request Body JQ | ⚠️ | Untested - Requires JQ on Gateway Docker Image |
| Transform - Response Body JQ | ⚠️ | Untested - Requires JQ on Gateway Docker Image |
| Transform - URL Rewrite | ⚠️ | Untested |
| [Validate - JSON Schema](../config/samples/httpbin_validate.yaml) | [❌️](https://github.com/TykTechnologies/tyk-operator/issues/59) |
| [Validate - Limit Request Size](../config/samples/httpbin_validate.yaml) | ✅️ | - |

## APIDefinition - Migrating Existing APIs

Please visit the [API migration page](./api_definitions/migration.md) for more info
