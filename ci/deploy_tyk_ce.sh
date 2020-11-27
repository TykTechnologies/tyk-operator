#!/bin/bash

set -e

NAMESPACE=tykce-control-plane
PRODIR=${PWD}/ci/tyk-ce

echo "creating namespace ${NAMESPACE}"
if OUTPUT=$(kubectl get namespaces 2> /dev/null | grep "${NAMESPACE}") ; then
   echo "namespace ${NAMESPACE} already exists"
else
  echo "creating namespace ${NAMESPACE}"
  kubectl create namespace ${NAMESPACE}
fi

echo "deploying gRPC plugin server"
kubectl apply -f "${PRODIR}/../grpc-plugin" -n ${NAMESPACE}
kubectl rollout status  deployment/grpc-plugin -n ${NAMESPACE}  

echo "deploying databases"
kubectl apply -f "${PRODIR}/redis" -n ${NAMESPACE}

echo "waiting for redis"
kubectl rollout status  deployment/redis -n ${NAMESPACE} 

echo "creating configmaps"
kubectl create configmap -n ${NAMESPACE} tyk-conf --from-file "${PRODIR}/gateway/confs/tyk.json"

echo "deploying gateway"
kubectl apply -f "${PRODIR}/gateway/gateway.yaml" -n ${NAMESPACE}
kubectl rollout status  deployment/tyk -n ${NAMESPACE} 

echo "gateway logs"
kubectl logs svc/tyk -n ${NAMESPACE}

echo "deploying httpbin as mock upstream to default ns"
kubectl apply -f "${PWD}/ci/upstreams"
kubectl rollout status  deployment/httpbin 
