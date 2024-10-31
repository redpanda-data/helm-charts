{{/*
Expand the name of the chart.
*/}}
{{- define "redpanda-connect.name" -}}
{{- default "redpanda-connect" .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "redpanda-connect.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default "redpanda-connect" .Values.nameOverride }}
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
{{- define "redpanda-connect.chart" -}}
{{- printf "%s-%s" "redpanda-connect" .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "redpanda-connect.labels" -}}
helm.sh/chart: {{ include "redpanda-connect.chart" . }}
{{ include "redpanda-connect.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{ with .Values.commonLabels }}
{{- toYaml . -}}
{{- end }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "redpanda-connect.selectorLabels" -}}
app.kubernetes.io/name: {{ include "redpanda-connect.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "redpanda-connect.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "redpanda-connect.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}
