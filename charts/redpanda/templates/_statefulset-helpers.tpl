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
