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
Strip out the suffixes on memory to pass to Redpanda
*/}}
{{- define "redpanda.parseMemory" -}}
{{- $type := typeOf .Values.statefulset.resources.limits.memory }}
{{- if eq $type "float64" }}
{{- .Values.statefulset.resources.limits.memory | int64 }}
{{- else if eq $type "int" }}
{{- .Values.statefulset.resources.limits.memory }}
{{- else }}
{{- $string := .Values.statefulset.resources.limits.memory | toString }}
{{- regexReplaceAll "(\\d+)(\\w?)i?" $string "${1}${2}" }}
{{- end }}
{{- end }}

{{/*
Generate configuration needed for rpk
*/}}

{{- define "redpanda.kafka.internal.listen.address" -}}
{{- "$(POD_IP)" -}}
{{- end -}}

{{- define "redpanda.kafka.internal.listen.port" -}}
{{- (first .Values.config.redpanda.kafka_api).port -}}
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
{{- (first .Values.config.redpanda.kafka_api).port -}}
{{- end -}}

{{- define "redpanda.kafka.external.listen.address" -}}
{{- "$(POD_IP)" -}}
{{- end -}}

{{- define "redpanda.kafka.external.listen.port" -}}
{{- add1 (first .Values.config.redpanda.kafka_api).port -}}
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
{{- printf "%s." (first .Values.config.redpanda.kafka_api).external.subdomain | trimSuffix "." | default "$(HOST_IP)" -}}
{{- end }}

{{- define "redpanda.kafka.external.advertise.address" -}}
{{- $host := "$(SERVICE_NAME)" -}}
{{- $domain := include "redpanda.kafka.external.domain" . -}}
{{- printf "%s.%s" $host $domain -}}
{{- end -}}

{{- define "redpanda.kafka.external.advertise.port" -}}
{{- add1 (first .Values.config.redpanda.kafka_api).port -}}
{{- end -}}

{{- define "redpanda.kafka.external.advertise.nodeport.address" -}}
{{- "$(HOST_IP)" -}}
{{- end -}}

{{- define "redpanda.kafka.external.advertise.nodeport.port" -}}
{{- 32005 -}}
{{- end -}}



















{{- define "redpanda.rpc.advertise.address" -}}
{{- $host := "$(SERVICE_NAME)" -}}
{{- $domain := include "redpanda.internal.domain" . -}}
{{- printf "%s.%s" $host $domain -}}
{{- end -}}

{{- define "redpanda.rpc.advertise.port" -}}
{{- .Values.config.redpanda.rpc_server.port -}}
{{- end -}}

{{- define "redpanda.rpc.listen.address" -}}
{{- "$(POD_IP)" -}}
{{- end -}}

{{ define "redpanda.rpc.listen.port" -}}
{{- .Values.config.redpanda.rpc_server.port -}}
{{- end -}}

{{- define "redpanda.admin.address" -}}
{{- "$(POD_IP)" -}}
{{- end -}}

{{ define "redpanda.admin.port" -}}
{{- .Values.config.redpanda.admin.port -}}
{{- end -}}

{{- define "redpanda.admin.external.address" -}}
{{- "$(HOST_IP)" -}}
{{- end -}}

{{ define "redpanda.admin.external.port" -}}
{{- (add1 .Values.config.redpanda.admin.port) -}}
{{- end -}}







{{- define "redpanda.pandaproxy.internal.advertise.address" -}}
{{- $host := "$(SERVICE_NAME)" -}}
{{- $domain := include "redpanda.internal.domain" . -}}
{{- printf "%s.%s" $host $domain -}}
{{- end -}}

{{- define "redpanda.pandaproxy.internal.advertise.port" -}}
{{- (first .Values.config.pandaproxy.pandaproxy_api).port -}}
{{- end -}}

{{- define "redpanda.pandaproxy.internal.listen.address" -}}
{{- "$(POD_IP)" -}}
{{- end -}}

{{- define "redpanda.pandaproxy.internal.listen.port" -}}
{{- (first .Values.config.pandaproxy.pandaproxy_api).port -}}
{{- end -}}

{{- define "redpanda.pandaproxy.external.listen.address" -}}
{{- "$(POD_IP)" -}}
{{- end -}}

{{- define "redpanda.pandaproxy.external.listen.port" -}}
{{- add1 (first .Values.config.pandaproxy.pandaproxy_api).port -}}
{{- end -}}

{{- define "redpanda.schemaregistry.internal.address" -}}
{{- "$(POD_IP)" -}}
{{- end -}}

{{- define "redpanda.schemaregistry.external.nodeport.address" -}}
{{- "0.0.0.0" -}}
{{- end -}}

{{- define "redpanda.schemaregistry.internal.port" -}}
{{- (first .Values.config.schema_registry.schema_registry_api).port -}}
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
{{- add1 (first .Values.config.pandaproxy.pandaproxy_api).port -}}
{{- end -}}

{{- define "redpanda.pandaproxy.external.advertise.nodeport.address" -}}
{{- "$(HOST_IP)" -}}
{{- end -}}

{{- define "redpanda.pandaproxy.external.advertise.nodeport.port" -}}
{{- add1 (first .Values.config.pandaproxy.pandaproxy_api).port -}}
{{- end -}}
