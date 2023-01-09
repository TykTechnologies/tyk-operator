# Policies

To check the supported features of the Security Policy CRD version you're currently using, please use the "Switch 
branches / tags" feature on GitHub and select corresponding version tag.

## Support for CE mode

Security Policy resources are supported in Tyk CE mode as of Tyk Operator v0.13.0.

> Note: Policies API on the Tyk Gateway was introduced in Tyk v4.1. 
Please ensure that your version of Tyk Gateway is compatible.

## Dashboard Version

In order to manage Policies with the Tyk Operator, you must use Tyk Dashboard version `3.x.x+`

## Quick start

Please visit [one of the examples](./policies/api_access.md) from the Supported Feature list below to get started

## Supported Features

| Feature                                                   | Supported                                                      |
|-----------------------------------------------------------|----------------------------------------------------------------|
| [API Access](./policies/api_access.md)                    | ✅                                                              |
| [Rate Limit, Throttling, Quotas](./policies/ratelimit.md) | ✅                                                              |
| [Meta Data & Tags](./policies/metadata_tags.md)           | ✅                                                              |
| Path based permissions                                    | [⚠️](# "Requires testing")                                     |
| Partitions                                                | [⚠️](# "Requires testing")                                     |
| Per API limit                                             | [❌](https://github.com/TykTechnologies/tyk-operator/issues/66) |


## Migrating existing policies

Please visit the [policy migration page](./policies/migration.md) for more info.
