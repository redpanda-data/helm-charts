{{- /* Generated from "configmap.tpl.go" */ -}}

{{- define "redpanda.ConfigMaps" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $cms := (list (get (fromJson (include "redpanda.RedpandaConfigMap" (dict "a" (list $dot true) ))) "r")) -}}
{{- $cms = (concat (default (list ) $cms) (default (list ) (get (fromJson (include "redpanda.RPKProfile" (dict "a" (list $dot) ))) "r"))) -}}
{{- (dict "r" $cms) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.ConfigMapsWithoutSeedServer" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $cms := (list (get (fromJson (include "redpanda.RedpandaConfigMap" (dict "a" (list $dot false) ))) "r")) -}}
{{- $cms = (concat (default (list ) $cms) (default (list ) (get (fromJson (include "redpanda.RPKProfile" (dict "a" (list $dot) ))) "r"))) -}}
{{- (dict "r" $cms) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.RedpandaConfigMap" -}}
{{- $dot := (index .a 0) -}}
{{- $includeSeedServer := (index .a 1) -}}
{{- range $_ := (list 1) -}}
{{- (dict "r" (mustMergeOverwrite (dict "metadata" (dict "creationTimestamp" (coalesce nil) ) ) (mustMergeOverwrite (dict ) (dict "kind" "ConfigMap" "apiVersion" "v1" )) (dict "metadata" (mustMergeOverwrite (dict "creationTimestamp" (coalesce nil) ) (dict "name" (get (fromJson (include "redpanda.Fullname" (dict "a" (list $dot) ))) "r") "namespace" $dot.Release.Namespace "labels" (get (fromJson (include "redpanda.FullLabels" (dict "a" (list $dot) ))) "r") )) "data" (dict "bootstrap.yaml" (get (fromJson (include "redpanda.BootstrapFile" (dict "a" (list $dot) ))) "r") "redpanda.yaml" (get (fromJson (include "redpanda.RedpandaConfigFile" (dict "a" (list $dot $includeSeedServer) ))) "r") ) ))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.BootstrapFile" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $values := $dot.Values.AsMap -}}
{{- $bootstrap := (dict "kafka_enable_authorization" (get (fromJson (include "redpanda.Auth.IsSASLEnabled" (dict "a" (list $values.auth) ))) "r") "enable_sasl" (get (fromJson (include "redpanda.Auth.IsSASLEnabled" (dict "a" (list $values.auth) ))) "r") "enable_rack_awareness" $values.rackAwareness.enabled "storage_min_free_bytes" (get (fromJson (include "redpanda.Storage.StorageMinFreeBytes" (dict "a" (list $values.storage) ))) "r") ) -}}
{{- $bootstrap = (merge (dict ) $bootstrap (get (fromJson (include "redpanda.AuditLogging.Translate" (dict "a" (list $values.auditLogging $dot (get (fromJson (include "redpanda.Auth.IsSASLEnabled" (dict "a" (list $values.auth) ))) "r")) ))) "r")) -}}
{{- $bootstrap = (merge (dict ) $bootstrap (get (fromJson (include "redpanda.Logging.Translate" (dict "a" (list $values.logging) ))) "r")) -}}
{{- $bootstrap = (merge (dict ) $bootstrap (get (fromJson (include "redpanda.TunableConfig.Translate" (dict "a" (list $values.config.tunable) ))) "r")) -}}
{{- $bootstrap = (merge (dict ) $bootstrap (get (fromJson (include "redpanda.ClusterConfig.Translate" (dict "a" (list $values.config.cluster ($values.statefulset.replicas | int) false) ))) "r")) -}}
{{- $bootstrap = (merge (dict ) $bootstrap (get (fromJson (include "redpanda.Auth.Translate" (dict "a" (list $values.auth (get (fromJson (include "redpanda.Auth.IsSASLEnabled" (dict "a" (list $values.auth) ))) "r")) ))) "r")) -}}
{{- (dict "r" (toYaml $bootstrap)) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.RedpandaConfigFile" -}}
{{- $dot := (index .a 0) -}}
{{- $includeSeedServer := (index .a 1) -}}
{{- range $_ := (list 1) -}}
{{- $values := $dot.Values.AsMap -}}
{{- $redpanda := (dict "kafka_enable_authorization" (get (fromJson (include "redpanda.Auth.IsSASLEnabled" (dict "a" (list $values.auth) ))) "r") "enable_sasl" (get (fromJson (include "redpanda.Auth.IsSASLEnabled" (dict "a" (list $values.auth) ))) "r") "empty_seed_starts_cluster" false "storage_min_free_bytes" (get (fromJson (include "redpanda.Storage.StorageMinFreeBytes" (dict "a" (list $values.storage) ))) "r") ) -}}
{{- if $includeSeedServer -}}
{{- $_ := (set $redpanda "seed_servers" (get (fromJson (include "redpanda.Listeners.CreateSeedServers" (dict "a" (list $values.listeners ($values.statefulset.replicas | int) (get (fromJson (include "redpanda.Fullname" (dict "a" (list $dot) ))) "r") (get (fromJson (include "redpanda.InternalDomain" (dict "a" (list $dot) ))) "r")) ))) "r")) -}}
{{- end -}}
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
{{- (dict "r" (toYaml $redpandaYaml)) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.RPKProfile" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $values := $dot.Values.AsMap -}}
{{- if (not $values.external.enabled) -}}
{{- (dict "r" (coalesce nil)) | toJson -}}
{{- break -}}
{{- end -}}
{{- (dict "r" (list (mustMergeOverwrite (dict "metadata" (dict "creationTimestamp" (coalesce nil) ) ) (mustMergeOverwrite (dict ) (dict "kind" "ConfigMap" "apiVersion" "v1" )) (dict "metadata" (mustMergeOverwrite (dict "creationTimestamp" (coalesce nil) ) (dict "name" (printf "%s-rpk" (get (fromJson (include "redpanda.Fullname" (dict "a" (list $dot) ))) "r")) "namespace" $dot.Release.Namespace "labels" (get (fromJson (include "redpanda.FullLabels" (dict "a" (list $dot) ))) "r") )) "data" (dict "profile" (toYaml (get (fromJson (include "redpanda.rpkProfile" (dict "a" (list $dot) ))) "r")) ) )))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.rpkProfile" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $values := $dot.Values.AsMap -}}
{{- $brokerList := (list ) -}}
{{- range $_, $i := untilStep ((0 | int)|int) (($values.statefulset.replicas | int)|int) (1|int) -}}
{{- $brokerList = (concat (default (list ) $brokerList) (list (printf "%s:%d" (get (fromJson (include "redpanda.advertisedHost" (dict "a" (list $dot $i) ))) "r") (((get (fromJson (include "redpanda.advertisedKafkaPort" (dict "a" (list $dot $i) ))) "r") | int) | int)))) -}}
{{- end -}}
{{- $adminAdvertisedList := (list ) -}}
{{- range $_, $i := untilStep ((0 | int)|int) (($values.statefulset.replicas | int)|int) (1|int) -}}
{{- $adminAdvertisedList = (concat (default (list ) $adminAdvertisedList) (list (printf "%s:%d" (get (fromJson (include "redpanda.advertisedHost" (dict "a" (list $dot $i) ))) "r") (((get (fromJson (include "redpanda.advertisedAdminPort" (dict "a" (list $dot $i) ))) "r") | int) | int)))) -}}
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
{{- $result := (dict "name" (get (fromJson (include "redpanda.getFirstExternalKafkaListener" (dict "a" (list $dot) ))) "r") "kafka_api" $ka "admin_api" $aa ) -}}
{{- (dict "r" $result) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.advertisedKafkaPort" -}}
{{- $dot := (index .a 0) -}}
{{- $i := (index .a 1) -}}
{{- range $_ := (list 1) -}}
{{- $values := $dot.Values.AsMap -}}
{{- $externalKafkaListenerName := (get (fromJson (include "redpanda.getFirstExternalKafkaListener" (dict "a" (list $dot) ))) "r") -}}
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

{{- define "redpanda.getFirstExternalKafkaListener" -}}
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
{{- $brokerList = (concat (default (list ) $brokerList) (list (printf "%s-%d.%s:%d" (get (fromJson (include "redpanda.Fullname" (dict "a" (list $dot) ))) "r") $i (get (fromJson (include "redpanda.InternalDomain" (dict "a" (list $dot) ))) "r") (($values.listeners.kafka.port | int) | int)))) -}}
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
{{- $result := (dict "overprovisioned" (get (fromJson (include "redpanda.RedpandaResources.GetOverProvisionValue" (dict "a" (list $values.resources) ))) "r") "enable_memory_locking" (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $values.resources.memory.enable_memory_locking false) ))) "r") "additional_start_flags" (get (fromJson (include "redpanda.RedpandaAdditionalStartFlags" (dict "a" (list $dot ((get (fromJson (include "redpanda.RedpandaSMP" (dict "a" (list $dot) ))) "r") | int)) ))) "r") "kafka_api" (dict "brokers" $brokerList "tls" $brokerTLS ) "admin_api" (dict "addresses" (get (fromJson (include "redpanda.Listeners.AdminList" (dict "a" (list $values.listeners ($values.statefulset.replicas | int) (get (fromJson (include "redpanda.Fullname" (dict "a" (list $dot) ))) "r") (get (fromJson (include "redpanda.InternalDomain" (dict "a" (list $dot) ))) "r")) ))) "r") "tls" $adminTLS ) ) -}}
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
{{- $truststore_5 := (get (fromJson (include "redpanda.InternalTLS.TrustStoreFilePath" (dict "a" (list $values.listeners.kafka.tls $values.tls) ))) "r") -}}
{{- if (ne $truststore_5 "/etc/ssl/certs/ca-certificates.crt") -}}
{{- $_ := (set $result "truststore_file" $truststore_5) -}}
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
{{- $truststore_6 := (get (fromJson (include "redpanda.InternalTLS.TrustStoreFilePath" (dict "a" (list $values.listeners.admin.tls $values.tls) ))) "r") -}}
{{- if (ne $truststore_6 "/etc/ssl/certs/ca-certificates.crt") -}}
{{- $_ := (set $result "truststore_file" $truststore_6) -}}
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
{{- $brokerList = (concat (default (list ) $brokerList) (list (dict "address" (printf "%s-%d.%s" (get (fromJson (include "redpanda.Fullname" (dict "a" (list $dot) ))) "r") $i (get (fromJson (include "redpanda.InternalDomain" (dict "a" (list $dot) ))) "r")) "port" ($values.listeners.kafka.port | int) ))) -}}
{{- end -}}
{{- $kafkaTLS := $values.listeners.kafka.tls -}}
{{- $brokerTLS := (coalesce nil) -}}
{{- if (get (fromJson (include "redpanda.InternalTLS.IsEnabled" (dict "a" (list $values.listeners.kafka.tls $values.tls) ))) "r") -}}
{{- $brokerTLS = (dict "enabled" true "cert_file" (printf "/etc/tls/certs/%s/tls.crt" $kafkaTLS.cert) "key_file" (printf "/etc/tls/certs/%s/tls.key" $kafkaTLS.cert) "require_client_auth" $kafkaTLS.requireClientAuth "truststore_file" (get (fromJson (include "redpanda.InternalTLS.TrustStoreFilePath" (dict "a" (list $kafkaTLS $values.tls) ))) "r") ) -}}
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
{{- $values := $dot.Values.AsMap -}}
{{- $_ := (set $redpanda "admin" (get (fromJson (include "redpanda.AdminListeners.Listeners" (dict "a" (list $values.listeners.admin) ))) "r")) -}}
{{- $_ := (set $redpanda "kafka_api" (get (fromJson (include "redpanda.KafkaListeners.Listeners" (dict "a" (list $values.listeners.kafka $values.auth) ))) "r")) -}}
{{- $_ := (set $redpanda "rpc_server" (get (fromJson (include "redpanda.rpcListeners" (dict "a" (list $dot) ))) "r")) -}}
{{- $_ := (set $redpanda "admin_api_tls" (coalesce nil)) -}}
{{- $tls_7 := (get (fromJson (include "redpanda.AdminListeners.ListenersTLS" (dict "a" (list $values.listeners.admin $values.tls) ))) "r") -}}
{{- if (gt ((get (fromJson (include "_shims.len" (dict "a" (list $tls_7) ))) "r") | int) (0 | int)) -}}
{{- $_ := (set $redpanda "admin_api_tls" $tls_7) -}}
{{- end -}}
{{- $_ := (set $redpanda "kafka_api_tls" (coalesce nil)) -}}
{{- $tls_8 := (get (fromJson (include "redpanda.KafkaListeners.ListenersTLS" (dict "a" (list $values.listeners.kafka $values.tls) ))) "r") -}}
{{- if (gt ((get (fromJson (include "_shims.len" (dict "a" (list $tls_8) ))) "r") | int) (0 | int)) -}}
{{- $_ := (set $redpanda "kafka_api_tls" $tls_8) -}}
{{- end -}}
{{- $tls_9 := (get (fromJson (include "redpanda.rpcListenersTLS" (dict "a" (list $dot) ))) "r") -}}
{{- if (gt ((get (fromJson (include "_shims.len" (dict "a" (list $tls_9) ))) "r") | int) (0 | int)) -}}
{{- $_ := (set $redpanda "rpc_server_tls" $tls_9) -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.pandaProxyListener" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $values := $dot.Values.AsMap -}}
{{- $pandaProxy := (dict ) -}}
{{- $_ := (set $pandaProxy "pandaproxy_api" (get (fromJson (include "redpanda.HTTPListeners.Listeners" (dict "a" (list $values.listeners.http (get (fromJson (include "redpanda.Auth.IsSASLEnabled" (dict "a" (list $values.auth) ))) "r")) ))) "r")) -}}
{{- $_ := (set $pandaProxy "pandaproxy_api_tls" (coalesce nil)) -}}
{{- $tls_10 := (get (fromJson (include "redpanda.HTTPListeners.ListenersTLS" (dict "a" (list $values.listeners.http $values.tls) ))) "r") -}}
{{- if (gt ((get (fromJson (include "_shims.len" (dict "a" (list $tls_10) ))) "r") | int) (0 | int)) -}}
{{- $_ := (set $pandaProxy "pandaproxy_api_tls" $tls_10) -}}
{{- end -}}
{{- (dict "r" $pandaProxy) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.schemaRegistry" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $values := $dot.Values.AsMap -}}
{{- $schemaReg := (dict ) -}}
{{- $_ := (set $schemaReg "schema_registry_api" (get (fromJson (include "redpanda.SchemaRegistryListeners.Listeners" (dict "a" (list $values.listeners.schemaRegistry (get (fromJson (include "redpanda.Auth.IsSASLEnabled" (dict "a" (list $values.auth) ))) "r")) ))) "r")) -}}
{{- $_ := (set $schemaReg "schema_registry_api_tls" (coalesce nil)) -}}
{{- $tls_11 := (get (fromJson (include "redpanda.SchemaRegistryListeners.ListenersTLS" (dict "a" (list $values.listeners.schemaRegistry $values.tls) ))) "r") -}}
{{- if (gt ((get (fromJson (include "_shims.len" (dict "a" (list $tls_11) ))) "r") | int) (0 | int)) -}}
{{- $_ := (set $schemaReg "schema_registry_api_tls" $tls_11) -}}
{{- end -}}
{{- (dict "r" $schemaReg) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.rpcListenersTLS" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $values := $dot.Values.AsMap -}}
{{- $r := $values.listeners.rpc -}}
{{- if (and (not ((or (or (get (fromJson (include "redpanda.RedpandaAtLeast_22_2_atleast_22_2_10" (dict "a" (list $dot) ))) "r") (get (fromJson (include "redpanda.RedpandaAtLeast_22_3_atleast_22_3_13" (dict "a" (list $dot) ))) "r")) (get (fromJson (include "redpanda.RedpandaAtLeast_23_1_2" (dict "a" (list $dot) ))) "r")))) ((or (and (eq $r.tls.enabled (coalesce nil)) $values.tls.enabled) (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $r.tls.enabled false) ))) "r")))) -}}
{{- $_ := (fail (printf "Redpanda version v%s does not support TLS on the RPC port. Please upgrade. See technical service bulletin 2023-01." (trimPrefix "v" (get (fromJson (include "redpanda.Tag" (dict "a" (list $dot) ))) "r")))) -}}
{{- end -}}
{{- if (not (get (fromJson (include "redpanda.InternalTLS.IsEnabled" (dict "a" (list $r.tls $values.tls) ))) "r")) -}}
{{- (dict "r" (dict )) | toJson -}}
{{- break -}}
{{- end -}}
{{- $certName := $r.tls.cert -}}
{{- (dict "r" (dict "enabled" true "cert_file" (printf "/etc/tls/certs/%s/tls.crt" $certName) "key_file" (printf "/etc/tls/certs/%s/tls.key" $certName) "require_client_auth" $r.tls.requireClientAuth "truststore_file" (get (fromJson (include "redpanda.InternalTLS.TrustStoreFilePath" (dict "a" (list $r.tls $values.tls) ))) "r") )) | toJson -}}
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

{{- define "redpanda.createInternalListenerTLSCfg" -}}
{{- $tls := (index .a 0) -}}
{{- $internal := (index .a 1) -}}
{{- range $_ := (list 1) -}}
{{- if (not (get (fromJson (include "redpanda.InternalTLS.IsEnabled" (dict "a" (list $internal $tls) ))) "r")) -}}
{{- (dict "r" (dict )) | toJson -}}
{{- break -}}
{{- end -}}
{{- (dict "r" (dict "name" "internal" "enabled" true "cert_file" (printf "/etc/tls/certs/%s/tls.crt" $internal.cert) "key_file" (printf "/etc/tls/certs/%s/tls.key" $internal.cert) "require_client_auth" $internal.requireClientAuth "truststore_file" (get (fromJson (include "redpanda.InternalTLS.TrustStoreFilePath" (dict "a" (list $internal $tls) ))) "r") )) | toJson -}}
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
{{- $flags = (concat (default (list ) $flags) (list (printf "--%s=%s" $key (index $chartFlags $key)))) -}}
{{- end -}}
{{- (dict "r" (concat (default (list ) $flags) (default (list ) $values.statefulset.additionalRedpandaCmdFlags))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

