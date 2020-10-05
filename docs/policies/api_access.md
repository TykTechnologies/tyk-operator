## API Access

This example deploys both an API and a Policy which protects that API.

## 1. Deploy a protected API and the policy which protects it.

```curl
$ kubectl apply -f api_access.yaml
apidefinition.tyk.tyk.io/httpbin created
securitypolicy.tyk.tyk.io/httpbin created
```

## 2. Deploy the [policy](./api_access.md) which protects it. 

```
$ kubectl apply -f httpbin_protected_policy.yaml
apidefinition.tyk.tyk.io/httpbin created
```

## 3. Done!

Create a key which grants access to the API and use it against the API.
