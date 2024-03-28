{{/*
Copyright 2020 Redpanda Data, Inc.

Use of this software is governed by the Business Source License
included in the file licenses/BSL.md

As of the Change Date specified in that file, in accordance with
the Business Source License, use of this software will be governed
by the Apache License, Version 2.0
*/}}

{{/*
Expand the name of the chart.
*/}}
{{- define "redpanda-operator.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "redpanda-operator.fullname" -}}
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
{{- define "redpanda-operator.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "redpanda-operator.webhook-cert" -}}
{{- printf .Values.webhookSecretName }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "redpanda-operator.labels" -}}
app.kubernetes.io/name: {{ include "redpanda-operator.name" . }}
helm.sh/chart: {{ include "redpanda-operator.chart" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ include "operator.tag" . | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{ with .Values.commonLabels }}
{{- toYaml . -}}
{{- end }}
{{- end -}}

{{/*
Create the name of the service account to use
*/}}
{{- define "redpanda-operator.serviceAccountName" -}}
{{ default (include "redpanda-operator.fullname" .) .Values.serviceAccount.name }}
{{- end -}}

{{- define "operator.tag" -}}
{{- $tag := default .Chart.AppVersion .Values.image.tag -}}
{{- $matchString := "^v(0|[1-9]\\d*)\\.(0|[1-9]\\d*)\\.(0|[1-9]\\d*)(?:-((?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\\.(?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\\+([0-9a-zA-Z-]+(?:\\.[0-9a-zA-Z-]+)*))?$|^latest$|^dev$" -}}
{{- $match := mustRegexMatch $matchString $tag -}} 
{{- if not $match -}}
  {{/*
  This error message is for end users. This can also occur if
  AppVersion doesn't start with a 'v' in Chart.yaml.
  */}}
  {{ fail "image.tag must start with a 'v' and be valid semver" }}
{{- end -}} 
{{- $tag -}}
{{- end -}}

{{- define "configurator.tag" -}}
{{- $tag := default .Chart.AppVersion .Values.configurator.tag -}}
{{- $matchString := "^v(0|[1-9]\\d*)\\.(0|[1-9]\\d*)\\.(0|[1-9]\\d*)(?:-((?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\\.(?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\\+([0-9a-zA-Z-]+(?:\\.[0-9a-zA-Z-]+)*))?$|^latest$|^dev$" -}}
{{- $match := mustRegexMatch $matchString $tag -}}
{{- if not $match -}}
  {{/*
  This error message is for end users. This can also occur if
  AppVersion doesn't start with a 'v' in Chart.yaml.
  */}}
  {{ fail "configurator.tag must start with a 'v' and be valid semver" }}
{{- end -}}
{{- $tag -}}
{{- end -}}
