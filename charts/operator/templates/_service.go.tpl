{{- /* Generated from "service.go" */ -}}

{{- define "operator.WebhookService" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $values := $dot.Values.AsMap -}}
{{- if (not ((and $values.webhook.enabled (eq $values.scope "Cluster")))) -}}
{{- $_is_returning = true -}}
{{- (dict "r" (coalesce nil)) | toJson -}}
{{- break -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" (mustMergeOverwrite (dict "metadata" (dict "creationTimestamp" (coalesce nil) ) "spec" (dict ) "status" (dict "loadBalancer" (dict ) ) ) (mustMergeOverwrite (dict ) (dict "apiVersion" "v1" "kind" "Service" )) (dict "metadata" (mustMergeOverwrite (dict "creationTimestamp" (coalesce nil) ) (dict "name" (printf "%s-webhook-service" (get (fromJson (include "operator.RedpandaOperatorName" (dict "a" (list $dot) ))) "r")) "namespace" $dot.Release.Namespace "labels" (get (fromJson (include "operator.Labels" (dict "a" (list $dot) ))) "r") "annotations" $values.annotations )) "spec" (mustMergeOverwrite (dict ) (dict "selector" (get (fromJson (include "operator.SelectorLabels" (dict "a" (list $dot) ))) "r") "ports" (list (mustMergeOverwrite (dict "port" 0 "targetPort" 0 ) (dict "port" ((443 | int) | int) "targetPort" (9443 | int) ))) )) ))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "operator.RedpandaOperatorName" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $values := $dot.Values.AsMap -}}
{{- if (ne $values.nameOverride "") -}}
{{- $_is_returning = true -}}
{{- (dict "r" (get (fromJson (include "operator.cleanForK8s" (dict "a" (list $values.nameOverride) ))) "r")) | toJson -}}
{{- break -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" (get (fromJson (include "operator.cleanForK8s" (dict "a" (list $dot.Chart.Name) ))) "r")) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "operator.MetricsService" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $values := $dot.Values.AsMap -}}
{{- $_is_returning = true -}}
{{- (dict "r" (mustMergeOverwrite (dict "metadata" (dict "creationTimestamp" (coalesce nil) ) "spec" (dict ) "status" (dict "loadBalancer" (dict ) ) ) (mustMergeOverwrite (dict ) (dict "apiVersion" "v1" "kind" "Service" )) (dict "metadata" (mustMergeOverwrite (dict "creationTimestamp" (coalesce nil) ) (dict "name" (get (fromJson (include "operator.cleanForK8sWithSuffix" (dict "a" (list (get (fromJson (include "operator.Fullname" (dict "a" (list $dot) ))) "r") "metrics-service") ))) "r") "namespace" $dot.Release.Namespace "labels" (get (fromJson (include "operator.Labels" (dict "a" (list $dot) ))) "r") "annotations" $values.annotations )) "spec" (mustMergeOverwrite (dict ) (dict "selector" (get (fromJson (include "operator.SelectorLabels" (dict "a" (list $dot) ))) "r") "ports" (list (mustMergeOverwrite (dict "port" 0 "targetPort" 0 ) (dict "name" "https" "port" ((8443 | int) | int) "targetPort" "https" ))) )) ))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

