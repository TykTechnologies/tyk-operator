#!/bin/bash

set -e

NAMESPACE=tykce-control-plane
PRODIR=${PWD}/ci/tyk-ce

echo "creating namespace ${NAMESPACE}"
kubectl create namespace ${NAMESPACE}

echo "deploying databases"
kubectl apply -f "${PRODIR}/redis" -n ${NAMESPACE}

echo "waiting for redis"
kubectl wait deployment/redis -n ${NAMESPACE} --for condition=available --timeout=60s

echo "creating configmaps"
kubectl create configmap -n ${NAMESPACE} tyk-conf --from-file "${PRODIR}/gateway/confs/tyk.json"

echo "deploying gateway"
kubectl apply -f "${PRODIR}/gateway/gateway.yaml" -n ${NAMESPACE}
kubectl wait deployment/tyk -n ${NAMESPACE} --for condition=available

echo "gateway logs"
kubectl logs svc/tyk -n ${NAMESPACE}

echo "deploying httpbin as mock upstream to default ns"
kubectl apply -f "${PWD}/ci/upstreams"
kubectl wait deployment/httpbin --for condition=available
