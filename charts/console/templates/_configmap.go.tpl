{{- /* Generated from "configmap.go" */ -}}

{{- define "console.ConfigMap" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $values := $dot.Values.AsMap -}}
{{- if (not $values.configmap.create) -}}
{{- $_is_returning = true -}}
{{- (dict "r" (coalesce nil)) | toJson -}}
{{- break -}}
{{- end -}}
{{- $data := (dict "config.yaml" (printf "# from .Values.config\n%s\n" (tpl (toYaml $values.config) $dot))) -}}
{{- $_is_returning = true -}}
{{- (dict "r" (mustMergeOverwrite (dict "metadata" (dict "creationTimestamp" (coalesce nil))) (mustMergeOverwrite (dict) (dict "apiVersion" "v1" "kind" "ConfigMap")) (dict "metadata" (mustMergeOverwrite (dict "creationTimestamp" (coalesce nil)) (dict "labels" (get (fromJson (include "console.Labels" (dict "a" (list $dot)))) "r") "name" (get (fromJson (include "console.Fullname" (dict "a" (list $dot)))) "r") "namespace" $dot.Release.Namespace)) "data" $data))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

