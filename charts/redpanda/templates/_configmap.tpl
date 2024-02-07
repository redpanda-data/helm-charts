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

{{- define "configmap-content-no-seed" -}}
{{- /*
  configmap content without seed list.
*/ -}}
{{- $root := . }}
{{- $values := .Values }}

{{- /*
  It's impossible to do a rolling upgrade from not-tls-enabled rpc to tls-enabled rpc.
*/ -}}
{{- $check := list
  (include "redpanda-atleast-23-1-2" .|fromJson).bool
  (include "redpanda-22-3-atleast-22-3-13" .|fromJson).bool
  (include "redpanda-22-2-atleast-22-2-10" .|fromJson).bool
-}}
{{- $wantedRPCTLS := (include "rpc-tls-enabled" . | fromJson).bool -}}
{{- if and (not (mustHas true $check)) $wantedRPCTLS -}}
  {{- fail (printf "Redpanda version v%s does not support TLS on the RPC port. Please upgrade. See technical service bulletin 2023-01." (include "redpanda.semver" .)) -}}
{{- end -}}
{{- $cm := lookup "v1" "ConfigMap" .Release.Namespace (include "redpanda.fullname" .) -}}
{{- $redpandaYAML := dig "data" "redpanda.yaml" "" $cm | fromYaml -}}
{{- $currentRPCTLS := dig "redpanda" "rpc_server_tls" "enabled" false $redpandaYAML -}}
{{- /* Lookup will return an empty map when running `helm template` or when `--dry-run` is passed. */ -}}
{{- if (and .Release.IsUpgrade $cm) -}}
  {{- if ne $currentRPCTLS $wantedRPCTLS -}}
    {{- if eq (get .Values "force" | default false) false -}}
      {{- fail (join "\n" (list
          (printf "\n\nError: Cannot do a rolling restart to enable or disable tls at the RPC layer: changing listeners.rpc.tls.enabled (redpanda.yaml:repdanda.rpc_server_tls.enabled) from %v to %v" $currentRPCTLS $wantedRPCTLS)
          "***WARNING The following instructions will result in a short period of downtime."
          "To accept this risk, run the upgrade again adding `--force=true` and do the following:\n"
          "While helm is upgrading the release, manually delete ALL the pods:"
          (printf "    kubectl -n %s delete pod -l app.kubernetes.io/component=redpanda-statefulset" .Release.Namespace)
          "\nIf you got here thinking rpc tls was already enabled, see technical service bulletin 2023-01."
          ))
      -}}
    {{- end -}}
  {{- end -}}
{{- end -}}

{{- $users := list -}}
{{- if (include "sasl-enabled" . | fromJson).bool -}}
  {{- range $user := .Values.auth.sasl.users -}}
    {{- $users = append $users $user.name -}}
  {{- end -}}
{{- end -}}

bootstrap.yaml: |
  kafka_enable_authorization: {{ (include "sasl-enabled" . | fromJson).bool }}
  enable_sasl: {{ (include "sasl-enabled" . | fromJson).bool }}
  enable_rack_awareness: {{ .Values.rackAwareness.enabled }}
{{- with $users }}
  superusers: {{ toYaml . | nindent 4 }}
{{- end }}
{{- with (dig "cluster" dict .Values.config) }}
    {{- range $key, $element := .}}
      {{- if eq $key "default_topic_replications" }}
        {{/* "sub (add $i (mod $i 2)) 1" calculates the closest odd number less than or equal to $element: 1=1, 2=1, 3=3, ... */}}
        {{- $r := $.Values.statefulset.replicas }}
        {{- $element = min $element (sub (add $r (mod $r 2)) 1) }}
      {{- end }}
      {{- if eq (typeOf $element) "bool" }}
        {{- dict $key $element | toYaml | nindent 2 }}
      {{- else if eq (typeOf $element) "[]interface {}" }}
        {{- if not ( empty $element ) }}
      {{ dict $key $element | toYaml | nindent 2 }}
        {{- end }}
      {{- else if $element }}
        {{- dict $key $element | toYaml | nindent 2 }}
      {{- end }}
    {{- end }}
  {{- end }}
  {{- include "tunable" . | nindent 2 }}
  {{- if and (not (hasKey .Values.config.cluster "storage_min_free_bytes")) ((include "redpanda-atleast-22-2-0" . | fromJson).bool) }}
  storage_min_free_bytes: {{ include "storage-min-free-bytes" . }}
  {{- end }}
{{/* AUDIT LOGS */}}
{{- if (include "redpanda-atleast-23-3-0" . | fromJson).bool }}
  {{- if and ( dig "enabled" "false" .Values.auditLogging ) (include "sasl-enabled" $root | fromJson).bool }}
  audit_enabled: true
  {{- if not (eq (int .Values.auditLogging.clientMaxBufferSize) 16777216 ) }}
  audit_client_max_buffer_size: {{ .Values.auditLogging.clientMaxBufferSize }}
  {{- end }}
  {{- if not (eq (int .Values.auditLogging.queueDrainIntervalMs) 500) }}
  audit_queue_drain_interval_ms: {{ .Values.auditLogging.queueDrainIntervalMs }}
  {{- end }}
  {{- if not (eq (int .Values.auditLogging.queueMaxBufferSizePerShard) 1048576) }}
  audit_queue_max_buffer_size_per_shard: {{ .Values.auditLogging.queueMaxBufferSizePerShard }}
  {{- end }}
  {{- if not (eq (int .Values.auditLogging.partitions) 12) }}
  audit_log_num_partitions: {{ .Values.auditLogging.partitions }}
  {{- end }}
  {{- if (dig "replicationFactor" "" .Values.auditLogging) }}
  audit_log_replication_factor: {{ .Values.auditLogging.replicationFactor }}
  {{- end }}
    {{- if dig "enabledEventTypes" "" .Values.auditLogging }}
  audit_enabled_event_types:
      {{- with .Values.auditLogging.enabledEventTypes }}
        {{- toYaml . | nindent 2 }}
      {{- end }}
    {{- end }}
    {{- if dig "excludedTopics" "" .Values.auditLogging }}
  audit_excluded_topics:
      {{- with .Values.auditLogging.excludedTopics }}
        {{- toYaml . | nindent 2 }}
      {{- end }}
    {{- end }}
    {{- if dig "excludedPrincipals" "" .Values.auditLogging }}
  audit_excluded_principals:
      {{- with .Values.auditLogging.excludedPrincipals }}
        {{- toYaml . | nindent 2 }}
      {{- end }}
    {{- end }}
  {{- else }}
  audit_enabled: false
  {{- end }}
{{- end }}

redpanda.yaml: |
  config_file: /etc/redpanda/redpanda.yaml
{{- if .Values.logging.usageStats.enabled }}
  {{- with (dig "usageStats" "organization" "" .Values.logging) }}
  organization: {{ . }}
  {{- end }}
  {{- with (dig "usageStats" "clusterId" "" .Values.logging) }}
  cluster_id: {{ . }}
  {{- end }}
{{- end }}
  redpanda:
{{- if (include "redpanda-atleast-22-3-0" . | fromJson).bool }}
    empty_seed_starts_cluster: false
{{- end }}
    kafka_enable_authorization: {{ (include "sasl-enabled" . | fromJson).bool }}
    enable_sasl: {{ (include "sasl-enabled" . | fromJson).bool }}
{{- if $users }}
    superusers: {{ toJson $users }}
{{- end }}
{{- with (dig "cluster" dict .Values.config) }}
  {{- range $key, $element := . }}
    {{- if eq (typeOf $element) "bool"  }}
    {{ $key }}: {{ $element | toYaml }}
    {{- else if eq (typeOf $element) "[]interface {}" }}
      {{- if not ( empty $element ) }}
    {{ $key }}: {{ $element | toYaml | nindent 4 }}
      {{- end }}
    {{- else if $element }}
    {{ $key }}: {{ $element | toYaml }}
    {{- end }}
  {{- end }}
{{- end }}
{{- with (dig "tunable" dict .Values.config) }}
  {{- range $key, $element := .}}
    {{- if or (eq (typeOf $element) "bool") $element }}
    {{ $key }}: {{ $element | toYaml }}
    {{- end }}
  {{- end }}
{{- end }}
{{- if not (hasKey .Values.config.cluster "storage_min_free_bytes") }}
    storage_min_free_bytes: {{ include "storage-min-free-bytes" . }}
{{- end }}
{{- with dig "node" dict .Values.config }}
  {{- range $key, $element := .}}
    {{- $line := dict $key (toYaml $element) }}
    {{- if and (eq $key "crash_loop_limit") (not (include "redpanda-atleast-23-1-1" $root | fromJson).bool) }}
      {{- $line = dict }}
    {{- end }}
    {{- if not (or (eq (typeOf $element) "bool") $element) }}
      {{- $line = dict }}
    {{- end }}
    {{- with $line }}
      {{  toYaml . | nindent 4 }}
    {{- end }}
  {{- end }}
{{- end -}}
{{/* AUDIT LOGS */}}
{{- if (include "redpanda-atleast-23-3-0" . | fromJson).bool }}
  {{- if and ( dig "enabled" "false" .Values.auditLogging ) (include "sasl-enabled" $root | fromJson).bool }}
    audit_enabled: true
  {{- if not (eq (int .Values.auditLogging.clientMaxBufferSize) 16777216) }}
    audit_client_max_buffer_size: {{ .Values.auditLogging.clientMaxBufferSize }}
  {{- end }}
  {{- if not (eq (int .Values.auditLogging.queueDrainIntervalMs) 500) }}
    audit_queue_drain_interval_ms: {{ .Values.auditLogging.queueDrainIntervalMs }}
  {{- end }}
  {{- if not (eq (int .Values.auditLogging.queueMaxBufferSizePerShard) 1048576) }}
    audit_queue_max_buffer_size_per_shard: {{ .Values.auditLogging.queueMaxBufferSizePerShard }}
  {{- end }}
  {{- if not (eq (int .Values.auditLogging.partitions) 12) }}
    audit_log_num_partitions: {{ .Values.auditLogging.partitions }}
  {{- end }}
    {{- if dig "enabledEventTypes" "" .Values.auditLogging }}
    audit_enabled_event_types:
      {{- with .Values.auditLogging.enabledEventTypes }}
        {{- toYaml . | nindent 4 }}
      {{- end }}
    {{- end }}
    {{- if dig "excludedTopics" "" .Values.auditLogging }}
    audit_excluded_topics:
      {{- with .Values.auditLogging.excludedTopics }}
        {{- toYaml . | nindent 4 }}
      {{- end }}
    {{- end }}
    {{- if dig "excludedPrincipals" "" .Values.auditLogging }}
    audit_excluded_principals:
      {{- with .Values.auditLogging.excludedPrincipals }}
        {{- toYaml . | nindent 4 }}
      {{- end }}
    {{- end }}
  {{- else }}
    audit_enabled: false
  {{- end }}
{{- end }}
{{/* LISTENERS */}}
{{/* Admin API */}}
{{- $service := .Values.listeners.admin }}
    admin:
      - name: internal
        address: 0.0.0.0
        port: {{ $service.port }}
{{- range $name, $listener := $service.external }}
  {{- if and $listener.port $name (dig "enabled" true $listener) }}
      - name: {{ $name }}
        address: 0.0.0.0
        port: {{ $listener.port }}
  {{- end }}
{{- end }}
    admin_api_tls:
{{- if (include "admin-internal-tls-enabled" . | fromJson).bool }}
      - name: internal
        enabled: true
        cert_file: /etc/tls/certs/{{ $service.tls.cert }}/tls.crt
        key_file: /etc/tls/certs/{{ $service.tls.cert }}/tls.key
        require_client_auth: {{ $service.tls.requireClientAuth }}
  {{- $cert := get .Values.tls.certs $service.tls.cert }}
  {{- if empty $cert }}
    {{- fail (printf "Certificate used but not defined")}}
  {{- end }}
  {{- if $cert.caEnabled }}
        truststore_file: /etc/tls/certs/{{ $service.tls.cert }}/ca.crt
  {{- else }}
        {{/* This is a required field so we use the default in the redpanda debian container */}}
        truststore_file: /etc/ssl/certs/ca-certificates.crt
  {{- end }}
{{- end }}
{{- range $name, $listener := $service.external }}
  {{- $k := dict "Values" $values "listener" $listener }}
  {{- if and (include "admin-external-tls-enabled" $k | fromJson).bool (dig "enabled" true $listener) }}
    {{- $mtls := dig "tls" "requireClientAuth" false $listener }}
    {{- $mtls = dig "tls" "requireClientAuth" $mtls $k }}
    {{- $certName := include "admin-external-tls-cert" $k }}
    {{- $certPath := printf "/etc/tls/certs/%s" $certName }}
    {{- $cert := get $values.tls.certs $certName }}
    {{- if empty $cert }}
      {{- fail (printf "Certificate, '%s', used but not defined" $certName)}}
    {{- end }}
      - name: {{ $name }}
        enabled: true
        cert_file: {{ $certPath }}/tls.crt
        key_file: {{ $certPath }}/tls.key
        require_client_auth: {{ $mtls }}
    {{- if $cert.caEnabled }}
        truststore_file: {{ $certPath }}/ca.crt
    {{- else }}
        {{- /* This is a required field so we use the default in the redpanda debian container */}}
        truststore_file: /etc/ssl/certs/ca-certificates.crt
    {{- end }}
  {{- end }}
{{- end -}}
{{/* Kafka API */}}
{{- $kafkaService := .Values.listeners.kafka }}
    kafka_api:
      - name: internal
        address: 0.0.0.0
        port: {{ $kafkaService.port }}
{{- if or (include "sasl-enabled" $root | fromJson).bool $kafkaService.authenticationMethod }}
        authentication_method: {{ default "sasl" $kafkaService.authenticationMethod }}
{{- end }}
{{- range $name, $listener := $kafkaService.external }}
  {{- if and $listener.port $name (dig "enabled" true $listener) }}
      - name: {{ $name }}
        address: 0.0.0.0
        port: {{ $listener.port }}
    {{- if or (include "sasl-enabled" $root | fromJson).bool $listener.authenticationMethod }}
        authentication_method: {{ default "sasl" $listener.authenticationMethod }}
    {{- end }}
  {{- end }}
{{- end }}
    kafka_api_tls:
{{- if (include "kafka-internal-tls-enabled" . | fromJson).bool }}
      - name: internal
        enabled: true
        cert_file: /etc/tls/certs/{{ $kafkaService.tls.cert }}/tls.crt
        key_file: /etc/tls/certs/{{ $kafkaService.tls.cert }}/tls.key
        require_client_auth: {{ $kafkaService.tls.requireClientAuth }}
  {{- $cert := get .Values.tls.certs $kafkaService.tls.cert }}
  {{- if empty $cert }}
    {{- fail (printf "Certificate used but not defined")}}
  {{- end }}
  {{- if $cert.caEnabled }}
        truststore_file: /etc/tls/certs/{{ $kafkaService.tls.cert }}/ca.crt
  {{- else }}
        {{/* This is a required field so we use the default in the redpanda debian container */}}
        truststore_file: /etc/ssl/certs/ca-certificates.crt
  {{- end }}
{{- end }}
{{- range $name, $listener := $kafkaService.external }}
  {{- $k := dict "Values" $values "listener" $listener }}
  {{- if and (include "kafka-external-tls-enabled" $k | fromJson).bool (dig "enabled" true $listener) }}
    {{- $mtls := dig "tls" "requireClientAuth" false $listener }}
    {{- $mtls = dig "tls" "requireClientAuth" $mtls $k }}
    {{- $certName := include "kafka-external-tls-cert" $k }}
    {{- $certPath := printf "/etc/tls/certs/%s" $certName }}
    {{- $cert := get $values.tls.certs $certName }}
    {{- if empty $cert }}
      {{- fail (printf "Certificate, '%s', used but not defined" $certName)}}
    {{- end }}
      - name: {{ $name }}
        enabled: true
        cert_file: {{ $certPath }}/tls.crt
        key_file: {{ $certPath }}/tls.key
        require_client_auth: {{ $mtls }}
    {{- if $cert.caEnabled }}
        truststore_file: {{ $certPath }}/ca.crt
    {{- else }}
        {{/* This is a required field so we use the default in the redpanda debian container */}}
        truststore_file: /etc/ssl/certs/ca-certificates.crt
    {{- end }}
  {{- end }}
{{- end -}}
{{/* RPC Server */}}
{{- $service = .Values.listeners.rpc }}
    rpc_server:
      address: 0.0.0.0
      port: {{ $service.port }}
{{- if (include "rpc-tls-enabled" . | fromJson).bool }}
    rpc_server_tls:
      enabled: true
      cert_file: /etc/tls/certs/{{ $service.tls.cert }}/tls.crt
      key_file: /etc/tls/certs/{{ $service.tls.cert }}/tls.key
      require_client_auth: {{ $service.tls.requireClientAuth }}
  {{- $cert := get .Values.tls.certs $service.tls.cert }}
  {{- if empty $cert }}
    {{- fail (printf "Certificate used but not defined")}}
  {{- end }}
  {{- if $cert.caEnabled }}
      truststore_file: /etc/tls/certs/{{ $service.tls.cert }}/ca.crt
  {{- else }}
      {{- /* This is a required field so we use the default in the redpanda debian container */}}
      truststore_file: /etc/ssl/certs/ca-certificates.crt
  {{- end }}
{{- end -}}
{{- with $root.tempConfigMapServerList }}
    seed_servers: {{ toYaml . | nindent 6 }}
{{- end }}
{{- if and (include "is-licensed" . | fromJson).bool (include "storage-tiered-config" .|fromJson).cloud_storage_enabled }}
  {{- $tieredStorageConfig := (include "storage-tiered-config" .|fromJson) }}
  {{- if not (include "redpanda-atleast-22-3-0" . | fromJson).bool }}
    {{- $tieredStorageConfig = unset $tieredStorageConfig "cloud_storage_credentials_source" }}
  {{- end }}
  {{- range $key, $element := $tieredStorageConfig }}
    {{- if or (eq (typeOf $element) "bool") $element }}
      {{- if eq $key "cloud_storage_cache_size" }}
        {{- dict $key (include "SI-to-bytes" $element) | toYaml | nindent 2 -}}
      {{- else }}
        {{- dict $key $element | toYaml | nindent 2 -}}
      {{- end }}
    {{- end }}
  {{- end }}
{{- end }}
{{/* Schema Registry API */}}
{{- if and .Values.listeners.schemaRegistry.enabled (include "redpanda-22-2-x-without-sasl" $root | fromJson).bool }}
  {{- $schemaRegistryService := .Values.listeners.schemaRegistry }}
  schema_registry_client:
    brokers:
    {{- range $id, $item := $root.tempConfigMapServerList }}
    - address: {{ $item.host.address }}
      port: {{  $kafkaService.port }}
    {{- end }}
    {{- if (include "kafka-internal-tls-enabled" . | fromJson).bool }}
    broker_tls:
      enabled: true
      require_client_auth: {{ $kafkaService.tls.requireClientAuth }}
      cert_file: /etc/tls/certs/{{ $kafkaService.tls.cert }}/tls.crt
      key_file: /etc/tls/certs/{{ $kafkaService.tls.cert }}/tls.key
      {{- $cert := get .Values.tls.certs $kafkaService.tls.cert }}
      {{- if empty $cert }}
        {{- fail (printf "Certificate used but not defined")}}
      {{- end }}
      {{- if $cert.caEnabled }}
      truststore_file: /etc/tls/certs/{{ $kafkaService.tls.cert }}/ca.crt
      {{- else }}
      {{- /* This is a required field so we use the default in the redpanda debian container */}}
      truststore_file: /etc/ssl/certs/ca-certificates.crt
      {{- end }}
    {{- end }}
      {{- with .Values.config.schema_registry_client }}
        {{- toYaml . | nindent 6 }}
      {{- end }}
  schema_registry:
    schema_registry_api:
      - name: internal
        address: 0.0.0.0
        port: {{ $schemaRegistryService.port }}
          {{- if or (include "sasl-enabled" $root | fromJson).bool $schemaRegistryService.authenticationMethod }}
        authentication_method: {{ default "http_basic" $schemaRegistryService.authenticationMethod }}
          {{- end }}
  {{- range $name, $listener := $schemaRegistryService.external }}
    {{- if dig "enabled" true $listener }}
      - name: {{ $name }}
        address: 0.0.0.0
          {{- /*
            when upgrading from an older version that had a missing port, fail if we cannot guess a default
            this should work in all cases as the older versions would have failed with multiple listeners anyway
          */}}
          {{- if and (empty $listener.port) (ne (len $schemaRegistryService.external) 1) }}
            {{- fail "missing required port for schemaRegistry listener $listener.name" }}
          {{- end }}
        port: {{ $listener.port }}
          {{- if or (include "sasl-enabled" $root | fromJson).bool $listener.authenticationMethod }}
        authentication_method: {{ default "http_basic" $listener.authenticationMethod }}
          {{- end }}
    {{- end }}
  {{- end }}
    schema_registry_api_tls:
  {{- if (include "schemaRegistry-internal-tls-enabled" . | fromJson).bool }}
      - name: internal
        enabled: true
        cert_file: /etc/tls/certs/{{ $schemaRegistryService.tls.cert }}/tls.crt
        key_file: /etc/tls/certs/{{ $schemaRegistryService.tls.cert }}/tls.key
        require_client_auth: {{ $schemaRegistryService.tls.requireClientAuth }}
    {{- $cert := get .Values.tls.certs $schemaRegistryService.tls.cert }}
    {{- if empty $cert }}
      {{- fail ( printf "Certificate used but not defined" )}}
    {{- end }}
    {{- if $cert.caEnabled }}
        truststore_file: /etc/tls/certs/{{ $schemaRegistryService.tls.cert }}/ca.crt
    {{- else }}
        {{- /* This is a required field so we use the default in the redpanda debian container */}}
        truststore_file: /etc/ssl/certs/ca-certificates.crt
    {{- end }}
  {{- end }}
  {{- range $name, $listener := $schemaRegistryService.external }}
    {{- $k := dict "Values" $values "listener" $listener }}
    {{- if and (include "schemaRegistry-external-tls-enabled" $k | fromJson).bool (dig "enabled" true $listener) }}
      {{- $mtls := dig "tls" "requireClientAuth" false $listener }}
      {{- $mtls = dig "tls" "requireClientAuth" $mtls $k }}
      {{- $certName := include "schemaRegistry-external-tls-cert" $k }}
      {{- $certPath := printf "/etc/tls/certs/%s" $certName }}
      {{- $cert := get $values.tls.certs $certName }}
      {{- if empty $cert }}
        {{- fail ( printf "Certificate, '%s', used but not defined" $certName )}}
      {{- end }}
      - name: {{ $name }}
        enabled: true
        cert_file: {{ $certPath }}/tls.crt
        key_file: {{ $certPath }}/tls.key
        require_client_auth: {{ $mtls }}
      {{- if $cert.caEnabled }}
        truststore_file: {{ $certPath }}/ca.crt
      {{- else }}
        {{- /* This is a required field so we use the default in the redpanda debian container */}}
        truststore_file: /etc/ssl/certs/ca-certificates.crt
      {{- end }}
    {{- end }}
  {{- end }}
{{- end -}}
{{/* AUDIT LOGS: Client Details */}}
{{- if (include "redpanda-atleast-23-3-0" . | fromJson).bool }}
  {{- if and ( dig "enabled" "false" .Values.auditLogging ) (include "sasl-enabled" $root | fromJson).bool }}
    {{- if not ( empty ( include "kafka-brokers-sasl-enabled" . | fromJson ) ) }}
  audit_log_client:
    {{- include "kafka-brokers-sasl-enabled" . | nindent 4 -}}
    {{- end }}
  {{- end }}
{{- end }}
{{/* HTTP Proxy */}}
{{- if and .Values.listeners.http.enabled (include "redpanda-22-2-x-without-sasl" $root | fromJson).bool }}
  {{- $HTTPService := .Values.listeners.http }}
  pandaproxy_client:
    brokers:
  {{- range $id, $item := $root.tempConfigMapServerList }}
    - address: {{ $item.host.address }}
      port: {{  $kafkaService.port }}
  {{- end }}
  {{- if (include "kafka-internal-tls-enabled" . | fromJson).bool }}
    broker_tls:
      enabled: true
      require_client_auth: {{ $kafkaService.tls.requireClientAuth }}
      cert_file: /etc/tls/certs/{{ $kafkaService.tls.cert }}/tls.crt
      key_file: /etc/tls/certs/{{ $kafkaService.tls.cert }}/tls.key
  {{- $cert := get .Values.tls.certs $kafkaService.tls.cert }}
  {{- if empty $cert }}
    {{- fail (printf "Certificate used but not defined")}}
  {{- end }}
  {{- if $cert.caEnabled }}
      truststore_file: /etc/tls/certs/{{ $kafkaService.tls.cert }}/ca.crt
  {{- else }}
        {{- /* This is a required field so we use the default in the redpanda debian container */}}
      truststore_file: /etc/ssl/certs/ca-certificates.crt
  {{- end }}
  {{- with .Values.config.pandaproxy_client }}
    {{- toYaml . | nindent 6 }}
  {{- end }}
{{- end }}
  pandaproxy:
    pandaproxy_api:
      - name: internal
        address: 0.0.0.0
        port: {{ $HTTPService.port }}
{{- if or (include "sasl-enabled" $root | fromJson).bool $HTTPService.authenticationMethod }}
        authentication_method: {{ default "http_basic" $HTTPService.authenticationMethod }}
{{- end }}
{{- range $name, $listener := $HTTPService.external }}
  {{- if and $listener.port $name (dig "enabled" true $listener) }}
      - name: {{ $name }}
        address: 0.0.0.0
        port: {{ $listener.port }}
    {{- if or (include "sasl-enabled" $root | fromJson).bool $listener.authenticationMethod }}
        authentication_method: {{ default "http_basic" $listener.authenticationMethod }}
    {{- end }}
  {{- end }}
{{- end }}
    pandaproxy_api_tls:
{{- if (include "http-internal-tls-enabled" . | fromJson).bool }}
      - name: internal
        enabled: true
        cert_file: /etc/tls/certs/{{ $HTTPService.tls.cert }}/tls.crt
        key_file: /etc/tls/certs/{{ $HTTPService.tls.cert }}/tls.key
        require_client_auth: {{ $HTTPService.tls.requireClientAuth }}
    {{- $cert := get .Values.tls.certs $HTTPService.tls.cert }}
    {{- if empty $cert }}
      {{- fail (printf "Certificate used but not defined")}}
    {{- end }}
    {{- if $cert.caEnabled }}
        truststore_file: /etc/tls/certs/{{ $HTTPService.tls.cert }}/ca.crt
    {{- else }}
        {{- /* This is a required field so we use the default in the redpanda debian container */}}
        truststore_file: /etc/ssl/certs/ca-certificates.crt
    {{- end }}
  {{- end }}
  {{- range $name, $listener := $HTTPService.external }}
    {{- $k := dict "Values" $values "listener" $listener }}
    {{- if and (include "http-external-tls-enabled" $k | fromJson).bool (dig "enabled" true $listener) }}
      {{- $mtls := dig "tls" "requireClientAuth" false $listener }}
      {{- $mtls = dig "tls" "requireClientAuth" $mtls $k }}
      {{- $certName := include "http-external-tls-cert" $k }}
      {{- $certPath := printf "/etc/tls/certs/%s" $certName }}
      {{- $cert := get $values.tls.certs $certName }}
      {{- if empty $cert }}
        {{- fail (printf "Certificate, '%s', used but not defined" $certName )}}
      {{- end }}
      - name: {{ $name }}
        enabled: true
        cert_file: {{ $certPath }}/tls.crt
        key_file: {{ $certPath }}/tls.key
        require_client_auth: {{ $mtls }}
      {{- if $cert.caEnabled }}
        truststore_file: {{ $certPath }}/ca.crt
      {{- else }}
        {{- /* This is a required field so we use the default in the redpanda debian container */}}
        truststore_file: /etc/ssl/certs/ca-certificates.crt
      {{- end }}
    {{- end }}
  {{- end }}
{{- end }}
{{/* END LISTENERS */}}
{{- end -}}

{{- define "rpk-config-internal" -}}
  {{- $brokers := list -}}
  {{- $admin := list -}}
  {{- range $i := untilStep 0 (.Values.statefulset.replicas|int) 1 -}}
    {{- $podName := printf "%s-%d.%s" (include "redpanda.fullname" $) $i (include "redpanda.internal.domain" $) -}}
    {{- $brokers = concat $brokers (list (printf "%s:%d" $podName (int $.Values.listeners.kafka.port))) -}}
    {{- $admin = concat $admin (list (printf "%s:%d" $podName (int $.Values.listeners.admin.port))) -}}
  {{- end -}}
rpk:
  # redpanda server configuration
  overprovisioned: {{ dig "cpu" "overprovisioned" false .Values.resources }}
  enable_memory_locking: {{ dig "memory" "enable_memory_locking" false .Values.resources }}
  additional_start_flags:
    - "--smp={{ include "redpanda-smp" . }}"
    - "--memory={{ template "redpanda-memory" . }}M"
    {{- if not .Values.config.node.developer_mode }}
    - "--reserve-memory={{ template "redpanda-reserve-memory" . }}M"
    {{- end }}
    - "--default-log-level={{ .Values.logging.logLevel }}"
  {{- with .Values.statefulset.additionalRedpandaCmdFlags -}}
  {{- toYaml . | nindent 4 }}
  {{- end }}

  {{- with dig "config" "rpk" dict .Values.AsMap }}
  # config.rpk entries
  {{- toYaml . | nindent 2 }}
  {{- end }}

  {{- with dig "tuning" dict .Values.AsMap }}
  # rpk tune entries
  {{- toYaml . | nindent 2 }}
  {{- end }}

  # kafka connection configuration
  kafka_api:
    brokers: {{ toYaml $brokers | nindent 6 }}
    tls:
  {{- if (include "kafka-internal-tls-enabled" . | fromJson).bool }}
    {{- $cert := get .Values.tls.certs .Values.listeners.kafka.tls.cert }}
    {{- if $cert.caEnabled }}
      truststore_file: {{ printf "/etc/tls/certs/%s/ca.crt" .Values.listeners.kafka.tls.cert }}
    {{- end }}
    {{- if .Values.listeners.kafka.tls.requireClientAuth }}
      cert_file: {{ printf "/etc/tls/certs/%s-client/tls.crt" (include "redpanda.fullname" .) }}
      key_file: {{ printf "/etc/tls/certs/%s-client/tls.key" (include "redpanda.fullname" .) }}
    {{- end }}
  {{- end }}
  admin_api:
    addresses: {{ toYaml $admin | nindent 6 }}
    tls:
  {{- if (include "admin-internal-tls-enabled" . | fromJson).bool }}
    {{- $cert := get .Values.tls.certs .Values.listeners.admin.tls.cert }}
    {{- if $cert.caEnabled }}
      truststore_file: {{ printf "/etc/tls/certs/%s/ca.crt" .Values.listeners.admin.tls.cert }}
    {{- end }}
    {{- if .Values.listeners.admin.tls.requireClientAuth }}
      cert_file: {{ printf "/etc/tls/certs/%s-client/tls.crt" (include "redpanda.fullname" .) }}
      key_file: {{ printf "/etc/tls/certs/%s-client/tls.key" (include "redpanda.fullname" .) }}
    {{- end }}
  {{- end }}
{{- end -}}

{{- define "configmap-server-list" -}}
  {{- $serverList := list -}}
  {{- range (include "seed-server-list" . | mustFromJson) -}}
    {{- $server := dict "host" (dict "address" . "port" $.Values.listeners.rpc.port) -}}
    {{- $serverList = append $serverList $server -}}
  {{- end -}}
  {{- toJson (dict "serverList" $serverList) -}}
{{- end -}}

{{- define "full-configmap" -}}
  {{- $serverList := (fromJson (include "configmap-server-list" .)).serverList -}}
  {{- $r := set . "tempConfigMapServerList" $serverList -}}
  {{ include "configmap-content-no-seed" $r | nindent 0 }}
  {{ include "rpk-config-internal" $ | nindent 2 }}
{{- end -}}

{{- define "rpk-config-external" -}}
  {{- $brokers := list -}}
  {{- $admin := list -}}
  {{- $profile := keys .Values.listeners.kafka.external | first -}}
  {{- $kafkaListener := get .Values.listeners.kafka.external $profile -}}
  {{- $adminListener := dict -}}
  {{- if .Values.listeners.admin.external -}}
      {{- $adminprofile := keys .Values.listeners.admin.external | first -}}
      {{- $adminListener = get .Values.listeners.admin.external $adminprofile -}}
  {{- end -}}
  {{- range $i := until (.Values.statefulset.replicas|int) -}}
    {{- $externalAdvertiseAddress := printf "%s-%d" (include "redpanda.fullname" $) $i -}}
    {{- if (tpl ($.Values.external.domain | default "") $) -}}
      {{- $externalAdvertiseAddress = printf "%s.%s" $externalAdvertiseAddress (tpl $.Values.external.domain $) -}}
    {{- end -}}
    {{- $tmplVals := dict "listenerVals" $.Values.listeners.kafka "externalVals" $kafkaListener "externalName" $profile "externalAdvertiseAddress" $externalAdvertiseAddress "values" $.Values "replicaIndex" $i -}}
    {{- $port := int (include "advertised-port" $tmplVals) -}}
    {{- $host := fromJson (include "advertised-host" (mustMerge $tmplVals (dict "port" $port) $)) -}}
    {{- $brokers = concat $brokers (list (printf "%s:%d" (get $host "address") (get $host "port" | int))) -}}
    {{- $tmplVals = dict "listenerVals" $.Values.listeners.admin "externalVals" $adminListener "externalName" $profile "externalAdvertiseAddress" $externalAdvertiseAddress "values" $.Values "replicaIndex" $i -}}
    {{- $port = int (include "advertised-port" $tmplVals) -}}
    {{- $host = fromJson (include "advertised-host" (mustMerge $tmplVals (dict "port" $port) $)) -}}
    {{- $admin = concat $admin (list (printf "%s:%d" (get $host "address") (get $host "port" | int))) -}}
  {{- end -}}
name: {{ $profile }}
kafka_api:
  brokers: {{ toYaml $brokers | nindent 6 }}
  tls:
  {{- if and (include "kafka-external-tls-enabled" (dict "Values" .Values "listener" $kafkaListener) | fromJson).bool (dig "enabled" true $adminListener) }}
    {{- $cert := get .Values.tls.certs .Values.listeners.kafka.tls.cert }}
    {{- if $cert.caEnabled }}
    ca_file: ca.crt
    {{- end }}
    {{- if .Values.listeners.kafka.tls.requireClientAuth }}
    cert_file: tls.crt
    key_file: tls.key
    {{- end }}
  {{- end }}
admin_api:
  addresses: {{ toYaml $admin | nindent 6 }}
  tls:
  {{- if and (include "admin-external-tls-enabled" (dict "Values" .Values "listener" $adminListener) | fromJson).bool (dig "enabled" true $adminListener) }}
    {{- $cert := get .Values.tls.certs .Values.listeners.admin.tls.cert }}
    {{- if $cert.caEnabled }}
    ca_file: ca.crt
    {{- end }}
    {{- if .Values.listeners.admin.tls.requireClientAuth }}
    cert_file: tls.crt
    key_file: tls.key
    {{- end }}
  {{- end }}
{{- end -}}
