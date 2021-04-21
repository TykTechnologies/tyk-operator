package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"unicode"
	"unicode/utf8"
)

func main() {
	m := map[string]string{
		"replicas: 1":                   "replicas: {{ .Values.replicaCount }}",
		"tyk-operator-conf":             "{{ .Values.confSecretName }}",
		"tykio/tyk-operator:latest":     "{{ .Values.image.repository }}:{{ .Values.image.tag }}",
		"imagePullPolicy: IfNotPresent": "imagePullPolicy: {{ .Values.image.pullPolicy }}",
		"name: default":                 "name: {{ include \"tyk-operator-helm.serviceAccountName\" . }}",
		"serviceAccountName: default":   "serviceAccountName: {{ include \"tyk-operator-helm.serviceAccountName\" . }}",
		annotationsSrc:                  annotationsDest,
		envHTTPSSrc:                     envHTTPSDEST,
		envHTTPSrc:                      envHTTPDEST,
	}
	b, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	for k, v := range m {
		b = bytes.ReplaceAll(b, []byte(k), []byte(v))
	}
	os.Stdout.Write(injectResources(b))
}

var resource = `{{- with .Values.resources}}
        resources:
{{- . | toYaml | nindent 10 }}
{{end}}
`

const annotationsSrc = `  template:
    metadata:
      labels:
        control-plane: controller-manager`

const annotationsDest = `  template:
    metadata:
    {{- with .Values.podAnnotations }}
      annotations:
        {{- toYaml . | nindent 8 }}
    {{- end }}
      labels:
        control-plane: controller-manager`

const envHTTPSSrc = `value: "433"`

const envHTTPSDEST = `value: {{default 8443 .Values.ingressHTTPSPort |quote}}`
const envHTTPSrc = `value: "8080"`

const envHTTPDEST = `value: {{.Values.ingressHTTPPort |quote}}`

func injectResources(b []byte) []byte {
	n := bytes.Index(b, []byte("kind: Deployment"))
	s := b[n:]
	w := bytes.Index(s, []byte("volumeMounts"))
	for ; w > 0; w-- {
		r, _ := utf8.DecodeLastRune(s[:w])
		if !unicode.IsSpace(r) {
			break
		}
	}
	w++
	return append(b[:n+w],
		append([]byte(resource), b[n+w:]...)...,
	)
}
