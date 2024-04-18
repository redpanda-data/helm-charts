{{- /* Generated from "servicemonitor.go" */ -}}

{{- define "redpanda.ServiceMonitor" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $values := $dot.Values.AsMap -}}
{{- $monitoringEnabled := false -}}
{{- if (ne $values.monitoring.enabled (coalesce nil)) -}}
{{- $monitoringEnabled = $values.monitoring.enabled -}}
{{- end -}}
{{- if (not $monitoringEnabled) -}}
{{- (dict "r" (coalesce nil)) | toJson -}}
{{- break -}}
{{- end -}}
{{- $monitorLabels := (merge (dict ) (get (fromJson (include "redpanda.FullLabels" (dict "a" (list $dot) ))) "r") $values.monitoring.labels) -}}
{{- $matchLabels := (dict "monitoring.redpanda.com/enabled" "true" "app.kubernetes.io/name" (get (fromJson (include "redpanda.Name" (dict "a" (list $dot) ))) "r") "app.kubernetes.io/instance" $dot.Release.Name ) -}}
{{- $tlsConfig := $values.monitoring.tlsConfig -}}
{{- if (ne $tlsConfig (coalesce nil)) -}}
{{- $_ := (set $tlsConfig.SafeTLSConfig "insecureSkipVerify" false) -}}
{{- else -}}
{{- $tlsConfig = (mustMergeOverwrite (mustMergeOverwrite (dict "ca" (dict ) "cert" (dict ) ) (dict )) (mustMergeOverwrite (dict "ca" (dict ) "cert" (dict ) ) (dict "insecureSkipVerify" true )) (dict )) -}}
{{- end -}}
{{- (dict "r" (mustMergeOverwrite (mustMergeOverwrite (dict ) (dict "metadata" (dict "creationTimestamp" (coalesce nil) ) "spec" (dict "endpoints" (coalesce nil) "selector" (dict ) "namespaceSelector" (dict ) ) )) (dict "metadata" (mustMergeOverwrite (dict "creationTimestamp" (coalesce nil) ) (dict "name" (get (fromJson (include "redpanda.Fullname" (dict "a" (list $dot) ))) "r") "namespace" $dot.Release.Namespace "labels" $monitorLabels )) "spec" (mustMergeOverwrite (dict "endpoints" (coalesce nil) "selector" (dict ) "namespaceSelector" (dict ) ) (dict "endpoints" (list (mustMergeOverwrite (dict ) (dict "interval" $values.monitoring.scrapeInterval "path" "/public_metrics" "port" "admin" "enableHttp2" $values.monitoring.enableHttp2 "scheme" "https" "tlsConfig" $tlsConfig ))) "selector" (mustMergeOverwrite (dict ) (dict "matchLabels" $matchLabels )) )) ))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

