# GraphQL Proxy
GraphQL Proxy Only is just a GraphQL API with a single datasource and read-only schema.

To create GraphQL Proxy Only API you need to set following fields in your APIDefinition spec:

```yaml
 graphql:
    enabled: true
    execution_mode: proxyOnly
    schema: { Graphql Schema }
```

You can create sample GraphQL API by applying samples manifests from the repository.
```bash
kubectl apply -f config/samples/graphql_proxy/trevorblades_graphql_proxy.yaml
```

```bash
kubectl get tykapis
NAME                DOMAIN   LISTENPATH   PROXY.TARGETURL                       ENABLED
trevorblades                 /trevorblades   https://countries.trevorblades.com    true
```

You have successfully created API for [Countries GraphqQL API](https://github.com/trevorblades/countries).



# Protected Upstream

If your upstream GraphQL server is protected, you can save authorization headers to be passed to upstream inside APIDefintion.

```yaml
graphql:
    proxy:
        auth_headers:
            Auth-Header1: "foo"
            Auth-Header2: "bar"
```

Now you don't need to provide upstream Authorization headers while making request to this API. Headers inside `auth_headers` field will be passed to upstream server on each request.

