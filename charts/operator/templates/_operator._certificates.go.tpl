{{- /* Generated from "certificates.go" */ -}}

{{- define "operator.Certificate" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $values := $dot.Values.AsMap -}}
{{- if (not $values.webhook.enabled) -}}
{{- $_is_returning = true -}}
{{- (dict "r" (coalesce nil)) | toJson -}}
{{- break -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" (mustMergeOverwrite (dict "metadata" (dict "creationTimestamp" (coalesce nil) ) "spec" (dict "secretName" "" "issuerRef" (dict "name" "" ) ) "status" (dict ) ) (mustMergeOverwrite (dict ) (dict "apiVersion" "cert-manager.io/v1" "kind" "Certificate" )) (dict "metadata" (mustMergeOverwrite (dict "creationTimestamp" (coalesce nil) ) (dict "name" "redpanda-serving-cert" "namespace" $dot.Release.Namespace "labels" (get (fromJson (include "operator.Labels" (dict "a" (list $dot) ))) "r") "annotations" $values.annotations )) "spec" (mustMergeOverwrite (dict "secretName" "" "issuerRef" (dict "name" "" ) ) (dict "dnsNames" (list (printf "%s-webhook-service.%s.svc" (get (fromJson (include "operator.RedpandaOperatorName" (dict "a" (list $dot) ))) "r") $dot.Release.Namespace) (printf "%s-webhook-service.%s.svc.%s" (get (fromJson (include "operator.RedpandaOperatorName" (dict "a" (list $dot) ))) "r") $dot.Release.Namespace $values.clusterDomain)) "issuerRef" (mustMergeOverwrite (dict "name" "" ) (dict "kind" "Issuer" "name" (get (fromJson (include "operator.cleanForK8sWithSuffix" (dict "a" (list (get (fromJson (include "operator.Fullname" (dict "a" (list $dot) ))) "r") "selfsigned-issuer") ))) "r") )) "secretName" $values.webhookSecretName "privateKey" (mustMergeOverwrite (dict ) (dict "rotationPolicy" "Never" )) )) ))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "operator.Issuer" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $values := $dot.Values.AsMap -}}
{{- if (not $values.webhook.enabled) -}}
{{- $_is_returning = true -}}
{{- (dict "r" (coalesce nil)) | toJson -}}
{{- break -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" (mustMergeOverwrite (dict "metadata" (dict "creationTimestamp" (coalesce nil) ) "spec" (dict ) "status" (dict ) ) (mustMergeOverwrite (dict ) (dict "apiVersion" "cert-manager.io/v1" "kind" "Issuer" )) (dict "metadata" (mustMergeOverwrite (dict "creationTimestamp" (coalesce nil) ) (dict "name" (get (fromJson (include "operator.cleanForK8sWithSuffix" (dict "a" (list (get (fromJson (include "operator.Fullname" (dict "a" (list $dot) ))) "r") "selfsigned-issuer") ))) "r") "namespace" $dot.Release.Namespace "labels" (get (fromJson (include "operator.Labels" (dict "a" (list $dot) ))) "r") "annotations" $values.annotations )) "spec" (mustMergeOverwrite (dict ) (mustMergeOverwrite (dict ) (dict "selfSigned" (mustMergeOverwrite (dict ) (dict )) )) (dict )) ))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

