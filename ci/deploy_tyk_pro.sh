#!/bin/bash

NAMESPACE=tykpro-control-plane
PRODIR=${PWD}/ci/tyk-pro

echo "creating namespace ${NAMESPACE}"
kubectl create namespace ${NAMESPACE}

echo "deploying databases"
kubectl apply -f "${PRODIR}/mongo/mongo.yaml" -n ${NAMESPACE}
kubectl apply -f "${PRODIR}/redis" -n ${NAMESPACE}

echo "waiting for redis"
while [[ $(kubectl get pods -l name=redis -n ${NAMESPACE} -o 'jsonpath={..status.conditions[?(@.type=="Ready")].status}') != "True" ]]; do echo "  ...redis starting" && sleep 10; done
echo "waiting for mongo"
while [[ $(kubectl get pods -l name=mongo -n ${NAMESPACE} -o 'jsonpath={..status.conditions[?(@.type=="Ready")].status}') != "True" ]]; do echo "  ...mongo starting" && sleep 10; done

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
while [[ $(kubectl get pods -l name=dashboard -n ${NAMESPACE} -o 'jsonpath={..status.conditions[?(@.type=="Ready")].status}') != "True" ]]; do echo "  ...dashboard starting" && sleep 10; done
echo "waiting for gateway"
while [[ $(kubectl get pods -l name=tyk -n ${NAMESPACE} -o 'jsonpath={..status.conditions[?(@.type=="Ready")].status}') != "True" ]]; do echo "  ...gateway starting" && sleep 10; done

kubectl logs svc/dashboard -n ${NAMESPACE}
echo "${TYK_DB_LICENSEKEY}"
