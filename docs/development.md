# Local Development Environment

__NOTE__: These instructions are for kind clusters which is the recommended way
for development

## Prerequisites
- [kind](https://kind.sigs.k8s.io/)
- [helm](https://helm.sh/)
- [Docker](https://www.docker.com/) or [podman](https://podman.io/)
- Golang (1.15.0+)

### 0 Create cluster

```shell
kind create cluster --config hack/kind.yml
```

This will create a 3 node cluster. 1 control-plane and 2 worker nodes.

### 1. boot strap the dev environment

The operator needs a couple of env vars so that it knows how to speak to the Tyk apis.

```shell
export TYK_AUTH=REPLACE_WITH_DASH_USER_KEY_OR_FOR_OSS_GW_SECRET
export TYK_ORG=REPLACE_WITH_ORG_ID
export TYK_MODE=pro|oss
export TYK_URL=REPLACE_WITH_DASHBOARD_URL_OR_GATEWAY_ADMIN_URL
```

#### Booting tyk community edition

```shell
make boot-ce IMG=tykio/tyk-operator:test
```
This will 
- deploy cert-manager
- deploy tyk-ce
- setup secret that will be used by the operator to access the deployed gateway


#### Booting tyk Pro editio

Set your license key inside an env var.
```
export TYK_DB_LICENSEKEY=REPLACE.WITH.YOUR.LICENSE
```

```shell
make boot-pro IMG=tykio/tyk-operator:test
```

<details><summary>SHOW EXPECTED OUTPUT</summary>
<p>
<pre>
===> installing cert-manager
kubectl apply --validate=false -f https://github.com/jetstack/cert-manager/releases/download/v1.0.4/cert-manager.yaml
customresourcedefinition.apiextensions.k8s.io/certificaterequests.cert-manager.io created
customresourcedefinition.apiextensions.k8s.io/certificates.cert-manager.io created
customresourcedefinition.apiextensions.k8s.io/challenges.acme.cert-manager.io created
customresourcedefinition.apiextensions.k8s.io/clusterissuers.cert-manager.io created
customresourcedefinition.apiextensions.k8s.io/issuers.cert-manager.io created
customresourcedefinition.apiextensions.k8s.io/orders.acme.cert-manager.io created
namespace/cert-manager created
serviceaccount/cert-manager-cainjector created
serviceaccount/cert-manager created
serviceaccount/cert-manager-webhook created
clusterrole.rbac.authorization.k8s.io/cert-manager-cainjector created
clusterrole.rbac.authorization.k8s.io/cert-manager-controller-issuers created
clusterrole.rbac.authorization.k8s.io/cert-manager-controller-clusterissuers created
clusterrole.rbac.authorization.k8s.io/cert-manager-controller-certificates created
clusterrole.rbac.authorization.k8s.io/cert-manager-controller-orders created
clusterrole.rbac.authorization.k8s.io/cert-manager-controller-challenges created
clusterrole.rbac.authorization.k8s.io/cert-manager-controller-ingress-shim created
clusterrole.rbac.authorization.k8s.io/cert-manager-view created
clusterrole.rbac.authorization.k8s.io/cert-manager-edit created
clusterrolebinding.rbac.authorization.k8s.io/cert-manager-cainjector created
clusterrolebinding.rbac.authorization.k8s.io/cert-manager-controller-issuers created
clusterrolebinding.rbac.authorization.k8s.io/cert-manager-controller-clusterissuers created
clusterrolebinding.rbac.authorization.k8s.io/cert-manager-controller-certificates created
clusterrolebinding.rbac.authorization.k8s.io/cert-manager-controller-orders created
clusterrolebinding.rbac.authorization.k8s.io/cert-manager-controller-challenges created
clusterrolebinding.rbac.authorization.k8s.io/cert-manager-controller-ingress-shim created
role.rbac.authorization.k8s.io/cert-manager-cainjector:leaderelection created
role.rbac.authorization.k8s.io/cert-manager:leaderelection created
role.rbac.authorization.k8s.io/cert-manager-webhook:dynamic-serving created
rolebinding.rbac.authorization.k8s.io/cert-manager-cainjector:leaderelection created
rolebinding.rbac.authorization.k8s.io/cert-manager:leaderelection created
rolebinding.rbac.authorization.k8s.io/cert-manager-webhook:dynamic-serving created
service/cert-manager created
service/cert-manager-webhook created
deployment.apps/cert-manager-cainjector created
deployment.apps/cert-manager created
deployment.apps/cert-manager-webhook created
mutatingwebhookconfiguration.admissionregistration.k8s.io/cert-manager-webhook created
validatingwebhookconfiguration.admissionregistration.k8s.io/cert-manager-webhook created
kubectl rollout status  deployment/cert-manager -n cert-manager
Waiting for deployment "cert-manager" rollout to finish: 0 of 1 updated replicas are available...
deployment "cert-manager" successfully rolled out
kubectl rollout status  deployment/cert-manager-cainjector -n cert-manager
Waiting for deployment "cert-manager-cainjector" rollout to finish: 0 of 1 updated replicas are available...
deployment "cert-manager-cainjector" successfully rolled out
kubectl rollout status  deployment/cert-manager-webhook -n cert-manager
deployment "cert-manager-webhook" successfully rolled out
===> installing tyk-pro
sh ./ci/deploy_tyk_pro.sh
creating namespace tykpro-control-plane
creating namespace tykpro-control-plane
namespace/tykpro-control-plane created
deploying gRPC plugin server
service/grpc-plugin created
deployment.apps/grpc-plugin created
Waiting for deployment spec update to be observed...
Waiting for deployment "grpc-plugin" rollout to finish: 0 out of 1 new replicas have been updated...
Waiting for deployment "grpc-plugin" rollout to finish: 0 of 1 updated replicas are available...
deployment "grpc-plugin" successfully rolled out
deploying databases
service/mongo created
deployment.apps/mongo created
deployment.apps/redis created
service/redis created
waiting for redis
Waiting for deployment "redis" rollout to finish: 0 of 1 updated replicas are available...
deployment "redis" successfully rolled out
waiting for mongo
Waiting for deployment "mongo" rollout to finish: 0 of 1 updated replicas are available...
deployment "mongo" successfully rolled out
creating configmaps
configmap/dash-conf created
configmap/tyk-conf created
setting dashboard secrets
secret/dashboard created
ZXlKaGJHY2lPaUpTVXpJMU5pSXNJblI1Y0NJNklrcFhWQ0o5LmV5SmhiR3h2ZDJWa1gyNXZaR1Z6SWpvaU1USmlZek5tTkRJdFltSmxOeTAwWkRWaExUUXlORGd0TlRBMk1XTXhZV1V5TjJZd0xERmpPVFprWVRZMUxXVmhOVEF0TkdVM01TMDNNVFJtTFdFMlpUZ3dOVGcxWXpNMlpDd3pNR013TVdJM05DMWhPREF5TFRReU56TXRORGMzWlMxak0ySmtOREU0T0daaE5tSXNNRFl3TXpZd00yRXRNRGxtWVMwME1HRm1MVFkwWmpRdE56VTNabVE0TWpRNFkySTVMR0V5TVdKaE1UaGxMVGMwWVdRdE5HVXhZeTAwTkdJMkxXSmlNamswWXpVeU9EazBaU3hpWVRObU1UazJOaTB4TURWbUxUUmxOV010Tm1FM1pDMDBNV015TnpobU5UZzJOMlFzTkRkaU56RTJORGt0TkRBNVpDMDBNMkk1TFRZNVltTXROR0V3WmpVNVl6VTNOakl5TERVeE9UUXpOR0V6TFRZeE9XRXRORGRsWmkwMlpqZGpMVFUyTlRGaE1qZGtPV1kzTlN3M09UTTVNRGN5TkMxbE56TTFMVFJoTlRZdE5qWXhOQzFoWmpBMlpEYzVPV0ZrTXpNc1lqRmxOemN4WXpRdE1XRTJaUzAwWkdaa0xUWTFORGt0WXpKbFlUaGtOV0l6WWpJMElpd2laWGh3SWpveE5qQTROekEyTURjeExDSnBZWFFpT2pFMk1EWXhNVFF3TnpFc0ltOTNibVZ5SWpvaU5UYzNPVGN4TVRrME5XWTVNbVUyTmpnNU1EQXdNVEkzSWl3aWMyTnZjR1VpT2lKdGRXeDBhVjkwWldGdExISmlZV01zWjNKaGNHZ2lMQ0oySWpvaU1pSjkuZ2VsUC1YRmFqUVNxOGxGSUFvU0pfQWZLU1QwTm9MNnNEdUdMd056d1NSS1Z6bHJOUTFLWmFBU045UlAycjR4Mm5nNk1uYWhFdUZYamxOQW95Z1lxWENRdWpoYi1PWlVQWmMtVmZaYThYNVM3eTUtNGZMNi1ISUxnWlphczUxMEJxcjlsMXhobFh1WkM0WTA0TWdvVkRkVzBzV05mTUtEOWE3cXU1X3A4T3d5ZzRzODgtRmV5bE9hbWVaNnZCVXBmV2pZVnlzaUZIeEpRNkYzemoxV0ZjcFNkZWxVMU9GVHNmRUFjaENEeXh5Z0U3OTA4SGQ3eW5nb3ItZlg5UnVxdjFsLW9MS1VLZGJLYnhfQm9kaTRtQUdWbmVQQVVFTWlnbHZ4VkM5aFdRenNPQ1NtbEpQY05FX2c3U2k1Z2NCNkR6SGVQcFFsYzRTN1JqT3lEcTNIdUFndeploying dashboard
service/dashboard created
deployment.apps/dashboard created
Waiting for deployment "dashboard" rollout to finish: 0 out of 1 new replicas have been updated...
Waiting for deployment "dashboard" rollout to finish: 0 of 1 updated replicas are available...
deployment "dashboard" successfully rolled out
deploying gateway
service/tyk created
service/gw created
deployment.apps/tyk created
Waiting for deployment "tyk" rollout to finish: 0 of 1 updated replicas are available...
deployment "tyk" successfully rolled out
dashboard logs
time="Nov 27 09:04:45" level=warning msg="toth/tothic: no TYK_IB_SESSION_SECRET environment variable is set. The default cookie store is not available and any calls will fail. Ignore this warning if you are using a different store."

time="Nov 27 09:04:45" level=info msg="Tyk Analytics Dashboard v3.0.1"
time="Nov 27 09:04:45" level=info msg="Copyright Tyk Technologies Ltd 2020"
time="Nov 27 09:04:45" level=info msg="https://www.tyk.io"
time="Nov 27 09:04:45" level=info msg="Using /etc/tyk-dashboard/dash.json for configuration"
time="Nov 27 09:04:45" level=info msg="Listening on port: 3000"
time="Nov 27 09:04:45" level=info msg="Connecting to MongoDB: [mongo.tykpro-control-plane.svc.cluster.local:27017]"
time="Nov 27 09:04:45" level=info msg="Mongo connection established"
time="Nov 27 09:04:45" level=info msg="Creating new Redis connection pool"
time="Nov 27 09:04:45" level=info msg="--> [REDIS] Creating single-node client"
time="Nov 27 09:04:45" level=info msg="Creating new Redis connection pool"
time="Nov 27 09:04:45" level=info msg="--> [REDIS] Creating single-node client"
time="Nov 27 09:04:45" level=info msg="Creating new Redis connection pool"
time="Nov 27 09:04:45" level=info msg="--> [REDIS] Creating single-node client"
time="Nov 27 09:04:45" level=info msg="Creating new Redis connection pool"
time="Nov 27 09:04:45" level=info msg="--> [REDIS] Creating single-node client"
time="Nov 27 09:04:45" level=info msg="Licensing: Setting new license"
time="Nov 27 09:04:45" level=info msg="Licensing: Registering nodes..."
time="Nov 27 09:04:45" level=info msg="Adding available nodes..."
time="Nov 27 09:04:45" level=info msg="Licensing: Checking capabilities"
time="Nov 27 09:04:45" level=info msg="Audit log is disabled in config"
time="Nov 27 09:04:45" level=info msg="Creating new Redis connection pool"
time="Nov 27 09:04:45" level=info msg="--> [REDIS] Creating single-node client"
time="Nov 27 09:04:45" level=info msg="--> Standard listener (http) for dashboard and API"
time="Nov 27 09:04:45" level=info msg="Starting zeroconf heartbeat"
time="Nov 27 09:04:45" level=info msg="Starting notification handler for gateway cluster"
time="Nov 27 09:04:45" level=info msg="Loading routes..."
time="Nov 27 09:04:45" level=info msg="Creating new Redis connection pool"
time="Nov 27 09:04:45" level=info msg="--> [REDIS] Creating single-node client"
time="Nov 27 09:04:45" level=info msg="Initializing Internal TIB"
time="Nov 27 09:04:45" level=info msg="Initializing Identity Cache" prefix="TIB INITIALIZER"
time="Nov 27 09:04:45" level=info msg="Set DB" prefix="TIB REDIS STORE"
time="Nov 27 09:04:45" level=info msg="Using internal Identity Broker. Routes are loaded and available."
gateway logs
time="Nov 27 09:06:01" level=info msg="Tyk API Gateway v3.0.0" prefix=main
time="Nov 27 09:06:01" level=warning msg="Insecure configuration allowed" config.allow_insecure_configs=true prefix=checkup
time="Nov 27 09:06:01" level=info msg="gRPC dispatcher was initialized" prefix=coprocess
time="Nov 27 09:06:01" level=info msg="PIDFile location set to: ./tyk-gateway.pid" prefix=main
time="Nov 27 09:06:01" level=info msg="Initialising Tyk REST API Endpoints" prefix=main
time="Nov 27 09:06:01" level=info msg="--> Standard listener (http)" port=":8001" prefix=main
time="Nov 27 09:06:01" level=warning msg="Starting HTTP server on:[::]:8001" prefix=main
time="Nov 27 09:06:01" level=info msg="--> Standard listener (http)" port=":8000" prefix=main
time="Nov 27 09:06:01" level=warning msg="Starting HTTP server on:[::]:8000" prefix=main
time="Nov 27 09:06:01" level=info msg="Registering gateway node with Dashboard" prefix=dashboard
time="Nov 27 09:06:01" level=info msg="Starting Poller" prefix=host-check-mgr
time="Nov 27 09:06:01" level=info msg="--> [REDIS] Creating single-node client"
time="Nov 27 09:06:01" level=info msg="--> [REDIS] Creating single-node client"
time="Nov 27 09:06:01" level=info msg="Node Registered" id=0603603a-09fa-40af-64f4-757fd8248cb9 prefix=dashboard
time="Nov 27 09:06:01" level=info msg="Initialising distributed rate limiter" prefix=main
time="Nov 27 09:06:01" level=info msg="Tyk Gateway started (v3.0.0)" prefix=main
time="Nov 27 09:06:01" level=info msg="--> Listening on address: (open interface)" prefix=main
time="Nov 27 09:06:01" level=info msg="--> Listening on port: 8000" prefix=main
time="Nov 27 09:06:01" level=info msg="--> PID: 1" prefix=main
time="Nov 27 09:06:01" level=info msg="Starting gateway rate limiter notifications..."
time="Nov 27 09:06:01" level=info msg="Loading policies" prefix=main
time="Nov 27 09:06:01" level=info msg="Using Policies from Dashboard Service" prefix=main
time="Nov 27 09:06:01" level=info msg="Mutex lock acquired... calling" prefix=policy
time="Nov 27 09:06:01" level=info msg="Calling dashboard service for policy list" prefix=policy
time="Nov 27 09:06:01" level=info msg="Processing policy list" prefix=policy
time="Nov 27 09:06:01" level=info msg="Policies found (0 total):" prefix=main
time="Nov 27 09:06:01" level=info msg="Detected 0 APIs" prefix=main
time="Nov 27 09:06:01" level=warning msg="No API Definitions found, not reloading" prefix=main
deploying httpbin as mock upstream to default ns
service/httpbin created
deployment.apps/httpbin created
Waiting for deployment "httpbin" rollout to finish: 0 of 1 updated replicas are available...
deployment "httpbin" successfully rolled out
===> bootstrapping tyk dashboard (initial org + user)
sh ./ci/bootstrap_org.sh
cat bootstrapped
[Nov 27 09:08:31]  WARN toth/tothic: no TYK_IB_SESSION_SECRET environment variable is set. The default cookie store is not available and any calls will fail. Ignore this warning if you are using a different store.

Loading configuration from /etc/tyk-dashboard/dash.json

*************** ORGANISATIONS ***************
ORG NAME	ORG ID
*********************************************
No organisation is found.

Creating New Organisation
ORG DATA: {"Status":"OK","Message":"Org created","Meta":"5fc0c2103d490400019647fd"}
ORG ID: 5fc0c2103d490400019647fd

Adding New User
USER AUTHENTICATION CODE: 138053dc3fb9414658a3e0d49cd12410
NEW ID: 5fc0c2103d7d1c87e44f14f4

DONE
************************************
Login at http://localhost:3000/
User: crvhlecz9x@default.com
Pass: b3m3vrfb
************************************
===> setting operator dash secrets
sh ./ci/operator_pro_secrets.sh
creating namespace tyk-operator-system
namespace/tyk-operator-system created
secret/tyk-operator-conf created
{
  "TYK_AUTH": "MTM4MDUzZGMzZmI5NDE0NjU4YTNlMGQ0OWNkMTI0MTA=",
  "TYK_MODE": "cHJv",
  "TYK_ORG": "NWZjMGMyMTAzZDQ5MDQwMDAxOTY0N2Zk",
  "TYK_URL": "aHR0cDovL2Rhc2hib2FyZC50eWtwcm8tY29udHJvbC1wbGFuZS5zdmMuY2x1c3Rlci5sb2NhbDozMDAw"
}
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -mod=vendor -a -o manager.linux main.go
docker build -f cross.Dockerfile . -t tykio/tyk-operator:test
Sending build context to Docker daemon  370.3MB
Step 1/5 : FROM gcr.io/distroless/static:nonroot
 ---> aa99000bc55d
Step 2/5 : WORKDIR /
 ---> Using cache
 ---> 2877cb3cb6fe
Step 3/5 : COPY manager.linux manager
 ---> Using cache
 ---> dfd51e339896
Step 4/5 : USER nonroot:nonroot
 ---> Using cache
 ---> 886af638ed5a
Step 5/5 : ENTRYPOINT ["/manager"]
 ---> Using cache
 ---> 34f5baa0810d
Successfully built 34f5baa0810d
Successfully tagged tykio/tyk-operator:test
/Volumes/code/gosrc/bin/controller-gen "crd:trivialVersions=true" rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases
/Volumes/code/gosrc/bin/kustomize build config/crd > ./helm/crds/crds.yaml
/Volumes/code/gosrc/bin/kustomize build config/helm |go run hack/pre_helm.go > ./helm/templates/all.yaml
===> installing operator with helmr
kind load docker-image tykio/tyk-operator:test
Image: "tykio/tyk-operator:test" with ID "sha256:34f5baa0810d4c04e20d5cfc22265aea6ac510b33ab04022dfb58fe06dbeec20" not yet present on node "kind-worker", loading...
Image: "tykio/tyk-operator:test" with ID "sha256:34f5baa0810d4c04e20d5cfc22265aea6ac510b33ab04022dfb58fe06dbeec20" not yet present on node "kind-worker2", loading...
Image: "tykio/tyk-operator:test" with ID "sha256:34f5baa0810d4c04e20d5cfc22265aea6ac510b33ab04022dfb58fe06dbeec20" not yet present on node "kind-control-plane", loading...
helm install ci ./helm --values ./ci/helm_values.yaml -n tyk-operator-system --wait
NAME: ci
LAST DEPLOYED: Fri Nov 27 12:09:39 2020
NAMESPACE: tyk-operator-system
STATUS: deployed
REVISION: 1
TEST SUITE: None
NOTES:
You have deployed the tyk-operator! See https://github.com/TykTechnologies/tyk-operator for more information.
******** Successful boot strapped pro dev env ************

</pre>
</p>
</details>

This will 
- deploy cert-manager
- deploy tyk-pro
- creates an org that you can use to log into your dashboard. run `cat ./bootstrapped` to see the org credentials
- setup secret that will be used by the operator to access the deployed dashboard

```
export TYK_AUTH=$(awk '/USER AUTHENTICATION CODE: /{print $NF}' bootstrapped)
export TYK_ORG=$(awk '/ORG ID: /{print $NF}' bootstrapped)
export TYK_MODE=pro
export TYK_URL=http://localhost:3000
```

In this example, we set the `TYK_URL=http://localhost:3000`

Lets expose the dashboard on port 3000 so that the operator may talk to it, and so we can log in.

```
kubectl port-forward -n tykpro-control-plane svc/dashboard 3000:3000
```

#### OSS Inside Cluster

If you already have a Gateway running on port 8080 in the cluster, skip to step 2.

Run all these commands from the root directory of this project.

A) Run Redis

```kubernetes
kubectl apply -f playground/redis/redis.yaml
```

B) Run HttpBin

```kubernetes
kubectl apply -f playground/httpbin/httpbin.yaml
```

C) Create ConfigMap

```kubernetes
kubectl create configmap tyk-conf --from-file ./playground/gateway/confs/tyk.json
```

D) Deploy GW and create Service

```kubernetes
$ kubectl apply -f playground/gateway/deployment.yaml
$ kubectl apply -f playground/gateway/service.yaml
```

E) Expose Gateway locally
```bash
kubectl port-forward svc/tyk 8000:8080
```

F) Test:
```bash
$ curl localhost:8000/hello
{
  "description": "Tyk GW",
  "details": {
    "redis": {
      "componentType": "datastore",
      "status": "pass",
      "time": "2020-09-08T19:42:05Z"
    }
  },
  "status": "pass",
  "version": "v3.0.0"
}
```

### 2. Start the operator (on host/dev machine)

```
make run ENABLE_WEBHOOKS=false
```

```
make run ENABLE_WEBHOOKS=false
/home/asoorm/go/bin/controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./..."
go fmt ./...
go vet ./...
/home/asoorm/go/bin/controller-gen "crd:trivialVersions=true,crdVersions=v1" rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases
 ENABLE_WEBHOOKS=false go run ./main.gopro TYK_TLS_INSECURE_SKIP_VERIFY= TYK_ADMIN_AUTH= TYK_AUTH=2dcc0707f5ff42764ecf2fb84ea23cd6
2020-10-15T11:32:04.789+0100    INFO    controller-runtime.metrics      metrics server is starting to listen    {"addr": ":8080"}
2020-10-15T11:32:04.790+0100    INFO    setup   starting manager
2020-10-15T11:32:04.790+0100    INFO    controller-runtime.manager      starting metrics server {"path": "/metrics"}
2020-10-15T11:32:04.790+0100    INFO    controller-runtime.controller   Starting EventSource    {"controller": "securitypolicy", "source": "kind source: /, Kind="}
2020-10-15T11:32:04.790+0100    INFO    controller-runtime.controller   Starting EventSource    {"controller": "apidefinition", "source": "kind source: /, Kind="}
2020-10-15T11:32:04.790+0100    INFO    controller-runtime.controller   Starting EventSource    {"controller": "webhook", "source": "kind source: /, Kind="}
2020-10-15T11:32:04.790+0100    INFO    controller-runtime.controller   Starting EventSource    {"controller": "organization", "source": "kind source: /, Kind="}
2020-10-15T11:32:04.890+0100    INFO    controller-runtime.controller   Starting Controller     {"controller": "organization"}
2020-10-15T11:32:04.890+0100    INFO    controller-runtime.controller   Starting workers        {"controller": "organization", "worker count": 1}
2020-10-15T11:32:04.890+0100    INFO    controller-runtime.controller   Starting Controller     {"controller": "securitypolicy"}
2020-10-15T11:32:04.890+0100    INFO    controller-runtime.controller   Starting workers        {"controller": "securitypolicy", "worker count": 1}
2020-10-15T11:32:04.890+0100    INFO    controller-runtime.controller   Starting Controller     {"controller": "apidefinition"}
2020-10-15T11:32:04.890+0100    INFO    controller-runtime.controller   Starting workers        {"controller": "apidefinition", "worker count": 1}
2020-10-15T11:32:04.890+0100    INFO    controller-runtime.controller   Starting Controller     {"controller": "webhook"}
2020-10-15T11:32:04.890+0100    INFO    controller-runtime.controller   Starting workers        {"controller": "webhook", "worker count": 1}
```

### 3. Generate & install the CRDs and run the Operator from source

A) in root of directory, run this command to install the CRDs into the cluster.
```bash
make install
```

### 4. Add API definition through command line

Add an API definition using kubectl apply

```bash
$ kubectl apply -f config/samples/httpbin.yaml
apidefinition.tyk.tyk.io/httpbin created
```

### 4. Get the new resource

```bash
kubectl get tykapis                                        
NAME      PROXY.LISTENPATH   PROXY.TARGETURL      ENABLED
httpbin   /httpbin           http://httpbin.org   true
```

### 5. Try it out

```bash
curl localhost:8000/httpbin/get
{
  "args": {}, 
  "headers": {
    "Accept": "*/*", 
    "Accept-Encoding": "gzip", 
    "Host": "httpbin.org", 
    "User-Agent": "curl/7.71.1", 
    "X-Amzn-Trace-Id": "Root=1-5f64fdeb-db917d73dd04463839339047"
  }, 
  "origin": "127.0.0.1, 94.14.220.241", 
  "url": "http://httpbin.org/get"
}
```

### 5. Enable Authentication

```bash
cat ./config/samples/httpbin.yaml | yq w - spec.use_keyless false > ./config/samples/httpbin_protected.yaml
```

### 6. Try it out

```bash
kubectl apply -f ./config/samples/httpbin_protected.yaml                            
apidefinition.tyk.tyk.io/httpbin configured

curl localhost:8000/httpbin/get                   
{
  "error": "Authorization field missing"
}                    
```

---

## Scrapbook

After modifying the *_types.go file always run the following command to update the generated code for that resource type:
```
make generate
```

Once the API is defined with spec/status fields and CRD validation markers, the CRD manifests can be generated and 
updated with the following command:

```
make manifests
```

Register the CRD

```
make install
```

Run the operator locally, outside the cluster

```
make run ENABLE_WEBHOOKS=false
```

### Compile and load local docker image:

Minikube:
```
# docker build with Docker daemon of minikube
eval $(minikube docker-env)
docker build . -t controller:latest
```

Kind:
```
docker build . -t controller:latest
kind load docker-image controller:latest
```

Deploy it to the cluster:
```
make deploy IMG=controller:latest
```


### Highlights

1. The Tyk-Operator uses the [finalizer](https://book.kubebuilder.io/reference/using-finalizers.html) pattern for deleting CRs from the cluster.
