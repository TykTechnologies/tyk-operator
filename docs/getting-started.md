# Getting Started

- [Install](#install)
- [Create an API](#create-an-api)

Tyk Operator extends Kubernetes API with Custom Resources. API Definitions, Security Policies, Authentication, 
Authorization, Rate Limits, and other Tyk features can be managed just like other native Kubernetes objects, leveraging 
all the features of Kubernetes like kubectl, security, API services, RBAC. You can use Tyk Customer Resources to manage 
and secure REST, TCP, gRPC, GraphQL, and SOAP services. You can even take an existing REST based API and define a GraphQL 
schema to tell the GraphQL execution engine how to map those REST responses into the GraphQL schema, without a line of code.

This tutorial walks through creating an API using the Tyk Operator.
You can find all available example ApiDefinition files in the [samples](../config/samples) directory.

## Install

Please follow the [installation documentation](./installation/installation.md) to set up Tyk Operator.

In order to complete this tutorial, you need to have a fully functioning & bootstrapped Tyk installation (CE or Pro Licensed)
that is accessible from the Kubernetes cluster that will host the Tyk Operator, as explained in the [installation documentation](./installation/installation.md).

## Create an API

Creating an API takes the same approach whether you are using Tyk CE or Tyk Pro. In the first place, an `ApiDefinition` 
file must be created in the YAML format. Then, Tyk Operator handles the creation of the API.

1. Define your API in the YAML format,
2. Create a Kubernetes resource based on this YAML file,
3. Tyk Operator handles the creation of your API.

We are going to create an ApiDefinition described in the [httpbin.yaml](../config/samples/httpbin.yaml) file, as follows: 

```bash
kubectl apply -f config/samples/httpbin.yaml
```

Or,

```bash
cat <<EOF | kubectl apply -f -
apiVersion: tyk.tyk.io/v1alpha1
kind: ApiDefinition
metadata:
    name: httpbin
spec:
    name: httpbin
    use_keyless: true
    protocol: http
    active: true
    proxy:
        target_url: http://httpbin.org
        listen_path: /httpbin
        strip_listen_path: true
EOF
```

Let's walk through the ApiDefinition that we created. We have an ApiDefinition called `httpbin`, as specified in `spec.name` 
field, that listens `/httpbin` and proxies requests to http://httpbin.org, as specified under `spec.proxy` field. Now, any 
requests coming to the `/httpbin` endpoint will be proxied to the target URL that we defined in `spec.proxy.target_url`, 
which is http://httpbin.org in our example.

To find out about available ApiDefinition objects in your cluster:
```bash
$ kubectl get tykapis
NAME      DOMAIN   LISTENPATH   PROXY.TARGETURL      ENABLED
httpbin            /httpbin     http://httpbin.org   true
```

We can see that our ApiDefinition has been created. Now let's verify that our API is working as expected.

**NOTE**: The verification step may vary based on your environment, such as the type of your Tyk installation and Kubernetes cluster.
- If you are using local Kubernetes cluster such as [KinD](https://kind.sigs.k8s.io/) and [Minikube](https://minikube.sigs.k8s.io/docs/start/), 
you can do port-forwarding to access objects within the cluster.
- If you are using Kubernetes clusters provided by cloud providers, you need to configure your cluster to make it accessible.
  
For the scope of this example, we are using a local Kubernetes cluster. Port-forwarding details can be found in the 
official [Kubernetes documentation](https://kubernetes.io/docs/tasks/access-application-cluster/port-forward-access-application-cluster/).

### Tyk CE

<details><summary>Example port forwarding for Tyk CE</summary>
<p>

If you have installed Tyk CE in the `<TYK_CE_NAMESPACE>` namespace, you will have the following services:
```bash
kubectl get svc -n <TYK_CE_NAMESPACE> 
NAME                              TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
gateway-svc-tyk-ce-tyk-headless   ClusterIP   10.96.38.138    <none>        8080/TCP   22m
redis                             ClusterIP   10.96.254.227   <none>        6379/TCP   22m
```

In order to access Tyk Gateway, you can use the following port-forwarding command:
```bash
kubectl port-forward service/gateway-svc-tyk-ce-tyk-headless -n <TYK_CE_NAMESPACE> 8080:8080
```

The Tyk Gateway is accessible from your local cluster's 8080 port (e.g., `localhost:8080`).

</p>
</details>

Since Tyk CE does not come with the Dashboard, you can list APIs using [Tyk Gateway API](https://tyk.io/docs/tyk-gateway-api/).

```bash
$ curl -H "x-tyk-authorization: foo" localhost:8080/tyk/apis/
```

Let's make a request to verify that our API is working.

```bash
$ curl -i localhost:8080/httpbin/get
{
  "args": {},
  "headers": {
    "Accept": "*/*",
    "Accept-Encoding": "gzip",
    "Host": "httpbin.org",
    "User-Agent": "curl/7.77.0",
    "X-Amzn-Trace-Id": "Root=1-62161e8c-2a1ece436633f2e42129be2a"
  },
  "origin": "127.0.0.1, 176.88.45.17",
  "url": "http://httpbin.org/get"
}
```

### Tyk PRO

<details><summary>Example port forwarding for Tyk Pro</summary>
<p>

If you have installed Tyk Pro in the `<TYK_PRO_NAMESPACE>` namespace, you will have the following services:
```bash
kubectl get svc -n <TYK_PRO_NAMESPACE>
NAME                    TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)          AGE
dashboard-svc-tyk-pro   NodePort    10.96.152.180   <none>        3000:30357/TCP   2d17h
gateway-svc-tyk-pro     NodePort    10.96.228.133   <none>        8080:31516/TCP   2d17h
mongo                   ClusterIP   10.96.12.192    <none>        27017/TCP        2d17h
redis                   ClusterIP   10.96.66.91     <none>        6379/TCP         2d17h
```

In order to access the Dashboard, you can use the following port-forwarding command:
```bash
kubectl port-forward service/dashboard-svc-tyk-pro 3000:3000 -n TYK_PRO_NAMESPACE
```

The Dashboard is accessible from your local cluster's 3000 port (e.g., `localhost:3000`).

</p>
</details>

If you head over to the Dashboard, we can see that an ApiDefinition called `httpbin` is created. 

![tyk-dashboard](./img/getting-started-dashboard_1.png)

Let's make a request to verify that our API is working.

```bash
$ curl -i localhost:8080/httpbin/get
{
  "args": {},
  "headers": {
    "Accept": "*/*",
    "Accept-Encoding": "gzip",
    "Host": "httpbin.org",
    "User-Agent": "curl/7.77.0",
    "X-Amzn-Trace-Id": "Root=1-62161e8c-2a1ece436633f2e42129be2a"
  },
  "origin": "127.0.0.1, 176.88.45.17",
  "url": "http://httpbin.org/get"
}
```
