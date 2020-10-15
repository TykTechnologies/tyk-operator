# Local Development Environment

## Prerequisites

- Have a Kubernetes (v.18.0+) cluster running locally
- Golang (1.15.0+)

## Helpful

- https://ngrok.com/
- yq https://mikefarah.gitbook.io/yq/

### 1. Start or Connect to Tyk

Decide whether you wish to run against a Pro or OSS installation.
You may use **any** existing Tyk installation. Whether running in-cluster or on host machine, or in Tyk Cloud.

The operator needs a couple of env vars so that it knows how to speak to the Tyk apis.

```
export TYK_AUTH=REPLACE_WITH_DASH_USER_KEY_OR_FOR_OSS_GW_SECRET
export TYK_ORG=REPLACE_WITH_ORG_ID
export TYK_MODE=pro|oss
export TYK_URL=REPLACE_WITH_DASHBOARD_URL_OR_GATEWAY_ADMIN_URL
```

#### Pro inside cluster

Set your license key inside an env var.
```
export TYK_DB_LICENSEKEY=REPLACE.WITH.YOUR.LICENSE
```

Deploy tyk using the CI script.
```
sh ./ci/deploy_tyk_pro.sh
```

<details><summary>SHOW EXPECTED OUTPUT</summary>
<p>

```
sh ./ci/deploy_tyk_pro.sh
creating namespace tykpro-control-plane
namespace/tykpro-control-plane created
deploying databases
service/mongo created
deployment.apps/mongo created
deployment.apps/redis created
service/redis created
waiting for redis
deployment.apps/redis condition met
waiting for mongo
deployment.apps/mongo condition met
creating configmaps
configmap/dash-conf created
configmap/tyk-conf created
setting dashboard secrets
secret/dashboard created
"OBFUSCATED"
deploying dashboard
service/dashboard created
deployment.apps/dashboard created
deployment.apps/dashboard condition met
deploying gateway
service/tyk created
deployment.apps/tyk created
deployment.apps/tyk condition met
dashboard logs
time="Oct 15 09:47:15" level=warning msg="toth/tothic: no TYK_IB_SESSION_SECRET environment variable is set. The default cookie store is not available and any calls will fail. Ignore this warning if you are using a different store."

time="Oct 15 09:47:15" level=info msg="Tyk Analytics Dashboard v3.0.1"
time="Oct 15 09:47:15" level=info msg="Copyright Tyk Technologies Ltd 2020"
time="Oct 15 09:47:15" level=info msg="https://www.tyk.io"
time="Oct 15 09:47:15" level=info msg="Using /etc/tyk-dashboard/dash.json for configuration"
time="Oct 15 09:47:15" level=info msg="Listening on port: 3000"
time="Oct 15 09:47:15" level=info msg="Connecting to MongoDB: [mongo.tykpro-control-plane.svc.cluster.local:27017]"
time="Oct 15 09:47:15" level=info msg="Mongo connection established"
time="Oct 15 09:47:15" level=info msg="Creating new Redis connection pool"
time="Oct 15 09:47:15" level=info msg="--> [REDIS] Creating single-node client"
time="Oct 15 09:47:15" level=info msg="Creating new Redis connection pool"
time="Oct 15 09:47:15" level=info msg="--> [REDIS] Creating single-node client"
time="Oct 15 09:47:15" level=info msg="Creating new Redis connection pool"
time="Oct 15 09:47:15" level=info msg="--> [REDIS] Creating single-node client"
time="Oct 15 09:47:15" level=info msg="Creating new Redis connection pool"
time="Oct 15 09:47:15" level=info msg="--> [REDIS] Creating single-node client"
time="Oct 15 09:47:15" level=info msg="Licensing: Setting new license"
time="Oct 15 09:47:15" level=info msg="Licensing: Registering nodes..."
time="Oct 15 09:47:15" level=info msg="Adding available nodes..."
time="Oct 15 09:47:15" level=info msg="Licensing: Checking capabilities"
time="Oct 15 09:47:15" level=info msg="Audit log is disabled in config"
time="Oct 15 09:47:15" level=info msg="Creating new Redis connection pool"
time="Oct 15 09:47:15" level=info msg="--> [REDIS] Creating single-node client"
time="Oct 15 09:47:15" level=info msg="--> Standard listener (http) for dashboard and API"
time="Oct 15 09:47:15" level=info msg="Creating new Redis connection pool"
time="Oct 15 09:47:15" level=info msg="--> [REDIS] Creating single-node client"
time="Oct 15 09:47:15" level=info msg="Starting zeroconf heartbeat"
time="Oct 15 09:47:15" level=info msg="Starting notification handler for gateway cluster"
time="Oct 15 09:47:15" level=info msg="Loading routes..."
time="Oct 15 09:47:15" level=info msg="Initializing Internal TIB"
time="Oct 15 09:47:15" level=info msg="Initializing Identity Cache" prefix="TIB INITIALIZER"
time="Oct 15 09:47:15" level=info msg="Set DB" prefix="TIB REDIS STORE"
time="Oct 15 09:47:15" level=info msg="Using internal Identity Broker. Routes are loaded and available."
gateway logs
time="Oct 15 09:47:18" level=info msg="Tyk API Gateway v3.0.0" prefix=main
time="Oct 15 09:47:18" level=warning msg="Insecure configuration allowed" config.allow_insecure_configs=true prefix=checkup
time="Oct 15 09:47:18" level=info msg="Rich plugins are disabled" prefix=coprocess
time="Oct 15 09:47:18" level=info msg="Starting Poller" prefix=host-check-mgr
time="Oct 15 09:47:18" level=info msg="PIDFile location set to: ./tyk-gateway.pid" prefix=main
time="Oct 15 09:47:18" level=info msg="Initialising Tyk REST API Endpoints" prefix=main
time="Oct 15 09:47:18" level=info msg="--> [REDIS] Creating single-node client"
time="Oct 15 09:47:18" level=info msg="--> Standard listener (http)" port=":8081" prefix=main
time="Oct 15 09:47:18" level=warning msg="Starting HTTP server on:[::]:8081" prefix=main
time="Oct 15 09:47:18" level=info msg="--> Standard listener (http)" port=":8080" prefix=main
time="Oct 15 09:47:18" level=warning msg="Starting HTTP server on:[::]:8080" prefix=main
time="Oct 15 09:47:18" level=info msg="Registering gateway node with Dashboard" prefix=dashboard
time="Oct 15 09:47:18" level=info msg="--> [REDIS] Creating single-node client"
time="Oct 15 09:47:18" level=info msg="Node Registered" id=1290bd4f-baca-4a42-4ab1-b7f96c80d85c prefix=dashboard
time="Oct 15 09:47:18" level=info msg="Initialising distributed rate limiter" prefix=main
time="Oct 15 09:47:18" level=info msg="Tyk Gateway started (v3.0.0)" prefix=main
time="Oct 15 09:47:18" level=info msg="--> Listening on address: (open interface)" prefix=main
time="Oct 15 09:47:18" level=info msg="--> Listening on port: 8080" prefix=main
time="Oct 15 09:47:18" level=info msg="--> PID: 1" prefix=main
time="Oct 15 09:47:18" level=info msg="Starting gateway rate limiter notifications..."
time="Oct 15 09:47:18" level=info msg="Loading policies" prefix=main
time="Oct 15 09:47:18" level=info msg="Using Policies from Dashboard Service" prefix=main
time="Oct 15 09:47:18" level=info msg="Mutex lock acquired... calling" prefix=policy
time="Oct 15 09:47:18" level=info msg="Calling dashboard service for policy list" prefix=policy
time="Oct 15 09:47:18" level=info msg="Processing policy list" prefix=policy
time="Oct 15 09:47:18" level=info msg="Policies found (0 total):" prefix=main
time="Oct 15 09:47:18" level=info msg="Detected 0 APIs" prefix=main
time="Oct 15 09:47:18" level=warning msg="No API Definitions found, not reloading" prefix=main
deploying httpbin as mock upstream to default ns
service/httpbin unchanged
deployment.apps/httpbin created
deployment.apps/httpbin condition met
```

</p>
</details>

Tyk Pro will be installed into the `tykpro-control-plane` namespace.

```
kubectl get all -n tykpro-control-plane
NAME                             READY   STATUS    RESTARTS   AGE
pod/dashboard-854554d94d-w6m8d   1/1     Running   0          3m55s
pod/mongo-57d77d59-4k8qt         1/1     Running   0          3m58s
pod/redis-7f7887fd75-smmfs       1/1     Running   0          3m58s
pod/tyk-5b78db8687-ddhw8         1/1     Running   0          3m52s

NAME                TYPE           CLUSTER-IP       EXTERNAL-IP   PORT(S)          AGE
service/dashboard   LoadBalancer   10.107.79.160    <pending>     3000:31458/TCP   3m55s
service/mongo       ClusterIP      10.103.163.183   <none>        27017/TCP        3m58s
service/redis       ClusterIP      10.103.92.28     <none>        6379/TCP         3m58s
service/tyk         LoadBalancer   10.103.134.195   <pending>     8080:31228/TCP   3m52s

NAME                        READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/dashboard   1/1     1            1           3m55s
deployment.apps/mongo       1/1     1            1           3m58s
deployment.apps/redis       1/1     1            1           3m58s
deployment.apps/tyk         1/1     1            1           3m52s

NAME                                   DESIRED   CURRENT   READY   AGE
replicaset.apps/dashboard-854554d94d   1         1         1       3m55s
replicaset.apps/mongo-57d77d59         1         1         1       3m58s
replicaset.apps/redis-7f7887fd75       1         1         1       3m58s
replicaset.apps/tyk-5b78db8687         1         1         1       3m52s
```

Bootstrap Tyk Dashboard

```
sh ./ci/bootstrap_org.sh
```

A file will be created `./bootstrapped`

```
sh ./ci/bootstrap_org.sh 
pop-os% cat ./bootstrapped 
[Oct 15 09:52:24]  WARN toth/tothic: no TYK_IB_SESSION_SECRET environment variable is set. The default cookie store is not available and any calls will fail. Ignore this warning if you are using a different store.

Loading configuration from /etc/tyk-dashboard/dash.json

*************** ORGANISATIONS ***************
ORG NAME        ORG ID
*********************************************
No organisation is found.

Creating New Organisation
ORG DATA: {"Status":"OK","Message":"Org created","Meta":"5f881bd8286b600001d56a4d"}
ORG ID: 5f881bd8286b600001d56a4d

Adding New User
USER AUTHENTICATION CODE: 2dcc0707f5ff42764ecf2fb84ea23cd6
NEW ID: 5f881bd8940c098198080ef4

DONE
************************************
Login at http://localhost:3000/
User: hc2khpjytf@default.com
Pass: 8qnz6tiz
************************************
```

The operator requires various env vars to be set, so that may interract with Tyk.

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
