{{- /* Generated from "statefulset.go" */ -}}

{{- define "redpanda.StatefulSetRedpandaEnv" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $values := $dot.Values.AsMap -}}
{{- $userEnv := (coalesce nil) -}}
{{- range $_, $container := $values.statefulset.podTemplate.spec.containers -}}
{{- if (eq $container.name "redpanda") -}}
{{- $userEnv = $container.env -}}
{{- end -}}
{{- end -}}
{{- (dict "r" (concat (default (list ) (list (mustMergeOverwrite (dict "name" "" ) (dict "name" "SERVICE_NAME" "valueFrom" (mustMergeOverwrite (dict ) (dict "fieldRef" (mustMergeOverwrite (dict "fieldPath" "" ) (dict "fieldPath" "metadata.name" )) )) )) (mustMergeOverwrite (dict "name" "" ) (dict "name" "POD_IP" "valueFrom" (mustMergeOverwrite (dict ) (dict "fieldRef" (mustMergeOverwrite (dict "fieldPath" "" ) (dict "fieldPath" "status.podIP" )) )) )) (mustMergeOverwrite (dict "name" "" ) (dict "name" "HOST_IP" "valueFrom" (mustMergeOverwrite (dict ) (dict "fieldRef" (mustMergeOverwrite (dict "fieldPath" "" ) (dict "fieldPath" "status.hostIP" )) )) )))) (default (list ) $userEnv))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.StatefulSetPodLabelsSelector" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- if $dot.Release.IsUpgrade -}}
{{- $tmp_tuple_1 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "_shims.lookup" (dict "a" (list "apps/v1" "StatefulSet" $dot.Release.Namespace (get (fromJson (include "redpanda.Fullname" (dict "a" (list $dot) ))) "r")) ))) "r")) ))) "r") -}}
{{- $ok_2 := $tmp_tuple_1.T2 -}}
{{- $existing_1 := $tmp_tuple_1.T1 -}}
{{- if (and $ok_2 (gt ((get (fromJson (include "_shims.len" (dict "a" (list $existing_1.spec.selector.matchLabels) ))) "r") | int) (0 | int))) -}}
{{- (dict "r" $existing_1.spec.selector.matchLabels) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}
{{- $values := $dot.Values.AsMap -}}
{{- $additionalSelectorLabels := (dict ) -}}
{{- if (ne $values.statefulset.additionalSelectorLabels (coalesce nil)) -}}
{{- $additionalSelectorLabels = $values.statefulset.additionalSelectorLabels -}}
{{- end -}}
{{- $component := (printf "%s-statefulset" (trimSuffix "-" (trunc (51 | int) (get (fromJson (include "redpanda.Name" (dict "a" (list $dot) ))) "r")))) -}}
{{- $defaults := (dict "app.kubernetes.io/component" $component "app.kubernetes.io/instance" $dot.Release.Name "app.kubernetes.io/name" (get (fromJson (include "redpanda.Name" (dict "a" (list $dot) ))) "r") ) -}}
{{- (dict "r" (merge (dict ) $additionalSelectorLabels $defaults)) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.StatefulSetPodLabels" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- if $dot.Release.IsUpgrade -}}
{{- $tmp_tuple_2 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "_shims.lookup" (dict "a" (list "apps/v1" "StatefulSet" $dot.Release.Namespace (get (fromJson (include "redpanda.Fullname" (dict "a" (list $dot) ))) "r")) ))) "r")) ))) "r") -}}
{{- $ok_4 := $tmp_tuple_2.T2 -}}
{{- $existing_3 := $tmp_tuple_2.T1 -}}
{{- if (and $ok_4 (gt ((get (fromJson (include "_shims.len" (dict "a" (list $existing_3.spec.template.metadata.labels) ))) "r") | int) (0 | int))) -}}
{{- (dict "r" $existing_3.spec.template.metadata.labels) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}
{{- $values := $dot.Values.AsMap -}}
{{- $statefulSetLabels := (dict ) -}}
{{- if (ne $values.statefulset.podTemplate.labels (coalesce nil)) -}}
{{- $statefulSetLabels = $values.statefulset.podTemplate.labels -}}
{{- end -}}
{{- $defaults := (dict "redpanda.com/poddisruptionbudget" (get (fromJson (include "redpanda.Fullname" (dict "a" (list $dot) ))) "r") ) -}}
{{- (dict "r" (merge (dict ) $statefulSetLabels (get (fromJson (include "redpanda.StatefulSetPodLabelsSelector" (dict "a" (list $dot) ))) "r") $defaults (get (fromJson (include "redpanda.FullLabels" (dict "a" (list $dot) ))) "r"))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.StatefulSetPodAnnotations" -}}
{{- $dot := (index .a 0) -}}
{{- $configMapChecksum := (index .a 1) -}}
{{- range $_ := (list 1) -}}
{{- $values := $dot.Values.AsMap -}}
{{- $configMapChecksumAnnotation := (dict "config.redpanda.com/checksum" $configMapChecksum ) -}}
{{- if (ne $values.statefulset.podTemplate.annotations (coalesce nil)) -}}
{{- (dict "r" (merge (dict ) $values.statefulset.podTemplate.annotations $configMapChecksumAnnotation)) | toJson -}}
{{- break -}}
{{- end -}}
{{- (dict "r" (merge (dict ) $values.statefulset.annotations $configMapChecksumAnnotation)) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.StatefulSetVolumes" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $fullname := (get (fromJson (include "redpanda.Fullname" (dict "a" (list $dot) ))) "r") -}}
{{- $volumes := (get (fromJson (include "redpanda.CommonVolumes" (dict "a" (list $dot) ))) "r") -}}
{{- $values := $dot.Values.AsMap -}}
{{- $volumes = (concat (default (list ) $volumes) (default (list ) (list (mustMergeOverwrite (dict "name" "" ) (mustMergeOverwrite (dict ) (dict "secret" (mustMergeOverwrite (dict ) (dict "secretName" (printf "%.50s-sts-lifecycle" $fullname) "defaultMode" (0o775 | int) )) )) (dict "name" "lifecycle-scripts" )) (mustMergeOverwrite (dict "name" "" ) (mustMergeOverwrite (dict ) (dict "configMap" (mustMergeOverwrite (dict ) (mustMergeOverwrite (dict ) (dict "name" $fullname )) (dict )) )) (dict "name" $fullname )) (mustMergeOverwrite (dict "name" "" ) (mustMergeOverwrite (dict ) (dict "emptyDir" (mustMergeOverwrite (dict ) (dict )) )) (dict "name" "config" )) (mustMergeOverwrite (dict "name" "" ) (mustMergeOverwrite (dict ) (dict "secret" (mustMergeOverwrite (dict ) (dict "secretName" (printf "%.51s-configurator" $fullname) "defaultMode" (0o775 | int) )) )) (dict "name" (printf "%.51s-configurator" $fullname) )) (mustMergeOverwrite (dict "name" "" ) (mustMergeOverwrite (dict ) (dict "secret" (mustMergeOverwrite (dict ) (dict "secretName" (printf "%s-config-watcher" $fullname) "defaultMode" (0o775 | int) )) )) (dict "name" (printf "%s-config-watcher" $fullname) )) (mustMergeOverwrite (dict "name" "" ) (mustMergeOverwrite (dict ) (dict "secret" (mustMergeOverwrite (dict ) (dict "secretName" (printf "%.49s-fs-validator" $fullname) "defaultMode" (0o775 | int) )) )) (dict "name" (printf "%.49s-fs-validator" $fullname) ))))) -}}
{{- $vol_5 := (get (fromJson (include "redpanda.Listeners.TrustStoreVolume" (dict "a" (list $values.listeners $values.tls) ))) "r") -}}
{{- if (ne $vol_5 (coalesce nil)) -}}
{{- $volumes = (concat (default (list ) $volumes) (list $vol_5)) -}}
{{- end -}}
{{- $volumes = (concat (default (list ) $volumes) (default (list ) $values.statefulset.extraVolumes)) -}}
{{- (dict "r" $volumes) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.StatefulSetVolumeMounts" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $mounts := (get (fromJson (include "redpanda.CommonMounts" (dict "a" (list $dot) ))) "r") -}}
{{- $values := $dot.Values.AsMap -}}
{{- $mounts = (concat (default (list ) $mounts) (default (list ) (list (mustMergeOverwrite (dict "name" "" "mountPath" "" ) (dict "name" "config" "mountPath" "/etc/redpanda" )) (mustMergeOverwrite (dict "name" "" "mountPath" "" ) (dict "name" (get (fromJson (include "redpanda.Fullname" (dict "a" (list $dot) ))) "r") "mountPath" "/tmp/base-config" )) (mustMergeOverwrite (dict "name" "" "mountPath" "" ) (dict "name" "lifecycle-scripts" "mountPath" "/var/lifecycle" )) (mustMergeOverwrite (dict "name" "" "mountPath" "" ) (dict "name" "datadir" "mountPath" "/var/lib/redpanda/data" ))))) -}}
{{- if (gt ((get (fromJson (include "_shims.len" (dict "a" (list (get (fromJson (include "redpanda.Listeners.TrustStores" (dict "a" (list $values.listeners $values.tls) ))) "r")) ))) "r") | int) (0 | int)) -}}
{{- $mounts = (concat (default (list ) $mounts) (list (mustMergeOverwrite (dict "name" "" "mountPath" "" ) (dict "name" "truststores" "mountPath" "/etc/truststores" "readOnly" true )))) -}}
{{- end -}}
{{- (dict "r" $mounts) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.StatefulSetInitContainers" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $values := $dot.Values.AsMap -}}
{{- $containers := (coalesce nil) -}}
{{- $c_6 := (get (fromJson (include "redpanda.statefulSetInitContainerTuning" (dict "a" (list $dot) ))) "r") -}}
{{- if (ne $c_6 (coalesce nil)) -}}
{{- $containers = (concat (default (list ) $containers) (list $c_6)) -}}
{{- end -}}
{{- $c_7 := (get (fromJson (include "redpanda.statefulSetInitContainerSetDataDirOwnership" (dict "a" (list $dot) ))) "r") -}}
{{- if (ne $c_7 (coalesce nil)) -}}
{{- $containers = (concat (default (list ) $containers) (list $c_7)) -}}
{{- end -}}
{{- $c_8 := (get (fromJson (include "redpanda.statefulSetInitContainerFSValidator" (dict "a" (list $dot) ))) "r") -}}
{{- if (ne $c_8 (coalesce nil)) -}}
{{- $containers = (concat (default (list ) $containers) (list $c_8)) -}}
{{- end -}}
{{- $c_9 := (get (fromJson (include "redpanda.statefulSetInitContainerSetTieredStorageCacheDirOwnership" (dict "a" (list $dot) ))) "r") -}}
{{- if (ne $c_9 (coalesce nil)) -}}
{{- $containers = (concat (default (list ) $containers) (list $c_9)) -}}
{{- end -}}
{{- $containers = (concat (default (list ) $containers) (list (get (fromJson (include "redpanda.statefulSetInitContainerConfigurator" (dict "a" (list $dot) ))) "r"))) -}}
{{- $containers = (concat (default (list ) $containers) (default (list ) $values.statefulset.initContainers.extraInitContainers)) -}}
{{- (dict "r" $containers) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.statefulSetInitContainerTuning" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $values := $dot.Values.AsMap -}}
{{- if (not $values.tuning.tune_aio_events) -}}
{{- (dict "r" (coalesce nil)) | toJson -}}
{{- break -}}
{{- end -}}
{{- (dict "r" (mustMergeOverwrite (dict "name" "" "resources" (dict ) ) (dict "name" "tuning" "image" (printf "%s:%s" $values.image.repository (get (fromJson (include "redpanda.Tag" (dict "a" (list $dot) ))) "r")) "command" (list `/bin/bash` `-c` `rpk redpanda tune all`) "securityContext" (mustMergeOverwrite (dict ) (dict "capabilities" (mustMergeOverwrite (dict ) (dict "add" (list `SYS_RESOURCE`) )) "privileged" true "runAsUser" ((0 | int64) | int64) "runAsGroup" ((0 | int64) | int64) )) "volumeMounts" (concat (default (list ) (concat (default (list ) (get (fromJson (include "redpanda.CommonMounts" (dict "a" (list $dot) ))) "r")) (default (list ) $values.statefulset.initContainers.tuning.extraVolumeMounts))) (list (mustMergeOverwrite (dict "name" "" "mountPath" "" ) (dict "name" (get (fromJson (include "redpanda.Fullname" (dict "a" (list $dot) ))) "r") "mountPath" "/etc/redpanda" )))) "resources" $values.statefulset.initContainers.tuning.resources ))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.statefulSetInitContainerSetDataDirOwnership" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $values := $dot.Values.AsMap -}}
{{- if (not $values.statefulset.initContainers.setDataDirOwnership.enabled) -}}
{{- (dict "r" (coalesce nil)) | toJson -}}
{{- break -}}
{{- end -}}
{{- $tmp_tuple_3 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "redpanda.securityContextUidGid" (dict "a" (list $dot "set-datadir-ownership") ))) "r")) ))) "r") -}}
{{- $gid := ($tmp_tuple_3.T2 | int64) -}}
{{- $uid := ($tmp_tuple_3.T1 | int64) -}}
{{- (dict "r" (mustMergeOverwrite (dict "name" "" "resources" (dict ) ) (dict "name" "set-datadir-ownership" "image" (printf "%s:%s" $values.statefulset.initContainerImage.repository $values.statefulset.initContainerImage.tag) "command" (list `/bin/sh` `-c` (printf `chown %d:%d -R /var/lib/redpanda/data` $uid $gid)) "volumeMounts" (concat (default (list ) (concat (default (list ) (get (fromJson (include "redpanda.CommonMounts" (dict "a" (list $dot) ))) "r")) (default (list ) $values.statefulset.initContainers.setDataDirOwnership.extraVolumeMounts))) (list (mustMergeOverwrite (dict "name" "" "mountPath" "" ) (dict "name" `datadir` "mountPath" `/var/lib/redpanda/data` )))) "resources" $values.statefulset.initContainers.setDataDirOwnership.resources ))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.securityContextUidGid" -}}
{{- $dot := (index .a 0) -}}
{{- $containerName := (index .a 1) -}}
{{- range $_ := (list 1) -}}
{{- $values := $dot.Values.AsMap -}}
{{- $uid := $values.statefulset.securityContext.runAsUser -}}
{{- if (and (ne $values.statefulset.podSecurityContext (coalesce nil)) (ne $values.statefulset.podSecurityContext.runAsUser (coalesce nil))) -}}
{{- $uid = $values.statefulset.podSecurityContext.runAsUser -}}
{{- end -}}
{{- if (eq $uid (coalesce nil)) -}}
{{- $_ := (fail (printf `%s container requires runAsUser to be specified` $containerName)) -}}
{{- end -}}
{{- $gid := $values.statefulset.securityContext.fsGroup -}}
{{- if (and (ne $values.statefulset.podSecurityContext (coalesce nil)) (ne $values.statefulset.podSecurityContext.fsGroup (coalesce nil))) -}}
{{- $gid = $values.statefulset.podSecurityContext.fsGroup -}}
{{- end -}}
{{- if (eq $gid (coalesce nil)) -}}
{{- $_ := (fail (printf `%s container requires fsGroup to be specified` $containerName)) -}}
{{- end -}}
{{- (dict "r" (list $uid $gid)) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.statefulSetInitContainerFSValidator" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $values := $dot.Values.AsMap -}}
{{- if (not $values.statefulset.initContainers.fsValidator.enabled) -}}
{{- (dict "r" (coalesce nil)) | toJson -}}
{{- break -}}
{{- end -}}
{{- (dict "r" (mustMergeOverwrite (dict "name" "" "resources" (dict ) ) (dict "name" "fs-validator" "image" (printf "%s:%s" $values.image.repository (get (fromJson (include "redpanda.Tag" (dict "a" (list $dot) ))) "r")) "command" (list `/bin/sh`) "args" (list `-c` (printf `trap "exit 0" TERM; exec /etc/secrets/fs-validator/scripts/fsValidator.sh %s & wait $!` $values.statefulset.initContainers.fsValidator.expectedFS)) "securityContext" (get (fromJson (include "redpanda.ContainerSecurityContext" (dict "a" (list $dot) ))) "r") "volumeMounts" (concat (default (list ) (concat (default (list ) (get (fromJson (include "redpanda.CommonMounts" (dict "a" (list $dot) ))) "r")) (default (list ) $values.statefulset.initContainers.fsValidator.extraVolumeMounts))) (list (mustMergeOverwrite (dict "name" "" "mountPath" "" ) (dict "name" (printf `%.49s-fs-validator` (get (fromJson (include "redpanda.Fullname" (dict "a" (list $dot) ))) "r")) "mountPath" `/etc/secrets/fs-validator/scripts/` )) (mustMergeOverwrite (dict "name" "" "mountPath" "" ) (dict "name" `datadir` "mountPath" `/var/lib/redpanda/data` )))) "resources" $values.statefulset.initContainers.fsValidator.resources ))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.statefulSetInitContainerSetTieredStorageCacheDirOwnership" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $values := $dot.Values.AsMap -}}
{{- if (not (get (fromJson (include "redpanda.Storage.IsTieredStorageEnabled" (dict "a" (list $values.storage) ))) "r")) -}}
{{- (dict "r" (coalesce nil)) | toJson -}}
{{- break -}}
{{- end -}}
{{- $tmp_tuple_4 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "redpanda.securityContextUidGid" (dict "a" (list $dot "set-tiered-storage-cache-dir-ownership") ))) "r")) ))) "r") -}}
{{- $gid := ($tmp_tuple_4.T2 | int64) -}}
{{- $uid := ($tmp_tuple_4.T1 | int64) -}}
{{- $cacheDir := (get (fromJson (include "redpanda.storageTieredCacheDirectory" (dict "a" (list $dot) ))) "r") -}}
{{- $mounts := (get (fromJson (include "redpanda.CommonMounts" (dict "a" (list $dot) ))) "r") -}}
{{- $mounts = (concat (default (list ) $mounts) (list (mustMergeOverwrite (dict "name" "" "mountPath" "" ) (dict "name" "datadir" "mountPath" "/var/lib/redpanda/data" )))) -}}
{{- if (ne (get (fromJson (include "redpanda.storageTieredMountType" (dict "a" (list $dot) ))) "r") "none") -}}
{{- $name := "tiered-storage-dir" -}}
{{- if (and (ne $values.storage.persistentVolume (coalesce nil)) (ne $values.storage.persistentVolume.nameOverwrite "")) -}}
{{- $name = $values.storage.persistentVolume.nameOverwrite -}}
{{- end -}}
{{- $mounts = (concat (default (list ) $mounts) (list (mustMergeOverwrite (dict "name" "" "mountPath" "" ) (dict "name" $name "mountPath" $cacheDir )))) -}}
{{- end -}}
{{- $mounts = (concat (default (list ) $mounts) (default (list ) $values.statefulset.initContainers.setTieredStorageCacheDirOwnership.extraVolumeMounts)) -}}
{{- (dict "r" (mustMergeOverwrite (dict "name" "" "resources" (dict ) ) (dict "name" `set-tiered-storage-cache-dir-ownership` "image" (printf `%s:%s` $values.statefulset.initContainerImage.repository $values.statefulset.initContainerImage.tag) "command" (list `/bin/sh` `-c` (printf `mkdir -p %s; chown %d:%d -R %s` $cacheDir $uid $gid $cacheDir)) "volumeMounts" $mounts "resources" $values.statefulset.initContainers.setTieredStorageCacheDirOwnership.resources ))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.storageTieredCacheDirectory" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $values := $dot.Values.AsMap -}}
{{- $config := (get (fromJson (include "redpanda.Storage.GetTieredStorageConfig" (dict "a" (list $values.storage) ))) "r") -}}
{{- $dir := (get (fromJson (include "_shims.typeassertion" (dict "a" (list "string" (dig `cloud_storage_cache_directory` "/var/lib/redpanda/data/cloud_storage_cache" $config)) ))) "r") -}}
{{- if (eq $dir "") -}}
{{- (dict "r" "/var/lib/redpanda/data/cloud_storage_cache") | toJson -}}
{{- break -}}
{{- end -}}
{{- (dict "r" $dir) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.storageTieredMountType" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $values := $dot.Values.AsMap -}}
{{- if (and (ne $values.storage.tieredStoragePersistentVolume (coalesce nil)) $values.storage.tieredStoragePersistentVolume.enabled) -}}
{{- (dict "r" "persistentVolume") | toJson -}}
{{- break -}}
{{- end -}}
{{- if (ne $values.storage.tieredStorageHostPath "") -}}
{{- (dict "r" "hostPath") | toJson -}}
{{- break -}}
{{- end -}}
{{- (dict "r" $values.storage.tiered.mountType) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.statefulSetInitContainerConfigurator" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $values := $dot.Values.AsMap -}}
{{- (dict "r" (mustMergeOverwrite (dict "name" "" "resources" (dict ) ) (dict "name" (printf `%.51s-configurator` (get (fromJson (include "redpanda.Name" (dict "a" (list $dot) ))) "r")) "image" (printf `%s:%s` $values.image.repository (get (fromJson (include "redpanda.Tag" (dict "a" (list $dot) ))) "r")) "command" (list `/bin/bash` `-c` `trap "exit 0" TERM; exec $CONFIGURATOR_SCRIPT "${SERVICE_NAME}" "${KUBERNETES_NODE_NAME}" & wait $!`) "env" (list (mustMergeOverwrite (dict "name" "" ) (dict "name" "CONFIGURATOR_SCRIPT" "value" "/etc/secrets/configurator/scripts/configurator.sh" )) (mustMergeOverwrite (dict "name" "" ) (dict "name" "SERVICE_NAME" "valueFrom" (mustMergeOverwrite (dict ) (dict "fieldRef" (mustMergeOverwrite (dict "fieldPath" "" ) (dict "fieldPath" "metadata.name" )) "resourceFieldRef" (coalesce nil) "configMapKeyRef" (coalesce nil) "secretKeyRef" (coalesce nil) )) )) (mustMergeOverwrite (dict "name" "" ) (dict "name" "KUBERNETES_NODE_NAME" "valueFrom" (mustMergeOverwrite (dict ) (dict "fieldRef" (mustMergeOverwrite (dict "fieldPath" "" ) (dict "fieldPath" "spec.nodeName" )) )) )) (mustMergeOverwrite (dict "name" "" ) (dict "name" "HOST_IP_ADDRESS" "valueFrom" (mustMergeOverwrite (dict ) (dict "fieldRef" (mustMergeOverwrite (dict "fieldPath" "" ) (dict "apiVersion" "v1" "fieldPath" "status.hostIP" )) )) ))) "securityContext" (get (fromJson (include "redpanda.ContainerSecurityContext" (dict "a" (list $dot) ))) "r") "volumeMounts" (concat (default (list ) (concat (default (list ) (get (fromJson (include "redpanda.CommonMounts" (dict "a" (list $dot) ))) "r")) (default (list ) $values.statefulset.initContainers.configurator.extraVolumeMounts))) (list (mustMergeOverwrite (dict "name" "" "mountPath" "" ) (dict "name" "config" "mountPath" "/etc/redpanda" )) (mustMergeOverwrite (dict "name" "" "mountPath" "" ) (dict "name" (get (fromJson (include "redpanda.Fullname" (dict "a" (list $dot) ))) "r") "mountPath" "/tmp/base-config" )) (mustMergeOverwrite (dict "name" "" "mountPath" "" ) (dict "name" (printf `%.51s-configurator` (get (fromJson (include "redpanda.Fullname" (dict "a" (list $dot) ))) "r")) "mountPath" "/etc/secrets/configurator/scripts/" )))) "resources" $values.statefulset.initContainers.configurator.resources ))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.StatefulSetContainers" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $containers := (coalesce nil) -}}
{{- $containers = (concat (default (list ) $containers) (list (get (fromJson (include "redpanda.statefulSetContainerRedpanda" (dict "a" (list $dot) ))) "r"))) -}}
{{- $c_10 := (get (fromJson (include "redpanda.statefulSetContainerConfigWatcher" (dict "a" (list $dot) ))) "r") -}}
{{- if (ne $c_10 (coalesce nil)) -}}
{{- $containers = (concat (default (list ) $containers) (list $c_10)) -}}
{{- end -}}
{{- $c_11 := (get (fromJson (include "redpanda.statefulSetContainerControllers" (dict "a" (list $dot) ))) "r") -}}
{{- if (ne $c_11 (coalesce nil)) -}}
{{- $containers = (concat (default (list ) $containers) (list $c_11)) -}}
{{- end -}}
{{- (dict "r" $containers) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.statefulSetContainerRedpanda" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $values := $dot.Values.AsMap -}}
{{- $internalAdvertiseAddress := (printf "%s.%s" "$(SERVICE_NAME)" (get (fromJson (include "redpanda.InternalDomain" (dict "a" (list $dot) ))) "r")) -}}
{{- $container := (mustMergeOverwrite (dict "name" "" "resources" (dict ) ) (dict "name" (get (fromJson (include "redpanda.Name" (dict "a" (list $dot) ))) "r") "image" (printf `%s:%s` $values.image.repository (get (fromJson (include "redpanda.Tag" (dict "a" (list $dot) ))) "r")) "env" (get (fromJson (include "redpanda.StatefulSetRedpandaEnv" (dict "a" (list $dot) ))) "r") "lifecycle" (mustMergeOverwrite (dict ) (dict "postStart" (mustMergeOverwrite (dict ) (dict "exec" (mustMergeOverwrite (dict ) (dict "command" (list `/bin/bash` `-c` (join "\n" (list (printf `timeout -v %d bash -x /var/lifecycle/postStart.sh` ((div ($values.statefulset.terminationGracePeriodSeconds | int) (2 | int)) | int)) `true` ``))) )) )) "preStop" (mustMergeOverwrite (dict ) (dict "exec" (mustMergeOverwrite (dict ) (dict "command" (list `/bin/bash` `-c` (join "\n" (list (printf `timeout -v %d bash -x /var/lifecycle/preStop.sh` ((div ($values.statefulset.terminationGracePeriodSeconds | int) (2 | int)) | int)) `true # do not fail and cause the pod to terminate` ``))) )) )) )) "startupProbe" (mustMergeOverwrite (dict ) (mustMergeOverwrite (dict ) (dict "exec" (mustMergeOverwrite (dict ) (dict "command" (list `/bin/sh` `-c` (join "\n" (list `set -e` (printf `RESULT=$(curl --silent --fail -k -m 5 %s "%s://%s/v1/status/ready")` (get (fromJson (include "redpanda.adminTLSCurlFlags" (dict "a" (list $dot) ))) "r") (get (fromJson (include "redpanda.adminInternalHTTPProtocol" (dict "a" (list $dot) ))) "r") (get (fromJson (include "redpanda.adminApiURLs" (dict "a" (list $dot) ))) "r")) `echo $RESULT` `echo $RESULT | grep ready` ``))) )) )) (dict "initialDelaySeconds" ($values.statefulset.startupProbe.initialDelaySeconds | int) "periodSeconds" ($values.statefulset.startupProbe.periodSeconds | int) "failureThreshold" ($values.statefulset.startupProbe.failureThreshold | int) )) "livenessProbe" (mustMergeOverwrite (dict ) (mustMergeOverwrite (dict ) (dict "exec" (mustMergeOverwrite (dict ) (dict "command" (list `/bin/sh` `-c` (printf `curl --silent --fail -k -m 5 %s "%s://%s/v1/status/ready"` (get (fromJson (include "redpanda.adminTLSCurlFlags" (dict "a" (list $dot) ))) "r") (get (fromJson (include "redpanda.adminInternalHTTPProtocol" (dict "a" (list $dot) ))) "r") (get (fromJson (include "redpanda.adminApiURLs" (dict "a" (list $dot) ))) "r"))) )) )) (dict "initialDelaySeconds" ($values.statefulset.livenessProbe.initialDelaySeconds | int) "periodSeconds" ($values.statefulset.livenessProbe.periodSeconds | int) "failureThreshold" ($values.statefulset.livenessProbe.failureThreshold | int) )) "command" (list `rpk` `redpanda` `start` (printf `--advertise-rpc-addr=%s:%d` $internalAdvertiseAddress ($values.listeners.rpc.port | int))) "volumeMounts" (concat (default (list ) (get (fromJson (include "redpanda.StatefulSetVolumeMounts" (dict "a" (list $dot) ))) "r")) (default (list ) $values.statefulset.extraVolumeMounts)) "securityContext" (get (fromJson (include "redpanda.ContainerSecurityContext" (dict "a" (list $dot) ))) "r") "resources" (mustMergeOverwrite (dict ) (dict )) )) -}}
{{- if (not (get (fromJson (include "_shims.typeassertion" (dict "a" (list "bool" (dig `recovery_mode_enabled` false $values.config.node)) ))) "r")) -}}
{{- $_ := (set $container "readinessProbe" (mustMergeOverwrite (dict ) (mustMergeOverwrite (dict ) (dict "exec" (mustMergeOverwrite (dict ) (dict "command" (list `/bin/sh` `-c` (join "\n" (list `set -x` `RESULT=$(rpk cluster health)` `echo $RESULT` `echo $RESULT | grep 'Healthy:.*true'` ``))) )) )) (dict "initialDelaySeconds" ($values.statefulset.readinessProbe.initialDelaySeconds | int) "timeoutSeconds" ($values.statefulset.readinessProbe.timeoutSeconds | int) "periodSeconds" ($values.statefulset.readinessProbe.periodSeconds | int) "successThreshold" ($values.statefulset.readinessProbe.successThreshold | int) "failureThreshold" ($values.statefulset.readinessProbe.failureThreshold | int) ))) -}}
{{- end -}}
{{- $_ := (set $container "ports" (concat (default (list ) $container.ports) (list (mustMergeOverwrite (dict "containerPort" 0 ) (dict "name" "admin" "containerPort" ($values.listeners.admin.port | int) ))))) -}}
{{- range $externalName, $external := $values.listeners.admin.external -}}
{{- if (get (fromJson (include "redpanda.AdminExternal.IsEnabled" (dict "a" (list $external) ))) "r") -}}
{{- $_ := (set $container "ports" (concat (default (list ) $container.ports) (list (mustMergeOverwrite (dict "containerPort" 0 ) (dict "name" (printf "admin-%.8s" (lower $externalName)) "containerPort" ($external.port | int) ))))) -}}
{{- end -}}
{{- end -}}
{{- $_ := (set $container "ports" (concat (default (list ) $container.ports) (list (mustMergeOverwrite (dict "containerPort" 0 ) (dict "name" "http" "containerPort" ($values.listeners.http.port | int) ))))) -}}
{{- range $externalName, $external := $values.listeners.http.external -}}
{{- if (get (fromJson (include "redpanda.HTTPExternal.IsEnabled" (dict "a" (list $external) ))) "r") -}}
{{- $_ := (set $container "ports" (concat (default (list ) $container.ports) (list (mustMergeOverwrite (dict "containerPort" 0 ) (dict "name" (printf "http-%.8s" (lower $externalName)) "containerPort" ($external.port | int) ))))) -}}
{{- end -}}
{{- end -}}
{{- $_ := (set $container "ports" (concat (default (list ) $container.ports) (list (mustMergeOverwrite (dict "containerPort" 0 ) (dict "name" "kafka" "containerPort" ($values.listeners.kafka.port | int) ))))) -}}
{{- range $externalName, $external := $values.listeners.kafka.external -}}
{{- if (get (fromJson (include "redpanda.KafkaExternal.IsEnabled" (dict "a" (list $external) ))) "r") -}}
{{- $_ := (set $container "ports" (concat (default (list ) $container.ports) (list (mustMergeOverwrite (dict "containerPort" 0 ) (dict "name" (printf "kafka-%.8s" (lower $externalName)) "containerPort" ($external.port | int) ))))) -}}
{{- end -}}
{{- end -}}
{{- $_ := (set $container "ports" (concat (default (list ) $container.ports) (list (mustMergeOverwrite (dict "containerPort" 0 ) (dict "name" "rpc" "containerPort" ($values.listeners.rpc.port | int) ))))) -}}
{{- $_ := (set $container "ports" (concat (default (list ) $container.ports) (list (mustMergeOverwrite (dict "containerPort" 0 ) (dict "name" "schemaregistry" "containerPort" ($values.listeners.schemaRegistry.port | int) ))))) -}}
{{- range $externalName, $external := $values.listeners.schemaRegistry.external -}}
{{- if (get (fromJson (include "redpanda.SchemaRegistryExternal.IsEnabled" (dict "a" (list $external) ))) "r") -}}
{{- $_ := (set $container "ports" (concat (default (list ) $container.ports) (list (mustMergeOverwrite (dict "containerPort" 0 ) (dict "name" (printf "schema-%.8s" (lower $externalName)) "containerPort" ($external.port | int) ))))) -}}
{{- end -}}
{{- end -}}
{{- if (and (get (fromJson (include "redpanda.Storage.IsTieredStorageEnabled" (dict "a" (list $values.storage) ))) "r") (ne (get (fromJson (include "redpanda.storageTieredMountType" (dict "a" (list $dot) ))) "r") "none")) -}}
{{- $name := "tiered-storage-dir" -}}
{{- if (and (ne $values.storage.persistentVolume (coalesce nil)) (ne $values.storage.persistentVolume.nameOverwrite "")) -}}
{{- $name = $values.storage.persistentVolume.nameOverwrite -}}
{{- end -}}
{{- $_ := (set $container "volumeMounts" (concat (default (list ) $container.volumeMounts) (list (mustMergeOverwrite (dict "name" "" "mountPath" "" ) (dict "name" $name "mountPath" (get (fromJson (include "redpanda.storageTieredCacheDirectory" (dict "a" (list $dot) ))) "r") ))))) -}}
{{- end -}}
{{- $_ := (set $container.resources "limits" (dict "cpu" $values.resources.cpu.cores "memory" $values.resources.memory.container.max )) -}}
{{- if (ne $values.resources.memory.container.min (coalesce nil)) -}}
{{- $_ := (set $container.resources "requests" (dict "cpu" $values.resources.cpu.cores "memory" $values.resources.memory.container.min )) -}}
{{- end -}}
{{- (dict "r" $container) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.adminApiURLs" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $values := $dot.Values.AsMap -}}
{{- (dict "r" (printf `${SERVICE_NAME}.%s:%d` (get (fromJson (include "redpanda.InternalDomain" (dict "a" (list $dot) ))) "r") ($values.listeners.admin.port | int))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.statefulSetContainerConfigWatcher" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $values := $dot.Values.AsMap -}}
{{- if (not $values.statefulset.sideCars.configWatcher.enabled) -}}
{{- (dict "r" (coalesce nil)) | toJson -}}
{{- break -}}
{{- end -}}
{{- (dict "r" (mustMergeOverwrite (dict "name" "" "resources" (dict ) ) (dict "name" "config-watcher" "image" (printf `%s:%s` $values.image.repository (get (fromJson (include "redpanda.Tag" (dict "a" (list $dot) ))) "r")) "command" (list `/bin/sh`) "args" (list `-c` `trap "exit 0" TERM; exec /etc/secrets/config-watcher/scripts/sasl-user.sh & wait $!`) "resources" $values.statefulset.sideCars.configWatcher.resources "securityContext" $values.statefulset.sideCars.configWatcher.securityContext "volumeMounts" (concat (default (list ) (concat (default (list ) (get (fromJson (include "redpanda.CommonMounts" (dict "a" (list $dot) ))) "r")) (list (mustMergeOverwrite (dict "name" "" "mountPath" "" ) (dict "name" "config" "mountPath" "/etc/redpanda" )) (mustMergeOverwrite (dict "name" "" "mountPath" "" ) (dict "name" (printf `%s-config-watcher` (get (fromJson (include "redpanda.Fullname" (dict "a" (list $dot) ))) "r")) "mountPath" "/etc/secrets/config-watcher/scripts" ))))) (default (list ) $values.statefulset.sideCars.configWatcher.extraVolumeMounts)) ))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.statefulSetContainerControllers" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $values := $dot.Values.AsMap -}}
{{- if (or (not $values.rbac.enabled) (not $values.statefulset.sideCars.controllers.enabled)) -}}
{{- (dict "r" (coalesce nil)) | toJson -}}
{{- break -}}
{{- end -}}
{{- (dict "r" (mustMergeOverwrite (dict "name" "" "resources" (dict ) ) (dict "name" "redpanda-controllers" "image" (printf `%s:%s` $values.statefulset.sideCars.controllers.image.repository $values.statefulset.sideCars.controllers.image.tag) "command" (list `/manager`) "args" (list `--operator-mode=false` (printf `--namespace=%s` $dot.Release.Namespace) (printf `--health-probe-bind-address=%s` $values.statefulset.sideCars.controllers.healthProbeAddress) (printf `--metrics-bind-address=%s` $values.statefulset.sideCars.controllers.metricsAddress) (printf `--additional-controllers=%s` (join "," $values.statefulset.sideCars.controllers.run))) "env" (list (mustMergeOverwrite (dict "name" "" ) (dict "name" "REDPANDA_HELM_RELEASE_NAME" "value" $dot.Release.Name ))) "resources" $values.statefulset.sideCars.controllers.resources "securityContext" $values.statefulset.sideCars.controllers.securityContext ))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

