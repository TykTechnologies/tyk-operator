#!/bin/bash

set -e

NAMESPACE=tykpro-control-plane
PRODIR=${PWD}/ci/tyk-pro
ADMINSECRET="54321"

echo "creating namespace ${NAMESPACE}"
kubectl create namespace ${NAMESPACE}

echo "deploying gRPC plugin server"
kubectl apply -f "${PRODIR}/../grpc-plugin" -n ${NAMESPACE}
kubectl wait deployment/grpc_plugin -n ${NAMESPACE} --for condition=available --timeout=10s

echo "deploying databases"
kubectl apply -f "${PRODIR}/mongo/mongo.yaml" -n ${NAMESPACE}
kubectl apply -f "${PRODIR}/redis" -n ${NAMESPACE}

echo "waiting for redis"
kubectl wait deployment/redis -n ${NAMESPACE} --for condition=available --timeout=60s
echo "waiting for mongo"
kubectl wait deployment/mongo -n ${NAMESPACE} --for condition=available --timeout=60s

echo "creating configmaps"
kubectl create configmap -n ${NAMESPACE} dash-conf --from-file "${PRODIR}/dashboard/confs/dash.json"
kubectl create configmap -n ${NAMESPACE} tyk-conf --from-file "${PRODIR}/gateway/confs/tyk.json"

echo "setting dashboard secrets"
kubectl create secret -n ${NAMESPACE} generic dashboard --from-literal "license=${TYK_DB_LICENSEKEY}" --from-literal "adminSecret=${ADMINSECRET}"
kubectl get secret/dashboard -n tykpro-control-plane -o json | jq '.data.license'

echo "deploying dashboard"
kubectl apply -f "${PRODIR}/dashboard/dashboard.yaml" -n ${NAMESPACE}
kubectl wait deployment/dashboard -n ${NAMESPACE} --for condition=available

echo "deploying gateway"
kubectl apply -f "${PRODIR}/gateway/gateway.yaml" -n ${NAMESPACE}
kubectl wait deployment/tyk -n ${NAMESPACE} --for condition=available

echo "dashboard logs"
kubectl logs svc/dashboard -n ${NAMESPACE}

echo "gateway logs"
kubectl logs svc/tyk -n ${NAMESPACE}

echo "deploying httpbin as mock upstream to default ns"
kubectl apply -f "${PWD}/ci/upstreams"
kubectl wait deployment/httpbin --for condition=available
