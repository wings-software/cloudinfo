{{/*
Expand the name of the chart.
*/}}
{{- define "cloud-info.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "cloud-info.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "cloud-info.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "cloud-info.labels" -}}
helm.sh/chart: {{ include "cloud-info.chart" . }}
{{ include "cloud-info.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "cloud-info.selectorLabels" -}}
app.kubernetes.io/name: {{ include "cloud-info.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "cloud-info.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "cloud-info.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{- define "cloud-info.deploymentEnv" -}}
- name: DB_PASSWORD
  valueFrom:
    secretKeyRef:
        name: postgres
        key: postgres-password
- { name: APP_DATABASE_DATASOURCE, value: "{{ printf "postgres://postgres:$(DB_PASSWORD)@postgres:5432" }}" }
- { name: APP_DB_MIGRATION_DATASOURCE, value: "{{ printf "postgres://postgres:$(DB_PASSWORD)@postgres:5432" }}" }
{{- end }}

{{- define "cloud-info.generateMountSecrets" }}
    {{- if not .Values.workloadIdentity.enabled }}
    gcp-creds: {{ include "harnesscommon.secrets.passwords.manage" (dict "secret" "cloud-info-secret-mount" "key" "gcp-creds" "providedValues" (list "CLOUD_INFO_GCP_CREDS") "length" 10 "context" $) }}
    {{- end }}
    config-file: {{ include "harnesscommon.secrets.passwords.manage" (dict "secret" "cloud-info-secret-mount" "key" "config-file" "providedValues" (list "CLOUD_INFO_CONFIG") "length" 10 "context" $) }}
{{- end }}

{{- define "cloud-info.pullSecrets" -}}
{{ include "common.images.pullSecrets" (dict "images" (list .Values.image) "global" .Values.global ) }}
{{- end -}}