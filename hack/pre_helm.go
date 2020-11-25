package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
)

func main() {
	m := map[string]string{
		"replicas: 1":                   "replicas: {{ .Values.replicaCount }}",
		"tyk-operator-conf":             "{{ .Values.confSecretName }}",
		"tykio/tyk-operator:latest":     "{{ .Values.image.repository }}:{{ .Values.image.tag }}",
		"imagePullPolicy: IfNotPresent": "imagePullPolicy: {{ .Values.image.pullPolicy }}",
		"name: default":                 "name: {{ include \"tyk-operator-helm.serviceAccountName\" . }}",
		"serviceAccountName: default":   "serviceAccountName: {{ include \"tyk-operator-helm.serviceAccountName\" . }}",
	}
	b, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	for k, v := range m {
		b = bytes.ReplaceAll(b, []byte(k), []byte(v))
	}
	os.Stdout.Write(b)
}
