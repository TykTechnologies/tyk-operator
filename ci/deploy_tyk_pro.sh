#!/bin/bash

set -e

NAMESPACE=tykpro-control-plane
PRODIR=${PWD}/ci/tyk-pro

echo "creating namespace ${NAMESPACE}"
kubectl create namespace ${NAMESPACE}

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

echo "getting dash license key"
echo -n "${TYK_DB_LICENSEKEY}" > ./license.txt
kubectl create secret -n ${NAMESPACE} generic dashboard --from-file=./license.txt

echo "deploying dashboard & gateway"
kubectl apply -f "${PRODIR}/dashboard/dashboard.yaml" -n ${NAMESPACE}
kubectl apply -f "${PRODIR}/gateway/gateway.yaml" -n ${NAMESPACE}

echo "waiting for dashboard"
kubectl wait deployment/dashboard -n ${NAMESPACE} --for condition=available
echo "waiting for gateway"
kubectl wait deployment/tyk -n ${NAMESPACE} --for condition=available

kubectl logs svc/dashboard -n ${NAMESPACE}

echo "creating an organization"
kubectl exec -n ${NAMESPACE} svc/dashboard -- /opt/tyk-dashboard/tyk-analytics bootstrap --conf=/etc/tyk-dashboard/dash.json --create-org
