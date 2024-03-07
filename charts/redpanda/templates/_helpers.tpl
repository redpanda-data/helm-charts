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
{{- get ((include "redpanda.Name" (dict "a" (list .))) | fromJson) "r" }}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
*/}}
{{- define "redpanda.fullname" -}}
{{- get ((include "redpanda.Fullname" (dict "a" (list .))) | fromJson) "r" }}
{{- end -}}

{{/*
Create a default service name
*/}}
{{- define "redpanda.servicename" -}}
{{- get ((include "redpanda.ServiceName" (dict "a" (list .))) | fromJson) "r" }}
{{- end -}}

{{/*
full helm labels + common labels
*/}}
{{- define "full.labels" -}}
{{- (get ((include "redpanda.FullLabels" (dict "a" (list .))) | fromJson) "r") | toYaml }}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "redpanda.chart" -}}
{{- get ((include "redpanda.Chart" (dict "a" (list .))) | fromJson) "r" }}
{{- end }}

{{/*
Get the version of redpanda being used as an image
*/}}
{{- define "redpanda.semver" -}}
{{ include "redpanda.tag" . | trimPrefix "v" }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "redpanda.serviceAccountName" -}}
{{- get ((include "redpanda.ServiceAccountName" (dict "a" (list .))) | fromJson) "r" }}
{{- end }}

{{/*
Use AppVersion if image.tag is not set
*/}}
{{- define "redpanda.tag" -}}
{{- get ((include "redpanda.Tag" (dict "a" (list .))) | fromJson) "r" }}
{{- end -}}

{{/* Generate internal fqdn */}}
{{- define "redpanda.internal.domain" -}}
{{- get ((include "redpanda.InternalDomain" (dict "a" (list .))) | fromJson) "r" }}
{{- end -}}

{{/* ConfigMap variables */}}
{{- define "admin-internal-tls-enabled" -}}
{{- $listener := .Values.listeners.admin -}}
{{- toJson (dict "bool" (and (dig "tls" "enabled" .Values.tls.enabled $listener) (not (empty (dig "tls" "cert" "" $listener))))) -}}
{{- end -}}

{{- define "admin-external-tls-enabled" -}}
{{- toJson (dict "bool" (and (dig "tls" "enabled" (include "admin-internal-tls-enabled" . | fromJson).bool .listener) (not (empty (include "admin-external-tls-cert" .))))) -}}
{{- end -}}

{{- define "admin-external-tls-cert" -}}
{{- dig "tls" "cert" .Values.listeners.admin.tls.cert .listener -}}
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
{{- $tlsenabled := get ((include "redpanda.TLSEnabled" (dict "a" (list .))) | fromJson) "r" }}
{{- toJson (dict "bool" $tlsenabled) -}}
{{- end -}}

{{- define "sasl-enabled" -}}
{{- toJson (dict "bool" (dig "enabled" false .Values.auth.sasl)) -}}
{{- end -}}

{{- define "SI-to-bytes" -}}
  {{/*
  This template converts the incoming SI value to whole number bytes.
  Input can be: b | B | k | K | m | M | g | G | Ki | Mi | Gi
  Or number without suffix
  */}}
  {{- $si := . -}}
  {{- if not (typeIs "string" . ) -}}
    {{- $si = int64 $si | toString -}}
  {{- end -}}
  {{- $bytes := 0 -}}
  {{- if or (hasSuffix "B" $si) (hasSuffix "b" $si) -}}
    {{- $bytes = $si | trimSuffix "B" | trimSuffix "b" | float64 | floor -}}
  {{- else if or (hasSuffix "K" $si) (hasSuffix "k" $si) -}}
    {{- $raw := $si | trimSuffix "K" | trimSuffix "k" | float64 -}}
    {{- $bytes = mulf $raw (mul 1000) | floor -}}
  {{- else if or (hasSuffix "M" $si) (hasSuffix "m" $si) -}}
    {{- $raw := $si | trimSuffix "M" | trimSuffix "m" | float64 -}}
    {{- $bytes = mulf $raw (mul 1000 1000) | floor -}}
  {{- else if or (hasSuffix "G" $si) (hasSuffix "g" $si) -}}
    {{- $raw := $si | trimSuffix "G" | trimSuffix "g" | float64 -}}
    {{- $bytes = mulf $raw (mul 1000 1000 1000) | floor -}}
  {{- else if hasSuffix "Ki" $si -}}
    {{- $raw := $si | trimSuffix "Ki" | float64 -}}
    {{- $bytes = mulf $raw (mul 1024) | floor -}}
  {{- else if hasSuffix "Mi" $si -}}
    {{- $raw := $si | trimSuffix "Mi" | float64 -}}
    {{- $bytes = mulf $raw (mul 1024 1024) | floor -}}
  {{- else if hasSuffix "Gi" $si -}}
    {{- $raw := $si | trimSuffix "Gi" | float64 -}}
    {{- $bytes = mulf $raw (mul 1024 1024 1024) | floor -}}
  {{- else if (mustRegexMatch "^[0-9]+$" $si) -}}
    {{- $bytes = $si -}}
  {{- else -}}
    {{- printf "\n%s is invalid SI quantity\nSuffixes can be: b | B | k | K | m | M | g | G | Ki | Mi | Gi or without any Suffixes" $si | fail -}}
  {{- end -}}
  {{- $bytes | int64 -}}
{{- end -}}

{{/* Resource variables */}}
{{- define "redpanda-memoryToMi" -}}
  {{/*
  This template converts the incoming memory value to whole number mebibytes.
  */}}
  {{- div (include "SI-to-bytes" .) (mul 1024 1024) -}}
{{- end -}}

{{- define "container-memory" -}}
  {{- $result := "" -}}
  {{- if (hasKey .Values.resources.memory.container "min") -}}
    {{- $result = .Values.resources.memory.container.min | include "redpanda-memoryToMi" -}}
  {{- else -}}
    {{- $result = .Values.resources.memory.container.max | include "redpanda-memoryToMi" -}}
  {{- end -}}
  {{- if eq $result "" -}}
    {{- "unable to get memory value from container" | fail -}}
  {{- end -}}
  {{- $result -}}
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

{{- define "external-loadbalancer-enabled" -}}
{{- $values := .Values -}}
{{- $enabled := and .Values.external.enabled (eq .Values.external.type "LoadBalancer") -}}
{{- range $listener := .Values.listeners -}}
  {{- range $external := $listener.external -}}
    {{- if and (dig "enabled" false $external) (eq (dig "type" $values.external.type $external) "LoadBalancer") -}}
      {{- $enabled = true -}}
    {{- end -}}
  {{- end -}}
{{- end -}}
{{- toJson (dict "bool" $enabled) -}}
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
    {{- "unable to get memory value redpanda-memory" | fail -}}
  {{- end -}}
  {{- if lt $result 256 -}}
    {{- printf "\n%d is below the minimum value for Redpanda" $result | fail -}}
  {{- end -}}
  {{- if not .Values.config.node.developer_mode }}
  {{- if gt (add $result (include "redpanda-reserve-memory" .)) (include "container-memory" . | int64) -}}
    {{- printf "\nNot enough container memory for Redpanda memory values\nredpanda: %d, reserve: %d, container: %d" $result (include "redpanda-reserve-memory" . | int64) (include "container-memory" . | int64) | fail -}}
  {{- end -}}
  {{- end -}}
  {{- $result -}}
{{- end -}}

{{/*
Returns the value of "resources.cpu.cores" in millicores. And ensures CPU units
are using known suffix (really only "m") or no suffix at all.
*/}}
{{- define "redpanda-cores-in-millis" -}}
  {{- $cores := .Values.resources.cpu.cores | toString -}}
  {{- $coresSuffix := regexReplaceAll "^[0-9.]+(.*)" $cores "${1}" -}}
  {{- if eq $coresSuffix "m" -}}
    {{- trimSuffix $coresSuffix .Values.resources.cpu.cores -}}
  {{- else -}}
    {{- if eq $coresSuffix "" -}}
      {{- mulf 1000.0 ($cores | float64) -}}
    {{- else -}}
      {{- printf "Unrecognized CPU unit '%s'" $coresSuffix | fail -}}
    {{- end -}}
  {{- end -}}
{{- end -}}

{{/*
Returns the SMP CPU count in whole cores, with minimum of 1, and sets
"resources.cpu.overprovisioned: true" when the "resources.cpu.cores" is less
than 1 core.
*/}}
{{- define "redpanda-smp" -}}
  {{- $coresInMillies := include "redpanda-cores-in-millis" . | int -}}
  {{- if lt $coresInMillies 1000 -}}
    {{- $_ := set $.Values.resources.cpu "overprovisioned" true -}}
    {{- 1 -}}
  {{- else -}}
    {{- floor (divf $coresInMillies 1000) -}}
  {{- end -}}
{{- end -}}

{{- define "admin-api-urls" -}}
{{ printf "${SERVICE_NAME}.%s" (include "redpanda.internal.domain" .) }}:{{.Values.listeners.admin.port }}
{{- end -}}

{{- define "admin-api-service-url" -}}
{{ include "redpanda.internal.domain" .}}:{{.Values.listeners.admin.port }}
{{- end -}}

{{- define "sasl-mechanism" -}}
{{- dig "sasl" "mechanism" "SCRAM-SHA-512" .Values.auth -}}
{{- end -}}

{{- define "storage-min-free-bytes" -}}
{{- $fiveGiB := 5368709120 -}}
{{- if dig "enabled" false .Values.storage.persistentVolume -}}
  {{- min $fiveGiB (mulf (include "SI-to-bytes" .Values.storage.persistentVolume.size) 0.05 | int64) -}}
{{- else -}}
{{- $fiveGiB -}}
{{- end -}}
{{- end -}}

{{- define "tunable" -}}
  {{- $tunable := dig "tunable" dict .Values.config -}}
  {{- if (include "redpanda-atleast-22-3-0" . | fromJson).bool -}}
  {{- range $key, $element := $tunable }}
    {{- if or (eq (typeOf $element) "bool") $element }}
{{ $key }}: {{ $element | toYaml }}
    {{- end }}
  {{- end }}
  {{- else if (include "redpanda-atleast-22-2-0" . | fromJson).bool -}}
  {{- $tunable = unset $tunable "log_segment_size_min" -}}
  {{- $tunable = unset $tunable "log_segment_size_max" -}}
  {{- $tunable = unset $tunable "kafka_batch_max_bytes" -}}
  {{- range $key, $element := $tunable }}
    {{- if or (eq (typeOf $element) "bool") $element }}
{{ $key }}: {{ $element | toYaml }}
    {{- end }}
  {{- end }}
  {{- end -}}
{{- end -}}

{{- define "fail-on-insecure-sasl-logging" -}}
{{- if (include "sasl-enabled" .|fromJson).bool -}}
  {{- $check := list
      (include "redpanda-atleast-23-1-1" .|fromJson).bool
      (include "redpanda-22-3-atleast-22-3-13" .|fromJson).bool
      (include "redpanda-22-2-atleast-22-2-10" .|fromJson).bool
  -}}
  {{- if not (mustHas true $check) -}}
    {{- fail "SASL is enabled and the redpanda version specified leaks secrets to the logs. Please choose a newer version of redpanda." -}}
  {{- end -}}
{{- end -}}
{{- end -}}

{{- define "fail-on-unsupported-helm-version" -}}
  {{- $helmVer := (fromYaml (toYaml .Capabilities.HelmVersion)).version -}}
  {{- if semverCompare "<3.8.0-0" $helmVer -}}
    {{- fail (printf "helm version %s is not supported. Please use helm version v3.8.0 or newer." $helmVer) -}}
  {{- end -}}
{{- end -}}

{{- define "redpanda-atleast-22-2-0" -}}
{{- toJson (dict "bool" (or (not (eq .Values.image.repository "docker.redpanda.com/redpandadata/redpanda")) (include "redpanda.semver" . | semverCompare ">=22.2.0-0 || <0.0.1-0"))) -}}
{{- end -}}
{{- define "redpanda-atleast-22-3-0" -}}
{{- toJson (dict "bool" (or (not (eq .Values.image.repository "docker.redpanda.com/redpandadata/redpanda")) (include "redpanda.semver" . | semverCompare ">=22.3.0-0 || <0.0.1-0"))) -}}
{{- end -}}
{{- define "redpanda-atleast-23-1-1" -}}
{{- toJson (dict "bool" (or (not (eq .Values.image.repository "docker.redpanda.com/redpandadata/redpanda")) (include "redpanda.semver" . | semverCompare ">=23.1.1-0 || <0.0.1-0"))) -}}
{{- end -}}
{{- define "redpanda-atleast-23-1-2" -}}
{{- toJson (dict "bool" (or (not (eq .Values.image.repository "docker.redpanda.com/redpandadata/redpanda")) (include "redpanda.semver" . | semverCompare ">=23.1.2-0 || <0.0.1-0"))) -}}
{{- end -}}
{{- define "redpanda-22-3-atleast-22-3-13" -}}
{{- toJson (dict "bool" (or (not (eq .Values.image.repository "docker.redpanda.com/redpandadata/redpanda")) (include "redpanda.semver" . | semverCompare ">=22.3.13-0,<22.4"))) -}}
{{- end -}}
{{- define "redpanda-22-2-atleast-22-2-10" -}}
{{- toJson (dict "bool" (or (not (eq .Values.image.repository "docker.redpanda.com/redpandadata/redpanda")) (include "redpanda.semver" . | semverCompare ">=22.2.10-0,<22.3"))) -}}
{{- end -}}
{{- define "redpanda-atleast-23-2-1" -}}
{{- toJson (dict "bool" (or (not (eq .Values.image.repository "docker.redpanda.com/redpandadata/redpanda")) (include "redpanda.semver" . | semverCompare ">=23.2.1-0 || <0.0.1-0"))) -}}
{{- end -}}
{{- define "redpanda-atleast-23-3-0" -}}
{{- toJson (dict "bool" (or (not (eq .Values.image.repository "docker.redpanda.com/redpandadata/redpanda")) (include "redpanda.semver" . | semverCompare ">=23.3.0-0 || <0.0.1-0"))) -}}
{{- end -}}

{{- define "redpanda-22-2-x-without-sasl" -}}
{{- $result :=  (include "redpanda-atleast-22-3-0" . | fromJson).bool -}}
{{- if or (include "sasl-enabled" . | fromJson).bool .Values.listeners.kafka.authenticationMethod -}}
{{-   $result := false -}}
{{- end -}}
{{- toJson (dict "bool" $result) -}}
{{- end -}}

# manage backward compatibility with renaming podSecurityContext to securityContext
{{- define "pod-security-context" -}}
fsGroup: {{ dig "podSecurityContext" "fsGroup" .Values.statefulset.securityContext.fsGroup .Values.statefulset }}
fsGroupChangePolicy: {{ dig "securityContext" "fsGroupChangePolicy" "OnRootMismatch" .Values.statefulset }}
{{- end -}}

# for backward compatibility, force a default on releases that didn't
# set the podSecurityContext.runAsUser before
{{- define "container-security-context" -}}
runAsUser: {{ dig "podSecurityContext" "runAsUser" .Values.statefulset.securityContext.runAsUser .Values.statefulset }}
runAsGroup: {{ dig "podSecurityContext" "fsGroup" .Values.statefulset.securityContext.fsGroup .Values.statefulset }}
{{- if hasKey .Values.statefulset.securityContext "allowPrivilegeEscalation" }}
allowPrivilegeEscalation: {{ dig "podSecurityContext" "allowPrivilegeEscalation" .Values.statefulset.securityContext.allowPrivilegeEscalation .Values.statefulset }}
{{- end -}}
{{- if hasKey .Values.statefulset.securityContext "runAsNonRoot" }}
runAsNonRoot: {{ dig "podSecurityContext" "runAsNonRoot" .Values.statefulset.securityContext.runAsNonRoot .Values.statefulset }}
{{- end -}}
{{- end -}}

{{- define "admin-tls-curl-flags" -}}
  {{- $result := "" -}}
  {{- if (include "admin-internal-tls-enabled" . | fromJson).bool -}}
    {{- $path := (printf "/etc/tls/certs/%s" .Values.listeners.admin.tls.cert) -}}
    {{- $result = (printf "--cacert %s/tls.crt" $path) -}}
    {{- if .Values.listeners.admin.tls.requireClientAuth -}}
      {{- $result = (printf "--cacert %s/ca.crt --cert %s/tls.crt --key %s/tls.key" $path $path $path) -}}
    {{- end -}}
  {{- end -}}
  {{- $result -}}
{{- end -}}

{{- define "admin-http-protocol" -}}
  {{- $result := "http" -}}
  {{- if (include "admin-internal-tls-enabled" . | fromJson).bool -}}
    {{- $result = "https" -}}
  {{- end -}}
  {{- $result -}}
{{- end -}}

{{- /*
advertised-port returns either the only advertised port if only one is specified,
or the port specified for this pod ordinal when there is a full list provided.

This will return a string int or panic if there is more than one port provided,
but not enough ports for the number of replicas requested.
*/ -}}
{{- define "advertised-port" -}}
  {{- $port := dig "port" .listenerVals.port .externalVals -}}
  {{- if .externalVals.advertisedPorts -}}
    {{- if eq (len .externalVals.advertisedPorts) 1 -}}
      {{- $port = mustFirst .externalVals.advertisedPorts -}}
    {{- else -}}
      {{- $port = index .externalVals.advertisedPorts .replicaIndex -}}
    {{- end -}}
  {{- end -}}
  {{ $port }}
{{- end -}}

{{- /*
advertised-host returns a json string with the data needed for configuring the advertised listener
*/ -}}
{{- define "advertised-host" -}}
  {{- $host := dict "name" .externalName "address" .externalAdvertiseAddress "port" .port -}}
  {{- if .values.external.addresses -}}
    {{- $address := "" -}}
    {{- if gt (len .values.external.addresses) 1 -}}
      {{- $address = (index .values.external.addresses .replicaIndex) -}}
    {{- else -}}
      {{- $address = (index .values.external.addresses 0) -}}
    {{- end -}}
    {{- if ( .values.external.domain | default "" ) }}
      {{- $host = dict "name" .externalName "address" (printf "%s.%s" $address .values.external.domain) "port" .port -}}
    {{- else -}}
      {{- $host = dict "name" .externalName  "address" $address "port" .port -}}
    {{- end -}}
  {{- end -}}
  {{- toJson $host -}}
{{- end -}}

{{- define "is-licensed" -}}
{{- toJson (dict "bool" (or (not (empty (include "enterprise-license" . ))) (not (empty (include "enterprise-secret" . ))))) -}}
{{- end -}}

{{/*
"warnings" is an aggregate that returns a list of warnings to be shown in NOTES.txt
*/}}
{{- define "warnings" -}}
  {{- $result := list -}}
  {{- $warnings := list "redpanda-memory-warning" "redpanda-cpu-warning" -}}
  {{- range $t := $warnings -}}
    {{- $warning := include $t $ -}}
      {{- if $warning -}}
        {{- $result = append $result (printf "**Warning**: %s" $warning) -}}
      {{- end -}}
  {{- end -}}
  {{/* fromJson cannot decode list */}}
  {{- toJson (dict "result" $result) -}}
{{- end -}}

{{/*
return a warning if the chart is configured with insufficient memory
*/}}
{{- define "redpanda-memory-warning" -}}
  {{- $result := (include "redpanda-memory" .) | int -}}
  {{- if lt $result 2000 -}}
    {{- printf "%d is below the minimum recommended value for Redpanda" $result -}}
  {{- end -}}
{{- end -}}

{{/*
return a warning if the chart is configured with insufficient CPU
*/}}
{{- define "redpanda-cpu-warning" -}}
  {{- $coresInMillies := include "redpanda-cores-in-millis" . | int -}}
  {{- if lt $coresInMillies 1000 -}}
    {{- printf "%dm is below the minimum recommended CPU value for Redpanda" $coresInMillies -}}
  {{- end -}}
{{- end -}}

{{- define "seed-server-list" -}}
  {{- $brokers := list -}}
  {{- range $ordinal := until (.Values.statefulset.replicas | int) -}}
    {{- $brokers = append $brokers (printf "%s-%d.%s"
        (include "redpanda.fullname" $)
        $ordinal
        (include "redpanda.internal.domain" $))
    -}}
  {{- end -}}
  {{- toJson $brokers -}}
{{- end -}}

{{- define "kafka-brokers-sasl-enabled" -}}
  {{- $root := . -}}
  {{- $kafkaService := .Values.listeners.kafka }}
  {{- $auditLogging := .Values.auditLogging }}
  {{- $brokers := list -}}
  {{- $broker_tls := dict -}}
  {{- $result := dict -}}
  {{- $tlsEnabled := .Values.tls.enabled -}}
  {{- $tlsCerts := .Values.tls.certs -}}
  {{- $trustStoreFile := "" -}}
  {{- $requireClientAuth := dig "tls" "requireClientAuth" false $kafkaService -}}
  {{- if and ( eq "internal" $auditLogging.listener ) ( eq (default "sasl" $kafkaService.authenticationMethod) "sasl" ) -}}
    {{- range $id, $item := $root.tempConfigMapServerList }}
      {{- $brokerItem := ( dict
        "address" $item.host.address
        "port" $kafkaService.port
        )
      -}}
    {{- $brokers = append $brokers $brokerItem -}}
    {{- end }}
    {{- if $brokers -}}
      {{- $result = set $result "brokers" $brokers -}}
    {{- end -}}
    {{- if dig "tls" "enabled" $tlsEnabled $kafkaService -}}
      {{- $cert := get $tlsCerts $kafkaService.tls.cert -}}
      {{- if empty $cert -}}
        {{- fail (printf "Certificate used but not defined") -}}
      {{- end -}}
      {{- if $cert.caEnabled -}}
        {{- $trustStoreFile =  ( printf "/etc/tls/certs/%s/ca.crt" $kafkaService.tls.cert ) -}}
      {{- else -}}
        {{- $trustStoreFile = "/etc/ssl/certs/ca-certificates.crt" -}}
      {{- end -}}
      {{- $broker_tls = ( dict
        "enabled" true
        "cert_file" ( printf "/etc/tls/certs/%s/tls.crt" $kafkaService.tls.cert )
        "key_file" ( printf "/etc/tls/certs/%s/tls.key" $kafkaService.tls.cert )
        "require_client_auth" $requireClientAuth
        )
      -}}
      {{- if $trustStoreFile -}}
        {{- $broker_tls = set $broker_tls "truststore_file" $trustStoreFile -}}
      {{- end -}}
      {{- if $broker_tls -}}
        {{- $result = set $result "broker_tls" $broker_tls -}}
      {{- end -}}
    {{- end -}}
  {{- else -}}
    {{- range $name, $listener := $kafkaService.external -}}
      {{- if and $listener.port $name (dig "enabled" true $listener) ( eq (default "sasl" $listener.authenticationMethod) "sasl" ) ( eq $name $auditLogging.listener ) -}}
        {{- range $id, $item := $root.tempConfigMapServerList }}
          {{- $brokerItem := ( dict
            "address" $item.host.address
            "port" $listener.port
            )
          -}}
        {{- $brokers = append $brokers $brokerItem -}}
        {{- end }}
        {{- if $brokers -}}
          {{- $result = set $result "brokers" $brokers -}}
        {{- end -}}
        {{- if dig "tls" "enabled" $tlsEnabled $listener -}}
          {{- $cert := get $tlsCerts $listener.tls.cert -}}
          {{- if empty $cert -}}
            {{- fail (printf "Certificate used but not defined") -}}
          {{- end -}}
          {{- if $cert.caEnabled -}}
            {{- $trustStoreFile =  ( printf "/etc/tls/certs/%s/ca.crt" $listener.tls.cert ) -}}
          {{- else -}}
            {{- $trustStoreFile = "/etc/ssl/certs/ca-certificates.crt" -}}
          {{- end -}}
          {{- $broker_tls = ( dict
            "enabled" true
            "cert_file" ( printf "/etc/tls/certs/%s/tls.crt" $listener.tls.cert )
            "key_file" ( printf "/etc/tls/certs/%s/tls.key" $listener.tls.cert )
            "require_client_auth" $requireClientAuth
            )
          -}}
          {{- if $trustStoreFile -}}
            {{- $broker_tls = set $broker_tls "truststore_file" $trustStoreFile -}}
          {{- end -}}
          {{- if $broker_tls -}}
            {{- $result = set $result "broker_tls" $broker_tls -}}
          {{- end -}}
        {{- end -}}
      {{- end -}}
    {{- end -}}
  {{- end -}}
{{- toYaml $result  -}}
{{- end -}}

{{/*
return correct secretName to use based if secretRef exists
*/}}
{{- define "cert-secret-name" -}}
  {{- if .tempCert.cert.secretRef -}}
    {{- .tempCert.cert.secretRef.name -}}
  {{- else -}}
    {{- include "redpanda.fullname" . }}-{{ .tempCert.name }}-cert
  {{- end -}}
{{- end -}}

{{/*
return license checks deprecated values if current values is empty
*/}}
{{- define "enterprise-license" -}}
{{- if dig "license" dict .Values.enterprise -}}
  {{- .Values.enterprise.license -}}
{{- else -}}
  {{- .Values.license_key -}}
{{- end -}}
{{- end -}}

{{/*
return licenseSecretRef checks deprecated values entry if current values empty
*/}}
{{- define "enterprise-secret" -}}
{{- if ( dig "licenseSecretRef" dict .Values.enterprise ) -}}
  {{- .Values.enterprise.licenseSecretRef -}}
{{- else if not (empty .Values.license_secret_ref ) -}}
  {{- .Values.license_secret_ref -}}
{{- end -}}
{{- end -}}

{{/*
return licenseSecretRef.name checks deprecated values entry if current values empty
*/}}
{{- define "enterprise-secret-name" -}}
{{- if ( dig "licenseSecretRef" dict .Values.enterprise ) -}}
  {{- dig "name" "" .Values.enterprise.licenseSecretRef -}}
{{- else if not (empty .Values.license_secret_ref ) -}}
  {{- dig "secret_name" "" .Values.license_secret_ref -}}
{{- end -}}
{{- end -}}

{{/*
return licenseSecretRef.key checks deprecated values entry if current values empty
*/}}
{{- define "enterprise-secret-key" -}}
{{- if ( dig "licenseSecretRef" dict .Values.enterprise ) -}}
  {{- dig "key" "" .Values.enterprise.licenseSecretRef -}}
{{- else if not (empty .Values.license_secret_ref ) -}}
  {{- dig "secret_key" "" .Values.license_secret_ref -}}
{{- end -}}
{{- end -}}

{{/* mounts that are common to all containers */}}
{{- define "common-mounts" -}}
  {{- if and .Values.auth.sasl.enabled (not (empty .Values.auth.sasl.secretRef )) }}
- name: users
  mountPath: /etc/secrets/users
  readOnly: true
  {{- end }}
  {{- if (include "tls-enabled" . | fromJson).bool }}
    {{- range $name, $cert := .Values.tls.certs }}
- name: redpanda-{{ $name }}-cert
  mountPath: {{ printf "/etc/tls/certs/%s" $name }}
    {{- end }}
    {{- if (include "client-auth-required" . | fromJson).bool }}
- name: mtls-client
  mountPath: /etc/tls/certs/{{ template "redpanda.fullname" $ }}-client
    {{- end }}
  {{- end }}
{{- end -}}

{{/* mounts that are common to most containers */}}
{{- define "default-mounts" -}}
- name: config
  mountPath: /etc/redpanda
{{- include "common-mounts" . }}
{{- end -}}

{{/* volumes that are common to all pods */}}
{{- define "common-volumes" -}}
  {{- if (include "tls-enabled" . | fromJson).bool -}}
    {{- range $name, $cert := .Values.tls.certs }}
      {{- $r :=  set $ "tempCert" ( dict "name" $name "cert" $cert ) }}
- name: redpanda-{{ $name }}-cert
  secret:
    secretName: {{ template "cert-secret-name" $r }}
    defaultMode: 0o440
    {{- end }}
    {{- if (include "client-auth-required" . | fromJson).bool }}
- name: mtls-client
  secret:
    secretName: {{ template "redpanda.fullname" $ }}-client
    defaultMode: 0o440
    {{- end }}
  {{- end -}}
  {{- if and .Values.auth.sasl.enabled (not (empty .Values.auth.sasl.secretRef )) }}
- name: users
  secret:
    secretName: {{ .Values.auth.sasl.secretRef }}
  {{- end }}
{{- end -}}

{{/* the default set of volumes for most pods, except the sts pod */}}
{{- define "default-volumes" -}}
- name: config
  configMap:
    name: {{ include "redpanda.fullname" . }}
{{- include "common-volumes" . }}
{{- end -}}

{{/* support legacy tiered storage type selection */}}
{{- define "storage-tiered-mountType" -}}
  {{- $mountType := .Values.storage.tiered.mountType -}}
  {{- if dig "tieredStoragePersistentVolume" "enabled" false .Values.storage -}}
    {{- $mountType = "persistentVolume" -}}
  {{- else if dig "tieredStorageHostPath" false .Values.storage -}}
    {{- $mountType = "hostPath" -}}
  {{- end -}}
  {{- $mountType -}}
{{- end -}}

{{/* support legacy storage.tieredStoragePersistentVolume */}}
{{- define "storage-tiered-persistentvolume" -}}
  {{- $pv := dig "tieredStoragePersistentVolume" .Values.storage.tiered.persistentVolume .Values.storage | toJson -}}
  {{- if empty $pv -}}
    {{- fail "storage.tiered.mountType is \"persistentVolume\" but storage.tiered.persistentVolume is not configured" -}}
  {{- end -}}
  {{- $pv -}}
{{- end -}}

{{/* support legacy storage.tieredStorageHostPath */}}
{{- define "storage-tiered-hostpath" -}}
  {{- $hp := dig "tieredStorageHostPath" .Values.storage.tiered.hostPath .Values.storage -}}
  {{- if empty $hp -}}
    {{- fail "storage.tiered.mountType is \"hostPath\" but storage.tiered.hostPath is empty" -}}
  {{- end -}}
  {{- $hp -}}
{{- end -}}

{{/* support legacy storage.tieredConfig */}}
{{- define "storage-tiered-config" -}}
  {{- dig "tieredConfig" .Values.storage.tiered.config .Values.storage | toJson -}}
{{- end -}}

{{/*
  rpk sasl environment variables

  this will return a string with the correct environment variables to use for SASL based on the
  version of the redpada container being used
*/}}
{{- define "rpk-sasl-environment-variables" -}}
{{- if (include "redpanda-atleast-23-2-1" . | fromJson).bool -}}
RPK_USER RPK_PASS RPK_SASL_MECHANISM
{{- else -}}
REDPANDA_SASL_USERNAME REDPANDA_SASL_PASSWORD REDPANDA_SASL_MECHANISM
{{- end -}}
{{- end -}}

{{- define "curl-options" -}}
{{- print " -svm3 --fail --retry \"120\" --retry-max-time \"120\" --retry-all-errors -o - -w \"\\nstatus=%{http_code} %{redirect_url} size=%{size_download} time=%{time_total} content-type=\\\"%{content_type}\\\"\\n\" "}}
{{- end -}}

{{- define "advertised-address-template" -}}
  {{- $prefixTemplate := dig "prefixTemplate" "" .externalListener -}}
  {{- if empty $prefixTemplate -}}
    {{- $prefixTemplate = dig "prefixTemplate" "" .externalVals -}}
  {{- end -}}
  {{ quote $prefixTemplate }}
{{- end -}}

{{/* check if client auth is enabled for any of the listeners */}}
{{- define "client-auth-required" -}}
{{- $requireClientAuth := get ((include "redpanda.ClientAuthRequired" (dict "a" (list .))) | fromJson) "r" }}
{{- toJson (dict "bool" $requireClientAuth) -}}
{{- end -}}

{{/* secret-ref-or-value
        in: {Value: string?, SecretKey: string?, SecretName: string?}
        out: corev1.Envvar | nil
    secret-ref-or-value converts a set of values into a structure suitable for
    use as an environment variable or nil.
*/}}
{{- define "secret-ref-or-value" -}}
    {{- if and (empty .Value) (or (empty .SecretName) (empty .SecretKey)) -}}
        {{- mustToJson nil -}}
    {{- else -}}
        {{- $out := (dict
            "name" .Name
            "value" .Value
            "valueFrom" (dict
                "secretKeyRef" (dict
                    "name" .SecretName
                    "key" .SecretKey
                )
            )
        ) -}}
        {{- if empty .Value -}}
            {{- $_ := unset $out "value" -}}
        {{- else -}}
            {{- $_ := unset $out "valueFrom" -}}
        {{- end -}}
        {{- mustToJson $out -}}
    {{- end -}}
{{- end -}}

{{- define "tiered-storage-env-vars" -}}
    {{- $config := (include "storage-tiered-config" . | fromJson) -}}
    [
        {{- if and (include "is-licensed" . | fromJson).bool (dig "cloud_storage_enabled" false $config) -}}
            {{include "secret-ref-or-value" (dict
                "Name" "RPK_CLOUD_STORAGE_SECRET_KEY"
                "Value" (dig "cloud_storage_secret_key" nil $config)
                "SecretName" (dig "tiered" "credentialsSecretRef" "secretKey" "name" nil .Values.storage)
                "SecretKey" (dig "tiered" "credentialsSecretRef" "secretKey" "key" nil .Values.storage)
            )}}
            ,
            {{include "secret-ref-or-value" (dict
                "Name" "RPK_CLOUD_STORAGE_ACCESS_KEY"
                "Value" (dig "cloud_storage_access_key" nil $config)
                "SecretName" (dig "tiered" "credentialsSecretRef" "accessKey" "name" nil .Values.storage)
                "SecretKey" (dig "tiered" "credentialsSecretRef" "accessKey" "key" nil .Values.storage)
            )}}

            {{/* Because these keys can be set via secrets, they're special
            cased above. Remove them so they don't get duplicated. */}}
            {{- $_ := unset $config "cloud_storage_access_key" -}}
            {{- $_ := unset $config "cloud_storage_secret_key" -}}

            {{/* iterate over the sorted keys of $config for deterministic output */}}
            {{- range $i, $key := ($config | keys | sortAlpha) -}}
                {{- $value := (get $config $key) -}}

                {{/* Special case for cache size */}}
                {{- if eq $key "cloud_storage_cache_size" -}}
                    {{- $value = (include "SI-to-bytes" $value | int64) -}}
                {{- end -}}

                ,

                {{/* Only include values that are truthy OR that are booleans */}}
                {{- if or (eq (typeOf $value) "bool") $value -}}
                    {{include "secret-ref-or-value" (dict
                        "Name" (printf "RPK_%s" ($key | upper))
                        "Value" ($value | toJson)
                    )}}
                {{- else -}}
                    null
                {{- end -}}
            {{- end -}}
        {{- end -}}
    ]
{{- end -}}
