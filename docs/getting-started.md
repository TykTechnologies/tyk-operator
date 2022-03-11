# Getting Started

| Features| CE | PRO
|-------|-----| ------
|- [Install](#install)| ✅ | ✅
|- [Create an API](#create-an-api)| ✅ | ✅
|- [Access an API](#access-an-api)| ✅ | ✅
|- [Secure an API](#secure-an-api)| ✅ | ✅
|- [Publish an API](#publish-an-api-to-the-tyk-portal)| ❌ | ✅

Tyk Operator extends Kubernetes API with Custom Resources. API Definitions, Security Policies, Authentication, 
Authorization, Rate Limits, and other Tyk features can be managed just like other native Kubernetes objects, leveraging 
all the features of Kubernetes like kubectl, security, API services, RBAC. You can use Tyk Custom Resources to manage 
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
Custom Resource should be created in the YAML format. Then, Tyk Operator handles the creation of the API.

1. Define APIDefinition resource in the YAML format,
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
you can do port-forwarding to access resources/services within the cluster.
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
$ curl -H "x-tyk-authorization: {your-secret}" localhost:8080/tyk/apis/
```

> Your Tyk Gateway API secret is stored in your `tyk.conf` file, the property is called `secret`, you will need to use this 
as a header called `x-tyk-authorization` to make calls to the Gateway API.

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

### Kubernetes Service as an Upstream Target

Tyk Operator allows accessing your Kubernetes service as an upstream proxy target.
Thus, you can set the `proxy.target_url` as a Kubernetes Service following [DNS for Services and Pods guideline](https://kubernetes.io/docs/concepts/services-networking/dns-pod-service/), 
so that the requests will be proxied to your service.

> In general, Kubernetes Services have `<service-name>.<namespace-name>.svc.cluster.local` DNS entry once they are created. 
For example, if you have a service called `httpbin` in `default` namespace, you can contact `httpbin` service with 
`httpbin.default.svc` DNS record in the cluster, instead of IP addresses. 
Please visit the official [Kubernetes documentation](https://kubernetes.io/docs/concepts/services-networking/dns-pod-service/) for more details.

Suppose we want to create a Deployment of [`httpbin`](https://hub.docker.com/r/kennethreitz/httpbin/) service
using [`ci/upstreams/httpbin.yaml`](../ci/upstreams/httpbin.yaml) file. We are going to expose the application through port `8000` as described under
the Service [specification](https://github.com/TykTechnologies/tyk-operator/blob/master/ci/upstreams/httpbin.yaml#L10).

First, let's create Service and Deployment by either applying the manifest defined in our repository

```bash
kubectl apply -f ci/upstreams/httpbin.yaml
```

Or,

```bash
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Service
metadata:
  name: httpbin
  labels:
    app: httpbin
spec:
  ports:
    - name: http
      port: 8000
      targetPort: 80
  selector:
    app: httpbin
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: httpbin
spec:
  replicas: 1
  selector:
    matchLabels:
      app: httpbin
      version: v1
  template:
    metadata:
      labels:
        app: httpbin
        version: v1
    spec:
      containers:
        - image: docker.io/kennethreitz/httpbin
          imagePullPolicy: IfNotPresent
          name: httpbin
          ports:
            - containerPort: 80
EOF
```

> Please wait awhile until all pods reach READY `1/1` and STATUS `Running` state.

Once the pod is ready, we can update our `httpbin` API's `target_url` field to proxy our requests to the Service that we've created above.

> You can check all Services in the `<ns>` namespace as follows;
```bash
kubectl get service -n <ns>
```

Let's update our `httpbin`:

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
    target_url: http://httpbin.default.svc:8000
    listen_path: /httpbin
    strip_listen_path: true
EOF
```

Please pay attention to the value of the `spec.proxy.target_url` field. 
It is set to `http://httpbin.default.svc:8000` by following the convention described above (`<service_name>.<namespace>.svc:<service_port>`).

Now, if you send your request to the `/httpbin` endpoint, the request will be proxied to the `httpbin Service`.

```bash
curl -sS http://localhost:8080/httpbin/headers
```
```json
{
  "headers": {
    "Accept": "*/*", 
    "Accept-Encoding": "gzip", 
    "Host": "httpbin.default.svc:8000", 
    "User-Agent": "curl/7.68.0"
  }
}
```

As you can see from the response, the host that our request should be proxied to is `httpbin.default.svc:8000`.

## Access an API

Our `httpbin` API is keyless, as you might already have realized. If you check the APIDefinition's spec, the `use_keyless` field is set to `true`.

> Tyk keyless access represents completely open access for your API and causes Tyk to bypass any session-based middleware 
(middleware that requires access to token-related metadata). Keyless access will allow all requests through. 
> 
You can disable keyless access by setting `use_keyless` to false. Let's update `httpbin` API to see it in action.

In order to update `httpbin` API, either update your `httpbin.yaml` file as follows:
```yaml
apiVersion: tyk.tyk.io/v1alpha1
kind: ApiDefinition
metadata:
  name: httpbin
spec:
  name: httpbin
  use_keyless: false
  protocol: http
  active: true
  proxy:
    target_url: http://httpbin.org
    listen_path: /httpbin
    strip_listen_path: true
```

and apply changes :
```bash
kubectl apply -f config/samples/httpbin.yaml

apidefinition.tyk.tyk.io/httpbin configured
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
    use_keyless: false
    protocol: http
    active: true
    proxy:
        target_url: http://httpbin.org
        listen_path: /httpbin
        strip_listen_path: true
EOF

apidefinition.tyk.tyk.io/httpbin configured
```

In order to access `httpbin` API, you need to include a key to the header.

> **Note**: All supported authentication types by Tyk Operator are listed [here](https://github.com/TykTechnologies/tyk-operator/blob/master/docs/api_definitions.md#client-to-gateway-authentication).
Authentication token is the default one if you have set `use_keyless` to false and haven't specified any other Authentication mode.

```bash
curl -i localhost:8080/httpbin/get
HTTP/1.1 401 Unauthorized
Content-Type: application/json
X-Generator: tyk.io
Date: Thu, 03 Mar 2022 15:47:30 GMT
Content-Length: 46

{
    "error": "Authorization field missing"
}%
```

We need to generate a key to access the `httpbin` API. Generating a key depends on which type of Tyk installation you have.

### Tyk CE

To create an API Key, we will need the API ID that we wish to grant the key access to, then creating the key is a very simple API call to the endpoint.

> **Prerequisite**: You will need your API secret, this is the `secret` property of the tyk.conf file.
Once you have this value, you can use them to access the Gateway API

In order to obtain API ID, we can describe ApiDefinition resource that represents our `httpbin` API.

```bash
kubectl describe tykapis httpbin
Name:         httpbin
Namespace:    default
Labels:       <none>
Annotations:  <none>
API Version:  tyk.tyk.io/v1alpha1
Kind:         ApiDefinition
Metadata:
  ...
Spec:
  ...
  api_id:  ZGVmYXVsdC9odHRwYmlu
  ...
Status:
  api_id:  ZGVmYXVsdC9odHRwYmlu
Events:    <none>
```

Once you describe `httpbin` ApiDefinition, the API ID is represented in the `status.api_id` field.

In our example,
- `{API-NAME}`: httpbin
- `{API-ID}`: ZGVmYXVsdC9odHRwYmlu

The below curl command will generate a key for our API:

```bash
curl -X POST -H "x-tyk-authorization: {API-SECRET}" \
  -s \
  -H "Content-Type: application/json" \
  -X POST \
  -d '{
    "allowance": 1000,
    "rate": 1000,
    "per": 1,
    "expires": -1,
    "quota_max": -1,
    "org_id": "1",
    "quota_renews": 1449051461,
    "quota_remaining": -1,
    "quota_renewal_rate": 60,
    "access_rights": {
      "{API-ID}": {
        "api_id": "{API-ID}",
        "api_name": "{API-NAME}",
        "versions": ["Default"]
      }
    },
    "meta_data": {}
  }' http://localhost:8080/tyk/keys/create | python -mjson.tool
```
```json
{
"action": "added",
"key": "eyJvcmciOiIxIiwiaWQiOiIxZTNhMTNhNGU4MGQ0ZWIxOGMzNjhlNzkyMjY5ODBmYiIsImgiOiJtdXJtdXIxMjgifQ==",
"key_hash": "9d8b101625d77f93153f2eeb0d2ae365",
"status": "ok"
}
```

That’s it, we have created a key - now we can try and use it.

```bash
curl -H "Authorization: Bearer eyJvcmciOiIxIiwiaWQiOiIxZTNhMTNhNGU4MGQ0ZWIxOGMzNjhlNzkyMjY5ODBmYiIsImgiOiJtdXJtdXIxMjgifQ==" localhost:8080/httpbin/get
```
```json
{
  "args": {},
  "headers": {
    "Accept": "*/*",
    "Accept-Encoding": "gzip",
    "Authorization": "Bearer eyJvcmciOiIxIiwiaWQiOiIxZTNhMTNhNGU4MGQ0ZWIxOGMzNjhlNzkyMjY5ODBmYiIsImgiOiJtdXJtdXIxMjgifQ==",
    "Host": "httpbin.org",
    "User-Agent": "curl/7.77.0",
    "X-Amzn-Trace-Id": "Root=1-6221e1e5-084139824f22e88542ab9246"
  },
  "origin": "127.0.0.1, 176.42.143.200",
  "url": "http://httpbin.org/get"
}
```

### Tyk Pro

Select `Keys` from the `System Management` section and click `ADD KEY` as shown in the image below.

![dashboard-keys-page](./img/getting-started-keys.png)


You have the option to create a new key either by selecting an existing Policy created for your API or by simply selecting your API. For this tutorial we are going to use an API.

To select `httpbin` API:

- Scroll through your `API Name` list,
- Use the `Search` field
- Select the `httpbin` API that was previously created.

You can leave all other options at their default settings.

![dashboard-keys-page-2](./img/getting-started-keys-2.png)

Now, we will add configurations details to set an expiry time after which the key will expire, using `Configuration` section indicated by `3` in the above image.

![dashboard-keys-page-3](./img/getting-started-keys-3.png)

Once you click `CREATE KEY`, a Key successfully generated pop-up will be displayed with the key. 
> You must save this somewhere for future reference as it will not be displayed again. 
Click `Copy to clipboard` and paste into a text document.

![dashboard-keys-page-4](./img/getting-started-keys-4.png)

That’s it, we have created a key - now we can try and use it.

```bash
curl -H "Authorization: Bearer {Key ID}" localhost:8080/httpbin/get
```

```json
{
  "args": {},
  "headers": {
    "Accept": "*/*",
    "Accept-Encoding": "gzip",
    "Authorization": "Bearer {Key ID}",
    "Host": "httpbin.org",
    "User-Agent": "curl/7.77.0",
    "X-Amzn-Trace-Id": "Root=1-6221de2a-01aa10dd56f6f13f420ba313"
  },
  "origin": "127.0.0.1, 176.42.143.200",
  "url": "http://httpbin.org/get"
}
```

Since we have provided a valid key along with our request, we do not have `HTTP 401 Unauthorized` response. 

## Secure an API

You can access a secured API by creating a key. Key can be created by specifing security policy. A security policy encapsulates several options that can be applied to a key. It acts as a template that can
override individual sections of an API key (or identity) in Tyk.

Imagine you have issued mutiple keys and later if you want to change access rights, rate limits or quotas, you will have to update all the keys manually. Security policy comes handy in this scenario. You just need to update security policy linked to the keys once. 

Let's create an API and security policy for that API. 

> **Security policy resources are currently only supported when using Tyk Pro mode. You can get round this by mounting the policy object as a volume into the gateway container.**

You can do so either by applying manifest defined in our repo 

```bash
kubectl apply -f docs/policies/ratelimit.yaml
```

Or create it by running following command

```bash 
cat <<EOF | kubectl apply -f -
apiVersion: tyk.tyk.io/v1alpha1
kind: ApiDefinition
metadata:
  name: httpbin
spec:
  name: httpbin protected
  protocol: http
  active: true
  proxy:
    target_url: http://httpbin.org
    listen_path: /httpbin
    strip_listen_path: true
  use_standard_auth: true
  auth_configs:
    authToken:
      auth_header_name: Authorization
---
apiVersion: tyk.tyk.io/v1alpha1
kind: SecurityPolicy
metadata:
  name: httpbin
spec:
  name: Rate Limit, Quota and Throttling policy
  state: active
  active: true
  access_rights_array:
    - name: httpbin
      namespace: default
      versions:
        - Default
  quota_max: 10
  quota_renewal_rate: 60
  rate: 5
  per: 5
  throttle_interval: 2
  throttle_retry_limit: 2
EOF
```

Ensure policy is created

```
$ kubectl get securitypolicy
NAME      AGE
httpbin   10s
```

We have successfully created `httpbin` security policy for `httpbin` API. Policy also sets global usage quota, rate limits and throttling.

Let's dive into fields we have set in the policy
- **name**: Name of security policy.
- **active**: Marks policy as active.
- **state**: It can have value `active`, `draft`,`deny`.
- **access_right_array**: List of APIs security policy has access to.

Usage Quota fields
- **quota_max**: The maximum number of allowed requests over a quota period.
- **quota_renewal_rate**: Time, in seconds, after which quota will be renewed.

Rate limiting fields
- **rate**: The number of requests to allow per period. 
- **per**: Time in seconds.

Throttling fields:
- **throttle_interval**:  Interval (in seconds) between each request retry.
- **throttle_retry_limit**: Total request retry number.


Now you can create a key, using this security policy, and access your API.

> We are continously adding support for new features which you can track [here](./policies.md)


## Publish an API to the Tyk Portal

Assuming that you have created the httpbin example API using the steps above, we can easily publish it to the Tyk Portal by applying a few specs:
1. SecurityPolicy spec
2. APIDescription spec
3. PortalAPICatalogue spec


### 1. Adding a security policy via Tyk Operator

When you publish an API to the Portal, Tyk actually publishes a way for developers to enrol in a policy, not into the API directly. Therefore, we must also apply a SecurityPolicy CRD before proceeding with the publishing.

For this we can use the following command:

```bash
cat <<EOF | kubectl apply -f -
apiVersion: tyk.tyk.io/v1alpha1
kind: SecurityPolicy
metadata:
 name: standard-pol
spec:
 name: standard-pol
 active: true
 state: active
 access_rights_array:
 - name: httpbin
   namespace: default
   versions:
     - Default
EOF
```

The above command will create the most basic security policy possible and attribute it to the httpbin API we previously created.


### 2. Creating an API description

The portal serves as a visual description of an API and therefore we need to let the Tyk Portal know (via Tyk Operator) some details about the API we want to publish.

For this we can run the following command:

```bash
cat <<EOF | kubectl apply -f -
apiVersion: tyk.tyk.io/v1alpha1
kind: APIDescription
metadata:
 name: standard-desc
spec:
 name: HTTPBIN API
 policyRef:
  name: standard-pol
  namespace: default
 docs: 
  doc_type: swagger
  documentation: "ewogICAgInN3YWdnZXIiOiAiMi4wIiwKICAgICJpbmZvIjogewogICAgICAgICJ0aXRsZSI6ICJodHRwYmluLm9yZyIsCiAgICAgICAgImRlc2NyaXB0aW9uIjogIkEgc2ltcGxlIEhUVFAgUmVxdWVzdCAmIFJlc3BvbnNlIFNlcnZpY2UuIiwKICAgICAgICAidmVyc2lvbiI6ICIxLjAiCiAgICB9LAogICAgImhvc3QiOiAiaHR0cGJpbi5vcmciLAogICAgInNjaGVtZXMiOiBbCiAgICAgICAgImh0dHAiLAogICAgICAgICJodHRwcyIKICAgIF0sCiAgICAicGF0aHMiOiB7CiAgICAgICAgIi9zdGF0dXMve2NvZGV9IjogewogICAgICAgICAgICAiZ2V0IjogewogICAgICAgICAgICAgICAgImRlc2NyaXB0aW9uIjogIlJldHVybnMgcHJvdmlkZWQgSFRUUCBTdGF0dXMgY29kZS4iLAogICAgICAgICAgICAgICAgIm9wZXJhdGlvbklkIjogIi9zdGF0dXMiLAogICAgICAgICAgICAgICAgInBhcmFtZXRlcnMiOiBbewogICAgICAgICAgICAgICAgICAgICJuYW1lIjogImNvZGUiLAogICAgICAgICAgICAgICAgICAgICJpbiI6ICJwYXRoIiwKICAgICAgICAgICAgICAgICAgICAiZGVzY3JpcHRpb24iOiAiSFRUUCBjb2RlIHRvIHJldHVybi4iLAogICAgICAgICAgICAgICAgICAgICJyZXF1aXJlZCI6IHRydWUsCiAgICAgICAgICAgICAgICAgICAgInR5cGUiOiAibnVtYmVyIiwKICAgICAgICAgICAgICAgICAgICAiZGVmYXVsdCI6IDIwMCwKICAgICAgICAgICAgICAgICAgICAiZW51bSI6IFsKICAgICAgICAgICAgICAgICAgICAgICAgMjAwCiAgICAgICAgICAgICAgICAgICAgXQogICAgICAgICAgICAgICAgfV0sCiAgICAgICAgICAgICAgICAicmVzcG9uc2VzIjoge30KICAgICAgICAgICAgfQogICAgICAgIH0sCiAgICAgICAgIi9nZXQiOiB7CiAgICAgICAgICAgICJnZXQiOiB7CiAgICAgICAgICAgICAgICAiZGVzY3JpcHRpb24iOiAiUmV0dXJucyBHRVQgZGF0YS4iLAogICAgICAgICAgICAgICAgIm9wZXJhdGlvbklkIjogIi9nZXQiLAogICAgICAgICAgICAgICAgInJlc3BvbnNlcyI6IHt9CiAgICAgICAgICAgIH0KICAgICAgICB9LAogICAgICAgICIvcG9zdCI6IHsKICAgICAgICAgICAgInBvc3QiOiB7CiAgICAgICAgICAgICAgICAiZGVzY3JpcHRpb24iOiAiUmV0dXJucyBQT1NUIGRhdGEuIiwKICAgICAgICAgICAgICAgICJvcGVyYXRpb25JZCI6ICIvcG9zdCIsCiAgICAgICAgICAgICAgICAicmVzcG9uc2VzIjoge30KICAgICAgICAgICAgfQogICAgICAgIH0sCiAgICAgICAgIi9wYXRjaCI6IHsKICAgICAgICAgICAgInBhdGNoIjogewogICAgICAgICAgICAgICAgImRlc2NyaXB0aW9uIjogIlJldHVybnMgUEFUQ0ggZGF0YS4iLAogICAgICAgICAgICAgICAgIm9wZXJhdGlvbklkIjogIi9wYXRjaCIsCiAgICAgICAgICAgICAgICAicmVzcG9uc2VzIjoge30KICAgICAgICAgICAgfQogICAgICAgIH0sCiAgICAgICAgIi9wdXQiOiB7CiAgICAgICAgICAgICJwdXQiOiB7CiAgICAgICAgICAgICAgICAiZGVzY3JpcHRpb24iOiAiUmV0dXJucyBQVVQgZGF0YS4iLAogICAgICAgICAgICAgICAgIm9wZXJhdGlvbklkIjogIi9wdXQiLAogICAgICAgICAgICAgICAgInJlc3BvbnNlcyI6IHt9CiAgICAgICAgICAgIH0KICAgICAgICB9LAogICAgICAgICIvZGVsZXRlIjogewogICAgICAgICAgICAiZGVsZXRlIjogewogICAgICAgICAgICAgICAgImRlc2NyaXB0aW9uIjogIlJldHVybnMgREVMRVRFIGRhdGEuIiwKICAgICAgICAgICAgICAgICJvcGVyYXRpb25JZCI6ICIvZGVsZXRlIiwKICAgICAgICAgICAgICAgICJyZXNwb25zZXMiOiB7fQogICAgICAgICAgICB9CiAgICAgICAgfSwKICAgICAgICAiL3htbCI6IHsKICAgICAgICAgICAgImdldCI6IHsKICAgICAgICAgICAgICAgICJkZXNjcmlwdGlvbiI6ICJSZXR1cm5zIHNvbWUgWE1MLiIsCiAgICAgICAgICAgICAgICAib3BlcmF0aW9uSWQiOiAiL3htbCIsCiAgICAgICAgICAgICAgICAicmVzcG9uc2VzIjoge30KICAgICAgICAgICAgfQogICAgICAgIH0sCiAgICAgICAgIi9pcCI6IHsKICAgICAgICAgICAgImdldCI6IHsKICAgICAgICAgICAgICAgICJkZXNjcmlwdGlvbiI6ICJSZXR1cm5zIG9yaWdpbiBJUC4iLAogICAgICAgICAgICAgICAgIm9wZXJhdGlvbklkIjogIi9pcCIsCiAgICAgICAgICAgICAgICAicmVzcG9uc2VzIjoge30KICAgICAgICAgICAgfQogICAgICAgIH0sCiAgICAgICAgIi91c2VyLWFnZW50IjogewogICAgICAgICAgICAiZ2V0IjogewogICAgICAgICAgICAgICAgImRlc2NyaXB0aW9uIjogIlJldHVybnMgdXNlciBhZ2VudCBzdHJpbmcuIiwKICAgICAgICAgICAgICAgICJvcGVyYXRpb25JZCI6ICIvdXNlci1hZ2VudCIsCiAgICAgICAgICAgICAgICAicmVzcG9uc2VzIjoge30KICAgICAgICAgICAgfQogICAgICAgIH0sCiAgICAgICAgIi9oZWFkZXJzIjogewogICAgICAgICAgICAiZ2V0IjogewogICAgICAgICAgICAgICAgImRlc2NyaXB0aW9uIjogIlJldHVybnMgaGVhZGVycyBkaWN0aW9uYXJ5LiIsCiAgICAgICAgICAgICAgICAib3BlcmF0aW9uSWQiOiAiL2hlYWRlcnMiLAogICAgICAgICAgICAgICAgInJlc3BvbnNlcyI6IHt9CiAgICAgICAgICAgIH0KICAgICAgICB9LAogICAgICAgICIvZGVsYXkve3NlY29uZHN9IjogewogICAgICAgICAgICAiZ2V0IjogewogICAgICAgICAgICAgICAgImRlc2NyaXB0aW9uIjogIkRlbGF5cyByZXNwb25kaW5nIGZvciBu4oCTMTAgc2Vjb25kcy4iLAogICAgICAgICAgICAgICAgIm9wZXJhdGlvbklkIjogIi9kZWxheSIsCiAgICAgICAgICAgICAgICAicGFyYW1ldGVycyI6IFt7CiAgICAgICAgICAgICAgICAgICAgIm5hbWUiOiAic2Vjb25kcyIsCiAgICAgICAgICAgICAgICAgICAgImluIjogInBhdGgiLAogICAgICAgICAgICAgICAgICAgICJkZXNjcmlwdGlvbiI6ICIiLAogICAgICAgICAgICAgICAgICAgICJyZXF1aXJlZCI6IHRydWUsCiAgICAgICAgICAgICAgICAgICAgInR5cGUiOiAic3RyaW5nIiwKICAgICAgICAgICAgICAgICAgICAiZGVmYXVsdCI6IDIsCiAgICAgICAgICAgICAgICAgICAgImVudW0iOiBbCiAgICAgICAgICAgICAgICAgICAgICAgIDIKICAgICAgICAgICAgICAgICAgICBdCiAgICAgICAgICAgICAgICB9XSwKICAgICAgICAgICAgICAgICJyZXNwb25zZXMiOiB7fQogICAgICAgICAgICB9CiAgICAgICAgfSwKICAgICAgICAiL2NhY2hlL3ttYXhBZ2V9IjogewogICAgICAgICAgICAiZ2V0IjogewogICAgICAgICAgICAgICAgIm9wZXJhdGlvbklkIjogIi9jYWNoZSIsCiAgICAgICAgICAgICAgICAiZGVzY3JpcHRpb24iOiAiU2V0cyBhIENhY2hlLUNvbnRyb2wgaGVhZGVyIGZvciBuIHNlY29uZHMuIiwKICAgICAgICAgICAgICAgICJwYXJhbWV0ZXJzIjogW3sKICAgICAgICAgICAgICAgICAgICAibmFtZSI6ICJtYXhBZ2UiLAogICAgICAgICAgICAgICAgICAgICJpbiI6ICJwYXRoIiwKICAgICAgICAgICAgICAgICAgICAiZGVzY3JpcHRpb24iOiAiIiwKICAgICAgICAgICAgICAgICAgICAicmVxdWlyZWQiOiB0cnVlLAogICAgICAgICAgICAgICAgICAgICJ0eXBlIjogInN0cmluZyIsCiAgICAgICAgICAgICAgICAgICAgImRlZmF1bHQiOiAxMCwKICAgICAgICAgICAgICAgICAgICAiZW51bSI6IFsKICAgICAgICAgICAgICAgICAgICAgICAgMTAKICAgICAgICAgICAgICAgICAgICBdCiAgICAgICAgICAgICAgICB9XSwKICAgICAgICAgICAgICAgICJyZXNwb25zZXMiOiB7fQogICAgICAgICAgICB9CiAgICAgICAgfSwKICAgICAgICAiL3V1aWQiOiB7CiAgICAgICAgICAgICJnZXQiOiB7CiAgICAgICAgICAgICAgICAib3BlcmF0aW9uSWQiOiAiL3V1aWQiLAogICAgICAgICAgICAgICAgImRlc2NyaXB0aW9uIjogIlJldHVybnMgVVVJRDQuIiwKICAgICAgICAgICAgICAgICJyZXNwb25zZXMiOiB7fQogICAgICAgICAgICB9CiAgICAgICAgfSwKICAgICAgICAiL2FueXRoaW5nIjogewogICAgICAgICAgICAiZ2V0IjogewogICAgICAgICAgICAgICAgIm9wZXJhdGlvbklkIjogIi9hbnl0aGluZyIsCiAgICAgICAgICAgICAgICAiZGVzY3JpcHRpb24iOiAiUmV0dXJucyByZXF1ZXN0IGRhdGEsIGluY2x1ZGluZyBtZXRob2QgdXNlZC4iLAogICAgICAgICAgICAgICAgInJlc3BvbnNlcyI6IHt9CiAgICAgICAgICAgIH0KICAgICAgICB9LAogICAgICAgICIvZ3ppcCI6IHsKICAgICAgICAgICAgImdldCI6IHsKICAgICAgICAgICAgICAgICJvcGVyYXRpb25JZCI6ICIvZ3ppcCIsCiAgICAgICAgICAgICAgICAiZGVzY3JpcHRpb24iOiAiUmV0dXJucyBnemlwLWVuY29kZWQgZGF0YS4iLAogICAgICAgICAgICAgICAgInJlc3BvbnNlcyI6IHt9CiAgICAgICAgICAgIH0KICAgICAgICB9LAogICAgICAgICIvYmFzZTY0L3t2YWx1ZX0iOiB7CiAgICAgICAgICAgICJnZXQiOiB7CiAgICAgICAgICAgICAgICAib3BlcmF0aW9uSWQiOiAiL2Jhc2U2NCIsCiAgICAgICAgICAgICAgICAiZGVzY3JpcHRpb24iOiAiRGVjb2RlcyBiYXNlNjR1cmwtZW5jb2RlZCBzdHJpbmcuIiwKICAgICAgICAgICAgICAgICJwYXJhbWV0ZXJzIjogW3sKICAgICAgICAgICAgICAgICAgICAibmFtZSI6ICJ2YWx1ZSIsCiAgICAgICAgICAgICAgICAgICAgImluIjogInBhdGgiLAogICAgICAgICAgICAgICAgICAgICJkZXNjcmlwdGlvbiI6ICJCYXNlNjQgZW5jb2RlZCBzdHJpbmcuIiwKICAgICAgICAgICAgICAgICAgICAicmVxdWlyZWQiOiB0cnVlLAogICAgICAgICAgICAgICAgICAgICJ0eXBlIjogInN0cmluZyIsCiAgICAgICAgICAgICAgICAgICAgImRlZmF1bHQiOiAiYUdWc2JHOGdkMjl5YkdRTkNnJTNEJTNEIgogICAgICAgICAgICAgICAgfV0sCiAgICAgICAgICAgICAgICAicmVzcG9uc2VzIjoge30KICAgICAgICAgICAgfQogICAgICAgIH0sCiAgICAgICAgIi9kZWZsYXRlIjogewogICAgICAgICAgICAiZ2V0IjogewogICAgICAgICAgICAgICAgIm9wZXJhdGlvbklkIjogIi9kZWZsYXRlIiwKICAgICAgICAgICAgICAgICJkZXNjcmlwdGlvbiI6ICJSZXR1cm5zIGRlZmxhdGUtZW5jb2RlZCBkYXRhLiIsCiAgICAgICAgICAgICAgICAicmVzcG9uc2VzIjoge30KICAgICAgICAgICAgfQogICAgICAgIH0sCiAgICAgICAgIi9icm90bGkiOiB7CiAgICAgICAgICAgICJnZXQiOiB7CiAgICAgICAgICAgICAgICAib3BlcmF0aW9uSWQiOiAiL2Jyb3RsaSIsCiAgICAgICAgICAgICAgICAiZGVzY3JpcHRpb24iOiAiUmV0dXJucyBicm90bGktZW5jb2RlZCBkYXRhLiIsCiAgICAgICAgICAgICAgICAicmVzcG9uc2VzIjoge30KICAgICAgICAgICAgfQogICAgICAgIH0sCiAgICAgICAgIi9yZXNwb25zZS1oZWFkZXJzIjogewogICAgICAgICAgICAiZ2V0IjogewogICAgICAgICAgICAgICAgIm9wZXJhdGlvbklkIjogIi9yZXNwb25zZS1oZWFkZXJzIiwKICAgICAgICAgICAgICAgICJkZXNjcmlwdGlvbiI6ICJSZXR1cm5zIGtleS12YWx1ZSBxdWVyeSBwYXJhbWV0ZXJzIGFzIHJlc3BvbnNlIGhlYWRlcnMuIiwKICAgICAgICAgICAgICAgICJyZXNwb25zZXMiOiB7fQogICAgICAgICAgICB9CiAgICAgICAgfSwKICAgICAgICAiL2J5dGVzL3tudW1iZXJ9IjogewogICAgICAgICAgICAiZ2V0IjogewogICAgICAgICAgICAgICAgIm9wZXJhdGlvbklkIjogIi9ieXRlcyIsCiAgICAgICAgICAgICAgICAiZGVzY3JpcHRpb24iOiAiR2VuZXJhdGVzIG4gcmFuZG9tIGJ5dGVzIG9mIGJpbmFyeSBkYXRhIiwKICAgICAgICAgICAgICAgICJwYXJhbWV0ZXJzIjogW3sKICAgICAgICAgICAgICAgICAgICAibmFtZSI6ICJudW1iZXIiLAogICAgICAgICAgICAgICAgICAgICJpbiI6ICJwYXRoIiwKICAgICAgICAgICAgICAgICAgICAiZGVzY3JpcHRpb24iOiAiTnVtYmVyIG9mIGJ5dGVzIHRvIHJldHVybi4iLAogICAgICAgICAgICAgICAgICAgICJyZXF1aXJlZCI6IHRydWUsCiAgICAgICAgICAgICAgICAgICAgInR5cGUiOiAic3RyaW5nIiwKICAgICAgICAgICAgICAgICAgICAiZGVmYXVsdCI6ICIxMDI0IgogICAgICAgICAgICAgICAgfV0sCiAgICAgICAgICAgICAgICAicmVzcG9uc2VzIjoge30KICAgICAgICAgICAgfQogICAgICAgIH0sCiAgICAgICAgIi9yZWRpcmVjdC10byI6IHsKICAgICAgICAgICAgImdldCI6IHsKICAgICAgICAgICAgICAgICJvcGVyYXRpb25JZCI6ICIvcmVkaXJlY3QtdG8iLAogICAgICAgICAgICAgICAgImRlc2NyaXB0aW9uIjogIjMwMiByZWRpcmVjdHMgdG8gYSBVUkwgcHJvdmlkZWQgaW4gcXVlcnkiLAogICAgICAgICAgICAgICAgInBhcmFtZXRlcnMiOiBbewogICAgICAgICAgICAgICAgICAgICJuYW1lIjogInVybCIsCiAgICAgICAgICAgICAgICAgICAgImluIjogInF1ZXJ5IiwKICAgICAgICAgICAgICAgICAgICAiZGVzY3JpcHRpb24iOiAiUmVkaXJlY3QgdGFyZ2V0IiwKICAgICAgICAgICAgICAgICAgICAicmVxdWlyZWQiOiB0cnVlLAogICAgICAgICAgICAgICAgICAgICJ0eXBlIjogInN0cmluZyIsCiAgICAgICAgICAgICAgICAgICAgImRlZmF1bHQiOiAiaHR0cDovL2V4YW1wbGUuY29tIgogICAgICAgICAgICAgICAgfV0sCiAgICAgICAgICAgICAgICAicmVzcG9uc2VzIjoge30KICAgICAgICAgICAgfQogICAgICAgIH0sCiAgICAgICAgIi9zdHJlYW0ve251bWJlcn0iOiB7CiAgICAgICAgICAgICJnZXQiOiB7CiAgICAgICAgICAgICAgICAib3BlcmF0aW9uSWQiOiAiL3N0cmVhbSIsCiAgICAgICAgICAgICAgICAiZGVzY3JpcHRpb24iOiAiU3RyZWFtcyBtaW4obnVtYmVyLCAxMDApIGxpbmVzLiIsCiAgICAgICAgICAgICAgICAicGFyYW1ldGVycyI6IFt7CiAgICAgICAgICAgICAgICAgICAgIm5hbWUiOiAibnVtYmVyIiwKICAgICAgICAgICAgICAgICAgICAiaW4iOiAicGF0aCIsCiAgICAgICAgICAgICAgICAgICAgImRlc2NyaXB0aW9uIjogIk51bWJlciBvZiBsaW5lcyB0byBzdHJlYW0uIiwKICAgICAgICAgICAgICAgICAgICAicmVxdWlyZWQiOiB0cnVlLAogICAgICAgICAgICAgICAgICAgICJ0eXBlIjogIm51bWJlciIsCiAgICAgICAgICAgICAgICAgICAgImRlZmF1bHQiOiAiMTAiCiAgICAgICAgICAgICAgICB9XSwKICAgICAgICAgICAgICAgICJyZXNwb25zZXMiOiB7fQogICAgICAgICAgICB9CiAgICAgICAgfQogICAgfQp9Cg=="
 show: true
 version: v2
EOF
```

Here we can observe that the API description is tied to the SecurityPolicy
in the lines

```yaml
 policyRef:
  name: standard-pol
```

Furthermore, in the "documentation" field we have the swagger API definition of httpbin encoded in base64. All documentation descriptions need to be encoded base64.

### 3. Publishing the API to the Tyk Portal

For this step all we need to do is to create a PortalAPICatalogue resource that ties everything together. In our case, the command we need to run is:
```bash
cat <<EOF | kubectl apply -f -
apiVersion: tyk.tyk.io/v1alpha1
kind: PortalAPICatalogue
metadata:
 name: test-httpbin-api
spec:
 apis:
 - apiDescriptionRef:
    name: standard-desc
    namespace: default
EOF
```

As we can observe above, we just need to reference the name of the apiDescription and the Tyk Operator will handle everything for us.

After this is all done, if we go to the default TykPortal implementation
(generally located at `tyk-portal.local`) we should observe something like this in the API Catalogue section:


![api-catalogue-httpbin](./img/api-catalogue-httpbin.png)


And if we click on `View documentation` we should have:


![httpbin-swagger-portal](./img/httpbin-swagger-portal.png)