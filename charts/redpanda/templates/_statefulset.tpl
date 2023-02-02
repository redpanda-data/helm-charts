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

{{- define "statefulset-pod-labels" -}}
{{- /*
  StatefulSets cannot change their selector. Use the existing one even if it's broken.
  New installs will get better selectors.
*/ -}}
{{- $sts := lookup "apps/v1" "StatefulSet" .Release.Namespace (include "redpanda.fullname" .) -}}
{{- $labels := dig "spec" "selector" "matchLabels" "" $sts -}}
{{- if not (empty $labels) -}}
{{ $labels | toYaml }}
{{- else -}}
app.kubernetes.io/name: {{ template "redpanda.name" . }}
app.kubernetes.io/instance: {{ .Release.Name | quote }}
app.kubernetes.io/component: {{ (include "redpanda.name" .) | trunc 51 }}-statefulset
{{- with .Values.commonLabels }}
{{ toYaml . }}
{{- end }}
{{- end -}}
{{- end -}}

{{/*
Set default path for tiered storage cache or use one provided
*/}}
{{- define "tieredStorage.cacheDirectory" -}}
{{- if empty .Values.storage.tieredConfig.cloud_storage_cache_directory -}}
  {{- printf "/var/lib/redpanda/data/cloud_storage_cache" -}}
{{- else -}}
  {{- .Values.storage.tieredConfig.cloud_storage_cache_directory -}}
{{- end -}}
{{- end -}}