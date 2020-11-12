# tyk-operator installation

## prerequisites

Before running the operator

* A fully functioning & bootstrapped Tyk installation (OSS or Pro Licensed) needs to be installed.
* A secret in the namespace of your operator deployment telling the operator how to communicate with Tyk
* The CRDs must be registered with the Kubernetes apiserver
* cert-manager must be installed

### Installing Tyk

We shall assume you already have a deployed and bootstrapped Tyk installation. If not, head over to 
[tyk-helm-chart](https://github.com/TykTechnologies/tyk-helm-chart/), to install Tyk.

{{box op="start" cssClass="boxed tipBox"}}
**Tip!**

The Tyk Installation does not need to be deployed inside K8s. You may already have a fully-functioning Tyk installation.

Using Tyk Operator, you can manage APIs in any Tyk installation whether self-hosted, K8s or Tyk Cloud. As long as the 
management URL is accessible by the operator.
{{box op="end"}}

### Configuring tyk-operator secrets

By default, the operator deploys to the `tyk-operator-system` namespace. This example sets the appropriate secrets 
tyk-operator needs to connect to a Tyk Pro deployment.

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


And after you run the command
```
k get secret/tyk-operator-conf -n tyk-operator-system -o json | jq '.data'
{
  "TYK_AUTH": "NWFhOTIyMTQwMTA0NGYxYzcwZDFjOTUwMDhkMzllZGE=",
  "TYK_MODE": "cHJv",
  "TYK_ORG": "NWY5MmQ5YWQyZGFiMWMwMDAxM2M3NDlm",
  "TYK_URL": "aHR0cDovL2Rhc2hib2FyZC50eWtwcm8tY29udHJvbC1wbGFuZS5zdmMuY2x1c3Rlci5sb2NhbDozMDAw"
}
```

### Installing CRDs

Installing CRDs is as simple as checking out this repo & running `make install`.

```bash
make install
/Users/ahmet/go/bin/controller-gen "crd:trivialVersions=true,crdVersions=v1" rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases
/Users/ahmet/go/bin/kustomize build config/crd | kubectl apply -f -
customresourcedefinition.apiextensions.k8s.io/apidefinitions.tyk.tyk.io configured
customresourcedefinition.apiextensions.k8s.io/organizations.tyk.tyk.io configured
customresourcedefinition.apiextensions.k8s.io/securitypolicies.tyk.tyk.io configured
customresourcedefinition.apiextensions.k8s.io/webhooks.tyk.tyk.io configured
```

### Installing cert-manager

Quick install

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

If you wish to change the namespace that tyk-operator is deployed from the default `tyk-operator-system`:

```bash
cd config/default/ && kustomize edit set namespace "yournamespace" && cd ../..
```

Run the following to deploy tyk-operator. This will also install the RBAC manifests from config/rbac. 

```bash
make deploy IMG=tykio/tyk-operator:latest
```

```bash
k get all -n tyk-operator-system
NAME                                                   READY   STATUS    RESTARTS   AGE
pod/tyk-operator-controller-manager-6f554ffb59-244tj   2/2     Running   1          2m14s

NAME                                                      TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)    AGE
service/tyk-operator-controller-manager-metrics-service   ClusterIP   10.245.198.147   <none>        8443/TCP   2m22s
service/tyk-operator-webhook-service                      ClusterIP   10.245.143.252   <none>        443/TCP    2m21s

NAME                                              READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/tyk-operator-controller-manager   1/1     1            1           2m18s

NAME                                                         DESIRED   CURRENT   READY   AGE
replicaset.apps/tyk-operator-controller-manager-6f554ffb59   1         1         1       2m17s
```

## Uninstall

Did we do something wrong? Create a [GH issue](https://github.com/TykTechnologies/tyk-operator/issues/new) / ticket and 
maybe we can try to improve your experience, or that of others. 

```
make uninstall
```
