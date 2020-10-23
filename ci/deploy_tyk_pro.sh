#!/bin/bash

set -e

NAMESPACE=tykpro-control-plane
PRODIR=${PWD}/ci/tyk-pro
ADMINSECRET="54321"
TIMEOUT=60s



if OUTPUT=$(kubectl get namespaces 2> /dev/null | grep "${NAMESPACE}") ; then
   echo "namespace ${NAMESPACE} already exists"
else
  echo "creating namespace ${NAMESPACE}"
  kubectl create namespace ${NAMESPACE}
fi

echo "deploying databases"
kubectl apply -f "${PRODIR}/mongo/mongo.yaml" -n ${NAMESPACE}
kubectl apply -f "${PRODIR}/redis" -n ${NAMESPACE}

echo "waiting for redis"
kubectl wait deployment/redis -n ${NAMESPACE} --for condition=available --timeout=${TIMEOUT}
echo "waiting for mongo"
kubectl wait deployment/mongo -n ${NAMESPACE} --for condition=available --timeout=${TIMEOUT}

echo "creating configmaps"
kubectl create configmap -n ${NAMESPACE} dash-conf --dry-run --from-file "${PRODIR}/dashboard/confs/dash.json" -o yaml | kubectl apply -f -
kubectl create configmap -n ${NAMESPACE} tyk-conf --dry-run --from-file "${PRODIR}/gateway/confs/tyk.json" -o yaml | kubectl apply -f -

echo "setting dashboard secrets"
kubectl create secret -n ${NAMESPACE} generic dashboard --dry-run --from-literal "license=${TYK_DB_LICENSEKEY}" --from-literal "adminSecret=${ADMINSECRET}" -o yaml | kubectl apply -f -
kubectl get secret/dashboard -n tykpro-control-plane -o json | jq '.data.license'

echo "deploying dashboard"
kubectl apply -f "${PRODIR}/dashboard/dashboard.yaml" -n ${NAMESPACE}
kubectl wait deployment/dashboard -n ${NAMESPACE} --for condition=available --timeout=${TIMEOUT}

echo "deploying gateway"
kubectl apply -f "${PRODIR}/gateway/gateway.yaml" -n ${NAMESPACE}
kubectl wait deployment/tyk -n ${NAMESPACE} --for condition=available --timeout=${TIMEOUT}

echo "dashboard logs"
kubectl logs svc/dashboard -n ${NAMESPACE}

echo "gateway logs"
kubectl logs svc/tyk -n ${NAMESPACE}

echo "deploying httpbin as mock upstream to default ns"
kubectl apply -f "${PWD}/ci/upstreams"
kubectl wait deployment/httpbin --for condition=available --timeout=${TIMEOUT}
