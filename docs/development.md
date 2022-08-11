# Local Development Environment

__NOTE__: These instructions are for kind clusters which is the recommended way
for development

## Prerequisites

Note that these are tools which we use for development of this project.

| tool | version|
|-------|---------|
| [kind](https://kind.sigs.k8s.io) | v0.9.0+ |
|[go](https://go.dev/doc/install) |1.16.3+|
| [operator-sdk](https://sdk.operatorframework.io/)| v1.7.2+ |
|controller-gen | v0.4.1+|
| [kustomize](https://kustomize.io/) |v3.8.7+|
| [helm](https://helm.sh/)|v3.5.4+|
| [Docker](https://docs.docker.com/)| 19.03.13+|


### 0. Create cluster

```shell
make create-kind-cluster
```

This will create a 3 node cluster. 1 control-plane and 2 worker nodes.

### 1. Boot strap the dev environment

#### Booting tyk community edition

```shell
make boot-ce IMG=tykio/tyk-operator:test
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
deployment "cert-manager-cainjector" successfully rolled out
kubectl rollout status  deployment/cert-manager-webhook -n cert-manager
Waiting for deployment "cert-manager-webhook" rollout to finish: 0 of 1 updated replicas are available...
deployment "cert-manager-webhook" successfully rolled out
===> installing tyk-ce
sh ./ci/deploy_tyk_ce.sh
creating namespace tykce-control-plane
creating namespace tykce-control-plane
namespace/tykce-control-plane created
deploying gRPC plugin server
service/grpc-plugin created
deployment.apps/grpc-plugin created
Waiting for deployment "grpc-plugin" rollout to finish: 0 out of 1 new replicas have been updated...
Waiting for deployment "grpc-plugin" rollout to finish: 0 of 1 updated replicas are available...
deployment "grpc-plugin" successfully rolled out
deploying databases
deployment.apps/redis created
service/redis created
waiting for redis
Waiting for deployment "redis" rollout to finish: 0 of 1 updated replicas are available...
deployment "redis" successfully rolled out
creating configmaps
configmap/tyk-conf created
deploying gateway
service/tyk created
service/gw created
deployment.apps/tyk created
Waiting for deployment "tyk" rollout to finish: 0 of 1 updated replicas are available...
deployment "tyk" successfully rolled out
gateway logs
time="Nov 27 10:03:01" level=info msg="Tyk API Gateway v3.0.0" prefix=main
time="Nov 27 10:03:01" level=warning msg="Insecure configuration allowed" config.allow_insecure_configs=true prefix=checkup
time="Nov 27 10:03:01" level=info msg="Starting Poller" prefix=host-check-mgr
time="Nov 27 10:03:01" level=info msg="gRPC dispatcher was initialized" prefix=coprocess
time="Nov 27 10:03:01" level=info msg="PIDFile location set to: ./tyk-gateway.pid" prefix=main
time="Nov 27 10:03:01" level=info msg="Initialising Tyk REST API Endpoints" prefix=main
time="Nov 27 10:03:01" level=info msg="--> Standard listener (http)" port=":8001" prefix=main
time="Nov 27 10:03:01" level=info msg="--> [REDIS] Creating single-node client"
time="Nov 27 10:03:01" level=warning msg="Starting HTTP server on:[::]:8001" prefix=main
time="Nov 27 10:03:01" level=info msg="--> Standard listener (http)" port=":8000" prefix=main
time="Nov 27 10:03:01" level=warning msg="Starting HTTP server on:[::]:8000" prefix=main
time="Nov 27 10:03:01" level=info msg="Initialising distributed rate limiter" prefix=main
time="Nov 27 10:03:01" level=info msg="Tyk Gateway started (v3.0.0)" prefix=main
time="Nov 27 10:03:01" level=info msg="--> Listening on address: (open interface)" prefix=main
time="Nov 27 10:03:01" level=info msg="--> Listening on port: 8000" prefix=main
time="Nov 27 10:03:01" level=info msg="--> PID: 1" prefix=main
time="Nov 27 10:03:01" level=info msg="Starting gateway rate limiter notifications..."
time="Nov 27 10:03:01" level=info msg="Loading policies" prefix=main
time="Nov 27 10:03:01" level=info msg="Policies found (1 total):" prefix=main
time="Nov 27 10:03:01" level=info msg=" - default" prefix=main
time="Nov 27 10:03:01" level=info msg="Loading API Specification from apps/app_sample.json"
time="Nov 27 10:03:01" level=info msg="Detected 1 APIs" prefix=main
time="Nov 27 10:03:01" level=info msg="Loading API configurations." prefix=main
time="Nov 27 10:03:01" level=info msg="Tracking hostname" api_name="Tyk Test API" domain="(no host)" prefix=main
time="Nov 27 10:03:01" level=info msg="Initialising Tyk REST API Endpoints" prefix=main
time="Nov 27 10:03:01" level=info msg="API bind on custom port:0" prefix=main
time="Nov 27 10:03:01" level=info msg="Checking security policy: Token" api_id=1 api_name="Tyk Test API" org_id=default
time="Nov 27 10:03:01" level=info msg="API Loaded" api_id=1 api_name="Tyk Test API" org_id=default prefix=gateway server_name=-- user_id=-- user_ip=--
time="Nov 27 10:03:01" level=info msg="Loading uptime tests..." prefix=host-check-mgr
time="Nov 27 10:03:01" level=info msg="Initialised API Definitions" prefix=main
time="Nov 27 10:03:01" level=info msg="API reload complete" prefix=main
time="Nov 27 10:03:01" level=info msg="--> [REDIS] Creating single-node client"
deploying httpbin as mock upstream to default ns
service/httpbin created
deployment.apps/httpbin created
Waiting for deployment "httpbin" rollout to finish: 0 of 1 updated replicas are available...
deployment "httpbin" successfully rolled out
setting operator secrets
sh ./ci/operator_ce_secrets.sh
creating namespace tyk-operator-system
namespace/tyk-operator-system created
secret/tyk-operator-conf created
{
  "TYK_AUTH": "Zm9v",
  "TYK_MODE": "b3Nz",
  "TYK_ORG": "bXlvcmc=",
  "TYK_URL": "aHR0cDovL3R5ay50eWtjZS1jb250cm9sLXBsYW5lLnN2Yy5jbHVzdGVyLmxvY2FsOjgwMDE="
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
Image: "tykio/tyk-operator:test" with ID "sha256:34f5baa0810d4c04e20d5cfc22265aea6ac510b33ab04022dfb58fe06dbeec20" not yet present on node "kind-worker2", loading...
Image: "tykio/tyk-operator:test" with ID "sha256:34f5baa0810d4c04e20d5cfc22265aea6ac510b33ab04022dfb58fe06dbeec20" not yet present on node "kind-control-plane", loading...
Image: "tykio/tyk-operator:test" with ID "sha256:34f5baa0810d4c04e20d5cfc22265aea6ac510b33ab04022dfb58fe06dbeec20" not yet present on node "kind-worker", loading...
helm install ci ./helm --values ./ci/helm_values.yaml -n tyk-operator-system --wait
NAME: ci
LAST DEPLOYED: Fri Nov 27 13:07:44 2020
NAMESPACE: tyk-operator-system
STATUS: deployed
REVISION: 1
TEST SUITE: None
NOTES:
You have deployed the tyk-operator! See https://github.com/TykTechnologies/tyk-operator for more information.
******** Successful boot strapped ce dev env ************
</pre>
</p>
</details>

This will 
- deploy cert-manager
- deploy tyk-ce
- setup secret that will be used by the operator to access the deployed gateway


#### Booting tyk Pro edition

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


### 2. Checking if the environment is working


#### For PRO mode

***expose our gateway locally***

Run this in a separate terminal

```
kubectl port-forward svc/gateway-svc-pro-tyk-pro 8080:8080  -n tykpro-control-plane
```
<details><summary>SHOW EXPECTED OUTPUT</summary>
<p>
<pre>
Forwarding from 127.0.0.1:8080 -> 8080
Forwarding from [::1]:8080 -> 8080
</pre>
</p>
</details>

***check if the gateway is properly configured***

In a separate terminal run

```
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
        "dashboard": {
            "status": "pass",
            "componentType": "system",
            "time": "2020-11-27T09:29:09Z"
        },
        "redis": {
            "status": "pass",
            "componentType": "datastore",
            "time": "2020-11-27T09:29:09Z"
        }
    }
}
</pre>
</p>
</details>

***check if the operator is working***

create your first api resource

```
kubectl apply -f config/samples/httpbin.yaml 
```
<details><summary>SHOW EXPECTED OUTPUT</summary>
<p>
<pre>
apidefinition.tyk.tyk.io/httpbin created
</pre>
</p>
</details>

check that your api definition was applied and it works

```
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


#### For CE(Community edition) mode

***expose our gateway locally***

Run this in a separate terminal

```
kubectl port-forward svc/gateway-svc-ce-tyk-headless 8080:8080  -n tykce-control-plane
```
<details><summary>SHOW EXPECTED OUTPUT</summary>
<p>
<pre>
Forwarding from 127.0.0.1:8080 -> 8080
Forwarding from [::1]:8080 -> 8080
</pre>
</p>
</details>

***check if the gateway is properly configured***

In a separate terminal run

```
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

***check if the operator is working***

create your first api resource

```
kubectl apply -f config/samples/httpbin.yaml 
```
<details><summary>SHOW EXPECTED OUTPUT</summary>
<p>
<pre>
apidefinition.tyk.tyk.io/httpbin created
</pre>
</p>
</details>

check that your api definition was applied and it works

```
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

---

## Scrapbook

After making changes to the controller

```
make scrap IMG=tykio/tyk-operator:test
```
This will 

- generate manifests
- build and tag the new changes into a docker image
- load the image to kind cluster
- uninstall the previous controller deployment
- install the new controller  deployment


## Delete cluster

Delete kind cluster using following command

```
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


## Loading images into kind cluster
You can load images in the kind cluster required for Tyk Pro/CE installation. It will download all the images once and will be reused during next installtions, unless cluster is deleted.

You just need to set `TYK_OPERATOR_PRELOAD_IMAGES` environment variable to true before running make boot-ce/make boot-pro.



## Highlights

1. The Tyk-Operator uses the [finalizer](https://book.kubebuilder.io/reference/using-finalizers.html) pattern for deleting CRs from the cluster.
