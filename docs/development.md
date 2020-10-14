# Local Development Environment

## Prerequisites

- Have a Kubernetes (v.18.0+) cluster running locally
- Golang (1.15.0+)

## Helpful

- https://ngrok.com/
- yq https://mikefarah.gitbook.io/yq/

### 1. Run an OSS Tyk Gateway

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

### 2. Generate & install the CRDs and run the Operator from source

A) in root of directory, run this command to generate the CRD.
```bash
make generate; make manifests; make install
```

B) Run the operator
```bash
make run ENABLE_WEBHOOKS=false
```

### 3. Add API definition through command line

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
