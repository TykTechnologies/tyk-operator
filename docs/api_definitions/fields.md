This contains some comments /explanations about fields that are in the CRD on what
they are used for.

If you need to link anything into tables that needs additional explanation and is 
a field in the APIDefinition resource it goes here.


# active

This determines whether the api definition will be loaded on the tyk gateway or not.

Example

```yaml
apiVersion: tyk.tyk.io/v1alpha1
kind: ApiDefinition
metadata:
  name: httpbin
spec:
  name: httpbin
  use_keyless: true
  protocol: http
  active: false
  proxy:
    target_url: http://httpbin.org
    listen_path: /httpbin
    strip_listen_path: true
```

A `httpbin` api will be created and stored in tyk pro but it will not be loaded on the gateways , calling the service will result in `404`

