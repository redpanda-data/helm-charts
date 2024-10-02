{{- /* Generated from "configmap.go" */ -}}

{{- define "operator.ConfigMap" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $values := $dot.Values.AsMap -}}
{{- $_is_returning = true -}}
{{- (dict "r" (mustMergeOverwrite (dict "metadata" (dict "creationTimestamp" (coalesce nil) ) ) (mustMergeOverwrite (dict ) (dict "apiVersion" "v1" "kind" "ConfigMap" )) (dict "metadata" (mustMergeOverwrite (dict "creationTimestamp" (coalesce nil) ) (dict "name" (get (fromJson (include "operator.cleanForK8sWithSuffix" (dict "a" (list (get (fromJson (include "operator.Fullname" (dict "a" (list $dot) ))) "r") "config") ))) "r") "namespace" $dot.Release.Namespace "labels" (get (fromJson (include "operator.Labels" (dict "a" (list $dot) ))) "r") "annotations" $values.annotations )) "data" (dict "controller_manager_config.yaml" (toYaml $values.config) ) ))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

