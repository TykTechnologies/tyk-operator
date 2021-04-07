# Looping 

You can target other ApiDefinition resources from an api defintion like what tyk
documentation describes at [Looping](https://tyk.io/docs/advanced-configuration/transform-traffic/looping/)

In tyk you target api's by `tyk://<API_ID>/<path>` scheme.
this requires prior knowledge of the API_ID. Since the operator treats api's as objects and it manages them including assigning `API_ID`, we introduce a typed api to refer to  other apis.

The operator will automatically generates the correct url before sending it to the gateway.

[Full sample file](../../config/samples/06-looping/dynamic_auth.yaml)

# URL Rewriting

A `rewrite_to_internal` field is used to target other api resources.

```yaml
                                      rewrite_to_internal:
                                        description: RewriteToInternal defines options that constructs a url that refers to an api that is loaded into the gateway.
                                        properties:
                                          path:
                                            description: "Path path on target , this does not include query parameters. \texample /myendpoint"
                                            type: string
                                          query:
                                            description: "Query url query string to add to target \texample check_limits=true"
                                            type: string
                                          target:
                                            description: API a namespaced/name to the api definition resource that you are targetting
                                            properties:
                                              name:
                                                type: string
                                              namespace:
                                                type: string
                                            required:
                                            - name
                                            - namespace
```

for instance 

```yaml
          url_rewrites:
            - path: "/{id}"
              match_pattern: "/basic/(.*)"
              method: GET
              rewrite_to_internal:
                target:
                  name: proxy-api
                  namespace: default
                path: proxy/$1
```

This api is targeting an ApiDefinition resource `proxy-api` in `default` namespace

# URL Rewriting triggers

A `rewrite_to_internal` used to target other api resources in `triggers`.
For example

```yaml
              triggers:
                - "on": "all"
                  options:
                    header_matches:
                      "Authorization":
                        match_rx: "^Basic"
                  rewrite_to_internal:
                    target:
                      name: basic-auth-internal
                      namespace: default
                    path: "basic/$2"
```


# Proxy to internal apis

A `target_internal` field on `proxy` object is used to target other api resources. This field properties are the same as the ones described for `rewrite_to_internal`.
