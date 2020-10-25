#!/bin/bash

set -e

NAMESPACE=tyk-operator-system

if OUTPUT=$(kubectl get namespaces 2> /dev/null | grep "${NAMESPACE}") ; then
   echo "namespace ``${NAMESPACE}`` already exists"
else
  echo "creating namespace ``${NAMESPACE}``"
  kubectl create namespace ``${NAMESPACE}``
fi

TYK_AUTH=$(awk -F ':' '/USER AUTHENTICATION CODE: /{ print $2 }' bootstrapped | tr -d '[:space:]')
TYK_ORG=$(awk -F ':' '/ORG ID: /{ print $2 }' bootstrapped | tr -d '[:space:]')
TYK_MODE=pro
TYK_URL=http://dashboard.tykpro-control-plane.svc.cluster.local:3000

kubectl create secret -n ${NAMESPACE} generic tyk-operator-conf \
  --from-literal "tykAuth=${TYK_AUTH}" \
  --from-literal "tykOrg=${TYK_ORG}" \
  --from-literal "tykMode=${TYK_MODE}" \
  --from-literal "tykURL=${TYK_URL}"

kubectl get secret/tyk-operator-conf -n ${NAMESPACE} -o json | jq '.data'
