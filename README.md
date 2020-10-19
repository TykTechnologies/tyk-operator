# Tyk Operator

Tyk Operator contains various controllers which watch resources inside Kubernetes and reconciles
a Tyk deployment with the desired state as set inside K8s. 

Custom Tyk Objects are available as [CRDs](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/):

- [API Definitions](./docs/api_definitions.md)
- [WebHooks](./docs/webhooks.md)
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
