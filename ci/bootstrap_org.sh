#!/bin/bash

set -e

kubectl exec -it -n tykpro-control-plane svc/dashboard -- ./tyk-analytics bootstrap --conf=/etc/tyk-dashboard/dash.json > bootstrapped

#Loading configuration from /etc/tyk-dashboard/dash.json
#
#*************** ORGANISATIONS ***************
#ORG NAME        ORG ID
#*********************************************
#No organisation is found.
#
#Creating New Organisation
#ORG DATA: {"Status":"OK","Message":"Org created","Meta":"5f85cd2a5b01020001417e5a"}
#ORG ID: 5f85cd2a5b01020001417e5a
#
#Adding New User
#USER AUTHENTICATION CODE: 0dd99726905f4d4c703b91fc6c95392b
#NEW ID: 5f85cd2aa625ac27f7347f60
#
#DONE
#************************************
#Login at http://localhost:3000/
#User: d44x2f1rwu@default.com
#Pass: hdph4t17
#************************************
