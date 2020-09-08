# dev instructions

## Prerequisites

- Have a Kubernetes (v.18.0+) cluster running locally
- Golang (1.15.0+)

## 1. Run the Tyk Dev environment

If you already have a Gateway running on port 8080 in the cluster, skip to step 2.

Run all these commands from the root directory of this project.

A) Run Redis
```kubernetes
kubectl apply -f playground/redis/redis.yaml
```

B) Run Httpbin
```kubernetes
kubectl apply -f playground/httpbin/httpbin.yaml
```

C) Create ConfigMap
```kubernetes
kubectl create configmap tyk-conf --from-file ./playground/gateway/tyk.json
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
$ curl localhost:8080/hello
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

## 2. Generate the CRDs and run the Operator from source

A) in root of directory, run this command to generate the CRD.
These commands are explained below
```bash
make generate;make manifests;make install
```

B) Run the operator
```bash
make run ENABLE_WEBHOOKS=false
```

### Explanations of make commands
After modifying the *_types.go file always run the following command to update the generated code for that resource type:
```
make generate
```


Generating CRD manifests

Once the API is defined with spec/status fields and CRD validation markers, the CRD manifests can be generated and updated with the following command:

```
make manifests
```

This makefile target will invoke controller-gen to generate the CRD manifests at config/crd/bases/cache.example.com_memcacheds.yaml


Register the CRD

```
make install
```

Run the operator locally, outside the cluster

```
make run ENABLE_WEBHOOKS=false
```
