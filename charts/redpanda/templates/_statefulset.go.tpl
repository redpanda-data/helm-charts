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

