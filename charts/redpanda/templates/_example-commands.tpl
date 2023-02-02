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
{{ .rpk }} acl user create myuser --new-password changeme --mechanism {{ include "sasl-mechanism" . }} {{ include "rpk-flags-no-sasl" . }}
{{- end -}}

{{- define "rpk-acl-create" -}}
{{ .rpk }} acl create --allow-principal 'myuser' --allow-host '*' --operation all --topic 'test-topic' {{ include "rpk-flags-no-admin-no-sasl" . }} {{ include "rpk-dummy-sasl" . }}
{{- end -}}

{{- define "rpk-cluster-info" -}}
{{ .rpk }} cluster info {{ include "rpk-flags-no-admin-no-sasl" . }} {{ include "rpk-dummy-sasl" . }}
{{- end -}}

{{- define "rpk-topic-create" -}}
{{- $flags := fromJson (include "rpk-flags" .) -}}
{{ .rpk }} topic create test-topic {{ include "rpk-flags-no-admin-no-sasl" . }} {{ include "rpk-dummy-sasl" . }}
{{- end -}}

{{- define "rpk-topic-describe" -}}
{{- $flags := fromJson (include "rpk-flags" .) -}}
{{ .rpk }} topic describe test-topic {{ include "rpk-flags-no-admin-no-sasl" . }} {{ include "rpk-dummy-sasl" . }}
{{- end -}}

{{- define "rpk-topic-delete" -}}
{{- $flags := fromJson (include "rpk-flags" .) -}}
{{ .rpk }} topic delete test-topic {{ include "rpk-flags-no-admin-no-sasl" . }} {{ include "rpk-dummy-sasl" . }}
{{- end -}}


{{/* tested in tests/test-kafka-sasl-status.yaml */}}
{{- define "rpk-acl-user-create-no-dummy-sasl" -}}
{{ .rpk }} acl user create myuser --new-password changeme --mechanism {{ include "sasl-mechanism" . }} {{ include "rpk-flags-no-sasl" . }}
{{- end -}}

{{/* tested in tests/test-kafka-sasl-status.yaml */}}
{{- define "rpk-acl-create-no-dummy-sasl" -}}
{{ .rpk }} acl create --allow-principal 'myuser' --allow-host '*' --operation all --topic 'test-topic' {{ include "rpk-flags-no-admin" . }}
{{- end -}}

{{/* tested in tests/test-kafka-sasl-status.yaml */}}
{{- define "rpk-cluster-info-no-dummy-sasl" -}}
{{ .rpk }} cluster info {{ include "rpk-flags-no-admin" . }}
{{- end -}}

{{/* tested in tests/test-kafka-sasl-status.yaml */}}
{{- define "rpk-topic-create-no-dummy-sasl" -}}
{{- $flags := fromJson (include "rpk-flags" .) -}}
{{ .rpk }} topic create test-topic {{ include "rpk-flags-no-admin" . }}
{{- end -}}

{{/* tested in tests/test-kafka-sasl-status.yaml */}}
{{- define "rpk-topic-describe-no-dummy-sasl" -}}
{{- $flags := fromJson (include "rpk-flags" .) -}}
{{ .rpk }} topic describe test-topic {{ include "rpk-flags-no-admin" . }}
{{- end -}}

{{/* tested in tests/test-kafka-sasl-status.yaml */}}
{{- define "rpk-topic-delete-no-dummy-sasl" -}}
{{- $flags := fromJson (include "rpk-flags" .) -}}
{{ .rpk }} topic delete test-topic {{ include "rpk-flags-no-admin" . }}
{{- end -}}

