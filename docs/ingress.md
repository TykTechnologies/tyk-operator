# Ingress

Tyk Operator offers an Ingress Controller, which dynamically manages ApiDefinition resources as per the ingress spec.
 Tyk Gateway can be configured as a drop-in replacement for any standard Kubernetes Ingress.

Most Ingress Controllers heavily rely on annotations to configure the ingress gateway. With Tyk Operator, our Ingress 
 Controller prefers to reference a strongly typed custom resource template.

## Motivation

The standard Ingress resource is very basic and does not natively support many advanced capabilities that are required
 for real-life use-cases. Despite this, the community have built tooling, capabilities, dependencies, and technical debt 
 on top of the ingress resource. These all rely on abusing the metadata annotations, and this practice is apparent 
 with the standard [Kubernetes NginX Ingress resource](https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/annotations/#annotations).

In order to decouple, gain all the benefits of Kubernetes, offer a Native & consistent API, we introduced the Tyk
 ApiDefinition custom resource, which brings Full-Lifecycle API Management capabilities to Kubernetes. A side effect of
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
    - ingress: true
spec:
  name: foo
  protocol: http
  use_standard_auth: true
  proxy:
    target_url: http://example.com
```

Pay particular attention to the ApiDefinition metadata. This specifies that we have an ApiDefinition object with the
 label `ingress: true`.

When applying this manifest, the ApiDefinition reconciler's predicate filter will skip reconciliation. This will allow
 the ApiDefinition to be stored inside Kubernetes as a resource, but will not reconcile the ApiDefinition inside Tyk.

All mandatory fields inside the ApiDefinition spec are still mandatory, but can be replaced with placeholders as they
 will be overwritten by the Ingress reconciler.

## Sample Ingress resource

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: httpbin-ingress
  annotations:
    cert-manager.io/issuer: my-issuer
    kubernetes.io/ingress.class: tyk # <--------------- REFERENCES TYK INGRESS CONTROLLER 
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
```

The above ingress resource will create an ApiDefinition inside Tyk's Gateway. The ApiDefinition will offer path-based
 routing listening on /httpbin. Because the referenced template is `myapideftemplate`, the IngressReconciler will
 retrieve the `myapideftemplate` resource and determine that the ApiDefinition object it creates needs to have standard
 auth enabled.
