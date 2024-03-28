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
		{serviceMonitorIfStarts, serviceMonitorIfStartsTPL},
		{serviceMonitorIfEnds, serviceMonitorIfEndsTPL},
		{extraVolume, extraVolumeTPL},
		{extraVolumeMounts, extraVolumeMountsTPL},
		{imagePullSecretsServiceAccount, imagePullSecretsServiceAccountTPL},
		{rbacPort, rbacPortTPL},
		{webhookPort, webhookPortTPL},

		{"OPERATOR_FULLNAME", `{{ include "tyk-operator-helm.fullname" . }}`},
		{"RELEASE_NAMESPACE", "{{ .Release.Namespace }}"},
		{"OPERATOR_ENV_CONFIG", "{{ .Values.confSecretName }}"},
		{"IfNotPresent", "{{ .Values.image.pullPolicy }}"},
		{"replicas: 1", "replicas: {{default 1 .Values.replicaCount }}"},
		{"tykio/tyk-operator:latest", "{{ .Values.image.repository }}:{{ .Values.image.tag }}"},
		{"CONTROLLER_MANAGER_HEALTH_PROBE_PORT", "{{ .Values.healthProbePort }}"},
		{"CONTROLLER_MANAGER_METRICS_PORT", "{{ .Values.metricsPort }}"},
		{"CONTROLLER_MANAGER_WEBHOOK_PORT", "{{ .Values.webhookPort }}"},
		{"CONTROLLER_MANAGER_RBAC_PORT", "{{ .Values.rbac.port }}"},
		{"CONTROLLER_MANAGER_HOST_NETWORK", "{{ .Values.hostNetwork | default false }}"},
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
    control-plane: tyk-operator-controller-manager
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

// Replaces hardcoded values for ServiceMonitor resource with helm templates.
const (
	serviceMonitorIfStarts    = `TYK_OPERATOR_PROMETHEUS_SERVICEMONITOR_IF_STARTS: null`
	serviceMonitorIfStartsTPL = `{{ if .Values.serviceMonitor }}`
	serviceMonitorIfEnds      = `status: TYK_OPERATOR_PROMETHEUS_SERVICEMONITOR_IF_ENDS`
	serviceMonitorIfEndsTPL   = `{{ end }} `
)

const extraVolume = `- name: CONTROLLER_MANAGER_EXTRA_VOLUME`

const extraVolumeTPL = `{{ if .Values.extraVolumes }}
       {{ toYaml .Values.extraVolumes | nindent 6 }}
        {{ end }}`

const extraVolumeMounts = `- mountPath: CONTROLLER_MANAGER_EXTRA_VOLUMEMOUNTS`

const extraVolumeMountsTPL = `{{ if .Values.extraVolumeMounts }}
            {{ toYaml .Values.extraVolumeMounts | nindent 8}}
          {{ end }}`

const imagePullSecretsServiceAccount = `imagePullSecrets:
- name: TYK_OPERATOR_SERVICEACCOUNT_IMAGEPULLSECRETS
`

const imagePullSecretsServiceAccountTPL = `{{ with  .Values.imagePullSecrets }}
imagePullSecrets:
  {{- toYaml . | nindent 2 }}
{{ end }}
`

const rbacPort = `port: 8443`

const rbacPortTPL = `port: {{ .Values.rbac.port }}`

const webhookPort = `targetPort: 9443`

const webhookPortTPL = `targetPort: {{ .Values.webhookPort }}`