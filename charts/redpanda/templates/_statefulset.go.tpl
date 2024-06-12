{{- /* Generated from "statefulset.go" */ -}}

{{- define "redpanda.StatefulSetRedpandaEnv" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $values := $dot.Values.AsMap -}}
{{- $userEnv := (list ) -}}
{{- range $_, $container := $values.statefulset.podTemplate.spec.containers -}}
{{- if (eq $container.name "redpanda") -}}
{{- $userEnv = $container.env -}}
{{- end -}}
{{- end -}}
{{- (dict "r" (concat (list (mustMergeOverwrite (dict "name" "" ) (dict "name" "SERVICE_NAME" "valueFrom" (mustMergeOverwrite (dict ) (dict "fieldRef" (mustMergeOverwrite (dict "fieldPath" "" ) (dict "fieldPath" "metadata.name" )) )) )) (mustMergeOverwrite (dict "name" "" ) (dict "name" "POD_IP" "valueFrom" (mustMergeOverwrite (dict ) (dict "fieldRef" (mustMergeOverwrite (dict "fieldPath" "" ) (dict "fieldPath" "status.podIP" )) )) )) (mustMergeOverwrite (dict "name" "" ) (dict "name" "HOST_IP" "valueFrom" (mustMergeOverwrite (dict ) (dict "fieldRef" (mustMergeOverwrite (dict "fieldPath" "" ) (dict "fieldPath" "status.hostIP" )) )) ))) $userEnv)) | toJson -}}
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
{{- $volumes := (get (fromJson (include "redpanda.CommonVolumes" (dict "a" (list $dot) ))) "r") -}}
{{- $fullname := (get (fromJson (include "redpanda.Fullname" (dict "a" (list $dot) ))) "r") -}}
{{- $volumes = (concat $volumes (list (mustMergeOverwrite (dict "name" "" ) (mustMergeOverwrite (dict ) (dict "secret" (mustMergeOverwrite (dict ) (dict "secretName" (printf "%.50s-sts-lifecycle" $fullname) "defaultMode" (0o775 | int) )) )) (dict "name" "lifecycle-scripts" )) (mustMergeOverwrite (dict "name" "" ) (mustMergeOverwrite (dict ) (dict "configMap" (mustMergeOverwrite (dict ) (mustMergeOverwrite (dict ) (dict "name" $fullname )) (dict )) )) (dict "name" $fullname )) (mustMergeOverwrite (dict "name" "" ) (mustMergeOverwrite (dict ) (dict "emptyDir" (mustMergeOverwrite (dict ) (dict )) )) (dict "name" "config" )) (mustMergeOverwrite (dict "name" "" ) (mustMergeOverwrite (dict ) (dict "secret" (mustMergeOverwrite (dict ) (dict "secretName" (printf "%.51s-configurator" $fullname) "defaultMode" (0o775 | int) )) )) (dict "name" (printf "%.51s-configurator" $fullname) )) (mustMergeOverwrite (dict "name" "" ) (mustMergeOverwrite (dict ) (dict "secret" (mustMergeOverwrite (dict ) (dict "secretName" (printf "%s-config-watcher" $fullname) "defaultMode" (0o775 | int) )) )) (dict "name" (printf "%s-config-watcher" $fullname) )) (mustMergeOverwrite (dict "name" "" ) (mustMergeOverwrite (dict ) (dict "secret" (mustMergeOverwrite (dict ) (dict "secretName" (printf "%.49s-fs-validator" $fullname) "defaultMode" (0o775 | int) )) )) (dict "name" (printf "%.49s-fs-validator" $fullname) )))) -}}
{{- (dict "r" $volumes) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.StatefulSetVolumeMounts" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $mounts := (get (fromJson (include "redpanda.CommonMounts" (dict "a" (list $dot) ))) "r") -}}
{{- $mounts = (concat $mounts (list (mustMergeOverwrite (dict "name" "" "mountPath" "" ) (dict "name" "config" "mountPath" "/etc/redpanda" )) (mustMergeOverwrite (dict "name" "" "mountPath" "" ) (dict "name" (get (fromJson (include "redpanda.Fullname" (dict "a" (list $dot) ))) "r") "mountPath" "/tmp/base-config" )) (mustMergeOverwrite (dict "name" "" "mountPath" "" ) (dict "name" "lifecycle-scripts" "mountPath" "/var/lifecycle" )) (mustMergeOverwrite (dict "name" "" "mountPath" "" ) (dict "name" "datadir" "mountPath" "/var/lib/redpanda/data" )))) -}}
{{- (dict "r" $mounts) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

