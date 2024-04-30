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
  {{- if typeIs "string" .Values.storage.persistentVolume.size -}}
    {{- min $fiveGiB (mulf (get ((include "redpanda.SIToBytes" (dict "a" (list .Values.storage.persistentVolume.size))) | fromJson) "r" ) 0.05 | int64) -}}
  {{- else -}}
    {{- min $fiveGiB (mulf .Values.storage.persistentVolume.size 0.05 | int64) -}}
  {{- end -}}
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
{{- toJson (dict "bool" (get ((include "redpanda.RedpandaAtLeast_22_2_0" (dict "a" (list .))) | fromJson) "r")) }}
{{- end -}}
{{- define "redpanda-atleast-22-3-0" -}}
{{- toJson (dict "bool" (get ((include "redpanda.RedpandaAtLeast_22_3_0" (dict "a" (list .))) | fromJson) "r")) }}
{{- end -}}
{{- define "redpanda-atleast-23-1-1" -}}
{{- toJson (dict "bool" (get ((include "redpanda.RedpandaAtLeast_23_1_1" (dict "a" (list .))) | fromJson) "r")) }}
{{- end -}}
{{- define "redpanda-atleast-23-1-2" -}}
{{- toJson (dict "bool" (get ((include "redpanda.RedpandaAtLeast_23_1_2" (dict "a" (list .))) | fromJson) "r")) }}
{{- end -}}
{{- define "redpanda-22-3-atleast-22-3-13" -}}
{{- toJson (dict "bool" (get ((include "redpanda.RedpandaAtLeast_22_3_atleast_22_3_13" (dict "a" (list .))) | fromJson) "r")) }}
{{- end -}}
{{- define "redpanda-22-2-atleast-22-2-10" -}}
{{- toJson (dict "bool" (get ((include "redpanda.RedpandaAtLeast_22_2_atleast_22_2_10" (dict "a" (list .))) | fromJson) "r")) }}
{{- end -}}
{{- define "redpanda-atleast-23-2-1" -}}
{{- toJson (dict "bool" (get ((include "redpanda.RedpandaAtLeast_23_2_1" (dict "a" (list .))) | fromJson) "r")) }}
{{- end -}}
{{- define "redpanda-atleast-23-3-0" -}}
{{- toJson (dict "bool" (get ((include "redpanda.RedpandaAtLeast_23_3_0" (dict "a" (list .))) | fromJson) "r")) }}
{{- end -}}

{{- define "redpanda-22-2-x-without-sasl" -}}
{{- $result :=  (include "redpanda-atleast-22-3-0" . | fromJson).bool -}}
{{- if or (include "sasl-enabled" . | fromJson).bool .Values.listeners.kafka.authenticationMethod -}}
{{-   $result := false -}}
{{- end -}}
{{- toJson (dict "bool" $result) -}}
{{- end -}}

{{- define "pod-security-context" -}}
{{- get ((include "redpanda.PodSecurityContext" (dict "a" (list .))) | fromJson) "r" | toYaml }}
{{- end -}}

{{- define "container-security-context" -}}
{{- get ((include "redpanda.ContainerSecurityContext" (dict "a" (list .))) | fromJson) "r" | toYaml }}
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
  {{- $warnings := list "redpanda-cpu-warning" -}}
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
{{- $mounts := get ((include "redpanda.CommonMounts" (dict "a" (list .))) | fromJson) "r" }}
{{- if $mounts -}}
{{- toYaml $mounts -}}
{{- end -}}
{{- end -}}

{{/* mounts that are common to most containers */}}
{{- define "default-mounts" -}}
{{- $mounts := get ((include "redpanda.DefaultMounts" (dict "a" (list .))) | fromJson) "r" }}
{{- if $mounts -}}
{{- toYaml $mounts -}}
{{- end -}}
{{- end -}}

{{/* volumes that are common to all pods */}}
{{- define "common-volumes" -}}
{{- $volumes := get ((include "redpanda.CommonVolumes" (dict "a" (list .))) | fromJson) "r" }}
{{- if $volumes -}}
{{- toYaml $volumes -}}
{{- end -}}
{{- end -}}

{{/* the default set of volumes for most pods, except the sts pod */}}
{{- define "default-volumes" -}}
{{- $volumes := get ((include "redpanda.DefaultVolumes" (dict "a" (list .))) | fromJson) "r" }}
{{- if $volumes -}}
{{- toYaml $volumes -}}
{{- end -}}
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
