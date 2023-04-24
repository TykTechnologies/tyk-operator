# Local Development

__NOTE__: These instructions are for kind clusters which is the recommended way
for development.

<!-- TOC -->
* [Local Development](#local-development)
  * [Prerequisites](#prerequisites)
  * [Create cluster](#create-cluster)
  * [Bootstrap the dev environment](#bootstrap-the-dev-environment)
    * [Deploying only Tyk Operator](#deploying-only-tyk-operator)
    * [Deploy Tyk Community Edition and Tyk Operator](#deploy-tyk-community-edition-and-tyk-operator)
      * [Deploy Only Tyk Gateway](#deploy-only-tyk-gateway)
    * [Deploy Tyk Pro and Tyk Operator](#deploy-tyk-pro-and-tyk-operator)
      * [Deploy Only Tyk Pro](#deploy-only-tyk-pro)
    * [Run Tyk Operator Locally](#run-tyk-operator-locally)
    * [Specify Tyk Version](#specify-tyk-version)
  * [Check if the environment is working](#check-if-the-environment-is-working)
    * [Check Tyk Operator logs](#check-tyk-operator-logs)
    * [Check Tyk Pro installation](#check-tyk-pro-installation)
    * [Check Tyk Headless (Community edition)](#check-tyk-headless-community-edition)
  * [Scrapbook](#scrapbook)
  * [Delete cluster](#delete-cluster)
  * [Run tests](#run-tests)
    * [CE](#ce)
    * [Pro](#pro)
  * [Helm Chart Development](#helm-chart-development)
  * [Loading images into kind cluster](#loading-images-into-kind-cluster)
  * [Useful make commands](#useful-make-commands)
  * [Highlights](#highlights)
<!-- TOC -->

## Prerequisites

Note that these are tools which we use for development of this project.

| tool                                               | version   |
|----------------------------------------------------|-----------|
| [kubectl](https://kubernetes.io/docs/tasks/tools/) |           |
| [kind](https://kind.sigs.k8s.io)                   |           |
| [go](https://go.dev/doc/install)                   | 1.19      |
| [operator-sdk](https://sdk.operatorframework.io/)  | v1.7.2+   |
| controller-gen                                     | v0.4.1+   |
| [kustomize](https://kustomize.io/)                 | v3.8.7+   |
| [helm](https://helm.sh/)                           | v3.5.4+   |
| [Docker](https://docs.docker.com/)                 | 19.03.13+ |


## Create cluster

```shell
make create-kind-cluster
```

This will create a 3 node cluster; 1 control-plane and 2 worker nodes by using `kind`.

## Bootstrap the dev environment

This section will describe how you can set up local development environment for Tyk Operator.

Tyk Operator manages Tyk resources such as ApiDefinition or SecurityPolicy. In order to do that, 
it needs to access your Tyk installation which can be any type of installation 
([tyk-headless](https://tyk.io/docs/tyk-oss/ce-helm-chart/), 
[tyk-pro](https://tyk.io/docs/tyk-self-managed/tyk-helm-chart/), 
or [Tyk Cloud](https://tyk.io/docs/tyk-cloud/)).

This guideline will go through how you can bootstrap a testing environment for Tyk Operator, including the deployment of 
Tyk Gateway / Dashboard and their dependencies (Redis or Mongo if you want to use Tyk Dashboard).

<hr/>

There are a couple of options to setting up development environment.
You can either deploy Tyk Operator to Kubernetes cluster or you can run Tyk Operator locally. 

### Deploying only Tyk Operator

If you only want to deploy Tyk Operator, without deploying Tyk Gateway / Dashboard, you can use:

```bash
make install-operator-helm IMG=tykio/tyk-operator:test
```

This will 
- build a Docker image, 
- generate manifest files, 
- load the new Docker image to your cluster and, 
- deploy Tyk Operator by using `Helm`.

Now, Tyk Operator is deployed in `tyk-operator-system` namespace with `ci` release name.

### Deploy Tyk Community Edition and Tyk Operator

> You do not need any license to deploy Tyk Community Edition.

You can deploy fully functional environment for testing Tyk Operator by the using following command:

```shell
TYK_VERSION=${TYK_VERSION} make boot-ce IMG=tykio/tyk-operator:test
```

> By default, we deploy Tyk Gateway v4.3 by using [`tyk-headless`](https://tyk.io/docs/tyk-oss/ce-helm-chart/)
chart.

This will 
- deploy `cert-manager`,
- deploy [tyk-headless](https://tyk.io/docs/tyk-oss/ce-helm-chart/),
- setup a k8s secret `tyk-operator-conf` in `tyk-operator-system` namespace that 
will be used by Tyk Operator to access the deployed Tyk Gateway.

> The secret will contain required credentials mentioned [here](https://tyk.io/docs/tyk-stack/tyk-operator/installing-tyk-operator/#tyk-open-source).

#### Deploy Only Tyk Gateway

```bash
TYK_VERSION=${TYK_VERSION} make setup-ce
```

This command will **only** deploy Tyk Gateway without deploying Tyk Operator.
It is useful if your Tyk Operator is running on your local machine.

### Deploy Tyk Pro and Tyk Operator

In order to deploy Tyk Pro, you need a license.

Set your license key as an environment variable.
```bash
export TYK_DB_LICENSEKEY=${REPLACE_WITH_YOUR_LICENSE}
```

```shell
make boot-pro IMG=tykio/tyk-operator:test
```

This will 
- deploy cert-manager,
- deploy [tyk-pro](https://tyk.io/docs/tyk-self-managed/tyk-helm-chart/) chart,
- creates an organisation that you can use to log into your dashboard, 
  - By default, the username is `operator@example.com` and the password is `1234testing`.
- setup secret that will be used by Tyk Operator to access the deployed dashboard.

> The secret will contain required credentials mentioned [here](https://tyk.io/docs/tyk-stack/tyk-operator/installing-tyk-operator/#tyk-self-managed-hybrid).


#### Deploy Only Tyk Pro

> Make sure that your Tyk Pro license credential is available as described [above section](#deploy-tyk-pro-and-tyk-operator).

```bash
TYK_VERSION=${TYK_VERSION} make setup-pro
```

This command will **only** deploy Tyk Pro without deploying Tyk Operator.
It is useful if your Tyk Operator is running on your local machine.

### Run Tyk Operator Locally

You can run Tyk Operator on your local machine without deploying it into your Kubernetes cluster.
This approach will be useful while debugging.

> Note: While running Tyk Operator locally, webhooks won't be working. 
Tyk Operator uses webhooks (Validation and Mutation) to validate and default some values in Custom Resources.

```bash
TYK_URL=${TYK_URL} TYK_MODE=${TYK_MODE} TYK_AUTH=${TYK_AUTH} TYK_ORG=${TYK_ORG} ENABLE_WEBHOOKS=${ENABLE_WEBHOOKS} make run
```

This command will generate latest manifests (CRs, RBACs etc.) and run Tyk Operator locally.

Since Tyk Operator is not deployed on your cluster, it cannot access your Tyk installation.
So, before setting up the credentials for `TYK_URL`, please make sure that your Tyk service is accessible
in your local machine. Please see [here](#check-if-the-environment-is-working) for details.

### Specify Tyk Version

If you would like to bootstrap specific version of Tyk, you can use `TYK_VERSION` environment variable.

```bash
TYK_VERSION=v3.2 make boot-ce IMG=tykio/tyk-operator:test
``` 

This will bootstrap Tyk Gateway v3.2.

## Check if the environment is working

### Check Tyk Operator logs

```bash
make log
```
which runs the following command for you,
```bash
kubectl logs <tyk-controller-manager-pod-name> -n tyk-operator-system manager
```

### Check Tyk Pro installation

1. Expose our Tyk Gateway and Tyk Dashboard locally

Run this in a separate terminal:

```bash
kubectl port-forward svc/gateway-svc-pro-tyk-pro 8080:8080  -n tykpro-control-plane
kubectl port-forward service/dashboard-svc-pro-tyk-pro 3000:3000 -n tykpro-control-plane
```

2. Check if Tyk Gateway is properly configured

```bash
curl http://localhost:8080/hello
```
<details><summary>SHOW EXPECTED OUTPUT</summary>
<p>
<pre>
{
    "status": "pass",
    "version": "4.3.3",
    "description": "Tyk GW",
    "details": {
        "dashboard": {
            "status": "pass",
            "componentType": "system",
            "time": "2023-02-28T07:04:55Z"
        },
        "redis": {
            "status": "pass",
            "componentType": "datastore",
            "time": "2023-02-28T07:04:55Z"
        }
    }
}
</pre>
</p>
</details>

3. Check if the Tyk Operator is working
   - Create your first ApiDefinition resource

        ```bash
        kubectl apply -f config/samples/httpbin.yaml 
        ```
        <details><summary>SHOW EXPECTED OUTPUT</summary>
                    <p>
                    <pre>
                    apidefinition.tyk.tyk.io/httpbin created
                    </pre>
                    </p>
                    </details>

   - Check that your ApiDefinition was applied and it works
        ```bash
        curl http://localhost:8080/httpbin/headers
        ```
        <details><summary>SHOW EXPECTED OUTPUT</summary>
        <p>
        <pre>
        {
          "headers": {
            "Accept": "*/*", 
            "Accept-Encoding": "gzip", 
            "Host": "httpbin.org", 
            "Referer": "", 
            "User-Agent": "Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; Trident/5.0)", 
            "X-Amzn-Trace-Id": "Root=1-5fc0c7f4-4fe147d429feac260d185764"
          }
        }
        </pre>
        </p>
        </details>
    - Login to Tyk Dashboard through `http://localhost:3000` and check if your ApiDefinition was created.
    
### Check Tyk Headless (Community edition)

1. Expose our Tyk Gateway locally

Run this in a separate terminal

```bash
kubectl port-forward svc/gateway-svc-ce-tyk-headless 8080:8080  -n tykce-control-plane
```

2. Check if Tyk Gateway is properly configured

```bash
curl http://localhost:8080/hello
```
<details><summary>SHOW EXPECTED OUTPUT</summary>
<p>
<pre>
{
    "status": "pass",
    "version": "v3.0.0",
    "description": "Tyk GW",
    "details": {
        "redis": {
            "status": "pass",
            "componentType": "datastore",
            "time": "2020-11-27T10:36:19Z"
        }
    }
}
</pre>
</p>
</details>

3. Check if Tyk Operator is working

   - Create your first ApiDefinition resource:
        ```bash
        kubectl apply -f config/samples/httpbin.yaml 
        ```

   - Check that your ApiDefinition was applied and it works
        ```bash
        curl http://localhost:8080/httpbin/headers
        ```
        <details><summary>SHOW EXPECTED OUTPUT</summary>
        <p>
        <pre>
        {
          "headers": {
            "Accept": "*/*", 
            "Accept-Encoding": "gzip", 
            "Host": "httpbin.org", 
            "Referer": "", 
            "User-Agent": "Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; Trident/5.0)", 
            "X-Amzn-Trace-Id": "Root=1-5fc0d6e9-1d6ca594424891b803e5260d"
          }
        }
        </pre>
        </p>
        </details>


## Scrapbook

After making changes to the controller, in order to update Tyk Operator controller, you can run

```bash
make scrap IMG=tykio/tyk-operator:test
```
This will 

- generate manifests,
- build and tag the new changes into a docker image,
- load the image to kind cluster,
- uninstall the previous controller deployment,
- install the new controller deployment.

## Delete cluster

Delete kind cluster using following command

```bash
make clean
```

## Run tests
To run tests, boot up either of the following setups:

### CE
 
```shell
TYK_MODE=ce make boot-ce test-all
```

### Pro 

```shell
TYK_MODE=pro make boot-pro test-all
```

## Helm Chart Development

This section describes how you can update Helm Chart of Tyk Operator.

As of today, Tyk Operator can be installed via Helm Chart. The chart is stored in Tyk Operator repository, [here](../helm).

We generate all manifests by using `kustomize`.

Direct changes to files (e.g., [values.yaml](../helm/values.yaml) or [all.yaml](../helm/templates/all.yaml)) in the 
Helm directory will not help you update Tyk Operator's Helm Chart because changes will be overwritten by `make helm` command.

What we do for creating a Helm manifests is to run `make helm` command. 
This command runs `kustomize` on [config/helm](../config/helm) and sends the output of `kustomize` to 
[hack/helm/pre_helm.go](../hack/helm/pre_helm.go) file which replaces certain fields with their templating correspondence.

For example, in order to update [all.yaml](../helm/templates/all.yaml), it'd be better to update either 
[hack/helm/pre_helm.go](../hack/helm/pre_helm.go) or [config/helm/kustomization.yaml](../config/helm/kustomization.yaml).

## Loading images into kind cluster
You can load images in the kind cluster required for Tyk Pro/CE installation. 
It will download all the images once and will be reused during next installations, unless cluster is deleted.

You just need to set `TYK_OPERATOR_PRELOAD_IMAGES` environment variable to true before running `make boot-ce/make boot-pro`.

## Useful make commands

- `make linters`

This will run all linters and formatters that we use in Tyk Operator project. You can use this command 
to lint your code changes.

- `make helm`
  
This will update Helm Chart of Tyk Operator. Please see [Helm Chart Development](#helm-chart-development) section for details.

- `make generate manifests helm install`

This will
- generate code,
- generate manifests,
- update Helm Chart of Tyk Operator,
- install CRDs to the cluster.

## Highlights

1. The Tyk-Operator uses the [finalizer](https://book.kubebuilder.io/reference/using-finalizers.html) pattern for deleting CRs from the cluster.
