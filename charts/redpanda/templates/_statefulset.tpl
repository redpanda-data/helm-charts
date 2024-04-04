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

{{- define "statefulset-pod-labels-selector" -}}
{{- /*
  StatefulSets cannot change their selector. Use the existing one even if it's broken.
  New installs will get better selectors.
*/ -}}
{{- $sts := lookup "apps/v1" "StatefulSet" .Release.Namespace (include "redpanda.fullname" .) -}}
{{- get ((include "redpanda.StatefulSetPodLabelsSelector" (dict "a" (list . $sts))) | fromJson) "r" | toYaml }}
{{- end -}}

{{- define "statefulset-pod-labels" -}}
{{- /*
  StatefulSets cannot change their selector. Use the existing one even if it's broken.
  New installs will get better selectors.
*/ -}}
{{- $sts := lookup "apps/v1" "StatefulSet" .Release.Namespace (include "redpanda.fullname" .) -}}
{{- get ((include "redpanda.StatefulSetPodLabels" (dict "a" (list . $sts))) | fromJson) "r" | toYaml }}
{{- end -}}

{{/*
Set default path for tiered storage cache or use one provided
*/}}
{{- define "tieredStorage.cacheDirectory" -}}
{{- if empty (include "storage-tiered-config" . | fromJson).cloud_storage_cache_directory -}}
  {{- printf "/var/lib/redpanda/data/cloud_storage_cache" -}}
{{- else -}}
  {{- (include "storage-tiered-config" . | fromJson).cloud_storage_cache_directory -}}
{{- end -}}
{{- end -}}

{{/*
Set tolerations for statefulset, defaults to global tolerations if not defined in statefulset
*/}}
{{- define "statefulset-tolerations" -}}
{{- $tolerations := .Values.tolerations -}}
{{- if not ( empty .Values.statefulset.tolerations ) -}}
{{- $tolerations = .Values.statefulset.tolerations -}}
{{- end -}}
{{- toYaml $tolerations -}}
{{- end -}}

{{/*
Set nodeSelector for statefulset, defaults to global nodeSelector if not defined in statefulset
*/}}
{{- define "statefulset-nodeSelectors" -}}
{{- $nodeSelectors := .Values.nodeSelector -}}
{{- if not ( empty .Values.statefulset.nodeSelector ) -}}
{{- $nodeSelectors = .Values.statefulset.nodeSelector -}}
{{- end -}}
{{- toYaml $nodeSelectors -}}
{{- end -}}

{{/*
Set affinity for statefulset, defaults to global affinity if not defined in statefulset
*/}}
{{- define "statefulset-affinity" -}}
{{- if not ( empty .Values.statefulset.nodeAffinity ) -}}
nodeAffinity: {{ toYaml .Values.statefulset.nodeAffinity | nindent 2 }}
{{- else if not ( empty .Values.affinity.nodeAffinity ) -}}
nodeAffinity: {{ toYaml .Values.affinity.nodeAffinity | nindent 2 }}
{{- end -}}
{{- if not ( empty .Values.statefulset.podAffinity ) -}}
podAffinity: {{ toYaml .Values.statefulset.podAffinity | nindent 2 }}
{{- else if not ( empty .Values.affinity.podAffinity ) -}}
podAffinity: {{ toYaml .Values.affinity.podAffinity | nindent 2 }}
{{- end -}}
{{- if not ( empty .Values.statefulset.podAntiAffinity ) }}
podAntiAffinity:
  {{- if eq .Values.statefulset.podAntiAffinity.type "hard" }}
  requiredDuringSchedulingIgnoredDuringExecution:
  - topologyKey: {{ .Values.statefulset.podAntiAffinity.topologyKey }}
    labelSelector:
      matchLabels: {{ include "statefulset-pod-labels-selector" . | nindent 8 }}
  {{- else if eq .Values.statefulset.podAntiAffinity.type "soft" }}
  preferredDuringSchedulingIgnoredDuringExecution:
  - weight: {{ .Values.statefulset.podAntiAffinity.weight | int64 }}
    podAffinityTerm:
      topologyKey: {{ .Values.statefulset.podAntiAffinity.topologyKey }}
      labelSelector:
        matchLabels: {{ include "statefulset-pod-labels-selector" . | nindent 8 }}
  {{- else if eq .Values.statefulset.podAntiAffinity.type "custom" -}}
    {{- toYaml .Values.statefulset.podAntiAffinity.custom | nindent 2 }}
  {{- end -}}
{{- else if not ( empty .Values.affinity.podAntiAffinity ) -}}
podAntiAffinity: {{ toYaml .Values.affinity.podAntiAffinity | nindent 2 }}
{{- end -}}
{{- end -}}

{{/*
statefulset-checksum-annotation calculates a checksum that is used
as the value for the annotation, "checksum/config". When this value
changes, kube-controller-manager will roll the pods.

Append any additional dependencies that require the pods to restart
to the $dependencies list.
*/}}
{{- define "statefulset-checksum-annotation" -}}
  {{- $dependencies := list -}}
  {{- $dependencies = append $dependencies (include "configmap-content-no-seed" .) -}}
  {{- if .Values.external.enabled -}}
    {{- $dependencies = append $dependencies (dig "domain" "" .Values.external) -}}
    {{- $dependencies = append $dependencies (dig "addresses" "" .Values.external) -}}
  {{- end -}}
  {{- toJson $dependencies | sha256sum -}}
{{- end -}}
