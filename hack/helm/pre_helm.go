package main

import (
	"bytes"
	"io"
	"os"
)

func main() {
	a, err := io.ReadAll(os.Stdin)
	if err != nil {
		os.Stderr.WriteString("Failed to read input")
	}

	m := []struct{ key, value string }{
		{namespace, ""},
		{envFrom, envFromTPL},
		{envVars, envVarsTPL},
		{resources, resourcesTPL},
		{resourcesRBAC, resourcesRBACTPL},
		{annotation, annotationTPL},
		{securityContext, securityContextTPL},
		{imageRBAC, imageRBACTPL},
		{nodeSelector, nodeSelectorTPL},

		{"OPERATOR_FULLNAME", `{{ include "tyk-operator-helm.fullname" . }}`},
		{"RELEASE_NAMESPACE", "{{ .Release.Namespace }}"},
		{"OPERATOR_ENV_CONFIG", "{{ .Values.confSecretName }}"},
		{"IfNotPresent", "{{ .Values.image.pullPolicy }}"},
		{"replicas: 1", "replicas: {{default 1 .Values.replicaCount }}"},
		{"tykio/tyk-operator:latest", "{{ .Values.image.repository }}:{{ .Values.image.tag }}"},
	}

	for _, v := range m {
		a = bytes.ReplaceAll(a, []byte(v.key), []byte(v.value))
	}

	os.Stdout.Write(a)
}

const namespace = `apiVersion: v1
kind: Namespace
metadata:
  labels:
    control-plane: controller-manager
  name: RELEASE_NAMESPACE
---`

const annotation = `      annotations:
        POD-ANNOTATION: POD-ANNOTATION`

const annotationTPL = `{{- with .Values.podAnnotations }}
      annotations:
{{- toYaml . | nindent 8 }}
{{- end }}`

const envFrom = `        envFrom:
        - secretRef:
            name: OPERATOR_ENV_CONFIG`

const envFromTPL = `{{- with .Values.envFrom }}
        envFrom:
{{- toYaml . | nindent 10 }}
{{- end }}`

const envVars = `        env:
        - name: TYK_HTTPS_INGRESS_PORT
          value: PORT_HTTPS_INGRESS
        - name: TYK_HTTP_INGRESS_PORT
          value: PORT_HTTP_INGRESS`

const envVarsTPL = `{{- with .Values.envVars }}
        env:
{{- toYaml . | nindent 10 }}
{{- end }}`

const resources = `        resources:
          limits:
            cpu: 100m
            memory: 30Mi
          requests:
            cpu: 100m
            memory: 20Mi`

const resourcesTPL = `{{- with .Values.resources }}
        resources:
{{- toYaml . | nindent 10 }}
{{- end }}`

const resourcesRBAC = `        resources:
          limits:
            cpu: 50m
            memory: 20Mi
          requests:
            cpu: 50m
            memory: 20Mi`

const resourcesRBACTPL = `{{- with .Values.rbac.resources }}
        resources:
{{- toYaml . | nindent 10 }}
{{- end }}`

const securityContext = `        securityContext:
          allowPrivilegeEscalation: false`

const securityContextTPL = `{{- with .Values.securityContext }}
        securityContext:
{{- toYaml . | nindent 10 }}
{{- end }}`

const imageRBAC = `        image: gcr.io/kubebuilder/kube-rbac-proxy:v0.8.0
        name: kube-rbac-proxy`

const imageRBACTPL = `        image: {{ .Values.rbac.image.repository }}:{{ .Values.rbac.image.tag }}
        imagePullPolicy: {{ .Values.rbac.image.pullPolicy }}
        name: kube-rbac-proxy`

const nodeSelector = `      nodeSelector:
        NODE_SELECTOR: NODE_SELECTOR`

const nodeSelectorTPL = `{{- if .Values.nodeSelector }}
      nodeSelector:
{{ toYaml .Values.nodeSelector | indent 8 }}
{{- end }}`
