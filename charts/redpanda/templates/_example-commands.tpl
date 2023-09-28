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
Any rpk command that's given to the user in NOTES.txt must be defined in this template file
and tested in a test.
*/}}

{{/* tested in tests/test-kafka-sasl-status.yaml */}}
{{- define "rpk-acl-user-create" -}}
{{ .rpk }} acl user create myuser --new-password changeme --mechanism {{ include "sasl-mechanism" . }} {{ include "rpk-acl-user-flags" . }}
{{- end -}}

{{/* tested in tests/test-kafka-sasl-status.yaml */}}
{{- define "rpk-acl-create" -}}
{{- $dummySasl := .dummySasl -}}
{{- if $dummySasl -}}
{{ .rpk }} acl create --allow-principal 'myuser' --allow-host '*' --operation all --topic 'test-topic' {{ include "rpk-flags-no-admin-no-sasl" . }} {{ include "rpk-dummy-sasl" . }}
{{- else -}}
{{ .rpk }} acl create --allow-principal 'myuser' --allow-host '*' --operation all --topic 'test-topic' {{ include "rpk-flags-no-admin" . }}
{{- end -}}
{{- end -}}

{{/* tested in tests/test-kafka-sasl-status.yaml */}}
{{- define "rpk-cluster-info" -}}
{{- $dummySasl := .dummySasl -}}
{{- if $dummySasl -}}
{{ .rpk }} cluster info {{ include "rpk-flags-no-admin-no-sasl" . }} {{ include "rpk-dummy-sasl" . }}
{{- else -}}
{{ .rpk }} cluster info {{ include "rpk-flags-no-admin" . }}
{{- end -}}
{{- end -}}

{{/* tested in tests/test-kafka-sasl-status.yaml */}}
{{- define "rpk-topic-create" -}}
{{- $flags := fromJson (include "rpk-flags" .) -}}
{{- $dummySasl := .dummySasl -}}
{{- if $dummySasl -}}
{{ .rpk }} topic create test-topic -p 3 -r {{ .Values.statefulset.replicas | int64 }} {{ include "rpk-flags-no-admin-no-sasl" . }} {{ include "rpk-dummy-sasl" . }}
{{- else -}}
{{ .rpk }} topic create test-topic -p 3 -r {{ .Values.statefulset.replicas | int64 }} {{ include "rpk-flags-no-admin" . }}
{{- end -}}
{{- end -}}

{{/* tested in tests/test-kafka-sasl-status.yaml */}}
{{- define "rpk-topic-describe" -}}
{{- $flags := fromJson (include "rpk-flags" .) -}}
{{- $dummySasl := .dummySasl -}}
{{- if $dummySasl -}}
{{ .rpk }} topic describe test-topic {{ include "rpk-flags-no-admin-no-sasl" . }} {{ include "rpk-dummy-sasl" . }}
{{- else -}}
{{ .rpk }} topic describe test-topic {{ include "rpk-flags-no-admin" . }}
{{- end -}}
{{- end -}}

{{/* tested in tests/test-kafka-sasl-status.yaml */}}
{{- define "rpk-topic-delete" -}}
{{- $flags := fromJson (include "rpk-flags" .) -}}
{{- $dummySasl := $.dummySasl -}}
{{- if $dummySasl -}}
{{ .rpk }} topic delete test-topic {{ include "rpk-flags-no-admin-no-sasl" . }} {{ include "rpk-dummy-sasl" . }}
{{- else -}}
{{ .rpk }} topic delete test-topic {{ include "rpk-flags-no-admin" . }}
{{- end -}}
{{- end -}}
