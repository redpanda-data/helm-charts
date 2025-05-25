{{- /* Generated from "servicemonitor.go" */ -}}

{{- define "operator.ServiceMonitor" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $values := $dot.Values.AsMap -}}
{{- if (not $values.monitoring.enabled) -}}
{{- $_is_returning = true -}}
{{- (dict "r" (coalesce nil)) | toJson -}}
{{- break -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" (mustMergeOverwrite (dict "metadata" (dict "creationTimestamp" (coalesce nil)) "spec" (dict "endpoints" (coalesce nil) "selector" (dict) "namespaceSelector" (dict))) (mustMergeOverwrite (dict) (dict "kind" "ServiceMonitor" "apiVersion" "monitoring.coreos.com/v1")) (dict "metadata" (mustMergeOverwrite (dict "creationTimestamp" (coalesce nil)) (dict "name" (get (fromJson (include "operator.cleanForK8sWithSuffix" (dict "a" (list (get (fromJson (include "operator.Fullname" (dict "a" (list $dot)))) "r") "metrics-monitor")))) "r") "labels" (get (fromJson (include "operator.Labels" (dict "a" (list $dot)))) "r") "namespace" $dot.Release.Namespace "annotations" $values.annotations)) "spec" (mustMergeOverwrite (dict "endpoints" (coalesce nil) "selector" (dict) "namespaceSelector" (dict)) (dict "endpoints" (list (mustMergeOverwrite (dict) (dict "port" "https" "path" "/metrics" "scheme" "https" "tlsConfig" (mustMergeOverwrite (dict "ca" (dict) "cert" (dict)) (mustMergeOverwrite (dict "ca" (dict) "cert" (dict)) (dict "insecureSkipVerify" true)) (dict)) "bearerTokenFile" "/var/run/secrets/kubernetes.io/serviceaccount/token"))) "namespaceSelector" (mustMergeOverwrite (dict) (dict "matchNames" (list $dot.Release.Namespace))) "selector" (mustMergeOverwrite (dict) (dict "matchLabels" (get (fromJson (include "operator.Labels" (dict "a" (list $dot)))) "r")))))))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

