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
Create chart name and version as used by the chart label.
*/}}
{{- define "console.chart" -}}
{{- get ((include "console.Chart" (dict "a" (list .))) | fromJson) "r" }}
{{- end }}

{{/*
Console Image
*/}}
{{- define "console.container.image" -}}
{{- if .Values.image.registry -}}
{{- printf "%s/%s:%s" .Values.image.registry .Values.image.repository ( .Values.image.tag | default .Chart.AppVersion )  }}
{{- else -}}
{{- printf "%s:%s" .Values.image.repository ( .Values.image.tag | default .Chart.AppVersion )  }}
{{- end -}}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "console.labels" -}}
{{- (get ((include "console.Labels" (dict "a" (list .))) | fromJson) "r") | toYaml -}}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "console.selectorLabels" -}}
{{- (get ((include "console.SelectorLabels" (dict "a" (list .))) | fromJson) "r") | toYaml -}}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "console.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "console.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Console's HTTP server Port.
The port is defined from the provided config but can be overridden
by setting service.targetPort and if that is missing defaults to 8080.
*/}}
{{- define "console.containerPort" -}}
{{- $listenPort := 8080 -}}
{{- if .Values.service.targetPort -}}
{{- $listenPort = .Values.service.targetPort -}}
{{- end -}}
{{- if and .Values.console .Values.console.config .Values.console.config.server -}}
  {{- .Values.console.config.server.listenPort | default $listenPort -}}
{{- else -}}
  {{- $listenPort -}}
{{- end -}}
{{- end -}}

{{/*
Some umbrella charts may use a global registry variable.
In order to be compatible with this, we will watch for a global.imageRegistry
variable or return the imageRegistry as specified via the values.
*/}}
{{- define "console.imageRegistry" -}}
{{- $registryName := .Values.image.registry -}}
{{- if .Values.global }}
    {{- if .Values.global.imageRegistry }}
        {{- printf "%s" .Values.global.imageRegistry -}}
    {{- else -}}
        {{- printf "%s" $registryName -}}
    {{- end -}}
{{- else -}}
    {{- printf "%s" $registryName -}}
{{- end -}}
{{- end -}}
