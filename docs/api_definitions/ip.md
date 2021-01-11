# whitelisting

To enable ip whitelisting set `enable_ip_whitelisting` to true and use `allowed_ips`
to list all allowed ip's

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
  enable_ip_whitelisting: true
  allowed_ips:
    - 127.0.0.2
  proxy:
    target_url: http://httpbin.default.svc:8000
    listen_path: /httpbin
    strip_listen_path: true
```

This will only allow `127.0.0.2` ip, any other ip will we forbidden


# blacklisting

To enable blacklisting set `enable_ip_blacklisting` to true and use `blacklisted_ips` to
list all ip addresses to block.

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
  enable_ip_blacklisting: true
  blacklisted_ips:
    - 127.0.0.2
  proxy:
    target_url: http://httpbin.default.svc:8000
    listen_path: /httpbin
    strip_listen_path: true
```

This will return `403` for `127.0.0.2`