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
Set affinity for post_install_job, defaults to global affinity if not defined in post_install_job
*/}}
{{- define "post-install-job-affinity" -}}
{{- $affinity := .Values.affinity -}}
{{- if not ( empty .Values.post_install_job.affinity ) -}}
  {{- $affinity = .Values.post_install_job.affinity -}}
{{- end -}}
{{- toYaml $affinity -}}
{{- end -}}

{{/*
Set affinity for post_upgrade_job, defaults to global affinity if not defined in post_upgrade_job
*/}}
{{- define "post-upgrade-job-affinity" -}}
{{- $affinity := .Values.affinity -}}
{{- if not ( empty .Values.post_upgrade_job.affinity ) -}}
  {{- $affinity = .Values.post_upgrade_job.affinity -}}
{{- end -}}
{{- toYaml $affinity -}}
{{- end -}}
