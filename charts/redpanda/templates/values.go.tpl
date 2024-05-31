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
{{- if (not $isSASLEnabled) -}}
{{- (dict "r" (coalesce nil)) | toJson -}}
{{- break -}}
{{- end -}}
{{- if (eq ((get (fromJson (include "_shims.len" (dict "a" (list $a.sasl.users) ))) "r") | int) (0 | int)) -}}
{{- (dict "r" (coalesce nil)) | toJson -}}
{{- break -}}
{{- end -}}
{{- $users := (list ) -}}
{{- range $_, $u := $a.sasl.users -}}
{{- $users = (concat (default (list ) $users) (list $u.name)) -}}
{{- end -}}
{{- (dict "r" (dict "superusers" $users )) | toJson -}}
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

{{- define "redpanda.RedpandaResources.GetOverProvisionValue" -}}
{{- $rr := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- if (lt ((get (fromJson (include "redpanda.RedpandaResources.RedpandaCoresInMillis" (dict "a" (list $rr) ))) "r") | int) (1000 | int)) -}}
{{- (dict "r" true) | toJson -}}
{{- break -}}
{{- end -}}
{{- (dict "r" (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $rr.cpu.overprovisioned false) ))) "r")) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.RedpandaResources.RedpandaCoresInMillis" -}}
{{- $rr := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $tmp_tuple_3 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "_shims.asnumeric" (dict "a" (list $rr.cpu.cores) ))) "r")) ))) "r") -}}
{{- $ok_3 := $tmp_tuple_3.T2 -}}
{{- $cores_2 := ($tmp_tuple_3.T1 | float64) -}}
{{- if $ok_3 -}}
{{- (dict "r" (((mulf $cores_2 1000.0) | float64) | int)) | toJson -}}
{{- break -}}
{{- end -}}
{{- $tmp_tuple_4 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "_shims.typetest" (dict "a" (list "string" $rr.cpu.cores "") ))) "r")) ))) "r") -}}
{{- $ok_5 := $tmp_tuple_4.T2 -}}
{{- $cores_4 := $tmp_tuple_4.T1 -}}
{{- if $ok_5 -}}
{{- $suffix := (regexReplaceAll "^[0-9.]+(.*)" $cores_4 "${1}") -}}
{{- if (eq $suffix "m") -}}
{{- $c := (trimSuffix $suffix $cores_4) -}}
{{- $tmp_tuple_5 := (get (fromJson (include "_shims.compact" (dict "a" (list (list (atoi $c) nil)) ))) "r") -}}
{{- $err := $tmp_tuple_5.T2 -}}
{{- $coreMilli := ($tmp_tuple_5.T1 | int) -}}
{{- if (ne $err (coalesce nil)) -}}
{{- $_ := (fail (printf "cores is not integer with 'm' suffix: %s" $cores_4)) -}}
{{- end -}}
{{- (dict "r" $coreMilli) | toJson -}}
{{- break -}}
{{- else -}}{{- if (eq $suffix "") -}}
{{- $tmp_tuple_6 := (get (fromJson (include "_shims.compact" (dict "a" (list (list (atoi $cores_4) nil)) ))) "r") -}}
{{- $err := $tmp_tuple_6.T2 -}}
{{- $c := ($tmp_tuple_6.T1 | int) -}}
{{- if (ne $err (coalesce nil)) -}}
{{- $_ := (fail (printf "cores is not integer with 'm' suffix: %s" $cores_4)) -}}
{{- end -}}
{{- (dict "r" ((mul $c (1000 | int)) | int)) | toJson -}}
{{- break -}}
{{- else -}}
{{- $_ := (fail (printf "Unrecognized CPU unit '%s'" $suffix)) -}}
{{- end -}}
{{- end -}}
{{- end -}}
{{- (dict "r" (1 | int)) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.Storage.IsTieredStorageEnabled" -}}
{{- $s := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $conf := (get (fromJson (include "redpanda.Storage.GetTieredStorageConfig" (dict "a" (list $s) ))) "r") -}}
{{- $tmp_tuple_7 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "_shims.dicttest" (dict "a" (list $conf "cloud_storage_enabled" (coalesce nil)) ))) "r")) ))) "r") -}}
{{- $ok := $tmp_tuple_7.T2 -}}
{{- $b := $tmp_tuple_7.T1 -}}
{{- (dict "r" (and $ok (get (fromJson (include "_shims.typeassertion" (dict "a" (list "bool" $b) ))) "r"))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.Storage.GetTieredStorageConfig" -}}
{{- $s := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- if (gt ((get (fromJson (include "_shims.len" (dict "a" (list $s.tieredConfig) ))) "r") | int) (0 | int)) -}}
{{- (dict "r" $s.tieredConfig) | toJson -}}
{{- break -}}
{{- end -}}
{{- (dict "r" $s.tiered.config) | toJson -}}
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
{{- $tmp_tuple_9 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "_shims.typetest" (dict "a" (list "string" $v "") ))) "r")) ))) "r") -}}
{{- $isStr_7 := $tmp_tuple_9.T2 -}}
{{- $asStr_6 := $tmp_tuple_9.T1 -}}
{{- if (and (and (eq $k "cloud_storage_cache_size") $isStr_7) (ne $asStr_6 "")) -}}
{{- $_ := (set $result $k (toJson ((get (fromJson (include "redpanda.SIToBytes" (dict "a" (list (get (fromJson (include "_shims.typeassertion" (dict "a" (list "string" $v) ))) "r")) ))) "r") | int))) -}}
{{- continue -}}
{{- end -}}
{{- if (eq $k "cloud_storage_cache_size") -}}
{{- $tmp_tuple_10 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "_shims.typetest" (dict "a" (list "string" $v "") ))) "r")) ))) "r") -}}
{{- $isStr_9 := $tmp_tuple_10.T2 -}}
{{- $str_8 := $tmp_tuple_10.T1 -}}
{{- $tmp_tuple_11 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "_shims.asnumeric" (dict "a" (list $v) ))) "r")) ))) "r") -}}
{{- $isFloat_11 := $tmp_tuple_11.T2 -}}
{{- $f_10 := ($tmp_tuple_11.T1 | float64) -}}
{{- if (and $isStr_9 (ne $str_8 "")) -}}
{{- $_ := (set $result $k (toJson ((get (fromJson (include "redpanda.SIToBytes" (dict "a" (list $str_8) ))) "r") | int))) -}}
{{- else -}}{{- if $isFloat_11 -}}
{{- $_ := (set $result $k (toJson ((get (fromJson (include "redpanda.SIToBytes" (dict "a" (list (toString ($f_10 | int))) ))) "r") | int))) -}}
{{- end -}}
{{- end -}}
{{- continue -}}
{{- end -}}
{{- $tmp_tuple_12 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "_shims.typetest" (dict "a" (list "string" $v "") ))) "r")) ))) "r") -}}
{{- $ok_13 := $tmp_tuple_12.T2 -}}
{{- $str_12 := $tmp_tuple_12.T1 -}}
{{- $tmp_tuple_13 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "_shims.typetest" (dict "a" (list "bool" $v false) ))) "r")) ))) "r") -}}
{{- $ok_15 := $tmp_tuple_13.T2 -}}
{{- $b_14 := $tmp_tuple_13.T1 -}}
{{- $tmp_tuple_14 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "_shims.asnumeric" (dict "a" (list $v) ))) "r")) ))) "r") -}}
{{- $isFloat_17 := $tmp_tuple_14.T2 -}}
{{- $f_16 := ($tmp_tuple_14.T1 | float64) -}}
{{- if $ok_13 -}}
{{- $_ := (set $result $k $str_12) -}}
{{- else -}}{{- if $ok_15 -}}
{{- $_ := (set $result $k $b_14) -}}
{{- else -}}{{- if $isFloat_17 -}}
{{- $_ := (set $result $k ($f_16 | int)) -}}
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

{{- define "redpanda.Storage.StorageMinFreeBytes" -}}
{{- $s := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- if (and (ne $s.persistentVolume (coalesce nil)) (not $s.persistentVolume.enabled)) -}}
{{- (dict "r" (5368709120 | int)) | toJson -}}
{{- break -}}
{{- end -}}
{{- $minimumFreeBytes := ((mulf (((get (fromJson (include "redpanda.SIToBytes" (dict "a" (list (toString $s.persistentVolume.size)) ))) "r") | int) | float64) 0.05) | float64) -}}
{{- (dict "r" (min (5368709120 | int) ($minimumFreeBytes | int64))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.Tuning.Translate" -}}
{{- $t := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $result := (dict ) -}}
{{- $s := (toJson $t) -}}
{{- $tune := (fromJson $s) -}}
{{- $tmp_tuple_15 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "_shims.typetest" (dict "a" (list (printf "map[%s]%s" "string" "interface {}") $tune (coalesce nil)) ))) "r")) ))) "r") -}}
{{- $ok := $tmp_tuple_15.T2 -}}
{{- $m := $tmp_tuple_15.T1 -}}
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
{{- $result := (coalesce nil) -}}
{{- range $_, $i := untilStep ((0 | int)|int) ($replicas|int) (1|int) -}}
{{- $result = (concat (default (list ) $result) (list (dict "host" (dict "address" (printf "%s-%d.%s" $fullname $i $internalDomain) "port" ($l.rpc.port | int) ) ))) -}}
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
{{- $result := (coalesce nil) -}}
{{- range $_, $i := untilStep ((0 | int)|int) ($replicas|int) (1|int) -}}
{{- $result = (concat (default (list ) $result) (list (printf "%s-%d.%s:%d" $fullname $i $internalDomain (($l.admin.port | int) | int)))) -}}
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

{{- define "redpanda.TLSCertMap.getTrustStoreFilePath" -}}
{{- $m := (index .a 0) -}}
{{- $certName := (index .a 1) -}}
{{- range $_ := (list 1) -}}
{{- if (eq $m (coalesce nil)) -}}
{{- $_ := (fail "TLS map is not defined") -}}
{{- end -}}
{{- if (eq $certName "") -}}
{{- (dict "r" "/etc/ssl/certs/ca-certificates.crt") | toJson -}}
{{- break -}}
{{- end -}}
{{- $tmp_tuple_18 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "_shims.dicttest" (dict "a" (list $m $certName (coalesce nil)) ))) "r")) ))) "r") -}}
{{- $ok_19 := $tmp_tuple_18.T2 -}}
{{- $crt_18 := $tmp_tuple_18.T1 -}}
{{- if (and $ok_19 $crt_18.caEnabled) -}}
{{- (dict "r" (printf "/etc/tls/certs/%s/ca.crt" $certName)) | toJson -}}
{{- break -}}
{{- else -}}{{- if (not $ok_19) -}}
{{- $_ := (fail (printf "Certificate name reference (%s) defined in listener, but not found in the tls.certs map" $certName)) -}}
{{- end -}}
{{- end -}}
{{- (dict "r" "/etc/ssl/certs/ca-certificates.crt") | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.TLSCertMap.MustGet" -}}
{{- $m := (index .a 0) -}}
{{- $name := (index .a 1) -}}
{{- range $_ := (list 1) -}}
{{- $tmp_tuple_19 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "_shims.dicttest" (dict "a" (list $m $name (coalesce nil)) ))) "r")) ))) "r") -}}
{{- $ok := $tmp_tuple_19.T2 -}}
{{- $cert := $tmp_tuple_19.T1 -}}
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

{{- define "redpanda.AdminListeners.Listeners" -}}
{{- $l := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $admin := (list (get (fromJson (include "redpanda.createInternalListenerCfg" (dict "a" (list ($l.port | int)) ))) "r")) -}}
{{- range $k, $lis := $l.external -}}
{{- if (not (get (fromJson (include "redpanda.AdminExternal.IsEnabled" (dict "a" (list $lis) ))) "r")) -}}
{{- continue -}}
{{- end -}}
{{- $admin = (concat (default (list ) $admin) (list (dict "name" $k "port" ($lis.port | int) "address" "0.0.0.0" ))) -}}
{{- end -}}
{{- (dict "r" $admin) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.AdminListeners.ListenersTLS" -}}
{{- $l := (index .a 0) -}}
{{- $tls := (index .a 1) -}}
{{- range $_ := (list 1) -}}
{{- $admin := (list ) -}}
{{- $internal := (get (fromJson (include "redpanda.createInternalListenerTLSCfg" (dict "a" (list $tls $l.tls) ))) "r") -}}
{{- if (gt ((get (fromJson (include "_shims.len" (dict "a" (list $internal) ))) "r") | int) (0 | int)) -}}
{{- $admin = (concat (default (list ) $admin) (list $internal)) -}}
{{- end -}}
{{- range $k, $lis := $l.external -}}
{{- if (or (not (get (fromJson (include "redpanda.AdminExternal.IsEnabled" (dict "a" (list $lis) ))) "r")) (not (get (fromJson (include "redpanda.ExternalTLS.IsEnabled" (dict "a" (list $lis.tls $l.tls $tls) ))) "r"))) -}}
{{- continue -}}
{{- end -}}
{{- $certName := (get (fromJson (include "redpanda.ExternalTLS.GetCertName" (dict "a" (list $lis.tls $l.tls) ))) "r") -}}
{{- $admin = (concat (default (list ) $admin) (list (dict "name" $k "enabled" true "cert_file" (printf "/etc/tls/certs/%s/tls.crt" $certName) "key_file" (printf "/etc/tls/certs/%s/tls.key" $certName) "require_client_auth" (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $lis.tls.requireClientAuth false) ))) "r") "truststore_file" (get (fromJson (include "redpanda.TLSCertMap.getTrustStoreFilePath" (dict "a" (list (deepCopy $tls.certs) $certName) ))) "r") ))) -}}
{{- end -}}
{{- (dict "r" $admin) | toJson -}}
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
{{- $am_20 := (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $l.authenticationMethod "") ))) "r") -}}
{{- if (ne $am_20 "") -}}
{{- $_ := (set $internal "authentication_method" $am_20) -}}
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
{{- $am_21 := (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $l.authenticationMethod "") ))) "r") -}}
{{- if (ne $am_21 "") -}}
{{- $_ := (set $listener "authentication_method" $am_21) -}}
{{- end -}}
{{- $result = (concat (default (list ) $result) (list $listener)) -}}
{{- end -}}
{{- (dict "r" $result) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.HTTPListeners.ListenersTLS" -}}
{{- $l := (index .a 0) -}}
{{- $tls := (index .a 1) -}}
{{- range $_ := (list 1) -}}
{{- $pp := (list ) -}}
{{- $internal := (get (fromJson (include "redpanda.createInternalListenerTLSCfg" (dict "a" (list $tls $l.tls) ))) "r") -}}
{{- if (gt ((get (fromJson (include "_shims.len" (dict "a" (list $internal) ))) "r") | int) (0 | int)) -}}
{{- $pp = (concat (default (list ) $pp) (list $internal)) -}}
{{- end -}}
{{- range $k, $lis := $l.external -}}
{{- if (or (not (get (fromJson (include "redpanda.HTTPExternal.IsEnabled" (dict "a" (list $lis) ))) "r")) (not (get (fromJson (include "redpanda.ExternalTLS.IsEnabled" (dict "a" (list $lis.tls $l.tls $tls) ))) "r"))) -}}
{{- continue -}}
{{- end -}}
{{- $certName := (get (fromJson (include "redpanda.ExternalTLS.GetCertName" (dict "a" (list $lis.tls $l.tls) ))) "r") -}}
{{- $pp = (concat (default (list ) $pp) (list (dict "name" $k "enabled" true "cert_file" (printf "/etc/tls/certs/%s/tls.crt" $certName) "key_file" (printf "/etc/tls/certs/%s/tls.key" $certName) "require_client_auth" (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $lis.tls.requireClientAuth false) ))) "r") "truststore_file" (get (fromJson (include "redpanda.TLSCertMap.getTrustStoreFilePath" (dict "a" (list (deepCopy $tls.certs) $certName) ))) "r") ))) -}}
{{- end -}}
{{- (dict "r" $pp) | toJson -}}
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

{{- define "redpanda.KafkaListeners.Listeners" -}}
{{- $l := (index .a 0) -}}
{{- $auth := (index .a 1) -}}
{{- range $_ := (list 1) -}}
{{- $internal := (get (fromJson (include "redpanda.createInternalListenerCfg" (dict "a" (list ($l.port | int)) ))) "r") -}}
{{- if (get (fromJson (include "redpanda.Auth.IsSASLEnabled" (dict "a" (list $auth) ))) "r") -}}
{{- $_ := (set $internal "authentication_method" "sasl") -}}
{{- end -}}
{{- $am_22 := (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $l.authenticationMethod "") ))) "r") -}}
{{- if (ne $am_22 "") -}}
{{- $_ := (set $internal "authentication_method" $am_22) -}}
{{- end -}}
{{- $kafka := (list $internal) -}}
{{- range $k, $l := $l.external -}}
{{- if (not (get (fromJson (include "redpanda.KafkaExternal.IsEnabled" (dict "a" (list $l) ))) "r")) -}}
{{- continue -}}
{{- end -}}
{{- $listener := (dict "name" $k "port" ($l.port | int) "address" "0.0.0.0" ) -}}
{{- if (get (fromJson (include "redpanda.Auth.IsSASLEnabled" (dict "a" (list $auth) ))) "r") -}}
{{- $_ := (set $listener "authentication_method" "sasl") -}}
{{- end -}}
{{- $am_23 := (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $l.authenticationMethod "") ))) "r") -}}
{{- if (ne $am_23 "") -}}
{{- $_ := (set $listener "authentication_method" $am_23) -}}
{{- end -}}
{{- $kafka = (concat (default (list ) $kafka) (list $listener)) -}}
{{- end -}}
{{- (dict "r" $kafka) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.KafkaListeners.ListenersTLS" -}}
{{- $l := (index .a 0) -}}
{{- $tls := (index .a 1) -}}
{{- range $_ := (list 1) -}}
{{- $kafka := (list ) -}}
{{- $internal := (get (fromJson (include "redpanda.createInternalListenerTLSCfg" (dict "a" (list $tls $l.tls) ))) "r") -}}
{{- if (gt ((get (fromJson (include "_shims.len" (dict "a" (list $internal) ))) "r") | int) (0 | int)) -}}
{{- $kafka = (concat (default (list ) $kafka) (list $internal)) -}}
{{- end -}}
{{- range $k, $lis := $l.external -}}
{{- if (or (not (get (fromJson (include "redpanda.KafkaExternal.IsEnabled" (dict "a" (list $lis) ))) "r")) (not (get (fromJson (include "redpanda.ExternalTLS.IsEnabled" (dict "a" (list $lis.tls $l.tls $tls) ))) "r"))) -}}
{{- continue -}}
{{- end -}}
{{- $certName := (get (fromJson (include "redpanda.ExternalTLS.GetCertName" (dict "a" (list $lis.tls $l.tls) ))) "r") -}}
{{- $kafka = (concat (default (list ) $kafka) (list (dict "name" $k "enabled" true "cert_file" (printf "/etc/tls/certs/%s/tls.crt" $certName) "key_file" (printf "/etc/tls/certs/%s/tls.key" $certName) "require_client_auth" (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $lis.tls.requireClientAuth false) ))) "r") "truststore_file" (get (fromJson (include "redpanda.TLSCertMap.getTrustStoreFilePath" (dict "a" (list (deepCopy $tls.certs) $certName) ))) "r") ))) -}}
{{- end -}}
{{- (dict "r" $kafka) | toJson -}}
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
{{- $am_24 := (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $sr.authenticationMethod "") ))) "r") -}}
{{- if (ne $am_24 "") -}}
{{- $_ := (set $internal "authentication_method" $am_24) -}}
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
{{- $am_25 := (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $l.authenticationMethod "") ))) "r") -}}
{{- if (ne $am_25 "") -}}
{{- $_ := (set $listener "authentication_method" $am_25) -}}
{{- end -}}
{{- $result = (concat (default (list ) $result) (list $listener)) -}}
{{- end -}}
{{- (dict "r" $result) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.SchemaRegistryListeners.ListenersTLS" -}}
{{- $l := (index .a 0) -}}
{{- $tls := (index .a 1) -}}
{{- range $_ := (list 1) -}}
{{- $listeners := (list ) -}}
{{- $internal := (get (fromJson (include "redpanda.createInternalListenerTLSCfg" (dict "a" (list $tls $l.tls) ))) "r") -}}
{{- if (gt ((get (fromJson (include "_shims.len" (dict "a" (list $internal) ))) "r") | int) (0 | int)) -}}
{{- $listeners = (concat (default (list ) $listeners) (list $internal)) -}}
{{- end -}}
{{- range $k, $lis := $l.external -}}
{{- if (or (not (get (fromJson (include "redpanda.SchemaRegistryExternal.IsEnabled" (dict "a" (list $lis) ))) "r")) (not (get (fromJson (include "redpanda.ExternalTLS.IsEnabled" (dict "a" (list $lis.tls $l.tls $tls) ))) "r"))) -}}
{{- continue -}}
{{- end -}}
{{- $certName := (get (fromJson (include "redpanda.ExternalTLS.GetCertName" (dict "a" (list $lis.tls $l.tls) ))) "r") -}}
{{- $listeners = (concat (default (list ) $listeners) (list (dict "name" $k "enabled" true "cert_file" (printf "/etc/tls/certs/%s/tls.crt" $certName) "key_file" (printf "/etc/tls/certs/%s/tls.key" $certName) "require_client_auth" (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $lis.tls.requireClientAuth false) ))) "r") "truststore_file" (get (fromJson (include "redpanda.TLSCertMap.getTrustStoreFilePath" (dict "a" (list (deepCopy $tls.certs) $certName) ))) "r") ))) -}}
{{- end -}}
{{- (dict "r" $listeners) | toJson -}}
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
{{- $tmp_tuple_22 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "_shims.asintegral" (dict "a" (list $v) ))) "r")) ))) "r") -}}
{{- $ok_27 := $tmp_tuple_22.T2 -}}
{{- $num_26 := ($tmp_tuple_22.T1 | int) -}}
{{- if $ok_27 -}}
{{- $input = $num_26 -}}
{{- end -}}
{{- $tmp_tuple_23 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "_shims.asnumeric" (dict "a" (list $v) ))) "r")) ))) "r") -}}
{{- $ok_29 := $tmp_tuple_23.T2 -}}
{{- $f_28 := ($tmp_tuple_23.T1 | float64) -}}
{{- if $ok_29 -}}
{{- $input = ($f_28 | int) -}}
{{- end -}}
{{- $_ := (set $result $k (min ($input | int64) (((sub ((add $r (((mod $r (2 | int)) | int))) | int) (1 | int)) | int) | int64))) -}}
{{- continue -}}
{{- end -}}
{{- $tmp_tuple_24 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "_shims.typetest" (dict "a" (list "bool" $v false) ))) "r")) ))) "r") -}}
{{- $ok_31 := $tmp_tuple_24.T2 -}}
{{- $b_30 := $tmp_tuple_24.T1 -}}
{{- if $ok_31 -}}
{{- $_ := (set $result $k $b_30) -}}
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

{{- define "redpanda.SecretRef.IsValid" -}}
{{- $sr := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- (dict "r" (and (and (ne $sr (coalesce nil)) (not (empty $sr.key))) (not (empty $sr.name)))) | toJson -}}
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

