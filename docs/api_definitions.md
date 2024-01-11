# API Definitions

An API Definition describes the configuration of an API. It instructs Tyk Gateway how to configure the API. 

To check the supported features of the API Definitions CRD version you're currently using, please use the "Switch branches / tags" feature on GitHub and select corresponding version tag.

## Implemented Capabilities / Support Status

| Symbol | Description                       |
|--------|-----------------------------------|
| ✅      | Fully supported                   |
| ⚠️     | Untested / Requires Documentation |
| ❌️     | Not currently supported           |

## API Types
| Type                           | Support | Supported From | Comments                     | Sample                                                                                     |
|--------------------------------|---------|----------------|------------------------------|--------------------------------------------------------------------------------------------|
| HTTP                           | ✅       | v0.1           | -                            |                                                                                            |
| HTTPS                          | ✅       | v0.4           | -                            | [Sample](./../config/samples/tls/example.yaml)️                                            |
| TCP                            | ✅       | v0.1           | -                            |                                                                                            |
| TLS                            | ✅       | v0.1           | -                            |                                                                                            |
| GraphQL - Proxy                | ✅       | v0.1           | -                            | [Sample](./../config/samples/graphql_proxy/trevorblades_graphql_proxy.yaml)                |
| Universal Data Graph v1        | ✅       | v0.1           | -                             | [Sample - with GraphQL V1 Engine (Tyk v3.1 or before) ](./../config/samples/udg_1.yaml) |
| Universal Data Graph v2        | ✅       | v0.12          | -                             | [Sample - with GraphQL V2 Engine (Tyk v3.2 and after) ](./../config/samples/udg2/) |
| GraphQL - Federation           | ✅       | v0.12          | -                             | [Sample](./../config/samples/federation/README.md)                                         |
| OAS           | ❌️       | -          | Coming Soon!                             |                                  |

## Routing

| Type                        | Supported | Supported From | Comments | Sample |
| --------------------------- | --------- | -------------- | -------- | ------ |
| Path-Based                  | ✅        | v0.1           | -        | [Sample](./../config/samples/httpbin.yaml) |
| Host-Based                  | ✅        | v0.1           | -        | [Sample](./../config/samples/httpbin_routing_by_hostname.yaml) |
| Version-Based (Header)      | ⚠️        | v0.1           | Untested |        |
| Version-Based (QueryString) | ⚠️        | v0.1           | Untested |        |
| Version-Based (Subdomain)   | ⚠️        | v0.1           | Untested |        |

## Client to Gateway Authentication

| Type                          | Supported | Supported From | Comments | Sample |
| ----------------------------- | --------- | -------------- | -------- | ------ |
| Keyless (Open)                | ✅        | v0.1           | -        | [Sample](./../config/samples/httpbin.yaml) |
| Static Bearer Token           | ✅        | v0.1           | -        | [Sample](./../config/samples/httpbin_protected.yaml) |
| JWT                           | ✅️        | v0.5           | -        | [Sample](./../config/samples/jwt-auth) |
| OAuth2 - Client Credentials   | ✅️        | v0.6           | -        | [Sample](./../config/samples/oauth2/client_credentials.yaml) |
| OAuth2 - Authorization Code                 | ⚠️        | v0.6           | Untested | |
| OAuth2 - Authorization Code + Refresh Token | ⚠️        | v0.6           | Untested | |
| OAuth2 - Implicit             | ⚠️        | v0.6           | Untested | |
| OAuth2 - Password             | ⚠️        | v0.6           | Untested | |
| OpenID Connect                | ❌        | -              | Not implemented | |
| mTLS                          | ✅      | v0.11              | Only static client mTLS is supported | [Sample](./../config/samples/mtls/client/) |
| HMAC                          | ❌        | -              | Not implemented | |
| Basic Authentication          | ✅        | v0.12          | Only enabling with default metadata values is supported  | [Sample](./../config/samples/basic-auth/httpbin_basic_authentication.yaml) |
| Multiple (Chained) Auth       | ✅        | v0.14          | - | [Sample](./../config/samples/multiple-auth/httpbin_basic_authentication_and_mTLS.yaml) |
| Plugin Auth - Go              | ✅        | v0.11          | - | [Sample](./api_definitions/custom_plugin_goauth.yaml) |
| Plugin Auth - gRPC            | ✅        | v0.1           | - | [Sample](./../bdd/features/api_http_grpc_plugin.feature) |
| IP Whitelisting               | ✅        | v0.5           | - | [Sample](./api_definitions/ip.md#whitelisting) |
| IP Blacklisting               | ✅        | v0.5           | - | [Sample](./api_definitions/ip.md#blacklisting) |


## Gateway to Upstream Authentication

| Type                                            | Supported | Supported From | Comments        | Sample |
|-------------------------------------------------|-----------|----------------|-----------------| ------ |
| Public Key Certificate Pinning                  | ✅        | v0.9           |                 | [Sample](../config/samples/httpbin_certificate_pinning.yaml) |
| Upstream Certificates mTLS                      | ✅        | v0.9           |                 | [From Secret](../config/samples/httpbin_upstream_cert.yaml) or [Manual Upload](../config/samples/httpbin_upstream_cert_manual.yaml) |
| Request Signing                                 | ❌        | -              | Not implemented | |

## Features

| Feature                              | Supported | Supported From | Comments                                                               | Sample                                                          |
|--------------------------------------|-----------|----------------|------------------------------------------------------------------------|-----------------------------------------------------------------|
| API Tagging                          | ✅         | v0.1           | -                                                                      |                                                                 |
| Config Data                          | ✅         | v0.8.2         | -                                                                      | [Sample](./../config/samples/config_data_virtual_endpoint.yaml) |
| Context Variables                    | ✅         | v0.1           | -                                                                      |
| Cross Origin Resource Sharing (CORS) | ✅        | v0.2           | - | [Sample](./../config/samples/httpbin_cors.yaml)                 |
| Custom Plugins - Go                  | ⚠️        | v0.1           | Untested                                                               |
| Custom Plugins - gRPC                | ✅         | v0.1           | -                                                                      | [Sample](./../bdd/features/api_http_grpc_plugin.feature)        |
| Custom Plugins - Javascript          | ✅         | v0.1           | -                                                                      | [Sample](./api_definitions/custom_plugin.md)                    |
| Custom Plugins - Lua                 | ⚠️        | v0.1           | Untested                                                               |
| Custom Plugins - Python              | ⚠️        | v0.1           | Untested                                                               |
| Global Rate Limit                    | ✅         | v0.10          | -                                                                      | [Sample](./../config/samples/httpbin_global_rate_limit.yaml)    |
| Segment Tags                         | ✅         | v0.1           | -                                                                      | [Sample](./../config/samples/httpbin_tagged.yaml)               |
| Tag Headers                          | ⚠️         | -              | Untested                                                               |
| Webhooks                             | ❌         | -              | [WIP #62](https://github.com/TykTechnologies/tyk-operator/issues/62)   | 
| Looping                              | ⚠️        | v0.6           | Untested                                                               | [Sample](./api_definitions/looping.md)                          |
| Active API                           | ✅         | v0.2           | Only available to Tyk Self Managed (Pro) users                         | [Sample](./api_definitions/fields.md#active)                    |
| Round Robin Load Balancing           | ✅         | -              | -                                                                     | [Sample](./../config/samples/enable_round_robin_load_balancing.yaml)                    |

## APIDefinition - Endpoint Middleware

| Endpoint Middleware               | Supported | Supported From | Comments                                       | Sample |
|-----------------------------------|-----------|----------------|------------------------------------------------|--------|
| Analytics - Endpoint Tracking     | ✅        | v0.1           |                                                | [Sample](../config/samples/httpbin_endpoint_tracking.yaml) |
| Availability - Circuit Breaker    | ✅        | v0.5           | -                                              | [Sample](./../config/samples/httpbin_timeout.yaml)  |
| Availability - Enforced Timeouts  | ✅        | v0.1           | -                                              | [Sample](./../config/samples/httpbin_timeout.yaml) |
| Headers - Global Request Add      | ✅        | v0.1           | -                                              | [Sample](../config/samples/httpbin_global-headers.yaml) |
| Headers - Global Request Remove   | ✅        | v0.1           | -                                              | [Sample](../config/samples/httpbin_global-headers.yaml) |
| Headers - Global Response Add     | ✅        | v0.1           | -                                              | [Sample](../config/samples/httpbin_global-headers.yaml) |
| Headers - Global Response Remove  | ✅        | v0.1           | -                                              | [Sample](../config/samples/httpbin_global-headers.yaml) |
| Performance - Cache               | ✅        | v0.1           | -                                              | [Sample](./../config/samples/httpbin_cache.yaml) |
| Plugin - Virtual Endpoint         | ✅        | v0.1           | -                                              | [Sample](./../config/samples/config_data_virtual_endpoint.yaml) |
| Security - Allow list             | ✅️        | v0.8.2         | -                                              | [Sample](./../config/samples/httpbin_whitelist.yaml) |
| Security - Block list             | ✅️        | v0.8.2         | -                                              | [Sample](./../config/samples/httpbin_blacklist.yaml) |
| Security - Ignore list            | ✅        | v0.8.2         | -                                              | [Sample](./../config/samples/httpbin_ignored.yaml) |
| Transform - Internal              | ⚠️        | v0.1           | Untested                                       | |
| Transform - Method                | ✅        | v0.5           | -                                              | [Sample](../bdd/custom_resources/transform/method.yaml) |
| Transform - Mock                  | ✅        | v0.1           | -                                             | [Sample](../config/samples/httpbin_mock.yaml)|
| Transform - Request Body          | ✅        | v0.1           | -                                              | [Sample](../config/samples/httpbin_transform.yaml) |
| Transform - Response Body         | ✅        | v0.1           | -                                              | [Sample](../config/samples/httpbin_transform.yaml) |
| Transform - Request Body JQ       | ⚠️        | v0.1           | Untested - Requires JQ on Gateway Docker Image | |
| Transform - Response Body JQ      | ⚠️        | v0.1           | Untested - Requires JQ on Gateway Docker Image | |
| Transform - URL Rewrite Basic     | ✅️        | v0.1           | -                                              | [Sample](../config/samples/url_rewrite_basic.yaml) |
| Transform - URL Rewrite Advanced  | ⚠️        | v0.1           | Untested                                       | |
| Validate - JSON Schema            | ✅        | v0.8.2         | -                                              | [Sample](../config/samples/httpbin_json_schema_validation.yaml) |
| Validate - Limit Request Size     | ✅️        | v0.1           | -                                              | [Sample](../config/samples/request_size.yaml) |

## APIDefinition - Migrating Existing APIs

Please visit the [API migration page](https://tyk.io/docs/tyk-stack/tyk-operator/migration/#migration-of-existing-api) for more info

## Understanding reconciliation status

Please visit [Latest Transaction Status](./api_definitions/latest-transaction.md) page to see how you can check latest APIDefinition reconciliation status.
