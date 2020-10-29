# Policies

Visit here to read more about Tyk security [policies](https://tyk.io/getting-started/key-concepts/what-is-a-security-policy/).

### Dashboard Version

In order to manage Policies with the Tyk Operator, you must use Tyk Dashboard version `3.x.x+`

#### Known Issues w/ Tyk Dashboard

- Created `SecurityPolicy` types will not allow you to create Keys through the UI. In the meantime, you can generate keys through Dashboard API Calls.
- Created `SecurityPolicy` types will not be publishable to the Developer Portal.  

Both above are being addressed in an upcoming Tyk Dashboard release.  please wait for an official release date and version.

### Quick start

Please visit [one of the examples](./policies/api_access.md) from the Supported Feature list below to get started

### Supported Features

| Feature  | Supported |
| ----------- | --------- |
| [API Access](./policies/api_access.md) | ✅ |
| [Rate Limit, Throttling, Quotas](./policies/ratelimit.md) | ✅ |
| [Meta Data & Tags](./policies/metadata_tags.md) | ✅ |
| Path based permissions | [⚠️](# "Requires testing") |
| Partitions | [⚠️](# "Requires testing") |
| Per API limit | [❌](https://github.com/TykTechnologies/tyk-operator/issues/66) |


### Migrating existing policies

Please visit the [policy migration page](./policies/migration.md) for more info