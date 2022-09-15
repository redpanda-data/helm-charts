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

{{- define "listen.address" -}}
{{- "$(POD_IP)" -}}
{{- end -}}

{{- define "nodeport.listen.address" -}}
{{- "$(HOST_IP)" -}}
{{- end -}}

{{- define "redpanda.internal.domain" -}}
{{- $service := include "redpanda.fullname" . -}}
{{- $ns := .Release.Namespace -}}
{{- $domain := .Values.clusterDomain | trimSuffix "." -}}
{{- printf "%s.%s.svc.%s." $service $ns $domain -}}
{{- end -}}

{{- define "redpanda.kafka.internal.advertise.address" -}}
{{- $host := "$(SERVICE_NAME)" -}}
{{- $domain := include "redpanda.internal.domain" . -}}
{{- printf "%s.%s" $host $domain -}}
{{- end -}}

{{/*
The external advertised address can change depending on the externalisation method.
If the method is to expose via load balancer this must be provided through the values
load balancers configuration for parent zone. If the load balancer is not enabled
then then services are externalised using NodePorts, in which case the external node
IP is required for the advertised address.
*/}}

{{- define "redpanda.kafka.external.domain-lb-bkp" -}}
{{- .Values.loadBalancer.parentZone | trimSuffix "." -}}
{{- end -}}

{{- define "redpanda.kafka.external.domain" -}}
{{- .Values.external.domain | trimSuffix "." | default "$(HOST_IP)" -}}
{{- end -}}

{{- define "redpanda.kafka.external.advertise.address" -}}
{{- $host := "$(SERVICE_NAME)" -}}
{{- $domain := include "redpanda.kafka.external.domain" . -}}
{{- printf "%s.%s" $host $domain -}}
{{- end -}}

{{- define "redpanda.rpc.advertise.address" -}}
{{- $host := "$(SERVICE_NAME)" -}}
{{- $domain := include "redpanda.internal.domain" . -}}
{{- printf "%s.%s" $host $domain -}}
{{- end -}}

{{- define "redpanda.pandaproxy.internal.advertise.address" -}}
{{- $host := "$(SERVICE_NAME)" -}}
{{- $domain := include "redpanda.internal.domain" . -}}
{{- printf "%s.%s" $host $domain -}}
{{- end -}}

{{- define "redpanda.pandaproxy.external.advertise.address" -}}
{{- $host := "$(SERVICE_NAME)" -}}
{{- $domain := include "redpanda.kafka.external.domain" . -}}
{{- printf "%s.%s" $host $domain -}}
{{- end -}}

{{/* ConfigMap variables */}}
{{- define "admin-internal-tls-enabled" -}}
{{- $listener := .Values.listeners.admin -}}
{{- toJson (dict "bool" (and (dig "tls" "enabled" .Values.tls.enabled $listener) (not (empty (dig "tls" "cert" "" $listener))))) -}}
{{- end -}}

{{- define "kafka-internal-tls-enabled" -}}
{{- $listener := .Values.listeners.kafka -}}
{{- toJson (dict "bool" (and (dig "tls" "enabled" .Values.tls.enabled $listener) (not (empty (dig "tls" "cert" "" $listener))))) -}}
{{- end -}}

{{- define "kafka-external-tls-enabled" -}}
{{- toJson (dict "bool" (and (dig "tls" "enabled" (include "kafka-internal-tls-enabled" . | fromJson).bool .listener) (not (empty (include "kafka-external-tls-cert" .))))) -}}
{{- end -}}

{{- define "kafka-external-tls-cert" -}}
{{- dig "tls" "cert" .Values.listeners.kafka.tls.cert .listener -}}
{{- end -}}

{{- define "http-internal-tls-enabled" -}}
{{- $listener := .Values.listeners.http -}}
{{- toJson (dict "bool" (and (dig "tls" "enabled" .Values.tls.enabled $listener) (not (empty (dig "tls" "cert" "" $listener))))) -}}
{{- end -}}

{{- define "http-external-tls-enabled" -}}
{{- $tlsEnabled := dig "tls" "enabled" (include "http-internal-tls-enabled" . | fromJson).bool .listener -}}
{{- toJson (dict "bool" (and $tlsEnabled (not (empty (include "http-external-tls-cert" .))))) -}}
{{- end -}}

{{- define "http-external-tls-cert" -}}
{{- dig "tls" "cert" .Values.listeners.http.tls.cert .listener -}}
{{- end -}}

{{- define "rpc-tls-enabled" -}}
{{- $listener := .Values.listeners.rpc -}}
{{- toJson (dict "bool" (and (dig "tls" "enabled" .Values.tls.enabled $listener) (not (empty (dig "tls" "cert" "" $listener))))) -}}
{{- end -}}

{{- define "schemaRegistry-internal-tls-enabled" -}}
{{- $listener := .Values.listeners.schemaRegistry -}}
{{- toJson (dict "bool" (and (dig "tls" "enabled" .Values.tls.enabled $listener) (not (empty (dig "tls" "cert" "" $listener))))) -}}
{{- end -}}

{{- define "schemaRegistry-external-tls-enabled" -}}
{{- $tlsEnabled := dig "tls" "enabled" (include "schemaRegistry-internal-tls-enabled" . | fromJson).bool .listener -}}
{{- toJson (dict "bool" (and $tlsEnabled (not (empty (include "schemaRegistry-external-tls-cert" .))))) -}}
{{- end -}}

{{- define "schemaRegistry-external-tls-cert" -}}
{{- dig "tls" "cert" .Values.listeners.schemaRegistry.tls.cert .listener -}}
{{- end -}}

{{- define "tls-enabled" -}}
{{- $tlsenabled := .Values.tls.enabled -}}
{{- if not $tlsenabled -}}
  {{- range $listener := .Values.listeners -}}
    {{- if and
        (dig "tls" "enabled" false $listener)
        (not (empty (dig "tls" "cert" "" $listener )))
    -}}
      {{- $tlsenabled = true -}}
    {{- end -}}
    {{- if not $tlsenabled -}}
      {{- range $external := $listener.external -}}
        {{- if and
            (dig "tls" "enabled" false $external)
            (not (empty (dig "tls" "cert" "" $external)))
          -}}
          {{- $tlsenabled = true -}}
        {{- end -}}
      {{- end -}}
    {{- end -}}
  {{- end -}}
{{- end -}}
{{- toJson (dict "bool" $tlsenabled) -}}
{{- end -}}

{{- define "sasl-enabled" -}}
{{- toJson (dict "bool" (dig "enabled" false .Values.auth.sasl)) -}}
{{- end -}}

{{- define "external-nodeport-enabled" -}}
{{- $values := .Values -}}
{{- $enabled := and .Values.external.enabled (eq .Values.external.type "NodePort") -}}
{{- range $listener := .Values.listeners -}}
  {{- range $external := $listener.external -}}
    {{- if and (dig "enabled" false $external) (eq (dig "type" $values.external.type $external) "NodePort") -}}
      {{- $enabled = true -}}
    {{- end -}}
  {{- end -}}
{{- end -}}
{{- toJson (dict "bool" $enabled) -}}
{{- end -}}

{{/* Resource variables */}}
{{- define "redpanda-memoryToMi" -}}
  {{/*
  This template converts the incoming memory value to whole number mebibytes.
  Input can be: k | K | m | M | g | G | Ki | Mi | Gi
  */}}
  {{- $mem := . -}}
  {{- $result := 0 -}}
  {{- if or (hasSuffix "K" $mem) (hasSuffix "k" $mem) -}}
    {{- $rawmem := $mem | trimSuffix "K" | trimSuffix "k" -}}
    {{- if contains "." $rawmem -}}
      {{- $rawmem = $rawmem | float64 -}}
      {{- $result = divf (mulf $rawmem (mul 8 1000)) (mul 8 1024 1024) -}}
    {{- else -}}
      {{- $rawmem = $rawmem | int64 -}}
      {{- $result = divf (mul $rawmem (mul 8 1000)) (mul 8 1024 1024) -}}
    {{- end -}}
    {{- $result = floor $result -}}
  {{- else if or (hasSuffix "M" $mem) (hasSuffix "m" $mem) -}}
    {{- $rawmem := $mem | trimSuffix "M" | trimSuffix "m" -}}
    {{- if contains "." $rawmem -}}
      {{- $rawmem = $rawmem | float64 -}}
      {{- $result = divf (mulf $rawmem (mul 8 1000 1000)) (mul 8 1024 1024) -}}
    {{- else -}}
      {{- $rawmem = $rawmem | int64 -}}
      {{- $result = divf (mul $rawmem (mul 8 1000 1000)) (mul 8 1024 1024) -}}
    {{- end -}}
    {{- $result = floor $result -}}
  {{- else if or (hasSuffix "G" $mem) (hasSuffix "g" $mem) -}}
    {{- $rawmem := $mem | trimSuffix "G" | trimSuffix "g" -}}
    {{- if contains "." $rawmem -}}
      {{- $rawmem = $rawmem | float64 -}}
      {{- $result = divf (mulf $rawmem (mul 8 1000 1000 1000)) (mul 8 1024 1024) -}}
    {{- else -}}
      {{- $rawmem = $rawmem | int64 -}}
      {{- $result = divf (mul $rawmem (mul 8 1000 1000 1000)) (mul 8 1024 1024) -}}
    {{- end -}}
    {{- $result = floor $result -}}
  {{- else if hasSuffix "Ki" $mem }}
    {{- $rawmem := $mem | trimSuffix "Ki" -}}
    {{- if contains "." $rawmem -}}
      {{- $rawmem = $rawmem | float64 -}}
      {{- $result = divf (mulf $rawmem (mul 8 1024)) (mul 8 1024 1024) -}}
    {{- else -}}
      {{- $rawmem = $rawmem | int64 -}}
      {{- $result = divf (mul $rawmem (mul 8 1024)) (mul 8 1024 1024) -}}
    {{- end -}}
    {{- $result = floor $result -}}
  {{- else if hasSuffix "Mi" $mem -}}
    {{- $result = $mem | trimSuffix "Mi" -}}
    {{- if contains "." $result -}}
      {{- $result = $result | float64 -}}
    {{- else -}}
      {{- $result = $result | int64 -}}
    {{- end -}}
  {{- else if hasSuffix "Gi" $mem -}}
    {{- $rawmem := $mem | trimSuffix "Gi" -}}
    {{- if contains "." $rawmem -}}
      {{- $rawmem = $rawmem | float64 -}}
      {{- $result = (mulf $rawmem 1024) | floor -}}
    {{- else -}}
      {{- $rawmem = $rawmem | int64 -}}
      {{- $result = (mul $rawmem 1024) -}}
    {{- end -}}
  {{- else }}
    {{- printf "\n%s is invalid memory amount\nSuffixes can be: k | K | m | M | g | G | Ki | Mi | Gi" $mem | fail -}}
  {{- end }}
  {{- $result -}}
{{- end -}}

{{- define "container-memory" -}}
  {{- $result := "" -}}
  {{- if (hasKey .Values.resources.memory.container "min") -}}
    {{- $result = .Values.resources.memory.container.min | include "redpanda-memoryToMi" -}}
  {{- else -}}
    {{- $result = .Values.resources.memory.container.max | include "redpanda-memoryToMi" -}}
  {{- end -}}
  {{- if eq $result "" -}}
    {{- "unable to get memory value" | fail -}}
  {{- end -}}
  {{- $result -}}
{{- end -}}

{{- define "redpanda-reserve-memory" -}}
  {{/*
  Determines the value of --reserve-memory flag (in mebibytes with M suffix, per Seastar).
  This template looks at all locations where memory could be set.
  These locations, in order of priority, are:
  - .Values.resources.memory.redpanda.reserveMemory (commented out by default, users could uncomment)
  - .Values.resources.memory.container.min (commented out by default, users could uncomment and
    change to something lower than .Values.resources.memory.container.max)
  - .Values.resources.memory.container.max (set by default)
  */}}
  {{- $result := 0 -}}
  {{- if (hasKey .Values.resources.memory "redpanda") -}}
    {{- $result = .Values.resources.memory.redpanda.reserveMemory | include "redpanda-memoryToMi" | int64 -}}
  {{- else if (hasKey .Values.resources.memory.container "min") -}}
    {{- $result = add (mulf (include "container-memory" .) 0.002) 200 -}}
    {{- if gt $result 1000 -}}
      {{- $result = 1000 -}}
    {{- end -}}
  {{- else -}}
    {{- $result = add (mulf (include "container-memory" .) 0.002) 200 -}}
    {{- if gt $result 1000 -}}
      {{- $result = 1000 -}}
    {{- end -}}
  {{- end -}}
  {{- if eq $result 0 -}}
    {{- "unable to get memory value" | fail -}}
  {{- end -}}
  {{- $result -}}
{{- end -}}

{{- define "redpanda-memory" -}}
  {{/*
  Determines the value of --memory flag (in mebibytes with M suffix, per Seastar).
  This template looks at all locations where memory could be set.
  These locations, in order of priority, are:
  - .Values.resources.memory.redpanda.memory (commented out by default, users could uncomment)
  - .Values.resources.memory.container.min (commented out by default, users could uncomment and
    change to something lower than .Values.resources.memory.container.max)
  - .Values.resources.memory.container.max (set by default)
  */}}
  {{- $result := 0 -}}
  {{- if (hasKey .Values.resources.memory "redpanda") -}}
    {{- $result = .Values.resources.memory.redpanda.memory | include "redpanda-memoryToMi" | int64 -}}
  {{- else -}}
    {{- $result = mulf (include "container-memory" .) 0.8 | int64 -}}
  {{- end -}}
  {{- if eq $result 0 -}}
    {{- "unable to get memory value" | fail -}}
  {{- end -}}
  {{- if lt $result 2000 -}}
    {{- printf "\n%d is below the minimum recommended value for Redpanda" $result | fail -}}
  {{- end -}}
  {{- if gt (add $result (include "redpanda-reserve-memory" .)) (include "container-memory" . | int64) -}}
    {{- printf "\nNot enough container memory for Redpanda memory values\nredpanda: %d, reserve: %d, container: %d" $result (include "redpanda-reserve-memory" . | int64) (include "container-memory" . | int64) | fail -}}
  {{- end -}}
  {{- $result -}}
{{- end -}}

{{- define "api-urls" -}}
{{ template "redpanda.fullname" . }}-0.{{ include "redpanda.internal.domain" .}}:{{ .Values.listeners.admin.port }}
{{- end -}}

{{- define "rpk-command" -}}
  {{- $command := list "rpk" -}}
  {{- $command = concat $command (list "--api-urls" (include "api-urls" . )) -}}
  {{- if (include "admin-internal-tls-enabled" . | fromJson).bool -}}
    {{- $command = concat $command (list
      "--admin-api-tls-enabled"
      "--admin-api-tls-truststore"
      (printf "/etc/tls/certs/%s/ca.crt" .Values.listeners.admin.tls.cert))
    -}}
  {{- end -}}
  {{- if (include "kafka-internal-tls-enabled" . | fromJson).bool -}}
    {{- $command = concat $command (list
      "--tls-enabled"
      "--tls-truststore"
      (printf "/etc/tls/certs/%s/ca.crt" .Values.listeners.kafka.tls.cert))
    -}}
  {{- end -}}
  {{- if (include "sasl-enabled" . | fromJson).bool -}}
    {{- $command = concat $command (list
      "--user" (first .Values.auth.sasl.users).name
      "--password" (first .Values.auth.sasl.users).password
      "--sasl-mechanism SCRAM-SHA-256")
    -}}
  {{- end -}}
{{ $command | join " " }}
{{- end -}}