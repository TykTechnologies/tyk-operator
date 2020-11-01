#!/bin/bash

set -e

NAMESPACE=tyk-operator-system

if OUTPUT=$(kubectl get namespaces 2> /dev/null | grep "${NAMESPACE}") ; then
   echo "namespace ``${NAMESPACE}`` already exists"
else
  echo "creating namespace ``${NAMESPACE}``"
  kubectl create namespace ``${NAMESPACE}``
fi

TYK_AUTH=foo
TYK_ORG=myorg
TYK_MODE=ce
TYK_URL=http://tyk.tykce-control-plane.svc.cluster.local:8001

kubectl create secret -n ${NAMESPACE} generic tyk-operator-conf \
  --from-literal "TYK_AUTH=${TYK_AUTH}" \
  --from-literal "TYK_ORG=${TYK_ORG}" \
  --from-literal "TYK_MODE=${TYK_MODE}" \
  --from-literal "TYK_URL=${TYK_URL}"

kubectl get secret/tyk-operator-conf -n ${NAMESPACE} -o json | jq '.data'
