# Policies

Visit here to read more about Tyk security [policies](https://tyk.io/getting-started/key-concepts/what-is-a-security-policy/).

### Known Issues w/ Tyk Dashboard

- Created `SecurityPolicy` types will not allow you to create Keys through the UI.  That is a bug addressed by a Dashboard ticket.  In the meantime, generate keys through API Calls.
- Created `SecurityPolicy` types will not be publishable to the Developer Portal.  

Both above are being addressed in an internal Tyk Dashboard ticket.  please wait for an official release date.

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

If you have an existing policy, simple set the `id` field in the SecurityPolicy YAML to the `_id` field in the existing SecurityPolicy.
This will allow the Operator to make the link.  

Note that the YAML becomes the source of truth and will overrwrite any changes between it and the existing Policy.

#### Example

1. You have an existing Policy
![Demo](./img/policy_migration_step1.png)

2. Stick the policy ID `5f8f3933f56e1a5ffe2cd58c` into the YAML

`my-security-policy.yaml`:
```yaml
apiVersion: tyk.tyk.io/v1alpha1
kind: SecurityPolicy
metadata:
  name: new-httpbin-policy
spec:
  _id: 5f8f3933f56e1a5ffe2cd58c
  name: My New HttpBin Policy
  state: active
  active: true
  access_rights_array:
    - name: new-httpbin-api
      namespace: default
      versions:
        - Default
```

3. And then apply this file:
```bash
$ kubectl apply -f my-security-policy.yaml
securitypolicy.tyk.tyk.io/new-httpbin-policy created
```

Now the changes in the YAML were applied to the existing Policy.  You can now manage this policy through the CRD moving forward.

This is extremely important because of indempotency.

### Idempotent

If you took a look at the migrating Policy examples, you can see that we have the ability to tell Tyk what the Policy ID is.  This is extremely important.

Imagine the use case where you have keys tied to policies, and policies tied to APIs.  Now imagine that all these resources are destroyed for whatever reason. Well, using the Tyk Operator will allow us to re-generate all our resources, with the Policy IDs being matched up.  That's because, if you explicitly state a Policy ID, it will be respected.
Alternatively, if you don't explicitly state it, it will be hard-coded for you by Base64 encoding the namespaced name of the CRD.

##### For example,
1. we have keys tied to policies tied to APIs in production.
2. Our production DB gets destroyed, all our Policies and APIs are wiped
3. The Tyk Operator can resync all the changes from our CRDs into Tyk, by explicitly defining the Policy IDs and API IDs as before.
4. This allows keys to continue to work normally as Tyk resources are generated idempotently through the Operator.
