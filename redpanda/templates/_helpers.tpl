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
Get the version of redpanda being used as an image
*/}}
{{- define "redpanda.semver" -}}
{{ .Values.image.tag | trimPrefix "v" }}
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
{{- (first .Values.listeners.http.endpoints).port -}}
{{- end -}}

{{- define "redpanda.pandaproxy.internal.listen.address" -}}
{{- "$(POD_IP)" -}}
{{- end -}}

{{- define "redpanda.pandaproxy.internal.listen.port" -}}
{{- (first .Values.listeners.http.endpoints).port -}}
{{- end -}}

{{- define "redpanda.pandaproxy.external.listen.address" -}}
{{- "$(POD_IP)" -}}
{{- end -}}

{{- define "redpanda.pandaproxy.external.listen.port" -}}
{{- add1 (first .Values.listeners.http.endpoints).port -}}
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
{{- add1 (first .Values.listeners.http.endpoints).port -}}
{{- end -}}

{{- define "redpanda.pandaproxy.external.advertise.nodeport.address" -}}
{{- "$(HOST_IP)" -}}
{{- end -}}

{{- define "redpanda.pandaproxy.external.advertise.nodeport.port" -}}
{{- add1 (first .Values.listeners.http.endpoints).port -}}
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

{{- define "http-tls-enabled" -}}
{{- $listener := first .Values.listeners.http.endpoints -}}
{{ print (or $listener.tls.enabled (and (not (hasKey $listener.tls "enabled")) .Values.auth.tls.enabled)) }}
{{- end -}}

{{- define "schemaregistry-tls-enabled" -}}
{{- $listener := first .Values.listeners.schemaRegistry.endpoints -}}
{{ print (or $listener.tls.enabled (and (not (hasKey $listener.tls "enabled")) .Values.auth.tls.enabled)) }}
{{- end -}}

{{- define "tls-enabled" -}}
{{- print (or (eq (include "admin-tls-enabled" .) "true") (eq (include "kafka-tls-enabled" .) "true") (eq (include "http-tls-enabled" .) "true") (eq (include "rpc-tls-enabled" .) "true") (eq (include "schemaregistry-tls-enabled" .) "true")) -}}
{{- end -}}

{{- define "sasl-enabled" -}}
{{- print .Values.auth.sasl.enabled | default "false" -}}
{{- end -}}

{{- define "admin-external-nodeport-enabled" -}}
  {{- $values := .Values -}}
  {{- if hasKey $values.listeners.admin "external" -}}
    {{- if hasKey $values.listeners.admin.external "type" -}}
      {{- print (and (or $values.listeners.admin.external.enabled (and (not (hasKey $values.listeners.admin.external "enabled")) $values.external.enabled)) (eq $values.listeners.admin.external.type "NodePort")) -}}
    {{- else -}}
      {{- print (and (or $values.listeners.admin.external.enabled (and (not (hasKey $values.listeners.admin.external "enabled")) $values.external.enabled)) (eq $values.external.type "NodePort")) -}}
    {{- end -}}
  {{- else -}}
    {{- print (and $values.external.enabled (eq $values.external.type "NodePort")) -}}
  {{- end -}}
{{- end -}}

{{- define "kafka-external-nodeport-enabled" -}}
  {{- $values := .Values -}}
  {{- $listenerExternalEnabled := false -}}
  {{- range $listener := $values.listeners.kafka.endpoints -}}
    {{- if hasKey $listener "external" -}}
      {{- if hasKey $listener.external "type" -}}
        {{- $listenerExternalEnabled = and (or $listener.external.enabled (and (not (hasKey $listener.external "enabled")) $values.external.enabled)) (eq $listener.external.type "NodePort") -}}
      {{- else -}}
        {{- $listenerExternalEnabled = and (or $listener.external.enabled (and (not (hasKey $listener.external "enabled")) $values.external.enabled)) (eq $values.external.type "NodePort") -}}
      {{- end -}}
    {{- else -}}
      {{- $listenerExternalEnabled = and $values.external.enabled (eq $values.external.type "NodePort") -}}
    {{- end -}}
  {{- end -}}
  {{ print $listenerExternalEnabled }}
{{- end -}}

{{- define "http-external-nodeport-enabled" -}}
  {{- $values := .Values -}}
  {{- $listenerExternalEnabled := false -}}
  {{- range $listener := $values.listeners.http.endpoints -}}
    {{- if hasKey $listener "external" -}}
      {{- if hasKey $listener.external "type" -}}
        {{- $listenerExternalEnabled = and (or $listener.external.enabled (and (not (hasKey $listener.external "enabled")) $values.external.enabled)) (eq $listener.external.type "NodePort") -}}
      {{- else -}}
        {{- $listenerExternalEnabled = and (or $listener.external.enabled (and (not (hasKey $listener.external "enabled")) $values.external.enabled)) (eq $values.external.type "NodePort") -}}
      {{- end -}}
    {{- else -}}
      {{- $listenerExternalEnabled = and $values.external.enabled (eq $values.external.type "NodePort") -}}
    {{- end -}}
  {{- end -}}
  {{ print $listenerExternalEnabled }}
{{- end -}}

{{- define "rpc-external-nodeport-enabled" -}}
  {{- $values := .Values -}}
  {{- if hasKey $values.listeners.rpc "external" -}}
    {{- if hasKey $values.listeners.rpc.external "type" -}}
      {{- print (and (or $values.listeners.rpc.external.enabled (and (not (hasKey $values.listeners.rpc.external "enabled")) $values.external.enabled)) (eq $values.listeners.rpc.external.type "NodePort")) -}}
    {{- else -}}
      {{- print (and (or $values.listeners.rpc.external.enabled (and (not (hasKey $values.listeners.rpc.external "enabled")) $values.external.enabled)) (eq $values.external.type "NodePort")) -}}
    {{- end -}}
  {{- else -}}
    {{- print (and $values.external.enabled (eq $values.external.type "NodePort")) -}}
  {{- end -}}
{{- end -}}

{{- define "schemaregistry-external-nodeport-enabled" -}}
  {{- $values := .Values -}}
  {{- $listenerExternalEnabled := false -}}
  {{- range $listener := $values.listeners.schemaRegistry.endpoints -}}
    {{- if hasKey $listener "external" -}}
      {{- if hasKey $listener.external "type" -}}
        {{- $listenerExternalEnabled = and (or $listener.external.enabled (and (not (hasKey $listener.external "enabled")) $values.external.enabled)) (eq $listener.external.type "NodePort") -}}
      {{- else -}}
        {{- $listenerExternalEnabled = and (or $listener.external.enabled (and (not (hasKey $listener.external "enabled")) $values.external.enabled)) (eq $values.external.type "NodePort") -}}
      {{- end -}}
    {{- else -}}
      {{- $listenerExternalEnabled = and $values.external.enabled (eq $values.external.type "NodePort") -}}
    {{- end -}}
  {{- end -}}
  {{ print $listenerExternalEnabled }}
{{- end -}}


{{- define "external-nodeport-enabled" -}}
{{- print (or (eq (include "admin-external-nodeport-enabled" .) "true") (eq (include "kafka-external-nodeport-enabled" .) "true") (eq (include "http-external-nodeport-enabled" .) "true") (eq (include "rpc-external-nodeport-enabled" .) "true") (eq (include "schemaregistry-external-nodeport-enabled" .) "true")) -}}
{{- end -}}

{{- define "admin-external-lb-enabled" -}}
  {{- $values := .Values -}}
  {{- if hasKey $values.listeners.admin "external" -}}
    {{- if hasKey $values.listeners.admin.external "type" -}}
      {{- print (and (or $values.listeners.admin.external.enabled (and (not (hasKey $values.listeners.admin.external "enabled")) $values.external.enabled)) (eq $values.listeners.admin.external.type "LoadBalancer")) -}}
    {{- else -}}
      {{- print (and (or $values.listeners.admin.external.enabled (and (not (hasKey $values.listeners.admin.external "enabled")) $values.external.enabled)) (eq $values.external.type "LoadBalancer")) -}}
    {{- end -}}
  {{- else -}}
    {{- print (and $values.external.enabled (eq $values.external.type "LoadBalancer")) -}}
  {{- end -}}
{{- end -}}

{{- define "kafka-external-lb-enabled" -}}
  {{- $values := .Values -}}
  {{- $listenerLbEnabled := false -}}
  {{- range $listener := $values.listeners.kafka.endpoints -}}
    {{- if hasKey $listener "external" -}}
      {{- if hasKey $listener.external "type" -}}
        {{- $listenerLbEnabled = and (or $listener.external.enabled (and (not (hasKey $listener.external "enabled")) $values.external.enabled)) (eq $listener.external.type "LoadBalancer") -}}
      {{- else -}}
        {{- $listenerLbEnabled = and (or $listener.external.enabled (and (not (hasKey $listener.external "enabled")) $values.external.enabled)) (eq $values.external.type "LoadBalancer") -}}
      {{- end -}}
    {{- else -}}
      {{- $listenerLbEnabled = and $values.external.enabled (eq $values.external.type "LoadBalancer") -}}
    {{- end -}}
  {{- end -}}
  {{ print $listenerLbEnabled }}
{{- end -}}

{{- define "http-external-lb-enabled" -}}
  {{- $values := .Values -}}
  {{- $listenerLbEnabled := false -}}
  {{- range $listener := $values.listeners.http.endpoints -}}
    {{- if hasKey $listener "external" -}}
      {{- if hasKey $listener.external "type" -}}
        {{- $listenerLbEnabled = and (or $listener.external.enabled (and (not (hasKey $listener.external "enabled")) $values.external.enabled)) (eq $listener.external.type "LoadBalancer") -}}
      {{- else -}}
        {{- $listenerLbEnabled = and (or $listener.external.enabled (and (not (hasKey $listener.external "enabled")) $values.external.enabled)) (eq $values.external.type "LoadBalancer") -}}
      {{- end -}}
    {{- else -}}
      {{- $listenerLbEnabled = and $values.external.enabled (eq $values.external.type "LoadBalancer") -}}
    {{- end -}}
  {{- end -}}
  {{ print $listenerLbEnabled }}
{{- end -}}

{{- define "rpc-external-lb-enabled" -}}
  {{- $values := .Values -}}
  {{- if hasKey $values.listeners.rpc "external" -}}
    {{- if hasKey $values.listeners.rpc.external "type" -}}
      {{- print (and (or $values.listeners.rpc.external.enabled (and (not (hasKey $values.listeners.rpc.external "enabled")) $values.external.enabled)) (eq $values.listeners.rpc.external.type "LoadBalancer")) -}}
    {{- else -}}
      {{- print (and (or $values.listeners.rpc.external.enabled (and (not (hasKey $values.listeners.rpc.external "enabled")) $values.external.enabled)) (eq $values.external.type "LoadBalancer")) -}}
    {{- end -}}
  {{- else -}}
    {{- print (and $values.external.enabled (eq $values.external.type "LoadBalancer")) -}}
  {{- end -}}
{{- end -}}

{{- define "schemaregistry-external-lb-enabled" -}}
  {{- $values := .Values -}}
  {{- $listenerLbEnabled := false -}}
  {{- range $listener := $values.listeners.schemaRegistry.endpoints -}}
    {{- if hasKey $listener "external" -}}
      {{- if hasKey $listener.external "type" -}}
        {{- $listenerLbEnabled = and (or $listener.external.enabled (and (not (hasKey $listener.external "enabled")) $values.external.enabled)) (eq $listener.external.type "LoadBalancer") -}}
      {{- else -}}
        {{- $listenerLbEnabled = and (or $listener.external.enabled (and (not (hasKey $listener.external "enabled")) $values.external.enabled)) (eq $values.external.type "LoadBalancer") -}}
      {{- end -}}
    {{- else -}}
      {{- $listenerLbEnabled = and $values.external.enabled (eq $values.external.type "LoadBalancer") -}}
    {{- end -}}
  {{- end -}}
  {{ print $listenerLbEnabled }}
{{- end -}}

{{- define "external-lb-enabled" -}}
{{- print (or (eq (include "admin-external-lb-enabled" .) "true") (eq (include "kafka-external-lb-enabled" .) "true") (eq (include "http-external-lb-enabled" .) "true") (eq (include "rpc-external-lb-enabled" .) "true") (eq (include "schemaregistry-external-lb-enabled" .) "true")) -}}
{{- end -}}
