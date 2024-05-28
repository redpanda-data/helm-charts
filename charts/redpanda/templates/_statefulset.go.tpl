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

