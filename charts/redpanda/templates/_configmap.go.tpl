{{- /* Generated from "configmap.tpl.go" */ -}}

{{- define "redpanda.ConfigMap" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $values := $dot.Values.AsMap -}}
{{- $bootstrap := (dict "kafka_enable_authorization" (get (fromJson (include "redpanda.Auth.IsSASLEnabled" (dict "a" (list $values.auth) ))) "r") "enable_sasl" (get (fromJson (include "redpanda.Auth.IsSASLEnabled" (dict "a" (list $values.auth) ))) "r") "enable_rack_awareness" $values.rackAwareness.enabled "storage_min_free_bytes" (get (fromJson (include "redpanda.Storage.StorageMinFreeBytes" (dict "a" (list $values.storage) ))) "r") ) -}}
{{- $bootstrap = (merge (dict ) $bootstrap (get (fromJson (include "redpanda.AuditLogging.Translate" (dict "a" (list $values.auditLogging $dot (get (fromJson (include "redpanda.Auth.IsSASLEnabled" (dict "a" (list $values.auth) ))) "r")) ))) "r")) -}}
{{- $bootstrap = (merge (dict ) $bootstrap (get (fromJson (include "redpanda.Logging.Translate" (dict "a" (list $values.logging) ))) "r")) -}}
{{- $bootstrap = (merge (dict ) $bootstrap (get (fromJson (include "redpanda.TunableConfig.Translate" (dict "a" (list $values.config.tunable) ))) "r")) -}}
{{- $bootstrap = (merge (dict ) $bootstrap (get (fromJson (include "redpanda.ClusterConfig.Translate" (dict "a" (list $values.config.cluster ($values.statefulset.replicas | int) false) ))) "r")) -}}
{{- $bootstrap = (merge (dict ) $bootstrap (get (fromJson (include "redpanda.Auth.Translate" (dict "a" (list $values.auth (get (fromJson (include "redpanda.Auth.IsSASLEnabled" (dict "a" (list $values.auth) ))) "r")) ))) "r")) -}}
{{- $redpanda := (dict "kafka_enable_authorization" (get (fromJson (include "redpanda.Auth.IsSASLEnabled" (dict "a" (list $values.auth) ))) "r") "enable_sasl" (get (fromJson (include "redpanda.Auth.IsSASLEnabled" (dict "a" (list $values.auth) ))) "r") "empty_seed_starts_cluster" false "storage_min_free_bytes" (get (fromJson (include "redpanda.Storage.StorageMinFreeBytes" (dict "a" (list $values.storage) ))) "r") "seed_servers" (get (fromJson (include "redpanda.Listeners.CreateSeedServers" (dict "a" (list $values.listeners ($values.statefulset.replicas | int) (get (fromJson (include "redpanda.Fullname" (dict "a" (list $dot) ))) "r") (get (fromJson (include "redpanda.InternalDomain" (dict "a" (list $dot) ))) "r")) ))) "r") ) -}}
{{- $redpanda = (merge (dict ) $redpanda (get (fromJson (include "redpanda.AuditLogging.Translate" (dict "a" (list $values.auditLogging $dot (get (fromJson (include "redpanda.Auth.IsSASLEnabled" (dict "a" (list $values.auth) ))) "r")) ))) "r")) -}}
{{- $redpanda = (merge (dict ) $redpanda (get (fromJson (include "redpanda.Logging.Translate" (dict "a" (list $values.logging) ))) "r")) -}}
{{- $redpanda = (merge (dict ) $redpanda (get (fromJson (include "redpanda.TunableConfig.Translate" (dict "a" (list $values.config.tunable) ))) "r")) -}}
{{- $redpanda = (merge (dict ) $redpanda (get (fromJson (include "redpanda.ClusterConfig.Translate" (dict "a" (list $values.config.cluster ($values.statefulset.replicas | int) true) ))) "r")) -}}
{{- $redpanda = (merge (dict ) $redpanda (get (fromJson (include "redpanda.Auth.Translate" (dict "a" (list $values.auth (get (fromJson (include "redpanda.Auth.IsSASLEnabled" (dict "a" (list $values.auth) ))) "r")) ))) "r")) -}}
{{- $redpanda = (merge (dict ) $redpanda (get (fromJson (include "redpanda.NodeConfig.Translate" (dict "a" (list $values.config.node) ))) "r")) -}}
{{- $_ := (get (fromJson (include "redpanda.configureListeners" (dict "a" (list $redpanda $dot) ))) "r") -}}
{{- $redpandaYaml := (dict "redpanda" $redpanda "schema_registry" (get (fromJson (include "redpanda.schemaRegistry" (dict "a" (list $dot) ))) "r") "schema_registry_client" (get (fromJson (include "redpanda.kafkaClient" (dict "a" (list $dot) ))) "r") "pandaproxy" (get (fromJson (include "redpanda.pandaProxyListener" (dict "a" (list $dot) ))) "r") "pandaproxy_client" (get (fromJson (include "redpanda.kafkaClient" (dict "a" (list $dot) ))) "r") "rpk" (get (fromJson (include "redpanda.rpkConfiguration" (dict "a" (list $dot) ))) "r") "config_file" "/etc/redpanda/redpanda.yaml" ) -}}
{{- if (and (and (get (fromJson (include "redpanda.RedpandaAtLeast_23_3_0" (dict "a" (list $dot) ))) "r") $values.auditLogging.enabled) (get (fromJson (include "redpanda.Auth.IsSASLEnabled" (dict "a" (list $values.auth) ))) "r")) -}}
{{- $_ := (set $redpandaYaml "audit_log_client" (get (fromJson (include "redpanda.kafkaClient" (dict "a" (list $dot) ))) "r")) -}}
{{- end -}}
{{- $redpandaYaml = (merge (dict ) $redpandaYaml (get (fromJson (include "redpanda.Storage.Translate" (dict "a" (list $values.storage) ))) "r")) -}}
{{- $cms := (list (mustMergeOverwrite (dict "metadata" (dict "creationTimestamp" (coalesce nil) ) ) (mustMergeOverwrite (dict ) (dict "kind" "ConfigMap" "apiVersion" "v1" )) (dict "metadata" (mustMergeOverwrite (dict "creationTimestamp" (coalesce nil) ) (dict "name" (get (fromJson (include "redpanda.Fullname" (dict "a" (list $dot) ))) "r") "namespace" $dot.Release.Namespace "labels" (get (fromJson (include "redpanda.FullLabels" (dict "a" (list $dot) ))) "r") )) "data" (dict "bootstrap.yaml" (toYaml $bootstrap) "redpanda.yaml" (toYaml $redpandaYaml) ) ))) -}}
{{- if $values.external.enabled -}}
{{- $cms = (mustAppend $cms (mustMergeOverwrite (dict "metadata" (dict "creationTimestamp" (coalesce nil) ) ) (mustMergeOverwrite (dict ) (dict "kind" "ConfigMap" "apiVersion" "v1" )) (dict "metadata" (mustMergeOverwrite (dict "creationTimestamp" (coalesce nil) ) (dict "name" (printf "%s-rpk" (get (fromJson (include "redpanda.Fullname" (dict "a" (list $dot) ))) "r")) "namespace" $dot.Release.Namespace "labels" (get (fromJson (include "redpanda.FullLabels" (dict "a" (list $dot) ))) "r") )) "data" (dict "profile" (toYaml (get (fromJson (include "redpanda.rpkProfile" (dict "a" (list $dot) ))) "r")) ) ))) -}}
{{- end -}}
{{- (dict "r" $cms) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.rpkProfile" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $values := $dot.Values.AsMap -}}
{{- $brokerList := (list ) -}}
{{- range $_, $i := untilStep ((0 | int)|int) (($values.statefulset.replicas | int)|int) (1|int) -}}
{{- $brokerList = (mustAppend $brokerList (printf "%s:%d" (get (fromJson (include "redpanda.advertisedHost" (dict "a" (list $dot $i) ))) "r") (((get (fromJson (include "redpanda.advertisedKafkaPort" (dict "a" (list $dot $i) ))) "r") | int) | int))) -}}
{{- end -}}
{{- $adminAdvertisedList := (list ) -}}
{{- range $_, $i := untilStep ((0 | int)|int) (($values.statefulset.replicas | int)|int) (1|int) -}}
{{- $adminAdvertisedList = (mustAppend $adminAdvertisedList (printf "%s:%d" (get (fromJson (include "redpanda.advertisedHost" (dict "a" (list $dot $i) ))) "r") (((get (fromJson (include "redpanda.advertisedAdminPort" (dict "a" (list $dot $i) ))) "r") | int) | int))) -}}
{{- end -}}
{{- $kafkaTLS := (get (fromJson (include "redpanda.brokersTLSConfiguration" (dict "a" (list $dot) ))) "r") -}}
{{- $tmp_tuple_1 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "_shims.dicttest" (dict "a" (list $kafkaTLS "truststore_file" (coalesce nil)) ))) "r")) ))) "r") -}}
{{- $ok_1 := $tmp_tuple_1.T2 -}}
{{- if $ok_1 -}}
{{- $_ := (set $kafkaTLS "ca_file" "ca.crt") -}}
{{- $_ := (unset $kafkaTLS "truststore_file") -}}
{{- end -}}
{{- $adminTLS := (get (fromJson (include "redpanda.adminTLSConfiguration" (dict "a" (list $dot) ))) "r") -}}
{{- $tmp_tuple_2 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "_shims.dicttest" (dict "a" (list $adminTLS "truststore_file" (coalesce nil)) ))) "r")) ))) "r") -}}
{{- $ok_2 := $tmp_tuple_2.T2 -}}
{{- if $ok_2 -}}
{{- $_ := (set $adminTLS "ca_file" "ca.crt") -}}
{{- $_ := (unset $adminTLS "truststore_file") -}}
{{- end -}}
{{- $ka := (dict "brokers" $brokerList "tls" (coalesce nil) ) -}}
{{- if (gt ((get (fromJson (include "_shims.len" (dict "a" (list $kafkaTLS) ))) "r") | int) (0 | int)) -}}
{{- $_ := (set $ka "tls" $kafkaTLS) -}}
{{- end -}}
{{- $aa := (dict "addresses" $adminAdvertisedList "tls" (coalesce nil) ) -}}
{{- if (gt ((get (fromJson (include "_shims.len" (dict "a" (list $adminTLS) ))) "r") | int) (0 | int)) -}}
{{- $_ := (set $aa "tls" $adminTLS) -}}
{{- end -}}
{{- $result := (dict "name" (get (fromJson (include "redpanda.getFistExternalKafkaListener" (dict "a" (list $dot) ))) "r") "kafka_api" $ka "admin_api" $aa ) -}}
{{- (dict "r" $result) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.advertisedKafkaPort" -}}
{{- $dot := (index .a 0) -}}
{{- $i := (index .a 1) -}}
{{- range $_ := (list 1) -}}
{{- $values := $dot.Values.AsMap -}}
{{- $externalKafkaListenerName := (get (fromJson (include "redpanda.getFistExternalKafkaListener" (dict "a" (list $dot) ))) "r") -}}
{{- $listener := (index $values.listeners.kafka.external $externalKafkaListenerName) -}}
{{- $port := (($values.listeners.kafka.port | int) | int) -}}
{{- if (gt (($listener.port | int) | int) ((1 | int) | int)) -}}
{{- $port = (($listener.port | int) | int) -}}
{{- end -}}
{{- if (gt ((get (fromJson (include "_shims.len" (dict "a" (list $listener.advertisedPorts) ))) "r") | int) (1 | int)) -}}
{{- $port = ((index $listener.advertisedPorts $i) | int) -}}
{{- else -}}{{- if (eq ((get (fromJson (include "_shims.len" (dict "a" (list $listener.advertisedPorts) ))) "r") | int) (1 | int)) -}}
{{- $port = ((index $listener.advertisedPorts (0 | int)) | int) -}}
{{- end -}}
{{- end -}}
{{- (dict "r" $port) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.advertisedAdminPort" -}}
{{- $dot := (index .a 0) -}}
{{- $i := (index .a 1) -}}
{{- range $_ := (list 1) -}}
{{- $values := $dot.Values.AsMap -}}
{{- $keys := (keys $values.listeners.admin.external) -}}
{{- $_ := (sortAlpha $keys) -}}
{{- $externalAdminListenerName := (first $keys) -}}
{{- $listener := (index $values.listeners.admin.external (get (fromJson (include "_shims.typeassertion" (dict "a" (list "string" $externalAdminListenerName) ))) "r")) -}}
{{- $port := (($values.listeners.admin.port | int) | int) -}}
{{- if (gt (($listener.port | int) | int) (1 | int)) -}}
{{- $port = (($listener.port | int) | int) -}}
{{- end -}}
{{- if (gt ((get (fromJson (include "_shims.len" (dict "a" (list $listener.advertisedPorts) ))) "r") | int) (1 | int)) -}}
{{- $port = ((index $listener.advertisedPorts $i) | int) -}}
{{- else -}}{{- if (eq ((get (fromJson (include "_shims.len" (dict "a" (list $listener.advertisedPorts) ))) "r") | int) (1 | int)) -}}
{{- $port = ((index $listener.advertisedPorts (0 | int)) | int) -}}
{{- end -}}
{{- end -}}
{{- (dict "r" $port) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.advertisedHost" -}}
{{- $dot := (index .a 0) -}}
{{- $i := (index .a 1) -}}
{{- range $_ := (list 1) -}}
{{- $values := $dot.Values.AsMap -}}
{{- $address := (printf "%s-%d" (get (fromJson (include "redpanda.Fullname" (dict "a" (list $dot) ))) "r") ($i | int)) -}}
{{- if (ne (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $values.external.domain "") ))) "r") "") -}}
{{- $address = (printf "%s.%s" $address (tpl $values.external.domain $dot)) -}}
{{- end -}}
{{- if (le ((get (fromJson (include "_shims.len" (dict "a" (list $values.external.addresses) ))) "r") | int) (0 | int)) -}}
{{- (dict "r" $address) | toJson -}}
{{- break -}}
{{- end -}}
{{- if (eq ((get (fromJson (include "_shims.len" (dict "a" (list $values.external.addresses) ))) "r") | int) (1 | int)) -}}
{{- $address = (index $values.external.addresses (0 | int)) -}}
{{- else -}}
{{- $address = (index $values.external.addresses $i) -}}
{{- end -}}
{{- if (ne (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $values.external.domain "") ))) "r") "") -}}
{{- $address = (printf "%s.%s" $address $values.external.domain) -}}
{{- end -}}
{{- (dict "r" $address) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.getFistExternalKafkaListener" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $values := $dot.Values.AsMap -}}
{{- $keys := (keys $values.listeners.kafka.external) -}}
{{- $_ := (sortAlpha $keys) -}}
{{- (dict "r" (get (fromJson (include "_shims.typeassertion" (dict "a" (list "string" (first $keys)) ))) "r")) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.rpkConfiguration" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $values := $dot.Values.AsMap -}}
{{- $brokerList := (list ) -}}
{{- $r := ($values.statefulset.replicas | int) -}}
{{- range $_, $i := untilStep ((0 | int)|int) (($values.statefulset.replicas | int)|int) (1|int) -}}
{{- $brokerList = (mustAppend $brokerList (printf "%s-%d.%s:%d" (get (fromJson (include "redpanda.Fullname" (dict "a" (list $dot) ))) "r") $i (get (fromJson (include "redpanda.InternalDomain" (dict "a" (list $dot) ))) "r") (($values.listeners.kafka.port | int) | int))) -}}
{{- end -}}
{{- $adminTLS := (coalesce nil) -}}
{{- $tls_3 := (get (fromJson (include "redpanda.adminTLSConfiguration" (dict "a" (list $dot) ))) "r") -}}
{{- if (gt ((get (fromJson (include "_shims.len" (dict "a" (list $tls_3) ))) "r") | int) (0 | int)) -}}
{{- $adminTLS = $tls_3 -}}
{{- end -}}
{{- $brokerTLS := (coalesce nil) -}}
{{- $tls_4 := (get (fromJson (include "redpanda.brokersTLSConfiguration" (dict "a" (list $dot) ))) "r") -}}
{{- if (gt ((get (fromJson (include "_shims.len" (dict "a" (list $tls_4) ))) "r") | int) (0 | int)) -}}
{{- $brokerTLS = $tls_4 -}}
{{- end -}}
{{- $result := (dict "overprovisioned" (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $values.resources.cpu.overprovisioned false) ))) "r") "enable_memory_locking" (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $values.resources.memory.enable_memory_locking false) ))) "r") "additional_start_flags" (get (fromJson (include "redpanda.RedpandaAdditionalStartFlags" (dict "a" (list $dot ((get (fromJson (include "redpanda.RedpandaSMP" (dict "a" (list $dot) ))) "r") | int)) ))) "r") "kafka_api" (dict "brokers" $brokerList "tls" $brokerTLS ) "admin_api" (dict "addresses" (get (fromJson (include "redpanda.Listeners.AdminList" (dict "a" (list $values.listeners ($values.statefulset.replicas | int) (get (fromJson (include "redpanda.Fullname" (dict "a" (list $dot) ))) "r") (get (fromJson (include "redpanda.InternalDomain" (dict "a" (list $dot) ))) "r")) ))) "r") "tls" $adminTLS ) ) -}}
{{- $result = (merge (dict ) $result (get (fromJson (include "redpanda.Tuning.Translate" (dict "a" (list $values.tuning) ))) "r")) -}}
{{- $result = (merge (dict ) $result (get (fromJson (include "redpanda.Config.CreateRPKConfiguration" (dict "a" (list $values.config) ))) "r")) -}}
{{- (dict "r" $result) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.brokersTLSConfiguration" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $values := $dot.Values.AsMap -}}
{{- if (not (get (fromJson (include "redpanda.InternalTLS.IsEnabled" (dict "a" (list $values.listeners.kafka.tls $values.tls) ))) "r")) -}}
{{- (dict "r" (dict )) | toJson -}}
{{- break -}}
{{- end -}}
{{- $result := (dict ) -}}
{{- $certName := $values.listeners.kafka.tls.cert -}}
{{- $tmp_tuple_3 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "_shims.dicttest" (dict "a" (list $values.tls.certs $certName (coalesce nil)) ))) "r")) ))) "r") -}}
{{- $ok_6 := $tmp_tuple_3.T2 -}}
{{- $cert_5 := $tmp_tuple_3.T1 -}}
{{- if (and $ok_6 $cert_5.caEnabled) -}}
{{- $_ := (set $result "truststore_file" (printf "/etc/tls/certs/%s/ca.crt" $values.listeners.kafka.tls.cert)) -}}
{{- end -}}
{{- if $values.listeners.kafka.tls.requireClientAuth -}}
{{- $_ := (set $result "cert_file" (printf "/etc/tls/certs/%s-client/tls.crt" (get (fromJson (include "redpanda.Fullname" (dict "a" (list $dot) ))) "r"))) -}}
{{- $_ := (set $result "key_file" (printf "/etc/tls/certs/%s-client/tls.key" (get (fromJson (include "redpanda.Fullname" (dict "a" (list $dot) ))) "r"))) -}}
{{- end -}}
{{- (dict "r" $result) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.adminTLSConfiguration" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $values := $dot.Values.AsMap -}}
{{- $result := (dict ) -}}
{{- if (not (get (fromJson (include "redpanda.InternalTLS.IsEnabled" (dict "a" (list $values.listeners.admin.tls $values.tls) ))) "r")) -}}
{{- (dict "r" $result) | toJson -}}
{{- break -}}
{{- end -}}
{{- $certName := $values.listeners.admin.tls.cert -}}
{{- $tmp_tuple_4 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "_shims.dicttest" (dict "a" (list $values.tls.certs $certName (coalesce nil)) ))) "r")) ))) "r") -}}
{{- $ok_8 := $tmp_tuple_4.T2 -}}
{{- $cert_7 := $tmp_tuple_4.T1 -}}
{{- if (and $ok_8 $cert_7.caEnabled) -}}
{{- $_ := (set $result "truststore_file" (printf "/etc/tls/certs/%s/ca.crt" $values.listeners.admin.tls.cert)) -}}
{{- end -}}
{{- if $values.listeners.admin.tls.requireClientAuth -}}
{{- $_ := (set $result "cert_file" (printf "/etc/tls/certs/%s-client/tls.crt" (get (fromJson (include "redpanda.Fullname" (dict "a" (list $dot) ))) "r"))) -}}
{{- $_ := (set $result "key_file" (printf "/etc/tls/certs/%s-client/tls.key" (get (fromJson (include "redpanda.Fullname" (dict "a" (list $dot) ))) "r"))) -}}
{{- end -}}
{{- (dict "r" $result) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.kafkaClient" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $values := $dot.Values.AsMap -}}
{{- $brokerList := (list ) -}}
{{- range $_, $i := untilStep ((0 | int)|int) (($values.statefulset.replicas | int)|int) (1|int) -}}
{{- $brokerList = (mustAppend $brokerList (dict "address" (printf "%s-%d.%s" (get (fromJson (include "redpanda.Fullname" (dict "a" (list $dot) ))) "r") $i (get (fromJson (include "redpanda.InternalDomain" (dict "a" (list $dot) ))) "r")) "port" ($values.listeners.kafka.port | int) )) -}}
{{- end -}}
{{- $kafkaTLS := $values.listeners.kafka.tls -}}
{{- $brokerTLS := (coalesce nil) -}}
{{- if (get (fromJson (include "redpanda.InternalTLS.IsEnabled" (dict "a" (list $values.listeners.kafka.tls $values.tls) ))) "r") -}}
{{- $brokerTLS = (dict "enabled" true "cert_file" (printf "/etc/tls/certs/%s/tls.crt" $kafkaTLS.cert) "key_file" (printf "/etc/tls/certs/%s/tls.key" $kafkaTLS.cert) "require_client_auth" $kafkaTLS.requireClientAuth "truststore_file" (get (fromJson (include "redpanda.getCertificate" (dict "a" (list $values.tls.certs $kafkaTLS.cert) ))) "r") ) -}}
{{- end -}}
{{- $cfg := (dict "brokers" $brokerList ) -}}
{{- if (gt ((get (fromJson (include "_shims.len" (dict "a" (list $brokerTLS) ))) "r") | int) (0 | int)) -}}
{{- $_ := (set $cfg "broker_tls" $brokerTLS) -}}
{{- end -}}
{{- (dict "r" $cfg) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.configureListeners" -}}
{{- $redpanda := (index .a 0) -}}
{{- $dot := (index .a 1) -}}
{{- range $_ := (list 1) -}}
{{- $_ := (set $redpanda "admin" (get (fromJson (include "redpanda.adminListeners" (dict "a" (list $dot) ))) "r")) -}}
{{- $_ := (set $redpanda "admin_api_tls" (coalesce nil)) -}}
{{- $tls := (get (fromJson (include "redpanda.adminListenersTLS" (dict "a" (list $dot) ))) "r") -}}
{{- if (gt ((get (fromJson (include "_shims.len" (dict "a" (list $tls) ))) "r") | int) (0 | int)) -}}
{{- $_ := (set $redpanda "admin_api_tls" $tls) -}}
{{- end -}}
{{- $_ := (set $redpanda "kafka_api" (get (fromJson (include "redpanda.kafkaListeners" (dict "a" (list $dot) ))) "r")) -}}
{{- $_ := (set $redpanda "kafka_api_tls" (coalesce nil)) -}}
{{- $tls = (get (fromJson (include "redpanda.kafkaListenersTLS" (dict "a" (list $dot) ))) "r") -}}
{{- if (gt ((get (fromJson (include "_shims.len" (dict "a" (list $tls) ))) "r") | int) (0 | int)) -}}
{{- $_ := (set $redpanda "kafka_api_tls" $tls) -}}
{{- end -}}
{{- $_ := (set $redpanda "rpc_server" (get (fromJson (include "redpanda.rpcListeners" (dict "a" (list $dot) ))) "r")) -}}
{{- $rpcTLS := (get (fromJson (include "redpanda.rpcListenersTLS" (dict "a" (list $dot) ))) "r") -}}
{{- if (gt ((get (fromJson (include "_shims.len" (dict "a" (list $rpcTLS) ))) "r") | int) (0 | int)) -}}
{{- $_ := (set $redpanda "rpc_server_tls" $rpcTLS) -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.pandaProxyListener" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $values := $dot.Values.AsMap -}}
{{- $pandaProxy := (dict ) -}}
{{- $_ := (set $pandaProxy "pandaproxy_api" (get (fromJson (include "redpanda.HTTPListeners.Listeners" (dict "a" (list $values.listeners.http (get (fromJson (include "redpanda.Auth.IsSASLEnabled" (dict "a" (list $values.auth) ))) "r")) ))) "r")) -}}
{{- $tls := (get (fromJson (include "redpanda.pandaProxyListenersTLS" (dict "a" (list $dot) ))) "r") -}}
{{- $_ := (set $pandaProxy "pandaproxy_api_tls" (coalesce nil)) -}}
{{- if (gt ((get (fromJson (include "_shims.len" (dict "a" (list $tls) ))) "r") | int) (0 | int)) -}}
{{- $_ := (set $pandaProxy "pandaproxy_api_tls" $tls) -}}
{{- end -}}
{{- (dict "r" $pandaProxy) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.pandaProxyListenersTLS" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $values := $dot.Values.AsMap -}}
{{- $pp := (list ) -}}
{{- $internal := (get (fromJson (include "redpanda.createInternalListenerTLSCfg" (dict "a" (list $values.tls $values.listeners.http.tls) ))) "r") -}}
{{- if (gt ((get (fromJson (include "_shims.len" (dict "a" (list $internal) ))) "r") | int) (0 | int)) -}}
{{- $pp = (mustAppend $pp $internal) -}}
{{- end -}}
{{- range $k, $l := $values.listeners.http.external -}}
{{- if (or (not (get (fromJson (include "redpanda.HTTPExternal.IsEnabled" (dict "a" (list $l) ))) "r")) (not (get (fromJson (include "redpanda.ExternalTLS.IsEnabled" (dict "a" (list $l.tls $values.listeners.http.tls $values.tls) ))) "r"))) -}}
{{- continue -}}
{{- end -}}
{{- $certName := (get (fromJson (include "redpanda.ExternalTLS.GetCertName" (dict "a" (list $l.tls $values.listeners.http.tls) ))) "r") -}}
{{- $pp = (mustAppend $pp (dict "name" $k "enabled" true "cert_file" (printf "/etc/tls/certs/%s/tls.crt" $certName) "key_file" (printf "/etc/tls/certs/%s/tls.key" $certName) "require_client_auth" (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $l.tls.requireClientAuth false) ))) "r") "truststore_file" (get (fromJson (include "redpanda.getCertificate" (dict "a" (list $values.tls.certs $certName) ))) "r") )) -}}
{{- end -}}
{{- (dict "r" $pp) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.schemaRegistry" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $values := $dot.Values.AsMap -}}
{{- $schemaReg := (dict ) -}}
{{- $_ := (set $schemaReg "schema_registry_api" (get (fromJson (include "redpanda.SchemaRegistryListeners.Listeners" (dict "a" (list $values.listeners.schemaRegistry (get (fromJson (include "redpanda.Auth.IsSASLEnabled" (dict "a" (list $values.auth) ))) "r")) ))) "r")) -}}
{{- $tls := (get (fromJson (include "redpanda.schemaRegistryListenersTLS" (dict "a" (list $dot) ))) "r") -}}
{{- $_ := (set $schemaReg "schema_registry_api_tls" (coalesce nil)) -}}
{{- if (gt ((get (fromJson (include "_shims.len" (dict "a" (list $tls) ))) "r") | int) (0 | int)) -}}
{{- $_ := (set $schemaReg "schema_registry_api_tls" $tls) -}}
{{- end -}}
{{- (dict "r" $schemaReg) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.schemaRegistryListenersTLS" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $values := $dot.Values.AsMap -}}
{{- $sr := (list ) -}}
{{- $internal := (get (fromJson (include "redpanda.createInternalListenerTLSCfg" (dict "a" (list $values.tls $values.listeners.schemaRegistry.tls) ))) "r") -}}
{{- if (gt ((get (fromJson (include "_shims.len" (dict "a" (list $internal) ))) "r") | int) (0 | int)) -}}
{{- $sr = (mustAppend $sr $internal) -}}
{{- end -}}
{{- range $k, $l := $values.listeners.schemaRegistry.external -}}
{{- if (or (not (get (fromJson (include "redpanda.SchemaRegistryExternal.IsEnabled" (dict "a" (list $l) ))) "r")) (not (get (fromJson (include "redpanda.ExternalTLS.IsEnabled" (dict "a" (list $l.tls $values.listeners.schemaRegistry.tls $values.tls) ))) "r"))) -}}
{{- continue -}}
{{- end -}}
{{- $certName := (get (fromJson (include "redpanda.ExternalTLS.GetCertName" (dict "a" (list $l.tls $values.listeners.schemaRegistry.tls) ))) "r") -}}
{{- $sr = (mustAppend $sr (dict "name" $k "enabled" true "cert_file" (printf "/etc/tls/certs/%s/tls.crt" $certName) "key_file" (printf "/etc/tls/certs/%s/tls.key" $certName) "require_client_auth" (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $l.tls.requireClientAuth false) ))) "r") "truststore_file" (get (fromJson (include "redpanda.getCertificate" (dict "a" (list $values.tls.certs $certName) ))) "r") )) -}}
{{- end -}}
{{- (dict "r" $sr) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.rpcListenersTLS" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $values := $dot.Values.AsMap -}}
{{- $r := $values.listeners.rpc -}}
{{- if (not (get (fromJson (include "redpanda.InternalTLS.IsEnabled" (dict "a" (list $r.tls $values.tls) ))) "r")) -}}
{{- (dict "r" (dict )) | toJson -}}
{{- break -}}
{{- end -}}
{{- $certName := $r.tls.cert -}}
{{- (dict "r" (dict "enabled" true "cert_file" (printf "/etc/tls/certs/%s/tls.crt" $certName) "key_file" (printf "/etc/tls/certs/%s/tls.key" $certName) "require_client_auth" $r.tls.requireClientAuth "truststore_file" (get (fromJson (include "redpanda.getCertificate" (dict "a" (list $values.tls.certs $certName) ))) "r") )) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.rpcListeners" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $values := $dot.Values.AsMap -}}
{{- (dict "r" (dict "address" "0.0.0.0" "port" ($values.listeners.rpc.port | int) )) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.kafkaListenersTLS" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $values := $dot.Values.AsMap -}}
{{- $kafka := (list ) -}}
{{- $internal := (get (fromJson (include "redpanda.createInternalListenerTLSCfg" (dict "a" (list $values.tls $values.listeners.kafka.tls) ))) "r") -}}
{{- if (gt ((get (fromJson (include "_shims.len" (dict "a" (list $internal) ))) "r") | int) (0 | int)) -}}
{{- $kafka = (mustAppend $kafka $internal) -}}
{{- end -}}
{{- range $k, $l := $values.listeners.kafka.external -}}
{{- if (or (not (get (fromJson (include "redpanda.KafkaExternal.IsEnabled" (dict "a" (list $l) ))) "r")) (not (get (fromJson (include "redpanda.ExternalTLS.IsEnabled" (dict "a" (list $l.tls $values.listeners.kafka.tls $values.tls) ))) "r"))) -}}
{{- continue -}}
{{- end -}}
{{- $certName := (get (fromJson (include "redpanda.ExternalTLS.GetCertName" (dict "a" (list $l.tls $values.listeners.kafka.tls) ))) "r") -}}
{{- $kafka = (mustAppend $kafka (dict "name" $k "enabled" true "cert_file" (printf "/etc/tls/certs/%s/tls.crt" $certName) "key_file" (printf "/etc/tls/certs/%s/tls.key" $certName) "require_client_auth" (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $l.tls.requireClientAuth false) ))) "r") "truststore_file" (get (fromJson (include "redpanda.getCertificate" (dict "a" (list $values.tls.certs $certName) ))) "r") )) -}}
{{- end -}}
{{- (dict "r" $kafka) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.getCertificate" -}}
{{- $certs := (index .a 0) -}}
{{- $certName := (index .a 1) -}}
{{- range $_ := (list 1) -}}
{{- $defaultTruststorePath := "/etc/ssl/certs/ca-certificates.crt" -}}
{{- if (eq $certs (coalesce nil)) -}}
{{- $_ := (fail "TLS map is not defined") -}}
{{- end -}}
{{- if (eq $certName "") -}}
{{- (dict "r" $defaultTruststorePath) | toJson -}}
{{- break -}}
{{- end -}}
{{- $c := $certs -}}
{{- $tmp_tuple_5 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "_shims.dicttest" (dict "a" (list $c $certName (coalesce nil)) ))) "r")) ))) "r") -}}
{{- $ok_10 := $tmp_tuple_5.T2 -}}
{{- $crt_9 := $tmp_tuple_5.T1 -}}
{{- if (and $ok_10 $crt_9.caEnabled) -}}
{{- (dict "r" (printf "/etc/tls/certs/%s/ca.crt" $certName)) | toJson -}}
{{- break -}}
{{- else -}}{{- if (not $ok_10) -}}
{{- $_ := (fail (printf "Certificate name reference (%s) defined in listener, but not found in the tls.certs map" $certName)) -}}
{{- end -}}
{{- end -}}
{{- (dict "r" $defaultTruststorePath) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.kafkaListeners" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $values := $dot.Values.AsMap -}}
{{- $kf := $values.listeners.kafka -}}
{{- $internal := (get (fromJson (include "redpanda.createInternalListenerCfg" (dict "a" (list ($values.listeners.kafka.port | int)) ))) "r") -}}
{{- if (get (fromJson (include "redpanda.Auth.IsSASLEnabled" (dict "a" (list $values.auth) ))) "r") -}}
{{- $_ := (set $internal "authentication_method" "sasl") -}}
{{- end -}}
{{- $am_11 := (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $kf.authenticationMethod "") ))) "r") -}}
{{- if (ne $am_11 "") -}}
{{- $_ := (set $internal "authentication_method" $am_11) -}}
{{- end -}}
{{- $kafka := (list $internal) -}}
{{- range $k, $l := $kf.external -}}
{{- if (not (get (fromJson (include "redpanda.KafkaExternal.IsEnabled" (dict "a" (list $l) ))) "r")) -}}
{{- continue -}}
{{- end -}}
{{- $listener := (dict "name" $k "port" ($l.port | int) "address" "0.0.0.0" ) -}}
{{- if (get (fromJson (include "redpanda.Auth.IsSASLEnabled" (dict "a" (list $values.auth) ))) "r") -}}
{{- $_ := (set $listener "authentication_method" "sasl") -}}
{{- end -}}
{{- $am_12 := (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $l.authenticationMethod "") ))) "r") -}}
{{- if (ne $am_12 "") -}}
{{- $_ := (set $listener "authentication_method" $am_12) -}}
{{- end -}}
{{- $kafka = (mustAppend $kafka $listener) -}}
{{- end -}}
{{- (dict "r" $kafka) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.adminListenersTLS" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $values := $dot.Values.AsMap -}}
{{- $admin := (list ) -}}
{{- $internal := (get (fromJson (include "redpanda.createInternalListenerTLSCfg" (dict "a" (list $values.tls $values.listeners.admin.tls) ))) "r") -}}
{{- if (gt ((get (fromJson (include "_shims.len" (dict "a" (list $internal) ))) "r") | int) (0 | int)) -}}
{{- $admin = (mustAppend $admin $internal) -}}
{{- end -}}
{{- range $k, $l := $values.listeners.admin.external -}}
{{- if (or (not (get (fromJson (include "redpanda.AdminExternal.IsEnabled" (dict "a" (list $l) ))) "r")) (not (get (fromJson (include "redpanda.ExternalTLS.IsEnabled" (dict "a" (list $l.tls $values.listeners.admin.tls $values.tls) ))) "r"))) -}}
{{- continue -}}
{{- end -}}
{{- $certName := (get (fromJson (include "redpanda.ExternalTLS.GetCertName" (dict "a" (list $l.tls $values.listeners.admin.tls) ))) "r") -}}
{{- $admin = (mustAppend $admin (dict "name" $k "enabled" true "cert_file" (printf "/etc/tls/certs/%s/tls.crt" $certName) "key_file" (printf "/etc/tls/certs/%s/tls.key" $certName) "require_client_auth" (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $l.tls.requireClientAuth false) ))) "r") "truststore_file" (get (fromJson (include "redpanda.getCertificate" (dict "a" (list $values.tls.certs $certName) ))) "r") )) -}}
{{- end -}}
{{- (dict "r" $admin) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.createInternalListenerTLSCfg" -}}
{{- $tls := (index .a 0) -}}
{{- $internal := (index .a 1) -}}
{{- range $_ := (list 1) -}}
{{- if (not (get (fromJson (include "redpanda.InternalTLS.IsEnabled" (dict "a" (list $internal $tls) ))) "r")) -}}
{{- (dict "r" (dict )) | toJson -}}
{{- break -}}
{{- end -}}
{{- (dict "r" (dict "name" "internal" "enabled" true "cert_file" (printf "/etc/tls/certs/%s/tls.crt" $internal.cert) "key_file" (printf "/etc/tls/certs/%s/tls.key" $internal.cert) "require_client_auth" $internal.requireClientAuth "truststore_file" (get (fromJson (include "redpanda.getCertificate" (dict "a" (list $tls.certs $internal.cert) ))) "r") )) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.createInternalListenerCfg" -}}
{{- $port := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- (dict "r" (dict "name" "internal" "address" "0.0.0.0" "port" $port )) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.adminListeners" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $values := $dot.Values.AsMap -}}
{{- $admin := (list (get (fromJson (include "redpanda.createInternalListenerCfg" (dict "a" (list ($values.listeners.admin.port | int)) ))) "r")) -}}
{{- range $k, $l := $values.listeners.admin.external -}}
{{- if (not (get (fromJson (include "redpanda.AdminExternal.IsEnabled" (dict "a" (list $l) ))) "r")) -}}
{{- continue -}}
{{- end -}}
{{- $admin = (mustAppend $admin (dict "name" $k "port" ($l.port | int) "address" "0.0.0.0" )) -}}
{{- end -}}
{{- (dict "r" $admin) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.RedpandaAdditionalStartFlags" -}}
{{- $dot := (index .a 0) -}}
{{- $smp := (index .a 1) -}}
{{- range $_ := (list 1) -}}
{{- $values := $dot.Values.AsMap -}}
{{- $chartFlags := (dict "smp" (printf "%d" ($smp | int)) "memory" (printf "%dM" (((get (fromJson (include "redpanda.RedpandaMemory" (dict "a" (list $dot) ))) "r") | int) | int)) "reserve-memory" (printf "%dM" (((get (fromJson (include "redpanda.RedpandaReserveMemory" (dict "a" (list $dot) ))) "r") | int) | int)) "default-log-level" $values.logging.logLevel ) -}}
{{- if (eq (index $values.config.node "developer_mode") true) -}}
{{- $_ := (unset $chartFlags "reserve-memory") -}}
{{- end -}}
{{- range $flag, $_ := $chartFlags -}}
{{- range $_, $userFlag := $values.statefulset.additionalRedpandaCmdFlags -}}
{{- if (regexMatch (printf "^--%s" $flag) $userFlag) -}}
{{- $_ := (unset $chartFlags $flag) -}}
{{- end -}}
{{- end -}}
{{- end -}}
{{- $keys := (keys $chartFlags) -}}
{{- $_ := (sortAlpha $keys) -}}
{{- $flags := (list ) -}}
{{- range $_, $key := $keys -}}
{{- $flags = (mustAppend $flags (printf "--%s=%s" $key (index $chartFlags $key))) -}}
{{- end -}}
{{- (dict "r" (concat $flags $values.statefulset.additionalRedpandaCmdFlags)) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

