## Rate Limit

This example deploys both an API and a Policy which protects that API.

The Policy has meta data and tags being inserted.

## 1. Deploy a protected API and the policy which protects it.

```curl
$ kubectl apply -f metadata_tags.yaml
apidefinition.tyk.tyk.io/httpbin created
securitypolicy.tyk.tyk.io/httpbin created
```

Here's the section we care about in the SecurityPolicy yaml:
```
  tags:
    - Hello
    - World

  meta_data:
    key: value
    hello: world
```

## 2. Done!

Create a key which grants access to the API and use it against the API.

This key now inherits the tags and the meta data from the policy.

![img](./metadata_tags.png)

Note: Ignore the `SecurityPolicy-default/httpbin` tag, that is automatically added by the operator and required.

## cleanup
Delete both the policy & httpbin CRDs:
```
$ kubectl delete tykpolicies httpbin;kubectl delete tykapis httpbin
```