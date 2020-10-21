## API Access

This example deploys both an API and a Policy which protects that API.

## 0. Create the `api_access.yaml` file 

Grab [this file](./api_access.yaml) or copy the following contents and save it to a file called `api_access.yaml`
```yaml
apiVersion: tyk.tyk.io/v1alpha1
kind: ApiDefinition
metadata:
  name: httpbin
spec:
  name: httpbin protected
  protocol: http
  active: true
  proxy:
    target_url: http://httpbin.org
    listen_path: /httpbin
    strip_listen_path: true
  use_standard_auth: true
  auth_configs:
    authToken:
      auth_header_name: Authorization

---
apiVersion: tyk.tyk.io/v1alpha1
kind: SecurityPolicy            # SecurityPolicy CRD
metadata:
  name: httpbin                 # Unique k8s name
spec:
  name: Httpbin Security Policy # Generic Name
  state: active                 # View securitypolicy_types for more info
  active: true                  # View securitypolicy_types for more info
  access_rights_array:          # Adding APIs to the Policy. More info just below
    - name: httpbin             # Metadata name of API
      namespace: default
      versions:
        - Default               # Mandatory, Default is created automatically
```

## 1. Deploy the protected API and the policy which protects it.

```bash
$ kubectl apply -f api_access.yaml
apidefinition.tyk.tyk.io/httpbin created
securitypolicy.tyk.tyk.io/httpbin created
```

## 2. Done!

Create a key which grants access to the API and use it against the API.

(special step, delete both the API and Policy, and then recreate them, the key should still work !) Read the idempotency section for more info.

## cleanup
Delete both the policy & httpbin CRDs:
```
$ kubectl delete tykpolicies httpbin;kubectl delete tykapis httpbin
```
