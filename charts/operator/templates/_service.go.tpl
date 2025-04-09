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
{{- (dict "r" (mustMergeOverwrite (dict "metadata" (dict "creationTimestamp" (coalesce nil) ) "spec" (dict ) "status" (dict "loadBalancer" (dict ) ) ) (mustMergeOverwrite (dict ) (dict "apiVersion" "v1" "kind" "Service" )) (dict "metadata" (mustMergeOverwrite (dict "creationTimestamp" (coalesce nil) ) (dict "name" (printf "%s-webhook-service" (get (fromJson (include "operator.Name" (dict "a" (list $dot) ))) "r")) "namespace" $dot.Release.Namespace "labels" (get (fromJson (include "operator.Labels" (dict "a" (list $dot) ))) "r") "annotations" $values.annotations )) "spec" (mustMergeOverwrite (dict ) (dict "selector" (get (fromJson (include "operator.SelectorLabels" (dict "a" (list $dot) ))) "r") "ports" (list (mustMergeOverwrite (dict "port" 0 "targetPort" 0 ) (dict "port" ((443 | int) | int) "targetPort" (9443 | int) ))) )) ))) | toJson -}}
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

{{- define "operator.MutatingWebhookConfiguration" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $values := $dot.Values.AsMap -}}
{{- if (or (not $values.webhook.enabled) (ne $values.scope "Cluster")) -}}
{{- $_is_returning = true -}}
{{- (dict "r" (coalesce nil)) | toJson -}}
{{- break -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" (mustMergeOverwrite (dict "metadata" (dict "creationTimestamp" (coalesce nil) ) ) (mustMergeOverwrite (dict ) (dict "apiVersion" "admissionregistration.k8s.io/v1" "kind" "MutatingWebhookConfiguration" )) (dict "metadata" (mustMergeOverwrite (dict "creationTimestamp" (coalesce nil) ) (dict "name" (printf "%s-mutating-webhook-configuration" (get (fromJson (include "operator.Fullname" (dict "a" (list $dot) ))) "r")) "namespace" $dot.Release.Namespace "annotations" (dict "cert-manager.io/inject-ca-from" (printf "%s/redpanda-serving-cert" $dot.Release.Namespace) ) )) "webhooks" (list (mustMergeOverwrite (dict "name" "" "clientConfig" (dict ) "sideEffects" (coalesce nil) "admissionReviewVersions" (coalesce nil) ) (dict "admissionReviewVersions" (list "v1" "v1beta1") "clientConfig" (mustMergeOverwrite (dict ) (dict "service" (mustMergeOverwrite (dict "namespace" "" "name" "" ) (dict "name" (printf "%s-webhook-service" (get (fromJson (include "operator.Name" (dict "a" (list $dot) ))) "r")) "namespace" $dot.Release.Namespace "path" "/mutate-redpanda-vectorized-io-v1alpha1-cluster" )) )) "failurePolicy" "Fail" "name" "mcluster.kb.io" "rules" (list (mustMergeOverwrite (dict ) (mustMergeOverwrite (dict ) (dict "apiGroups" (list "redpanda.vectorized.io") "apiVersions" (list "v1alpha1") "resources" (list "clusters") )) (dict "operations" (list "CREATE" "UPDATE") ))) "sideEffects" "None" ))) ))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "operator.ValidatingWebhookConfiguration" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $values := $dot.Values.AsMap -}}
{{- if (or (not $values.webhook.enabled) (ne $values.scope "Cluster")) -}}
{{- $_is_returning = true -}}
{{- (dict "r" (coalesce nil)) | toJson -}}
{{- break -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" (mustMergeOverwrite (dict "metadata" (dict "creationTimestamp" (coalesce nil) ) ) (mustMergeOverwrite (dict ) (dict "apiVersion" "admissionregistration.k8s.io/v1" "kind" "ValidatingWebhookConfiguration" )) (dict "metadata" (mustMergeOverwrite (dict "creationTimestamp" (coalesce nil) ) (dict "name" (printf "%s-validating-webhook-configuration" (get (fromJson (include "operator.Fullname" (dict "a" (list $dot) ))) "r")) "namespace" $dot.Release.Namespace "annotations" (dict "cert-manager.io/inject-ca-from" (printf "%s/redpanda-serving-cert" $dot.Release.Namespace) ) )) "webhooks" (list (mustMergeOverwrite (dict "name" "" "clientConfig" (dict ) "sideEffects" (coalesce nil) "admissionReviewVersions" (coalesce nil) ) (dict "admissionReviewVersions" (list "v1" "v1beta1") "clientConfig" (mustMergeOverwrite (dict ) (dict "service" (mustMergeOverwrite (dict "namespace" "" "name" "" ) (dict "name" (printf "%s-webhook-service" (get (fromJson (include "operator.Name" (dict "a" (list $dot) ))) "r")) "namespace" $dot.Release.Namespace "path" "/validate-redpanda-vectorized-io-v1alpha1-cluster" )) )) "failurePolicy" "Fail" "name" "mcluster.kb.io" "rules" (list (mustMergeOverwrite (dict ) (mustMergeOverwrite (dict ) (dict "apiGroups" (list "redpanda.vectorized.io") "apiVersions" (list "v1alpha1") "resources" (list "clusters") )) (dict "operations" (list "CREATE" "UPDATE") ))) "sideEffects" "None" ))) ))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

