# tyk-operator installation

- [Prerequisites](#prerequisites)
- [Installing Tyk](#installing-tyk)
- [Operator Configuration](#tyk-operator-configuration)
- [Installing CRDs](#installing-crds)
- [Installing cert-manager](#installing-cert-manager)
- [Installing tyk-operator](#installing-tyk-operator)
- [Uninstall](#uninstall)

### prerequisites

Before running the operator

* A fully functioning & bootstrapped Tyk installation (OSS or Pro Licensed) needs to be installed.
* A secret in the namespace of your operator deployment telling the operator how to communicate with Tyk
* The CRDs must be registered with the Kubernetes apiserver
* cert-manager must be installed
* If you are using `pro` edition make sure in your gateway's `tyk.conf` `policies.allow_explicit_policy_id` is set to true
```json
    "policies": {
        "allow_explicit_policy_id": true
    },
```

### Installing Tyk

We shall assume you already have a deployed and bootstrapped Tyk installation. If not, head over to 
[tyk-helm-chart](https://github.com/TykTechnologies/tyk-helm-chart/), to install Tyk.

The Tyk Installation does not need to be deployed inside K8s. You may already have a fully-functioning Tyk installation.

Using Tyk Operator, you can manage APIs in any Tyk installation whether self-hosted, K8s or Tyk Cloud. As long as the 
management URL is accessible by the operator.

### tyk-operator configuration

Operator configurations are all stored in the secret `tyk-operator-conf`.

#### connecting to Tyk

tyk-operator needs to connect to a Tyk Pro deployment. And it needs to know whether it is talking to a Community
 Edition Gateway or Pro installation.

`TYK_MODE` can be `oss` or `pro`.

```bash
kubectl create namespace tyk-operator-system

kubectl create secret -n tyk-operator-system generic tyk-operator-conf \
  --from-literal "TYK_AUTH=${TYK_AUTH}" \
  --from-literal "TYK_ORG=${TYK_ORG}" \
  --from-literal "TYK_MODE=${TYK_MODE}" \
  --from-literal "TYK_URL=${TYK_URL}"
```

Examples of these values:

|         | TYK_ORG     | TYK_AUTH       | TYK_URL            | TYK_MODE |
|---------|-------------|----------------|--------------------|----------|
| Tyk Pro | User Org ID, ie "5e9d9544a1dcd60001d0ed20" | User API Key, ie "2d095c2155774fe36d77e5cbe3ac963b"   | Dashboard Base URL, ie "http://localhost:3000" | "pro"    |
| Tyk Hybrid | User Org ID | User API Key   | "https://admin.cloud.tyk.io/" | "pro"    |
| Tyk OSS |      "foo"       | Gateway secret | Gateway Base URL   | "oss"    |

And after you run the command, the values get automatically Base64 encoded:

```
k get secret/tyk-operator-conf -n tyk-operator-system -o json | jq '.data'
{
  "TYK_AUTH": "NWFhOTIyMTQwMTA0NGYxYzcwZDFjOTUwMDhkMzllZGE=",
  "TYK_MODE": "cHJv",
  "TYK_ORG": "NWY5MmQ5YWQyZGFiMWMwMDAxM2M3NDlm",
  "TYK_URL": "aHR0cDovL2Rhc2hib2FyZC50eWtwcm8tY29udHJvbC1wbGFuZS5zdmMuY2x1c3Rlci5sb2NhbDozMDAw"
}
```

#### Watching Namespaces

Tyk Operator installs with cluster permissions, however you can optionally control which namespaces it watches by
 by setting the `WATCH_NAMESPACE` environment variable.
 
`WATCH_NAMESPACE` can be omitted entirely, or a comma separated list of k8s namespaces.

- `WATCH_NAMESPACE=""` will watch for resources across the entire cluster.
- `WATCH_NAMESPACE="foo"` will watch for resources in the `foo` namespace.
- `WATCH_NAMESPACE="foo,bar"` will watch for resources in the `foo` and `bar` namespace.

example:

```
kubectl create secret -n tyk-operator-system generic tyk-operator-conf \
  --from-literal "TYK_AUTH=${TYK_AUTH}" \
  --from-literal "TYK_ORG=${TYK_ORG}" \
  --from-literal "TYK_MODE=${TYK_MODE}" \
  --from-literal "TYK_URL=${TYK_URL}" \
  --from-literal "WATCH_NAMESPACE=foo,bar"
```

#### Watching custom ingress class

The value of the `kubernetes.io/ingress.class` annotation that identifies Ingress objects to be processed.

Tyk Operator by default looks for the value `tyk` and will ignore all other ingress classes. If you wish to override this default behaviour,
 you may do so by setting the environment variable `WATCH_INGRESS_CLASS` in the operator manager deployment.

example:

```
kubectl create secret -n tyk-operator-system generic tyk-operator-conf \
  --from-literal "TYK_AUTH=${TYK_AUTH}" \
  --from-literal "TYK_ORG=${TYK_ORG}" \
  --from-literal "TYK_MODE=${TYK_MODE}" \
  --from-literal "TYK_URL=${TYK_URL}" \
  --from-literal "WATCH_INGRESS_CLASS=foo"
```

### Installing CRDs

Installing CRDs is as simple as checking out this repo & running `kubectl apply`.

```bash
kubectl apply -f ./helm/crds
customresourcedefinition.apiextensions.k8s.io/apidefinitions.tyk.tyk.io configured
customresourcedefinition.apiextensions.k8s.io/securitypolicies.tyk.tyk.io configured
customresourcedefinition.apiextensions.k8s.io/webhooks.tyk.tyk.io configured
```

### Installing cert-manager

If you don't have cert-manager installed: Quick install

```bash
kubectl apply --validate=false -f https://github.com/jetstack/cert-manager/releases/download/v1.0.3/cert-manager.yaml
```

[cert-manager documentation](https://cert-manager.io/docs/installation/kubernetes/)

Please wait for cert-manager to become available.

```
k get all -n cert-manager
NAME                                           READY   STATUS    RESTARTS   AGE
pod/cert-manager-79c5f9946-d5hfv               1/1     Running   0          14s
pod/cert-manager-cainjector-76c9d55b6f-qmpmv   1/1     Running   0          14s
pod/cert-manager-webhook-6d4c5c44bb-q9n9k      0/1     Running   0          14s

NAME                           TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
service/cert-manager           ClusterIP   10.245.61.87    <none>        9402/TCP   15s
service/cert-manager-webhook   ClusterIP   10.245.96.198   <none>        443/TCP    15s

NAME                                      READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/cert-manager              1/1     1            1           14s
deployment.apps/cert-manager-cainjector   1/1     1            1           14s
deployment.apps/cert-manager-webhook      0/1     1            0           14s

NAME                                                 DESIRED   CURRENT   READY   AGE
replicaset.apps/cert-manager-79c5f9946               1         1         1       14s
replicaset.apps/cert-manager-cainjector-76c9d55b6f   1         1         1       14s
replicaset.apps/cert-manager-webhook-6d4c5c44bb      1         1         0       14s
```

## Installing tyk-operator

Run the following to deploy tyk-operator. 

```
$ helm install foo ./helm -n tyk-operator-system

NAME: foo
LAST DEPLOYED: Tue Nov 10 18:38:32 2020
NAMESPACE: tyk-operator-system
STATUS: deployed
REVISION: 1
TEST SUITE: None
NOTES:
You have deployed the tyk-operator! See https://github.com/TykTechnologies/tyk-operator for more information.
```

## Uninstall

Did we do something wrong? Create a [GH issue](https://github.com/TykTechnologies/tyk-operator/issues/new) / ticket and 
maybe we can try to improve your experience, or that of others. 

```
helm delete foo
```
