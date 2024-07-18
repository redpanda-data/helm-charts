{{/*
Expand the name of the chart.
*/}}
{{- define "console.name" -}}
{{- get ((include "console.Name" (dict "a" (list .))) | fromJson) "r" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "console.fullname" -}}
{{- get ((include "console.Fullname" (dict "a" (list .))) | fromJson) "r" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "console.labels" -}}
{{- (get ((include "console.Labels" (dict "a" (list .))) | fromJson) "r") | toYaml -}}
{{- end }}
