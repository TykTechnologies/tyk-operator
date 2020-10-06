# API Definitions

An [API Definitions](https://tyk.io/docs/getting-started/key-concepts/what-is-an-api-definition/) is a reverse proxy definition in the world of Tyk.

Here is the list of currently supported plugins/middleware that you can apply in an [API definition](https://tyk.io/docs/getting-started/key-concepts/what-is-an-api-definition/).

| Middleware  | Supported |
| ----------- | --------- |
| [Cache](./../config/samples/httpbin_cache.yaml) | ✅ |
| [Circuit Breaker](./../config/samples/httpbin_timeout.yaml) | [❌️](# "Incompatible types string vs float64") |
| [Custom Plugins](./api_definitions/custom_plugin.md) | ✅ |
| [Enforced Timeouts](./../config/samples/httpbin_timeout.yaml) | ✅ |
| [Headers - Global Request Add](../config/samples/httpbin_global-headers.yaml) | ✅ |
| [Headers - Global Request Remove](../config/samples/httpbin_global-headers.yaml) | ✅ |
| [Headers - Global Response Add](../config/samples/httpbin_global-headers.yaml) | ✅ |
| [Headers - Global Response Remove](../config/samples/httpbin_global-headers.yaml) | ✅ |
| [JSON Schema Validation](../config/samples/httpbin_validate.yaml) | [❌️](https://github.com/TykTechnologies/tyk-operator/issues/59) |
| [Transform - Request Body](../config/samples/httpbin_transform.yaml) | ✅ |
| [Transform - Response Body](../config/samples/httpbin_transform.yaml) | ✅ |
| [Transform - Request Body JQ](../config/samples/httpbin_transform.yaml) | [⚠️](# "Requires JQ on Gateway Host & Testing") |
| [Transform - Response Body JQ](../config/samples/httpbin_transform.yaml) | [⚠️](# "Requires JQ on Gateway Host & Testing") |
