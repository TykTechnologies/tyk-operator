## Rate Limit

This example deploys both an API and a Policy which protects that API.

The Policy has rate limiting, quotas, and throttling turned on.

## 1. Deploy a protected API and the policy which protects it.

```curl
$ kubectl apply -f ratelimit.yaml
apidefinition.tyk.tyk.io/httpbin created
securitypolicy.tyk.tyk.io/httpbin created
```

## 2. Done!

Create a key which grants access to the API and use it against the API.

This key now inherits the rate limit values from the poolicy.


## cleanup
Delete both the policy & httpbin CRDs:
```
$ kubectl delete tykpolicies httpbin;kubectl delete tykapis httpbin
```