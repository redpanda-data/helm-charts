{{- /* Generated from "values.go" */ -}}

{{- define "redpanda.AuditLogging.Translate" -}}
{{- $a := (index .a 0) -}}
{{- $dot := (index .a 1) -}}
{{- $isSASLEnabled := (index .a 2) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $result := (dict ) -}}
{{- if (not (get (fromJson (include "redpanda.RedpandaAtLeast_23_3_0" (dict "a" (list $dot) ))) "r")) -}}
{{- $_is_returning = true -}}
{{- (dict "r" $result) | toJson -}}
{{- break -}}
{{- end -}}
{{- $enabled := (and $a.enabled $isSASLEnabled) -}}
{{- $_ := (set $result "audit_enabled" $enabled) -}}
{{- if (not $enabled) -}}
{{- $_is_returning = true -}}
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
{{- $_is_returning = true -}}
{{- (dict "r" $result) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.Auth.IsSASLEnabled" -}}
{{- $a := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- if (eq $a.sasl (coalesce nil)) -}}
{{- $_is_returning = true -}}
{{- (dict "r" false) | toJson -}}
{{- break -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" $a.sasl.enabled) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.Auth.Translate" -}}
{{- $a := (index .a 0) -}}
{{- $isSASLEnabled := (index .a 1) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- if (not $isSASLEnabled) -}}
{{- $_is_returning = true -}}
{{- (dict "r" (coalesce nil)) | toJson -}}
{{- break -}}
{{- end -}}
{{- $users := (list (get (fromJson (include "redpanda.BootstrapUser.Username" (dict "a" (list $a.sasl.bootstrapUser) ))) "r")) -}}
{{- range $_, $u := $a.sasl.users -}}
{{- $users = (concat (default (list ) $users) (list $u.name)) -}}
{{- end -}}
{{- if $_is_returning -}}
{{- break -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" (dict "superusers" $users )) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.Logging.Translate" -}}
{{- $l := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $result := (dict ) -}}
{{- $clusterID_1 := (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $l.usageStats.clusterId "") ))) "r") -}}
{{- if (ne $clusterID_1 "") -}}
{{- $_ := (set $result "cluster_id" $clusterID_1) -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" $result) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.RedpandaResources.GetOverProvisionValue" -}}
{{- $rr := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- if (lt ((get (fromJson (include "_shims.resource_MilliValue" (dict "a" (list $rr.cpu.cores) ))) "r") | int64) (1000 | int64)) -}}
{{- $_is_returning = true -}}
{{- (dict "r" true) | toJson -}}
{{- break -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $rr.cpu.overprovisioned false) ))) "r")) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.Storage.IsTieredStorageEnabled" -}}
{{- $s := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $conf := (get (fromJson (include "redpanda.Storage.GetTieredStorageConfig" (dict "a" (list $s) ))) "r") -}}
{{- $tmp_tuple_3 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "_shims.dicttest" (dict "a" (list $conf "cloud_storage_enabled" (coalesce nil)) ))) "r")) ))) "r") -}}
{{- $ok := $tmp_tuple_3.T2 -}}
{{- $b := $tmp_tuple_3.T1 -}}
{{- $_is_returning = true -}}
{{- (dict "r" (and $ok (get (fromJson (include "_shims.typeassertion" (dict "a" (list "bool" $b) ))) "r"))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.Storage.GetTieredStorageConfig" -}}
{{- $s := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- if (gt ((get (fromJson (include "_shims.len" (dict "a" (list $s.tieredConfig) ))) "r") | int) (0 | int)) -}}
{{- $_is_returning = true -}}
{{- (dict "r" $s.tieredConfig) | toJson -}}
{{- break -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" $s.tiered.config) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.Storage.GetTieredStorageHostPath" -}}
{{- $s := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $hp := $s.tieredStorageHostPath -}}
{{- if (and (empty $hp) (ne $s.tiered (coalesce nil))) -}}
{{- $hp = $s.tiered.hostPath -}}
{{- end -}}
{{- if (empty $hp) -}}
{{- $_ := (fail (printf `storage.tiered.mountType is "%s" but storage.tiered.hostPath is empty` $s.tiered.mountType)) -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" $hp) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.Storage.CloudStorageCacheSize" -}}
{{- $s := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $tmp_tuple_4 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "_shims.dicttest" (dict "a" (list (get (fromJson (include "redpanda.Storage.GetTieredStorageConfig" (dict "a" (list $s) ))) "r") `cloud_storage_cache_size` (coalesce nil)) ))) "r")) ))) "r") -}}
{{- $ok := $tmp_tuple_4.T2 -}}
{{- $value := $tmp_tuple_4.T1 -}}
{{- if (not $ok) -}}
{{- $_is_returning = true -}}
{{- (dict "r" (coalesce nil)) | toJson -}}
{{- break -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" $value) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.Storage.TieredCacheDirectory" -}}
{{- $s := (index .a 0) -}}
{{- $dot := (index .a 1) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $values := $dot.Values.AsMap -}}
{{- $tmp_tuple_5 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "_shims.typetest" (dict "a" (list "string" (index $values.config.node "cloud_storage_cache_directory") "") ))) "r")) ))) "r") -}}
{{- $ok_3 := $tmp_tuple_5.T2 -}}
{{- $dir_2 := $tmp_tuple_5.T1 -}}
{{- if $ok_3 -}}
{{- $_is_returning = true -}}
{{- (dict "r" $dir_2) | toJson -}}
{{- break -}}
{{- end -}}
{{- $tieredConfig := (get (fromJson (include "redpanda.Storage.GetTieredStorageConfig" (dict "a" (list $values.storage) ))) "r") -}}
{{- $tmp_tuple_6 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "_shims.typetest" (dict "a" (list "string" (index $tieredConfig "cloud_storage_cache_directory") "") ))) "r")) ))) "r") -}}
{{- $ok_5 := $tmp_tuple_6.T2 -}}
{{- $dir_4 := $tmp_tuple_6.T1 -}}
{{- if $ok_5 -}}
{{- $_is_returning = true -}}
{{- (dict "r" $dir_4) | toJson -}}
{{- break -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" "/var/lib/redpanda/data/cloud_storage_cache") | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.Storage.TieredMountType" -}}
{{- $s := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- if (and (ne $s.tieredStoragePersistentVolume (coalesce nil)) $s.tieredStoragePersistentVolume.enabled) -}}
{{- $_is_returning = true -}}
{{- (dict "r" "persistentVolume") | toJson -}}
{{- break -}}
{{- end -}}
{{- if (not (empty $s.tieredStorageHostPath)) -}}
{{- $_is_returning = true -}}
{{- (dict "r" "hostPath") | toJson -}}
{{- break -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" $s.tiered.mountType) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.Storage.TieredPersistentVolumeLabels" -}}
{{- $s := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- if (ne $s.tieredStoragePersistentVolume (coalesce nil)) -}}
{{- $_is_returning = true -}}
{{- (dict "r" $s.tieredStoragePersistentVolume.labels) | toJson -}}
{{- break -}}
{{- end -}}
{{- if (ne $s.tiered (coalesce nil)) -}}
{{- $_is_returning = true -}}
{{- (dict "r" $s.tiered.persistentVolume.labels) | toJson -}}
{{- break -}}
{{- end -}}
{{- $_ := (fail `storage.tiered.mountType is "persistentVolume" but storage.tiered.persistentVolume is not configured`) -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.Storage.TieredPersistentVolumeAnnotations" -}}
{{- $s := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- if (ne $s.tieredStoragePersistentVolume (coalesce nil)) -}}
{{- $_is_returning = true -}}
{{- (dict "r" $s.tieredStoragePersistentVolume.annotations) | toJson -}}
{{- break -}}
{{- end -}}
{{- if (ne $s.tiered (coalesce nil)) -}}
{{- $_is_returning = true -}}
{{- (dict "r" $s.tiered.persistentVolume.annotations) | toJson -}}
{{- break -}}
{{- end -}}
{{- $_ := (fail `storage.tiered.mountType is "persistentVolume" but storage.tiered.persistentVolume is not configured`) -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.Storage.TieredPersistentVolumeStorageClass" -}}
{{- $s := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- if (ne $s.tieredStoragePersistentVolume (coalesce nil)) -}}
{{- $_is_returning = true -}}
{{- (dict "r" $s.tieredStoragePersistentVolume.storageClass) | toJson -}}
{{- break -}}
{{- end -}}
{{- if (ne $s.tiered (coalesce nil)) -}}
{{- $_is_returning = true -}}
{{- (dict "r" $s.tiered.persistentVolume.storageClass) | toJson -}}
{{- break -}}
{{- end -}}
{{- $_ := (fail `storage.tiered.mountType is "persistentVolume" but storage.tiered.persistentVolume is not configured`) -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.Storage.Translate" -}}
{{- $s := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $result := (dict ) -}}
{{- if (not (get (fromJson (include "redpanda.Storage.IsTieredStorageEnabled" (dict "a" (list $s) ))) "r")) -}}
{{- $_is_returning = true -}}
{{- (dict "r" $result) | toJson -}}
{{- break -}}
{{- end -}}
{{- $tieredStorageConfig := (get (fromJson (include "redpanda.Storage.GetTieredStorageConfig" (dict "a" (list $s) ))) "r") -}}
{{- range $k, $v := $tieredStorageConfig -}}
{{- if (or (eq $v (coalesce nil)) (empty $v)) -}}
{{- continue -}}
{{- end -}}
{{- if (eq $k "cloud_storage_cache_size") -}}
{{- $_ := (set $result $k (printf "%d" ((get (fromJson (include "_shims.resource_Value" (dict "a" (list $v) ))) "r") | int64))) -}}
{{- continue -}}
{{- end -}}
{{- $tmp_tuple_8 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "_shims.typetest" (dict "a" (list "string" $v "") ))) "r")) ))) "r") -}}
{{- $ok_7 := $tmp_tuple_8.T2 -}}
{{- $str_6 := $tmp_tuple_8.T1 -}}
{{- $tmp_tuple_9 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "_shims.typetest" (dict "a" (list "bool" $v false) ))) "r")) ))) "r") -}}
{{- $ok_9 := $tmp_tuple_9.T2 -}}
{{- $b_8 := $tmp_tuple_9.T1 -}}
{{- $tmp_tuple_10 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "_shims.asnumeric" (dict "a" (list $v) ))) "r")) ))) "r") -}}
{{- $isFloat_11 := $tmp_tuple_10.T2 -}}
{{- $f_10 := ($tmp_tuple_10.T1 | float64) -}}
{{- if $ok_7 -}}
{{- $_ := (set $result $k $str_6) -}}
{{- else -}}{{- if $ok_9 -}}
{{- $_ := (set $result $k $b_8) -}}
{{- else -}}{{- if $isFloat_11 -}}
{{- $_ := (set $result $k ($f_10 | int)) -}}
{{- else -}}
{{- $_ := (set $result $k (mustToJson $v)) -}}
{{- end -}}
{{- end -}}
{{- end -}}
{{- end -}}
{{- if $_is_returning -}}
{{- break -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" $result) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.Storage.StorageMinFreeBytes" -}}
{{- $s := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- if (and (ne $s.persistentVolume (coalesce nil)) (not $s.persistentVolume.enabled)) -}}
{{- $_is_returning = true -}}
{{- (dict "r" (5368709120 | int)) | toJson -}}
{{- break -}}
{{- end -}}
{{- $minimumFreeBytes := ((mulf (((get (fromJson (include "_shims.resource_Value" (dict "a" (list $s.persistentVolume.size) ))) "r") | int64) | float64) 0.05) | float64) -}}
{{- $_is_returning = true -}}
{{- (dict "r" (min (5368709120 | int) ($minimumFreeBytes | int64))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.Tuning.Translate" -}}
{{- $t := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $result := (dict ) -}}
{{- $s := (toJson $t) -}}
{{- $tune := (fromJson $s) -}}
{{- $tmp_tuple_11 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "_shims.typetest" (dict "a" (list (printf "map[%s]%s" "string" "interface {}") $tune (coalesce nil)) ))) "r")) ))) "r") -}}
{{- $ok := $tmp_tuple_11.T2 -}}
{{- $m := $tmp_tuple_11.T1 -}}
{{- if (not $ok) -}}
{{- $_is_returning = true -}}
{{- (dict "r" (dict )) | toJson -}}
{{- break -}}
{{- end -}}
{{- range $k, $v := $m -}}
{{- $_ := (set $result $k $v) -}}
{{- end -}}
{{- if $_is_returning -}}
{{- break -}}
{{- end -}}
{{- $_is_returning = true -}}
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
{{- $_is_returning := false -}}
{{- $result := (coalesce nil) -}}
{{- range $_, $i := untilStep (((0 | int) | int)|int) ($replicas|int) (1|int) -}}
{{- $result = (concat (default (list ) $result) (list (dict "host" (dict "address" (printf "%s-%d.%s" $fullname $i $internalDomain) "port" ($l.rpc.port | int) ) ))) -}}
{{- end -}}
{{- if $_is_returning -}}
{{- break -}}
{{- end -}}
{{- $_is_returning = true -}}
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
{{- $_is_returning := false -}}
{{- $_is_returning = true -}}
{{- (dict "r" (get (fromJson (include "redpanda.ServerList" (dict "a" (list $replicas "" $fullname $internalDomain ($l.admin.port | int)) ))) "r")) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.ServerList" -}}
{{- $replicas := (index .a 0) -}}
{{- $prefix := (index .a 1) -}}
{{- $fullname := (index .a 2) -}}
{{- $internalDomain := (index .a 3) -}}
{{- $port := (index .a 4) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $result := (coalesce nil) -}}
{{- range $_, $i := untilStep (((0 | int) | int)|int) ($replicas|int) (1|int) -}}
{{- $result = (concat (default (list ) $result) (list (printf "%s%s-%d.%s:%d" $prefix $fullname $i $internalDomain ($port | int)))) -}}
{{- end -}}
{{- if $_is_returning -}}
{{- break -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" $result) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.Listeners.TrustStoreVolume" -}}
{{- $l := (index .a 0) -}}
{{- $tls := (index .a 1) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $cmSources := (dict ) -}}
{{- $secretSources := (dict ) -}}
{{- range $_, $ts := (get (fromJson (include "redpanda.Listeners.TrustStores" (dict "a" (list $l $tls) ))) "r") -}}
{{- $projection := (get (fromJson (include "redpanda.TrustStore.VolumeProjection" (dict "a" (list $ts) ))) "r") -}}
{{- if (ne $projection.secret (coalesce nil)) -}}
{{- $_ := (set $secretSources $projection.secret.name (concat (default (list ) (index $secretSources $projection.secret.name)) (default (list ) $projection.secret.items))) -}}
{{- else -}}
{{- $_ := (set $cmSources $projection.configMap.name (concat (default (list ) (index $cmSources $projection.configMap.name)) (default (list ) $projection.configMap.items))) -}}
{{- end -}}
{{- end -}}
{{- if $_is_returning -}}
{{- break -}}
{{- end -}}
{{- $sources := (coalesce nil) -}}
{{- range $_, $name := (sortAlpha (keys $cmSources)) -}}
{{- $keys := (index $cmSources $name) -}}
{{- $sources = (concat (default (list ) $sources) (list (mustMergeOverwrite (dict ) (dict "configMap" (mustMergeOverwrite (dict ) (mustMergeOverwrite (dict ) (dict "name" $name )) (dict "items" (get (fromJson (include "redpanda.dedupKeyToPaths" (dict "a" (list $keys) ))) "r") )) )))) -}}
{{- end -}}
{{- if $_is_returning -}}
{{- break -}}
{{- end -}}
{{- range $_, $name := (sortAlpha (keys $secretSources)) -}}
{{- $keys := (index $secretSources $name) -}}
{{- $sources = (concat (default (list ) $sources) (list (mustMergeOverwrite (dict ) (dict "secret" (mustMergeOverwrite (dict ) (mustMergeOverwrite (dict ) (dict "name" $name )) (dict "items" (get (fromJson (include "redpanda.dedupKeyToPaths" (dict "a" (list $keys) ))) "r") )) )))) -}}
{{- end -}}
{{- if $_is_returning -}}
{{- break -}}
{{- end -}}
{{- if (lt ((get (fromJson (include "_shims.len" (dict "a" (list $sources) ))) "r") | int) (1 | int)) -}}
{{- $_is_returning = true -}}
{{- (dict "r" (coalesce nil)) | toJson -}}
{{- break -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" (mustMergeOverwrite (dict "name" "" ) (mustMergeOverwrite (dict ) (dict "projected" (mustMergeOverwrite (dict "sources" (coalesce nil) ) (dict "sources" $sources )) )) (dict "name" "truststores" ))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.dedupKeyToPaths" -}}
{{- $items := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $seen := (dict ) -}}
{{- $deduped := (coalesce nil) -}}
{{- range $_, $item := $items -}}
{{- $tmp_tuple_12 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "_shims.dicttest" (dict "a" (list $seen $item.key (coalesce nil)) ))) "r")) ))) "r") -}}
{{- $ok_12 := $tmp_tuple_12.T2 -}}
{{- if $ok_12 -}}
{{- continue -}}
{{- end -}}
{{- $deduped = (concat (default (list ) $deduped) (list $item)) -}}
{{- $_ := (set $seen $item.key true) -}}
{{- end -}}
{{- if $_is_returning -}}
{{- break -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" $deduped) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.Listeners.TrustStores" -}}
{{- $l := (index .a 0) -}}
{{- $tls := (index .a 1) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $tss := (get (fromJson (include "redpanda.KafkaListeners.TrustStores" (dict "a" (list $l.kafka $tls) ))) "r") -}}
{{- $tss = (concat (default (list ) $tss) (default (list ) (get (fromJson (include "redpanda.AdminListeners.TrustStores" (dict "a" (list $l.admin $tls) ))) "r"))) -}}
{{- $tss = (concat (default (list ) $tss) (default (list ) (get (fromJson (include "redpanda.HTTPListeners.TrustStores" (dict "a" (list $l.http $tls) ))) "r"))) -}}
{{- $tss = (concat (default (list ) $tss) (default (list ) (get (fromJson (include "redpanda.SchemaRegistryListeners.TrustStores" (dict "a" (list $l.schemaRegistry $tls) ))) "r"))) -}}
{{- $_is_returning = true -}}
{{- (dict "r" $tss) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.Config.CreateRPKConfiguration" -}}
{{- $c := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $result := (dict ) -}}
{{- range $k, $v := $c.rpk -}}
{{- $_ := (set $result $k $v) -}}
{{- end -}}
{{- if $_is_returning -}}
{{- break -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" $result) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.TLSCertMap.MustGet" -}}
{{- $m := (index .a 0) -}}
{{- $name := (index .a 1) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $tmp_tuple_13 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "_shims.dicttest" (dict "a" (list $m $name (coalesce nil)) ))) "r")) ))) "r") -}}
{{- $ok := $tmp_tuple_13.T2 -}}
{{- $cert := $tmp_tuple_13.T1 -}}
{{- if (not $ok) -}}
{{- $_ := (fail (printf "Certificate %q referenced, but not found in the tls.certs map" $name)) -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" $cert) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.BootstrapUser.BootstrapEnvironment" -}}
{{- $b := (index .a 0) -}}
{{- $fullname := (index .a 1) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $_is_returning = true -}}
{{- (dict "r" (concat (default (list ) (get (fromJson (include "redpanda.BootstrapUser.RpkEnvironment" (dict "a" (list $b $fullname) ))) "r")) (list (mustMergeOverwrite (dict "name" "" ) (dict "name" "RP_BOOTSTRAP_USER" "value" "$(RPK_USER):$(RPK_PASS):$(RPK_SASL_MECHANISM)" ))))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.BootstrapUser.Username" -}}
{{- $b := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- if (ne $b.name (coalesce nil)) -}}
{{- $_is_returning = true -}}
{{- (dict "r" $b.name) | toJson -}}
{{- break -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" "kubernetes-controller") | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.BootstrapUser.RpkEnvironment" -}}
{{- $b := (index .a 0) -}}
{{- $fullname := (index .a 1) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $_is_returning = true -}}
{{- (dict "r" (list (mustMergeOverwrite (dict "name" "" ) (dict "name" "RPK_PASS" "valueFrom" (mustMergeOverwrite (dict ) (dict "secretKeyRef" (get (fromJson (include "redpanda.BootstrapUser.SecretKeySelector" (dict "a" (list $b $fullname) ))) "r") )) )) (mustMergeOverwrite (dict "name" "" ) (dict "name" "RPK_USER" "value" (get (fromJson (include "redpanda.BootstrapUser.Username" (dict "a" (list $b) ))) "r") )) (mustMergeOverwrite (dict "name" "" ) (dict "name" "RPK_SASL_MECHANISM" "value" (get (fromJson (include "redpanda.BootstrapUser.GetMechanism" (dict "a" (list $b) ))) "r") )))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.BootstrapUser.GetMechanism" -}}
{{- $b := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- if (eq $b.mechanism "") -}}
{{- $_is_returning = true -}}
{{- (dict "r" "SCRAM-SHA-256") | toJson -}}
{{- break -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" $b.mechanism) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.BootstrapUser.SecretKeySelector" -}}
{{- $b := (index .a 0) -}}
{{- $fullname := (index .a 1) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- if (ne $b.secretKeyRef (coalesce nil)) -}}
{{- $_is_returning = true -}}
{{- (dict "r" $b.secretKeyRef) | toJson -}}
{{- break -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" (mustMergeOverwrite (dict "key" "" ) (mustMergeOverwrite (dict ) (dict "name" (printf "%s-bootstrap-user" $fullname) )) (dict "key" "password" ))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.TrustStore.TrustStoreFilePath" -}}
{{- $t := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $_is_returning = true -}}
{{- (dict "r" (printf "%s/%s" "/etc/truststores" (get (fromJson (include "redpanda.TrustStore.RelativePath" (dict "a" (list $t) ))) "r"))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.TrustStore.RelativePath" -}}
{{- $t := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- if (ne $t.configMapKeyRef (coalesce nil)) -}}
{{- $_is_returning = true -}}
{{- (dict "r" (printf "configmaps/%s-%s" $t.configMapKeyRef.name $t.configMapKeyRef.key)) | toJson -}}
{{- break -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" (printf "secrets/%s-%s" $t.secretKeyRef.name $t.secretKeyRef.key)) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.TrustStore.VolumeProjection" -}}
{{- $t := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- if (ne $t.configMapKeyRef (coalesce nil)) -}}
{{- $_is_returning = true -}}
{{- (dict "r" (mustMergeOverwrite (dict ) (dict "configMap" (mustMergeOverwrite (dict ) (mustMergeOverwrite (dict ) (dict "name" $t.configMapKeyRef.name )) (dict "items" (list (mustMergeOverwrite (dict "key" "" "path" "" ) (dict "key" $t.configMapKeyRef.key "path" (get (fromJson (include "redpanda.TrustStore.RelativePath" (dict "a" (list $t) ))) "r") ))) )) ))) | toJson -}}
{{- break -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" (mustMergeOverwrite (dict ) (dict "secret" (mustMergeOverwrite (dict ) (mustMergeOverwrite (dict ) (dict "name" $t.secretKeyRef.name )) (dict "items" (list (mustMergeOverwrite (dict "key" "" "path" "" ) (dict "key" $t.secretKeyRef.key "path" (get (fromJson (include "redpanda.TrustStore.RelativePath" (dict "a" (list $t) ))) "r") ))) )) ))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.InternalTLS.IsEnabled" -}}
{{- $t := (index .a 0) -}}
{{- $tls := (index .a 1) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $_is_returning = true -}}
{{- (dict "r" (and (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $t.enabled $tls.enabled) ))) "r") (ne $t.cert ""))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.InternalTLS.TrustStoreFilePath" -}}
{{- $t := (index .a 0) -}}
{{- $tls := (index .a 1) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- if (ne $t.trustStore (coalesce nil)) -}}
{{- $_is_returning = true -}}
{{- (dict "r" (get (fromJson (include "redpanda.TrustStore.TrustStoreFilePath" (dict "a" (list $t.trustStore) ))) "r")) | toJson -}}
{{- break -}}
{{- end -}}
{{- if (get (fromJson (include "redpanda.TLSCertMap.MustGet" (dict "a" (list (deepCopy $tls.certs) $t.cert) ))) "r").caEnabled -}}
{{- $_is_returning = true -}}
{{- (dict "r" (printf "/etc/tls/certs/%s/ca.crt" $t.cert)) | toJson -}}
{{- break -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" "/etc/ssl/certs/ca-certificates.crt") | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.InternalTLS.ServerCAPath" -}}
{{- $t := (index .a 0) -}}
{{- $tls := (index .a 1) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- if (get (fromJson (include "redpanda.TLSCertMap.MustGet" (dict "a" (list (deepCopy $tls.certs) $t.cert) ))) "r").caEnabled -}}
{{- $_is_returning = true -}}
{{- (dict "r" (printf "/etc/tls/certs/%s/ca.crt" $t.cert)) | toJson -}}
{{- break -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" (printf "/etc/tls/certs/%s/tls.crt" $t.cert)) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.ExternalTLS.GetCert" -}}
{{- $t := (index .a 0) -}}
{{- $i := (index .a 1) -}}
{{- $tls := (index .a 2) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $_is_returning = true -}}
{{- (dict "r" (get (fromJson (include "redpanda.TLSCertMap.MustGet" (dict "a" (list (deepCopy $tls.certs) (get (fromJson (include "redpanda.ExternalTLS.GetCertName" (dict "a" (list $t $i) ))) "r")) ))) "r")) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.ExternalTLS.GetCertName" -}}
{{- $t := (index .a 0) -}}
{{- $i := (index .a 1) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $_is_returning = true -}}
{{- (dict "r" (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $t.cert $i.cert) ))) "r")) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.ExternalTLS.TrustStoreFilePath" -}}
{{- $t := (index .a 0) -}}
{{- $i := (index .a 1) -}}
{{- $tls := (index .a 2) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- if (ne $t.trustStore (coalesce nil)) -}}
{{- $_is_returning = true -}}
{{- (dict "r" (get (fromJson (include "redpanda.TrustStore.TrustStoreFilePath" (dict "a" (list $t.trustStore) ))) "r")) | toJson -}}
{{- break -}}
{{- end -}}
{{- if (get (fromJson (include "redpanda.ExternalTLS.GetCert" (dict "a" (list $t $i $tls) ))) "r").caEnabled -}}
{{- $_is_returning = true -}}
{{- (dict "r" (printf "/etc/tls/certs/%s/ca.crt" (get (fromJson (include "redpanda.ExternalTLS.GetCertName" (dict "a" (list $t $i) ))) "r"))) | toJson -}}
{{- break -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" "/etc/ssl/certs/ca-certificates.crt") | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.ExternalTLS.IsEnabled" -}}
{{- $t := (index .a 0) -}}
{{- $i := (index .a 1) -}}
{{- $tls := (index .a 2) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- if (eq $t (coalesce nil)) -}}
{{- $_is_returning = true -}}
{{- (dict "r" false) | toJson -}}
{{- break -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" (and (ne (get (fromJson (include "redpanda.ExternalTLS.GetCertName" (dict "a" (list $t $i) ))) "r") "") (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $t.enabled (get (fromJson (include "redpanda.InternalTLS.IsEnabled" (dict "a" (list $i $tls) ))) "r")) ))) "r"))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.AdminListeners.ConsoleTLS" -}}
{{- $l := (index .a 0) -}}
{{- $tls := (index .a 1) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $t := (mustMergeOverwrite (dict "enabled" false "caFilepath" "" "certFilepath" "" "keyFilepath" "" "insecureSkipTlsVerify" false ) (dict "enabled" (get (fromJson (include "redpanda.InternalTLS.IsEnabled" (dict "a" (list $l.tls $tls) ))) "r") )) -}}
{{- if (not $t.enabled) -}}
{{- $_is_returning = true -}}
{{- (dict "r" $t) | toJson -}}
{{- break -}}
{{- end -}}
{{- $adminAPIPrefix := "/mnt/cert/adminapi" -}}
{{- $_ := (set $t "caFilepath" (printf "%s/%s/ca.crt" $adminAPIPrefix $l.tls.cert)) -}}
{{- if (not $l.tls.requireClientAuth) -}}
{{- $_is_returning = true -}}
{{- (dict "r" $t) | toJson -}}
{{- break -}}
{{- end -}}
{{- $_ := (set $t "certFilepath" (printf "%s/%s/tls.crt" $adminAPIPrefix $l.tls.cert)) -}}
{{- $_ := (set $t "keyFilepath" (printf "%s/%s/tls.key" $adminAPIPrefix $l.tls.cert)) -}}
{{- $_is_returning = true -}}
{{- (dict "r" $t) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.AdminListeners.Listeners" -}}
{{- $l := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $admin := (list (get (fromJson (include "redpanda.createInternalListenerCfg" (dict "a" (list ($l.port | int)) ))) "r")) -}}
{{- range $k, $lis := $l.external -}}
{{- if (not (get (fromJson (include "redpanda.AdminExternal.IsEnabled" (dict "a" (list $lis) ))) "r")) -}}
{{- continue -}}
{{- end -}}
{{- $admin = (concat (default (list ) $admin) (list (dict "name" $k "port" ($lis.port | int) "address" "0.0.0.0" ))) -}}
{{- end -}}
{{- if $_is_returning -}}
{{- break -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" $admin) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.AdminListeners.ListenersTLS" -}}
{{- $l := (index .a 0) -}}
{{- $tls := (index .a 1) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
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
{{- $admin = (concat (default (list ) $admin) (list (dict "name" $k "enabled" true "cert_file" (printf "/etc/tls/certs/%s/tls.crt" $certName) "key_file" (printf "/etc/tls/certs/%s/tls.key" $certName) "require_client_auth" (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $lis.tls.requireClientAuth false) ))) "r") "truststore_file" (get (fromJson (include "redpanda.ExternalTLS.TrustStoreFilePath" (dict "a" (list $lis.tls $l.tls $tls) ))) "r") ))) -}}
{{- end -}}
{{- if $_is_returning -}}
{{- break -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" $admin) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.AdminListeners.TrustStores" -}}
{{- $l := (index .a 0) -}}
{{- $tls := (index .a 1) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $tss := (list ) -}}
{{- if (and (get (fromJson (include "redpanda.InternalTLS.IsEnabled" (dict "a" (list $l.tls $tls) ))) "r") (ne $l.tls.trustStore (coalesce nil))) -}}
{{- $tss = (concat (default (list ) $tss) (list $l.tls.trustStore)) -}}
{{- end -}}
{{- range $_, $key := (sortAlpha (keys $l.external)) -}}
{{- $lis := (index $l.external $key) -}}
{{- if (or (or (not (get (fromJson (include "redpanda.AdminExternal.IsEnabled" (dict "a" (list $lis) ))) "r")) (not (get (fromJson (include "redpanda.ExternalTLS.IsEnabled" (dict "a" (list $lis.tls $l.tls $tls) ))) "r"))) (eq $lis.tls.trustStore (coalesce nil))) -}}
{{- continue -}}
{{- end -}}
{{- $tss = (concat (default (list ) $tss) (list $lis.tls.trustStore)) -}}
{{- end -}}
{{- if $_is_returning -}}
{{- break -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" $tss) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.AdminExternal.IsEnabled" -}}
{{- $l := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $_is_returning = true -}}
{{- (dict "r" (and (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $l.enabled true) ))) "r") (gt ($l.port | int) (0 | int)))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.HTTPListeners.Listeners" -}}
{{- $l := (index .a 0) -}}
{{- $saslEnabled := (index .a 1) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $internal := (get (fromJson (include "redpanda.createInternalListenerCfg" (dict "a" (list ($l.port | int)) ))) "r") -}}
{{- if $saslEnabled -}}
{{- $_ := (set $internal "authentication_method" "http_basic") -}}
{{- end -}}
{{- $am_13 := (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $l.authenticationMethod "") ))) "r") -}}
{{- if (ne $am_13 "") -}}
{{- $_ := (set $internal "authentication_method" $am_13) -}}
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
{{- $am_14 := (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $l.authenticationMethod "") ))) "r") -}}
{{- if (ne $am_14 "") -}}
{{- $_ := (set $listener "authentication_method" $am_14) -}}
{{- end -}}
{{- $result = (concat (default (list ) $result) (list $listener)) -}}
{{- end -}}
{{- if $_is_returning -}}
{{- break -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" $result) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.HTTPListeners.ListenersTLS" -}}
{{- $l := (index .a 0) -}}
{{- $tls := (index .a 1) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
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
{{- $pp = (concat (default (list ) $pp) (list (dict "name" $k "enabled" true "cert_file" (printf "/etc/tls/certs/%s/tls.crt" $certName) "key_file" (printf "/etc/tls/certs/%s/tls.key" $certName) "require_client_auth" (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $lis.tls.requireClientAuth false) ))) "r") "truststore_file" (get (fromJson (include "redpanda.ExternalTLS.TrustStoreFilePath" (dict "a" (list $lis.tls $l.tls $tls) ))) "r") ))) -}}
{{- end -}}
{{- if $_is_returning -}}
{{- break -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" $pp) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.HTTPListeners.TrustStores" -}}
{{- $l := (index .a 0) -}}
{{- $tls := (index .a 1) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $tss := (coalesce nil) -}}
{{- if (and (get (fromJson (include "redpanda.InternalTLS.IsEnabled" (dict "a" (list $l.tls $tls) ))) "r") (ne $l.tls.trustStore (coalesce nil))) -}}
{{- $tss = (concat (default (list ) $tss) (list $l.tls.trustStore)) -}}
{{- end -}}
{{- range $_, $key := (sortAlpha (keys $l.external)) -}}
{{- $lis := (index $l.external $key) -}}
{{- if (or (or (not (get (fromJson (include "redpanda.HTTPExternal.IsEnabled" (dict "a" (list $lis) ))) "r")) (not (get (fromJson (include "redpanda.ExternalTLS.IsEnabled" (dict "a" (list $lis.tls $l.tls $tls) ))) "r"))) (eq $lis.tls.trustStore (coalesce nil))) -}}
{{- continue -}}
{{- end -}}
{{- $tss = (concat (default (list ) $tss) (list $lis.tls.trustStore)) -}}
{{- end -}}
{{- if $_is_returning -}}
{{- break -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" $tss) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.HTTPExternal.IsEnabled" -}}
{{- $l := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $_is_returning = true -}}
{{- (dict "r" (and (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $l.enabled true) ))) "r") (gt ($l.port | int) (0 | int)))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.KafkaListeners.Listeners" -}}
{{- $l := (index .a 0) -}}
{{- $auth := (index .a 1) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $internal := (get (fromJson (include "redpanda.createInternalListenerCfg" (dict "a" (list ($l.port | int)) ))) "r") -}}
{{- if (get (fromJson (include "redpanda.Auth.IsSASLEnabled" (dict "a" (list $auth) ))) "r") -}}
{{- $_ := (set $internal "authentication_method" "sasl") -}}
{{- end -}}
{{- $am_15 := (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $l.authenticationMethod "") ))) "r") -}}
{{- if (ne $am_15 "") -}}
{{- $_ := (set $internal "authentication_method" $am_15) -}}
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
{{- $am_16 := (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $l.authenticationMethod "") ))) "r") -}}
{{- if (ne $am_16 "") -}}
{{- $_ := (set $listener "authentication_method" $am_16) -}}
{{- end -}}
{{- $kafka = (concat (default (list ) $kafka) (list $listener)) -}}
{{- end -}}
{{- if $_is_returning -}}
{{- break -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" $kafka) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.KafkaListeners.ListenersTLS" -}}
{{- $l := (index .a 0) -}}
{{- $tls := (index .a 1) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
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
{{- $kafka = (concat (default (list ) $kafka) (list (dict "name" $k "enabled" true "cert_file" (printf "/etc/tls/certs/%s/tls.crt" $certName) "key_file" (printf "/etc/tls/certs/%s/tls.key" $certName) "require_client_auth" (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $lis.tls.requireClientAuth false) ))) "r") "truststore_file" (get (fromJson (include "redpanda.ExternalTLS.TrustStoreFilePath" (dict "a" (list $lis.tls $l.tls $tls) ))) "r") ))) -}}
{{- end -}}
{{- if $_is_returning -}}
{{- break -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" $kafka) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.KafkaListeners.TrustStores" -}}
{{- $l := (index .a 0) -}}
{{- $tls := (index .a 1) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $tss := (coalesce nil) -}}
{{- if (and (get (fromJson (include "redpanda.InternalTLS.IsEnabled" (dict "a" (list $l.tls $tls) ))) "r") (ne $l.tls.trustStore (coalesce nil))) -}}
{{- $tss = (concat (default (list ) $tss) (list $l.tls.trustStore)) -}}
{{- end -}}
{{- range $_, $key := (sortAlpha (keys $l.external)) -}}
{{- $lis := (index $l.external $key) -}}
{{- if (or (or (not (get (fromJson (include "redpanda.KafkaExternal.IsEnabled" (dict "a" (list $lis) ))) "r")) (not (get (fromJson (include "redpanda.ExternalTLS.IsEnabled" (dict "a" (list $lis.tls $l.tls $tls) ))) "r"))) (eq $lis.tls.trustStore (coalesce nil))) -}}
{{- continue -}}
{{- end -}}
{{- $tss = (concat (default (list ) $tss) (list $lis.tls.trustStore)) -}}
{{- end -}}
{{- if $_is_returning -}}
{{- break -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" $tss) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.KafkaListeners.ConsolemTLS" -}}
{{- $k := (index .a 0) -}}
{{- $tls := (index .a 1) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $t := (mustMergeOverwrite (dict "enabled" false "caFilepath" "" "certFilepath" "" "keyFilepath" "" "insecureSkipTlsVerify" false ) (dict "enabled" (get (fromJson (include "redpanda.InternalTLS.IsEnabled" (dict "a" (list $k.tls $tls) ))) "r") )) -}}
{{- if (not $t.enabled) -}}
{{- $_is_returning = true -}}
{{- (dict "r" $t) | toJson -}}
{{- break -}}
{{- end -}}
{{- $kafkaPathPrefix := "/mnt/cert/kafka" -}}
{{- $_ := (set $t "caFilepath" (printf "%s/%s/ca.crt" $kafkaPathPrefix $k.tls.cert)) -}}
{{- if (not $k.tls.requireClientAuth) -}}
{{- $_is_returning = true -}}
{{- (dict "r" $t) | toJson -}}
{{- break -}}
{{- end -}}
{{- $_ := (set $t "certFilepath" (printf "%s/%s/tls.crt" $kafkaPathPrefix $k.tls.cert)) -}}
{{- $_ := (set $t "keyFilepath" (printf "%s/%s/tls.key" $kafkaPathPrefix $k.tls.cert)) -}}
{{- $_is_returning = true -}}
{{- (dict "r" $t) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.KafkaExternal.IsEnabled" -}}
{{- $l := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $_is_returning = true -}}
{{- (dict "r" (and (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $l.enabled true) ))) "r") (gt ($l.port | int) (0 | int)))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.SchemaRegistryListeners.Listeners" -}}
{{- $sr := (index .a 0) -}}
{{- $saslEnabled := (index .a 1) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $internal := (get (fromJson (include "redpanda.createInternalListenerCfg" (dict "a" (list ($sr.port | int)) ))) "r") -}}
{{- if $saslEnabled -}}
{{- $_ := (set $internal "authentication_method" "http_basic") -}}
{{- end -}}
{{- $am_17 := (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $sr.authenticationMethod "") ))) "r") -}}
{{- if (ne $am_17 "") -}}
{{- $_ := (set $internal "authentication_method" $am_17) -}}
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
{{- $am_18 := (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $l.authenticationMethod "") ))) "r") -}}
{{- if (ne $am_18 "") -}}
{{- $_ := (set $listener "authentication_method" $am_18) -}}
{{- end -}}
{{- $result = (concat (default (list ) $result) (list $listener)) -}}
{{- end -}}
{{- if $_is_returning -}}
{{- break -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" $result) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.SchemaRegistryListeners.ListenersTLS" -}}
{{- $l := (index .a 0) -}}
{{- $tls := (index .a 1) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
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
{{- $listeners = (concat (default (list ) $listeners) (list (dict "name" $k "enabled" true "cert_file" (printf "/etc/tls/certs/%s/tls.crt" $certName) "key_file" (printf "/etc/tls/certs/%s/tls.key" $certName) "require_client_auth" (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $lis.tls.requireClientAuth false) ))) "r") "truststore_file" (get (fromJson (include "redpanda.ExternalTLS.TrustStoreFilePath" (dict "a" (list $lis.tls $l.tls $tls) ))) "r") ))) -}}
{{- end -}}
{{- if $_is_returning -}}
{{- break -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" $listeners) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.SchemaRegistryListeners.TrustStores" -}}
{{- $l := (index .a 0) -}}
{{- $tls := (index .a 1) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $tss := (coalesce nil) -}}
{{- if (and (get (fromJson (include "redpanda.InternalTLS.IsEnabled" (dict "a" (list $l.tls $tls) ))) "r") (ne $l.tls.trustStore (coalesce nil))) -}}
{{- $tss = (concat (default (list ) $tss) (list $l.tls.trustStore)) -}}
{{- end -}}
{{- range $_, $key := (sortAlpha (keys $l.external)) -}}
{{- $lis := (index $l.external $key) -}}
{{- if (or (or (not (get (fromJson (include "redpanda.SchemaRegistryExternal.IsEnabled" (dict "a" (list $lis) ))) "r")) (not (get (fromJson (include "redpanda.ExternalTLS.IsEnabled" (dict "a" (list $lis.tls $l.tls $tls) ))) "r"))) (eq $lis.tls.trustStore (coalesce nil))) -}}
{{- continue -}}
{{- end -}}
{{- $tss = (concat (default (list ) $tss) (list $lis.tls.trustStore)) -}}
{{- end -}}
{{- if $_is_returning -}}
{{- break -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" $tss) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.SchemaRegistryListeners.ConsoleTLS" -}}
{{- $sr := (index .a 0) -}}
{{- $tls := (index .a 1) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $t := (mustMergeOverwrite (dict "enabled" false "caFilepath" "" "certFilepath" "" "keyFilepath" "" "insecureSkipTlsVerify" false ) (dict "enabled" (get (fromJson (include "redpanda.InternalTLS.IsEnabled" (dict "a" (list $sr.tls $tls) ))) "r") )) -}}
{{- if (not $t.enabled) -}}
{{- $_is_returning = true -}}
{{- (dict "r" $t) | toJson -}}
{{- break -}}
{{- end -}}
{{- $schemaRegistryPrefix := "/mnt/cert/schemaregistry" -}}
{{- $_ := (set $t "caFilepath" (printf "%s/%s/ca.crt" $schemaRegistryPrefix $sr.tls.cert)) -}}
{{- if (not $sr.tls.requireClientAuth) -}}
{{- $_is_returning = true -}}
{{- (dict "r" $t) | toJson -}}
{{- break -}}
{{- end -}}
{{- $_ := (set $t "certFilepath" (printf "%s/%s/tls.crt" $schemaRegistryPrefix $sr.tls.cert)) -}}
{{- $_ := (set $t "keyFilepath" (printf "%s/%s/tls.key" $schemaRegistryPrefix $sr.tls.cert)) -}}
{{- $_is_returning = true -}}
{{- (dict "r" $t) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.SchemaRegistryExternal.IsEnabled" -}}
{{- $l := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $_is_returning = true -}}
{{- (dict "r" (and (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $l.enabled true) ))) "r") (gt ($l.port | int) (0 | int)))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.TunableConfig.Translate" -}}
{{- $c := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- if (eq $c (coalesce nil)) -}}
{{- $_is_returning = true -}}
{{- (dict "r" (coalesce nil)) | toJson -}}
{{- break -}}
{{- end -}}
{{- $result := (dict ) -}}
{{- range $k, $v := $c -}}
{{- if (not (empty $v)) -}}
{{- $_ := (set $result $k $v) -}}
{{- end -}}
{{- end -}}
{{- if $_is_returning -}}
{{- break -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" $result) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.NodeConfig.Translate" -}}
{{- $c := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $result := (dict ) -}}
{{- range $k, $v := $c -}}
{{- if (not (empty $v)) -}}
{{- $tmp_tuple_16 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "_shims.asnumeric" (dict "a" (list $v) ))) "r")) ))) "r") -}}
{{- $ok_19 := $tmp_tuple_16.T2 -}}
{{- if $ok_19 -}}
{{- $_ := (set $result $k $v) -}}
{{- else -}}{{- if (kindIs "bool" $v) -}}
{{- $_ := (set $result $k $v) -}}
{{- else -}}
{{- $_ := (set $result $k (toYaml $v)) -}}
{{- end -}}
{{- end -}}
{{- end -}}
{{- end -}}
{{- if $_is_returning -}}
{{- break -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" $result) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.ClusterConfig.Translate" -}}
{{- $c := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $result := (dict ) -}}
{{- range $k, $v := $c -}}
{{- $tmp_tuple_17 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "_shims.typetest" (dict "a" (list "bool" $v false) ))) "r")) ))) "r") -}}
{{- $ok_21 := $tmp_tuple_17.T2 -}}
{{- $b_20 := $tmp_tuple_17.T1 -}}
{{- if $ok_21 -}}
{{- $_ := (set $result $k $b_20) -}}
{{- continue -}}
{{- end -}}
{{- if (not (empty $v)) -}}
{{- $_ := (set $result $k $v) -}}
{{- end -}}
{{- end -}}
{{- if $_is_returning -}}
{{- break -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" $result) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.SecretRef.IsValid" -}}
{{- $sr := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $_is_returning = true -}}
{{- (dict "r" (and (and (ne $sr (coalesce nil)) (not (empty $sr.key))) (not (empty $sr.name)))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.TieredStorageCredentials.IsAccessKeyReferenceValid" -}}
{{- $tsc := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $_is_returning = true -}}
{{- (dict "r" (and (and (ne $tsc.accessKey (coalesce nil)) (ne $tsc.accessKey.name "")) (ne $tsc.accessKey.key ""))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.TieredStorageCredentials.IsSecretKeyReferenceValid" -}}
{{- $tsc := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $_is_returning = true -}}
{{- (dict "r" (and (and (ne $tsc.secretKey (coalesce nil)) (ne $tsc.secretKey.name "")) (ne $tsc.secretKey.key ""))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

