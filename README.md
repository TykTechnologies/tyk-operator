# Tyk Operator

The Tyk Operator is the [operator-sdk](https://github.com/operator-framework/operator-sdk) pattern for managing your Tyk 
environment via k8s native tooling.

Tyk objects available as [CRDs](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/):
- [API definitions](./docs/api_definitions.md)
- [Security Policies](./docs/policies.md)

![Demo](./docs/img/demo.svg)

## Sample Configurations

### HTTP Proxy

```yaml
apiVersion: tyk.tyk.io/v1alpha1
kind: ApiDefinition
metadata:
  name: httpbin
spec:
  name: httpbin
  use_keyless: true
  protocol: http
  active: true
  org_id: acme.com
  proxy:
    target_url: http://httpbin.org
    listen_path: /httpbin
    strip_listen_path: true
```

### TCP Proxy

```yaml
apiVersion: tyk.tyk.io/v1alpha1
kind: ApiDefinition
metadata:
  name: redis-tcp
spec:
  name: redis-tcp
  active: true
  protocol: tcp
  listen_port: 6380
  proxy:
    target_url: tcp://localhost:6379
```

## Dev

[Development Environment](./docs/development.md)
