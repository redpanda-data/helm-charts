{{/*
Licensed to the Apache Software Foundation (ASF) under one or more
contributor license agreements.  See the NOTICE file distributed with
this work for additional information regarding copyright ownership.
The ASF licenses this file to You under the Apache License, Version 2.0
(the "License"); you may not use this file except in compliance with
the License.  You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/}}
{{/*
Expand the name of the chart.
*/}}
{{- define "connectors.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
*/}}
{{- define "connectors.fullname" }}
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
full helm labels + common labels
*/}}
{{- define "full.labels" -}}
{{ $required := dict
"helm.sh/chart" ( include "connectors.chart" . )
"app.kubernetes.io/managed-by" ( .Release.Service ) }}
{{- toYaml ( merge $required (fromYaml (include "connectors-pod-labels" .))) }}
{{- end -}}

{{/*
pod labels merged with common labels
*/}}
{{- define "connectors-pod-labels" -}}
{{ $required := dict
"app.kubernetes.io/name" ( include "connectors.name" . )
"app.kubernetes.io/instance" ( .Release.Name )
"app.kubernetes.io/component" ( include "connectors.name" . ) }}
{{- toYaml ( merge $required .Values.commonLabels ) }}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "connectors.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Get the version of redpanda being used as an image
*/}}
{{- define "connectors.semver" -}}
{{ include "connectors.tag" . | trimPrefix "v" }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "connectors.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "connectors.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create the name of the service to use
*/}}
{{- define "connectors.serviceName" -}}
{{- default (include "connectors.fullname" .) .Values.service.name }}
{{- end }}

{{/*
Use AppVersion if image.tag is not set
*/}}
{{- define "connectors.tag" -}}
{{- $tag := default .Chart.AppVersion .Values.image.tag -}}
{{- $matchString := "^v(0|[1-9]\\d*)\\.(0|[1-9]\\d*)\\.(0|[1-9]\\d*)(?:-((?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\\.(?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\\+([0-9a-zA-Z-]+(?:\\.[0-9a-zA-Z-]+)*))?$" -}}
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

{{- define "curl-options" -}}
{{- print " -svm3 --fail --retry \"120\" --retry-max-time \"120\" --retry-all-errors -o - -w \"\\nstatus=%{http_code} %{redirect_url} size=%{size_download} time=%{time_total} content-type=\\\"%{content_type}\\\"\\n\" "}}
{{- end -}}
