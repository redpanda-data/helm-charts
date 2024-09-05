{{- /* Generated from "issuer.go" */ -}}

{{- define "operator.Issuer" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $values := $dot.Values.AsMap -}}
{{- $_is_returning = true -}}
{{- (dict "r" (mustMergeOverwrite (dict "metadata" (dict "creationTimestamp" (coalesce nil) ) "spec" (dict ) "status" (dict ) ) (mustMergeOverwrite (dict ) (dict "apiVersion" "cert-manager.io/v1" "kind" "Issuer" )) (dict "metadata" (mustMergeOverwrite (dict "creationTimestamp" (coalesce nil) ) (dict "name" (printf "%s-selfsigned-issuer" (trunc ((get (fromJson (include "_shims.len" (dict "a" (list "-selfsigned-issuer") ))) "r") | int) (get (fromJson (include "operator.Fullname" (dict "a" (list $dot) ))) "r"))) "namespace" $dot.Release.Namespace "labels" (get (fromJson (include "operator.Labels" (dict "a" (list $dot) ))) "r") "annotations" $values.annotations )) "spec" (mustMergeOverwrite (dict ) (mustMergeOverwrite (dict ) (dict "selfSigned" (mustMergeOverwrite (dict ) (dict )) )) (dict )) ))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

