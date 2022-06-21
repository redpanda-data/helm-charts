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
{{- define "redpanda.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
*/}}
{{- define "redpanda.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s" .Release.Name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "redpanda.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "redpanda.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "redpanda.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Generate configuration needed for rpk
*/}}

{{- define "redpanda.kafka.internal.listen.address" -}}
{{- "$(POD_IP)" -}}
{{- end -}}

{{- define "redpanda.kafka.internal.listen.port" -}}
{{- (first .Values.listeners.kafka.endpoints).port -}}
{{- end -}}

{{- define "redpanda.internal.domain" -}}
{{- $service := include "redpanda.fullname" . -}}
{{- $ns := .Release.Namespace -}}
{{- $domain := .Values.clusterDomain | trimSuffix "." -}}
{{- printf "%s.%s.svc.%s." $service $ns $domain -}}
{{- end }}

{{- define "redpanda.kafka.internal.advertise.address" -}}
{{- $host := "$(SERVICE_NAME)" -}}
{{- $domain := include "redpanda.internal.domain" . -}}
{{- printf "%s.%s" $host $domain -}}
{{- end -}}

{{- define "redpanda.kafka.internal.advertise.port" -}}
{{- (first .Values.listeners.kafka.endpoints).port -}}
{{- end -}}

{{- define "redpanda.kafka.external.listen.address" -}}
{{- "$(POD_IP)" -}}
{{- end -}}

{{- define "redpanda.kafka.external.listen.port" -}}
{{- (first .Values.listeners.kafka.endpoints).external.port | default (add1 (first .Values.listeners.kafka.endpoints).port) -}}
{{- end -}}

{{/*
The external advertised address can change depending on the externalisation method.
If the method is to expose via load balancer this must be provided through the values
load balancers configuration for parent zone. If the load balancer is not enabled
then then services are externalised using NodePorts, in which case the external node
IP is required for the advertised address. 
*/}}

{{- define "redpanda.kafka.external.domain-lb-bkp" -}}
{{- printf "%s." (.Values.loadBalancer.parentZone | trimSuffix ".")  -}}
{{- end }}

{{- define "redpanda.kafka.external.domain" -}}
{{- printf "%s." (first .Values.listeners.kafka.endpoints).external.subdomain | trimSuffix "." | default "$(HOST_IP)" -}}
{{- end }}

{{- define "redpanda.kafka.external.advertise.address" -}}
{{- $host := "$(SERVICE_NAME)" -}}
{{- $domain := include "redpanda.kafka.external.domain" . -}}
{{- printf "%s.%s" $host $domain -}}
{{- end -}}

{{- define "redpanda.kafka.external.advertise.port" -}}
{{- (first .Values.listeners.kafka.endpoints).external.port | default (add1 (first .Values.listeners.kafka.endpoints).port) -}}
{{- end -}}

{{- define "redpanda.kafka.external.advertise.nodeport.address" -}}
{{- "$(HOST_IP)" -}}
{{- end -}}

{{- define "redpanda.kafka.external.advertise.nodeport.port" -}}
{{- (first .Values.listeners.kafka.endpoints).external.port | default 32005 -}}
{{- end -}}

{{- define "redpanda.rpc.advertise.address" -}}
{{- $host := "$(SERVICE_NAME)" -}}
{{- $domain := include "redpanda.internal.domain" . -}}
{{- printf "%s.%s" $host $domain -}}
{{- end -}}

{{- define "redpanda.rpc.advertise.port" -}}
{{- .Values.listeners.rpc.port -}}
{{- end -}}

{{- define "redpanda.rpc.listen.address" -}}
{{- "$(POD_IP)" -}}
{{- end -}}

{{ define "redpanda.rpc.listen.port" -}}
{{- .Values.listeners.rpc.port -}}
{{- end -}}

{{- define "redpanda.admin.address" -}}
{{- "$(POD_IP)" -}}
{{- end -}}

{{ define "redpanda.admin.port" -}}
{{- .Values.listeners.admin.port -}}
{{- end -}}

{{- define "redpanda.admin.external.address" -}}
{{- "$(HOST_IP)" -}}
{{- end -}}

{{ define "redpanda.admin.external.port" -}}
{{- (add1 .Values.listeners.admin.port) -}}
{{- end -}}

{{- define "redpanda.pandaproxy.internal.advertise.address" -}}
{{- $host := "$(SERVICE_NAME)" -}}
{{- $domain := include "redpanda.internal.domain" . -}}
{{- printf "%s.%s" $host $domain -}}
{{- end -}}

{{- define "redpanda.pandaproxy.internal.advertise.port" -}}
{{- (first .Values.listeners.rest.endpoints).port -}}
{{- end -}}

{{- define "redpanda.pandaproxy.internal.listen.address" -}}
{{- "$(POD_IP)" -}}
{{- end -}}

{{- define "redpanda.pandaproxy.internal.listen.port" -}}
{{- (first .Values.listeners.rest.endpoints).port -}}
{{- end -}}

{{- define "redpanda.pandaproxy.external.listen.address" -}}
{{- "$(POD_IP)" -}}
{{- end -}}

{{- define "redpanda.pandaproxy.external.listen.port" -}}
{{- add1 (first .Values.listeners.rest.endpoints).port -}}
{{- end -}}

{{- define "redpanda.schemaregistry.internal.address" -}}
{{- "$(POD_IP)" -}}
{{- end -}}

{{- define "redpanda.schemaregistry.external.nodeport.address" -}}
{{- "0.0.0.0" -}}
{{- end -}}

{{- define "redpanda.schemaregistry.internal.port" -}}
{{- (first .Values.listeners.schemaRegistry.endpoints).port -}}
{{- end -}}

{{- define "redpanda.schemaregistry.external.port" -}}
{{- 18081 -}}
{{- end -}}

{{- define "redpanda.pandaproxy.external.advertise.address" -}}
{{- $host := "$(SERVICE_NAME)" -}}
{{- $domain := include "redpanda.kafka.external.domain" . -}}
{{- printf "%s.%s" $host $domain -}}
{{- end -}}

{{- define "redpanda.pandaproxy.external.advertise.port" -}}
{{- add1 (first .Values.listeners.rest.endpoints).port -}}
{{- end -}}

{{- define "redpanda.pandaproxy.external.advertise.nodeport.address" -}}
{{- "$(HOST_IP)" -}}
{{- end -}}

{{- define "redpanda.pandaproxy.external.advertise.nodeport.port" -}}
{{- add1 (first .Values.listeners.rest.endpoints).port -}}
{{- end -}}

{{/* ConfigMap variables */}}
{{- define "admin-tls-enabled" -}}
{{- $adminTlsEnabled1 := and .Values.auth.tls.enabled (or .Values.listeners.admin.tls.enabled (not (hasKey .Values.listeners.admin.tls "enabled"))) -}}
{{- $adminTlsEnabled2 := and (not .Values.auth.tls.enabled) .Values.listeners.admin.tls.enabled -}}
{{- print (and (or $adminTlsEnabled1 $adminTlsEnabled2) (not (empty .Values.listeners.admin.tls.cert))) -}}
{{- end -}}

{{- define "kafka-tls-enabled" -}}
{{- $listener := first .Values.listeners.kafka.endpoints -}}
{{ print (or $listener.tls.enabled (and (not (hasKey $listener.tls "enabled")) .Values.auth.tls.enabled)) }}
{{- end -}}

{{- define "rpc-tls-enabled" -}}
{{- print (or .Values.listeners.rpc.tls.enabled (and (not (hasKey .Values.listeners.rpc.tls "enabled")) .Values.auth.tls.enabled)) -}}
{{- end -}}

{{- define "rest-tls-enabled" -}}
{{- $listener := first .Values.listeners.rest.endpoints -}}
{{ print (or $listener.tls.enabled (and (not (hasKey $listener.tls "enabled")) .Values.auth.tls.enabled)) }}
{{- end -}}

{{- define "schemaregistry-tls-enabled" -}}
{{- $listener := first .Values.listeners.schemaRegistry.endpoints -}}
{{ print (or $listener.tls.enabled (and (not (hasKey $listener.tls "enabled")) .Values.auth.tls.enabled)) }}
{{- end -}}

{{- define "tls-enabled" -}}
{{- print (or (eq (include "admin-tls-enabled" .) "true") (eq (include "kafka-tls-enabled" .) "true") (eq (include "rest-tls-enabled" .) "true") (eq (include "rpc-tls-enabled" .) "true") (eq (include "schemaregistry-tls-enabled" .) "true")) -}}
{{- end -}}
