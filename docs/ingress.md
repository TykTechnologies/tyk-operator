# Ingress

Tyk Operator offers an Ingress Controller, which dynamically manages ApiDefinition resources as per the ingress spec.
 Tyk Gateway can be configured as a drop-in replacement for a standard Kubernetes Ingress.

Most Ingress Controllers heavily rely on annotations to configure the ingress gateway. With Tyk Operator, our Ingress 
 Controller prefers to reference a strongly typed custom resource template.
 
## Ingress Class

The value of the `kubernetes.io/ingress.class` annotation that identifies Ingress objects to be processed.

Tyk Operator by default looks for the value `tyk` and will ignore all other ingress classes. If you wish to override this default behaviour,
 you may do so by setting the environment variable `WATCH_INGRESS_CLASS` in the operator manager deployment. See https://github.com/TykTechnologies/tyk-operator/blob/master/docs/installation/installation.md for further info.

## Ingress Path Types

Each path in an Ingress must have its own particular path type. Kubernetes offers three types of path types: `ImplementationSpecific`, `Exact`, and `Prefix`. Currently, not all path types are supported. The below table shows the unsupported path types for [Sample HTTP Ingress Resource](#sample-http-ingress-resource) based on the examples in the [Kubernetes Ingress documentation](https://kubernetes.io/docs/concepts/services-networking/ingress/#examples).

| Kind   | Path(s)   | Request path(s) | Expected to match?               | Works as Expected                       |
|--------|-----------|-----------------|----------------------------------|-----------------------------------------|
| Exact  | /foo      | /foo/           | No                               | No.                                     |
| Prefix | /foo/     | /foo, /foo/     | Yes                              | No, /foo/ matches, /foo does not match. | 
| Prefix | /aaa/bb   | /aaa/bbb        | No                               | No, the request forwarded to service.   |
| Prefix | /aaa/bbb/ | /aaa/bbb        | Yes, ignores trailing slash      | No, /aaa/bbb does not match.            | 
| Prefix | /aaa/bbb  | /aaa/bbbxyz     | No, does not match string prefix | No, the request forwarded to service.   |

Please bear in mind that if `proxy.strip_listen_path` is set to true on API Definition, Tyk strips the listen-path (for example, the listen-path for the Ingress under [Sample HTTP Ingress Resource](#sample-http-ingress-resource) is /httpbin) with an empty string.

The following table shows an example of path matching if the listen-path is set to `/httpbin` or `/httpbin/`.

| Kind                   | Path(s)   | Request path(s)           | Matches?                                              |
|------------------------|-----------|---------------------------|-------------------------------------------------------|
| Exact                  | /httpbin  | /httpbin, /httpbin/       | Yes. The request forwarded as `/` to your service.    |
| Prefix                 | /httpbin  | /httpbin, /httpbin/       | Yes. The request forwarded as `/` to your service.    | 
| ImplementationSpecific | /httpbin  | /httpbin, /httpbin/       | Yes. The request forwarded as `/` to your service.    |
| Exact                  | /httpbin  | /httpbinget, /httpbin/get | Yes. The request forwarded as `/get` to your service. |
| Prefix                 | /httpbin  | /httpbinget, /httpbin/get | Yes. The request forwarded as `/get` to your service. | 
| ImplementationSpecific | /httpbin  | /httpbinget, /httpbin/get | Yes. The request forwarded as `/get` to your service. |
| Exact                  | /httpbin/ | /httpbin/,  /httpbin/get  | Yes. The request forwarded as `/get` to your service. |
| Prefix                 | /httpbin/ | /httpbin/,  /httpbin/get  | Yes. The request forwarded as `/get` to your service. | 
| ImplementationSpecific | /httpbin/ | /httpbin/,  /httpbin/get  | Yes. The request forwarded as `/get` to your service. |
| Exact                  | /httpbin/ | /httpbin                  | No. Ingress cannot find referenced service.           |
| Prefix                 | /httpbin/ | /httpbin                  | No. Ingress cannot find referenced service.           |  
| ImplementationSpecific | /httpbin/ | /httpbin                  | No. Ingress cannot find referenced service.           | 

## Quickstart / Samples

* [HTTP Host-Based](./../config/samples/ingress/ingress-httpbin/)
* [HTTP Path Based](./../config/samples/ingress/ingress-httpbin/)
* [HTTP Host and Path](./../config/samples/ingress/ingress-httpbin/)
* [HTTPS with Cert-Manager Integration](./../config/samples/ingress/ingress-tls)
* [Multiple Ingress Resources](./../config/samples/ingress/ingress-multi)
* [Wildcard Hosts](./../config/samples/ingress/ingress-wildcard-host)
* [Istio Ingress Gateway](./../config/samples/ingress/istio-ingress-bookinfo)

## Motivation

The standard Ingress resource is very basic and does not natively support many advanced capabilities that are required
 for API Management use-cases. Despite this, the community have built tooling, capabilities, dependencies,
 on top of the ingress resource. These all rely on abuse of the metadata annotations. This practice is apparent 
 with the standard [Kubernetes NginX Ingress resource](https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/annotations/#annotations).

In order to decouple, gain all the benefits of Kubernetes, offer a Native & consistent API, we introduced the Tyk
 ApiDefinition custom resource, which brings Full-Lifecycle API Management capabilities to Kubernetes Ingress. A side effect of
 this meant a potential trade-off between offering a K8s native experience & integrating with Ingress which would 
 facilitate clean integration with 3rd party tooling built on-top of and dependent on the Ingress Resource.

As a compromise & attempt to propose an alternative & more scalable solution, we have introduced the concept of the 
 ingress template ApiDefinition resource. The Template ApiDefinition resource offers a means to extend the capabilities 
 of the standard Ingress Resource, by merging features of the ingress specification with that of the template.

## Sample Template ApiDefinition resource

```yaml
apiVersion: tyk.tyk.io/v1alpha1
kind: ApiDefinition
metadata:
 name: myapideftemplate
 labels:
  template: "true"
spec:
 name: foo
 protocol: http
 use_keyless: true
 proxy:
  target_url: http://example.com
```

Pay particular attention to the ApiDefinition metadata. This specifies that we have an ApiDefinition object with the
 label `template: true`.

When applying this manifest, the ApiDefinition controller will skip reconciliation. This will allow
 the ApiDefinition to be stored inside Kubernetes as a resource, but will not reconcile the ApiDefinition inside Tyk.

All mandatory fields inside the ApiDefinition spec are still mandatory, but can be replaced with placeholders as they
 will be overwritten by the Ingress reconciler.

## Sample HTTP Ingress resource

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
 name: httpbin-ingress
 annotations:
  kubernetes.io/ingress.class: tyk # <----------------- REFERENCES TYK INGRESS CONTROLLER
  tyk.io/template: myapideftemplate # <---------------- REFERENCE TO APIDEFINITION IN SAME NAMESPACE
spec:
 rules:
  - http:
     paths:
      - path: /httpbin
        pathType: Prefix
        backend:
         service:
          name: httpbin
          port:
           number: 8000
```

Tyk Ingress Controller will create APIs in Tyk for each path defined for a specific rule in Ingress resource. Each API 
created inside Tyk will follow a special naming convention as follows:
```
<ingress_namespace>-<ingress_name>-<hash(Host + Path)>
```


The above ingress resource will create an ApiDefinition called `default-httpbin-ingress-78acd160d` inside Tyk's Gateway.
ApiDefinition's name comes from:
- `default`: The namespace of this Ingress resource,
- `httpbin-ingress`: The name of this Ingress resource,
- `78acd160d`: Short hash (first 9 characters) of Host (`""`) and Path (`/httpbin`). The hash algorithm is SHA256.

The ApiDefinition will offer path-based routing listening on `/httpbin`. Because the referenced template is `myapideftemplate`, the IngressReconciler will retrieve the `myapideftemplate` resource and determine that the ApiDefinition object it creates needs to have standard auth enabled.

## Sample HTTPS Ingress resource

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: httpbin-ingress-tls
  annotations:
    kubernetes.io/ingress.class: tyk # <----------------- REFERENCES TYK INGRESS CONTROLLER
    tyk.io/template: myapideftemplate # <---------------- REFERENCE TO APIDEFINITION IN SAME NAMESPACE
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
    acme.cert-manager.io/http01-edit-in-place: "true"
spec:
  tls:
    - hosts:
        - myingress.do.poc.tyk.technology
      secretName: httpbin-ingress-tls
  rules:
    - host: myingress.do.poc.tyk.technology
      http:
        paths:
          - path: /httpbin
            pathType: Prefix
            backend:
              service:
                name: httpbin
                port:
                  number: 8000
```

Assuming you already have a `letsencrypt-prod` cluster issuer, it is possible to automatically provision TLS certificates
 issued by LetsEncrypt.

```yaml
metadata:
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
    acme.cert-manager.io/http01-edit-in-place: "true"
```

Tyk ingress controller can then handle the acme challenge where cert-manager edits the ingress resource.
