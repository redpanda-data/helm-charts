{{- /* Generated from "values.go" */ -}}

{{- define "redpanda.AuditLogging.Translate" -}}
{{- $a := (index .a 0) -}}
{{- $dot := (index .a 1) -}}
{{- $isSASLEnabled := (index .a 2) -}}
{{- range $_ := (list 1) -}}
{{- $result := (dict ) -}}
{{- if (not (get (fromJson (include "redpanda.RedpandaAtLeast_23_3_0" (dict "a" (list $dot) ))) "r")) -}}
{{- (dict "r" $result) | toJson -}}
{{- break -}}
{{- end -}}
{{- $enabled := (and $a.enabled $isSASLEnabled) -}}
{{- $_ := (set $result "audit_enabled" $enabled) -}}
{{- if (not $enabled) -}}
{{- (dict "r" $result) | toJson -}}
{{- break -}}
{{- end -}}
{{- if (ne (($a.clientMaxBufferSize | int) | int) (16777216 | int)) -}}
{{- $_ := (set $result "audit_client_max_buffer_size" ($a.clientMaxBufferSize | int)) -}}
{{- end -}}
{{- if (ne (($a.queueDrainIntervalMs | int) | int) (500 | int)) -}}
{{- $_ := (set $result "audit_queue_drain_interval_ms" ($a.queueDrainIntervalMs | int)) -}}
{{- end -}}
{{- if (ne (($a.queueMaxBufferSizePerShard | int) | int) (1048576 | int)) -}}
{{- $_ := (set $result "audit_queue_max_buffer_size_per_shard" ($a.queueMaxBufferSizePerShard | int)) -}}
{{- end -}}
{{- if (ne (($a.partitions | int) | int) (12 | int)) -}}
{{- $_ := (set $result "audit_log_num_partitions" ($a.partitions | int)) -}}
{{- end -}}
{{- if (ne ($a.replicationFactor | int) (0 | int)) -}}
{{- $_ := (set $result "audit_log_replication_factor" ($a.replicationFactor | int)) -}}
{{- end -}}
{{- if (gt ((get (fromJson (include "_shims.len" (dict "a" (list $a.enabledEventTypes) ))) "r") | int) (0 | int)) -}}
{{- $_ := (set $result "audit_enabled_event_types" $a.enabledEventTypes) -}}
{{- end -}}
{{- if (gt ((get (fromJson (include "_shims.len" (dict "a" (list $a.excludedTopics) ))) "r") | int) (0 | int)) -}}
{{- $_ := (set $result "audit_excluded_topics" $a.excludedTopics) -}}
{{- end -}}
{{- if (gt ((get (fromJson (include "_shims.len" (dict "a" (list $a.excludedPrincipals) ))) "r") | int) (0 | int)) -}}
{{- $_ := (set $result "audit_excluded_principals" $a.excludedPrincipals) -}}
{{- end -}}
{{- (dict "r" $result) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.Auth.IsSASLEnabled" -}}
{{- $a := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- if (eq $a.sasl (coalesce nil)) -}}
{{- (dict "r" false) | toJson -}}
{{- break -}}
{{- end -}}
{{- (dict "r" $a.sasl.enabled) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.Auth.Translate" -}}
{{- $a := (index .a 0) -}}
{{- $isSASLEnabled := (index .a 1) -}}
{{- range $_ := (list 1) -}}
{{- $result := (dict ) -}}
{{- if (or (eq $a.sasl (coalesce nil)) (not $isSASLEnabled)) -}}
{{- (dict "r" $result) | toJson -}}
{{- break -}}
{{- end -}}
{{- if (eq ((get (fromJson (include "_shims.len" (dict "a" (list $a.sasl.users) ))) "r") | int) (0 | int)) -}}
{{- (dict "r" $result) | toJson -}}
{{- break -}}
{{- end -}}
{{- $users := (list ) -}}
{{- range $_, $u := $a.sasl.users -}}
{{- $users = (mustAppend $users $u.name) -}}
{{- end -}}
{{- $_ := (set $result "superusers" $users) -}}
{{- (dict "r" $result) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.Logging.Translate" -}}
{{- $l := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $result := (dict ) -}}
{{- $clusterID_1 := (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $l.usageStats.clusterId "") ))) "r") -}}
{{- if (ne $clusterID_1 "") -}}
{{- $_ := (set $result "cluster_id" $clusterID_1) -}}
{{- end -}}
{{- (dict "r" $result) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.Storage.Translate" -}}
{{- $s := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $result := (dict ) -}}
{{- if (not (get (fromJson (include "redpanda.Storage.IsTieredStorageEnabled" (dict "a" (list $s) ))) "r")) -}}
{{- (dict "r" $result) | toJson -}}
{{- break -}}
{{- end -}}
{{- $tieredStorageConfig := (get (fromJson (include "redpanda.Storage.GetTieredStorageConfig" (dict "a" (list $s) ))) "r") -}}
{{- range $k, $v := $tieredStorageConfig -}}
{{- if (or (eq $v (coalesce nil)) (empty $v)) -}}
{{- continue -}}
{{- end -}}
{{- $tmp_tuple_3 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "_shims.typetest" (dict "a" (list "string" $v "") ))) "r")) ))) "r") -}}
{{- $isStr_3 := $tmp_tuple_3.T2 -}}
{{- $asStr_2 := $tmp_tuple_3.T1 -}}
{{- if (and (and (eq $k "cloud_storage_cache_size") $isStr_3) (ne $asStr_2 "")) -}}
{{- $_ := (set $result $k (toJson ((get (fromJson (include "redpanda.SIToBytes" (dict "a" (list (get (fromJson (include "_shims.typeassertion" (dict "a" (list "string" $v) ))) "r")) ))) "r") | int))) -}}
{{- continue -}}
{{- end -}}
{{- if (eq $k "cloud_storage_cache_size") -}}
{{- $tmp_tuple_4 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "_shims.typetest" (dict "a" (list "string" $v "") ))) "r")) ))) "r") -}}
{{- $isStr_5 := $tmp_tuple_4.T2 -}}
{{- $str_4 := $tmp_tuple_4.T1 -}}
{{- $tmp_tuple_5 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "_shims.asnumeric" (dict "a" (list $v) ))) "r")) ))) "r") -}}
{{- $isFloat_7 := $tmp_tuple_5.T2 -}}
{{- $f_6 := ($tmp_tuple_5.T1 | float64) -}}
{{- if (and $isStr_5 (ne $str_4 "")) -}}
{{- $_ := (set $result $k (toJson ((get (fromJson (include "redpanda.SIToBytes" (dict "a" (list $str_4) ))) "r") | int))) -}}
{{- else -}}{{- if $isFloat_7 -}}
{{- $_ := (set $result $k (toJson ((get (fromJson (include "redpanda.SIToBytes" (dict "a" (list (toString ($f_6 | int))) ))) "r") | int))) -}}
{{- end -}}
{{- end -}}
{{- continue -}}
{{- end -}}
{{- $tmp_tuple_6 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "_shims.typetest" (dict "a" (list "string" $v "") ))) "r")) ))) "r") -}}
{{- $ok_9 := $tmp_tuple_6.T2 -}}
{{- $str_8 := $tmp_tuple_6.T1 -}}
{{- $tmp_tuple_7 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "_shims.typetest" (dict "a" (list "bool" $v false) ))) "r")) ))) "r") -}}
{{- $ok_11 := $tmp_tuple_7.T2 -}}
{{- $b_10 := $tmp_tuple_7.T1 -}}
{{- $tmp_tuple_8 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "_shims.asnumeric" (dict "a" (list $v) ))) "r")) ))) "r") -}}
{{- $isFloat_13 := $tmp_tuple_8.T2 -}}
{{- $f_12 := ($tmp_tuple_8.T1 | float64) -}}
{{- if $ok_9 -}}
{{- $_ := (set $result $k $str_8) -}}
{{- else -}}{{- if $ok_11 -}}
{{- $_ := (set $result $k $b_10) -}}
{{- else -}}{{- if $isFloat_13 -}}
{{- $_ := (set $result $k ($f_12 | int)) -}}
{{- else -}}
{{- $_ := (set $result $k (mustToJson $v)) -}}
{{- end -}}
{{- end -}}
{{- end -}}
{{- end -}}
{{- (dict "r" $result) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.Storage.IsTieredStorageEnabled" -}}
{{- $s := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $conf := (get (fromJson (include "redpanda.Storage.GetTieredStorageConfig" (dict "a" (list $s) ))) "r") -}}
{{- $tmp_tuple_9 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "_shims.dicttest" (dict "a" (list $conf "cloud_storage_enabled" (coalesce nil)) ))) "r")) ))) "r") -}}
{{- $ok := $tmp_tuple_9.T2 -}}
{{- $b := $tmp_tuple_9.T1 -}}
{{- (dict "r" (and $ok (get (fromJson (include "_shims.typeassertion" (dict "a" (list "bool" $b) ))) "r"))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.Storage.GetTieredStorageConfig" -}}
{{- $s := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $result := $s.tiered.config -}}
{{- if (gt ((get (fromJson (include "_shims.len" (dict "a" (list $s.tieredConfig) ))) "r") | int) (0 | int)) -}}
{{- $result = $s.tieredConfig -}}
{{- end -}}
{{- (dict "r" $result) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.Storage.StorageMinFreeBytes" -}}
{{- $s := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- if (and (ne $s.persistentVolume (coalesce nil)) (not $s.persistentVolume.enabled)) -}}
{{- (dict "r" (5368709120 | int)) | toJson -}}
{{- break -}}
{{- end -}}
{{- $minimumFreeBytes := ((mulf (((get (fromJson (include "redpanda.SIToBytes" (dict "a" (list (toString $s.persistentVolume.size)) ))) "r") | int) | float64) 0.05) | float64) -}}
{{- (dict "r" (min (5368709120 | int) ($minimumFreeBytes | int))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.Tuning.Translate" -}}
{{- $t := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $result := (dict ) -}}
{{- $s := (toJson $t) -}}
{{- $tune := (fromJson $s) -}}
{{- $tmp_tuple_11 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "_shims.typetest" (dict "a" (list (printf "map[%s]%s" "string" "interface {}") $tune (coalesce nil)) ))) "r")) ))) "r") -}}
{{- $ok := $tmp_tuple_11.T2 -}}
{{- $m := $tmp_tuple_11.T1 -}}
{{- if (not $ok) -}}
{{- (dict "r" (dict )) | toJson -}}
{{- break -}}
{{- end -}}
{{- range $k, $v := $m -}}
{{- $_ := (set $result $k $v) -}}
{{- end -}}
{{- (dict "r" $result) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.Listeners.CreateSeedServers" -}}
{{- $l := (index .a 0) -}}
{{- $replicas := (index .a 1) -}}
{{- $fullname := (index .a 2) -}}
{{- $internalDomain := (index .a 3) -}}
{{- range $_ := (list 1) -}}
{{- $result := (list ) -}}
{{- range $_, $i := untilStep ((0 | int)|int) ($replicas|int) (1|int) -}}
{{- $result = (mustAppend $result (dict "host" (dict "address" (printf "%s-%d.%s" $fullname $i $internalDomain) "port" ($l.rpc.port | int) ) )) -}}
{{- end -}}
{{- (dict "r" $result) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.Listeners.AdminList" -}}
{{- $l := (index .a 0) -}}
{{- $replicas := (index .a 1) -}}
{{- $fullname := (index .a 2) -}}
{{- $internalDomain := (index .a 3) -}}
{{- range $_ := (list 1) -}}
{{- $result := (list ) -}}
{{- range $_, $i := untilStep ((0 | int)|int) ($replicas|int) (1|int) -}}
{{- $result = (mustAppend $result (printf "%s-%d.%s:%d" $fullname $i $internalDomain (($l.admin.port | int) | int))) -}}
{{- end -}}
{{- (dict "r" $result) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.Config.CreateRPKConfiguration" -}}
{{- $c := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $result := (dict ) -}}
{{- range $k, $v := $c.rpk -}}
{{- $_ := (set $result $k $v) -}}
{{- end -}}
{{- (dict "r" $result) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.TLSCertMap.MustGet" -}}
{{- $m := (index .a 0) -}}
{{- $name := (index .a 1) -}}
{{- range $_ := (list 1) -}}
{{- $tmp_tuple_14 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "_shims.dicttest" (dict "a" (list $m $name (coalesce nil)) ))) "r")) ))) "r") -}}
{{- $ok := $tmp_tuple_14.T2 -}}
{{- $cert := $tmp_tuple_14.T1 -}}
{{- if (not $ok) -}}
{{- $_ := (fail "TODO") -}}
{{- end -}}
{{- (dict "r" $cert) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.InternalTLS.IsEnabled" -}}
{{- $t := (index .a 0) -}}
{{- $tls := (index .a 1) -}}
{{- range $_ := (list 1) -}}
{{- (dict "r" (and (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $t.enabled $tls.enabled) ))) "r") (ne $t.cert ""))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.ExternalTLS.GetCert" -}}
{{- $t := (index .a 0) -}}
{{- $i := (index .a 1) -}}
{{- $tls := (index .a 2) -}}
{{- range $_ := (list 1) -}}
{{- (dict "r" (get (fromJson (include "redpanda.TLSCertMap.MustGet" (dict "a" (list (deepCopy $tls.certs) (get (fromJson (include "redpanda.ExternalTLS.GetCertName" (dict "a" (list $t $i) ))) "r")) ))) "r")) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.ExternalTLS.GetCertName" -}}
{{- $t := (index .a 0) -}}
{{- $i := (index .a 1) -}}
{{- range $_ := (list 1) -}}
{{- (dict "r" (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $t.cert $i.cert) ))) "r")) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.ExternalTLS.IsEnabled" -}}
{{- $t := (index .a 0) -}}
{{- $i := (index .a 1) -}}
{{- $tls := (index .a 2) -}}
{{- range $_ := (list 1) -}}
{{- if (eq $t (coalesce nil)) -}}
{{- (dict "r" false) | toJson -}}
{{- break -}}
{{- end -}}
{{- (dict "r" (and (ne (get (fromJson (include "redpanda.ExternalTLS.GetCertName" (dict "a" (list $t $i) ))) "r") "") (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $t.enabled (get (fromJson (include "redpanda.InternalTLS.IsEnabled" (dict "a" (list $i $tls) ))) "r")) ))) "r"))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.AdminExternal.IsEnabled" -}}
{{- $l := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- (dict "r" (and (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $l.enabled true) ))) "r") (gt ($l.port | int) (0 | int)))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.HTTPListeners.Listeners" -}}
{{- $l := (index .a 0) -}}
{{- $saslEnabled := (index .a 1) -}}
{{- range $_ := (list 1) -}}
{{- $internal := (get (fromJson (include "redpanda.createInternalListenerCfg" (dict "a" (list ($l.port | int)) ))) "r") -}}
{{- if $saslEnabled -}}
{{- $_ := (set $internal "authentication_method" "http_basic") -}}
{{- end -}}
{{- $am_14 := (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $l.authenticationMethod "") ))) "r") -}}
{{- if (ne $am_14 "") -}}
{{- $_ := (set $internal "authentication_method" $am_14) -}}
{{- end -}}
{{- $result := (list $internal) -}}
{{- range $k, $l := $l.external -}}
{{- if (not (get (fromJson (include "redpanda.HTTPExternal.IsEnabled" (dict "a" (list $l) ))) "r")) -}}
{{- continue -}}
{{- end -}}
{{- $listener := (dict "name" $k "port" ($l.port | int) "address" "0.0.0.0" ) -}}
{{- if $saslEnabled -}}
{{- $_ := (set $listener "authentication_method" "http_basic") -}}
{{- end -}}
{{- $am_15 := (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $l.authenticationMethod "") ))) "r") -}}
{{- if (ne $am_15 "") -}}
{{- $_ := (set $listener "authentication_method" $am_15) -}}
{{- end -}}
{{- $result = (mustAppend $result $listener) -}}
{{- end -}}
{{- (dict "r" $result) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.HTTPExternal.IsEnabled" -}}
{{- $l := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- (dict "r" (and (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $l.enabled true) ))) "r") (gt ($l.port | int) (0 | int)))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.KafkaExternal.IsEnabled" -}}
{{- $l := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- (dict "r" (and (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $l.enabled true) ))) "r") (gt ($l.port | int) (0 | int)))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.SchemaRegistryListeners.Listeners" -}}
{{- $sr := (index .a 0) -}}
{{- $saslEnabled := (index .a 1) -}}
{{- range $_ := (list 1) -}}
{{- $internal := (get (fromJson (include "redpanda.createInternalListenerCfg" (dict "a" (list ($sr.port | int)) ))) "r") -}}
{{- if $saslEnabled -}}
{{- $_ := (set $internal "authentication_method" "http_basic") -}}
{{- end -}}
{{- $am_16 := (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $sr.authenticationMethod "") ))) "r") -}}
{{- if (ne $am_16 "") -}}
{{- $_ := (set $internal "authentication_method" $am_16) -}}
{{- end -}}
{{- $result := (list $internal) -}}
{{- range $k, $l := $sr.external -}}
{{- if (not (get (fromJson (include "redpanda.SchemaRegistryExternal.IsEnabled" (dict "a" (list $l) ))) "r")) -}}
{{- continue -}}
{{- end -}}
{{- $listener := (dict "name" $k "port" ($l.port | int) "address" "0.0.0.0" ) -}}
{{- if $saslEnabled -}}
{{- $_ := (set $listener "authentication_method" "http_basic") -}}
{{- end -}}
{{- $am_17 := (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $l.authenticationMethod "") ))) "r") -}}
{{- if (ne $am_17 "") -}}
{{- $_ := (set $listener "authentication_method" $am_17) -}}
{{- end -}}
{{- $result = (mustAppend $result $listener) -}}
{{- end -}}
{{- (dict "r" $result) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.SchemaRegistryExternal.IsEnabled" -}}
{{- $l := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- (dict "r" (and (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $l.enabled true) ))) "r") (gt ($l.port | int) (0 | int)))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.TunableConfig.Translate" -}}
{{- $c := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- if (eq $c (coalesce nil)) -}}
{{- (dict "r" (coalesce nil)) | toJson -}}
{{- break -}}
{{- end -}}
{{- $result := (dict ) -}}
{{- range $k, $v := $c -}}
{{- if (not (empty $v)) -}}
{{- $_ := (set $result $k $v) -}}
{{- end -}}
{{- end -}}
{{- (dict "r" $result) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.NodeConfig.Translate" -}}
{{- $c := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $result := (dict ) -}}
{{- range $k, $v := $c -}}
{{- if (not (empty $v)) -}}
{{- $_ := (set $result $k (toYaml $v)) -}}
{{- end -}}
{{- end -}}
{{- (dict "r" $result) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.ClusterConfig.Translate" -}}
{{- $c := (index .a 0) -}}
{{- $replicas := (index .a 1) -}}
{{- $skipDefaultTopic := (index .a 2) -}}
{{- range $_ := (list 1) -}}
{{- $result := (dict ) -}}
{{- range $k, $v := $c -}}
{{- if (and (eq $k "default_topic_replications") (not $skipDefaultTopic)) -}}
{{- $r := ($replicas | int) -}}
{{- $input := ($r | int) -}}
{{- $tmp_tuple_17 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "_shims.asintegral" (dict "a" (list $v) ))) "r")) ))) "r") -}}
{{- $ok_19 := $tmp_tuple_17.T2 -}}
{{- $num_18 := ($tmp_tuple_17.T1 | int) -}}
{{- if $ok_19 -}}
{{- $input = $num_18 -}}
{{- end -}}
{{- $tmp_tuple_18 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "_shims.asnumeric" (dict "a" (list $v) ))) "r")) ))) "r") -}}
{{- $ok_21 := $tmp_tuple_18.T2 -}}
{{- $f_20 := ($tmp_tuple_18.T1 | float64) -}}
{{- if $ok_21 -}}
{{- $input = ($f_20 | int) -}}
{{- end -}}
{{- $_ := (set $result $k (min $input ((sub ((add $r (((mod $r (2 | int)) | int))) | int) (1 | int)) | int))) -}}
{{- continue -}}
{{- end -}}
{{- $tmp_tuple_19 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "_shims.typetest" (dict "a" (list "bool" $v false) ))) "r")) ))) "r") -}}
{{- $ok_23 := $tmp_tuple_19.T2 -}}
{{- $b_22 := $tmp_tuple_19.T1 -}}
{{- if $ok_23 -}}
{{- $_ := (set $result $k $b_22) -}}
{{- continue -}}
{{- end -}}
{{- if (not (empty $v)) -}}
{{- $_ := (set $result $k $v) -}}
{{- end -}}
{{- end -}}
{{- (dict "r" $result) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.TieredStorageCredentials.IsAccessKeyReferenceValid" -}}
{{- $tsc := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- (dict "r" (and (and (ne $tsc.accessKey (coalesce nil)) (ne $tsc.accessKey.name "")) (ne $tsc.accessKey.key ""))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.TieredStorageCredentials.IsSecretKeyReferenceValid" -}}
{{- $tsc := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- (dict "r" (and (and (ne $tsc.secretKey (coalesce nil)) (ne $tsc.secretKey.name "")) (ne $tsc.secretKey.key ""))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

