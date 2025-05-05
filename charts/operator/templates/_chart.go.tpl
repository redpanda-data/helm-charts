{{- /* Generated from "chart.go" */ -}}

{{- define "operator.render" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $manifests := (list (get (fromJson (include "operator.Issuer" (dict "a" (list $dot)))) "r") (get (fromJson (include "operator.Certificate" (dict "a" (list $dot)))) "r") (get (fromJson (include "operator.ConfigMap" (dict "a" (list $dot)))) "r") (get (fromJson (include "operator.MetricsService" (dict "a" (list $dot)))) "r") (get (fromJson (include "operator.WebhookService" (dict "a" (list $dot)))) "r") (get (fromJson (include "operator.MutatingWebhookConfiguration" (dict "a" (list $dot)))) "r") (get (fromJson (include "operator.ValidatingWebhookConfiguration" (dict "a" (list $dot)))) "r") (get (fromJson (include "operator.ServiceAccount" (dict "a" (list $dot)))) "r") (get (fromJson (include "operator.ServiceMonitor" (dict "a" (list $dot)))) "r") (get (fromJson (include "operator.Deployment" (dict "a" (list $dot)))) "r")) -}}
{{- range $_, $role := (get (fromJson (include "operator.Roles" (dict "a" (list $dot)))) "r") -}}
{{- $manifests = (concat (default (list) $manifests) (list $role)) -}}
{{- end -}}
{{- if $_is_returning -}}
{{- break -}}
{{- end -}}
{{- range $_, $cr := (get (fromJson (include "operator.ClusterRoles" (dict "a" (list $dot)))) "r") -}}
{{- $manifests = (concat (default (list) $manifests) (list $cr)) -}}
{{- end -}}
{{- if $_is_returning -}}
{{- break -}}
{{- end -}}
{{- range $_, $rb := (get (fromJson (include "operator.RoleBindings" (dict "a" (list $dot)))) "r") -}}
{{- $manifests = (concat (default (list) $manifests) (list $rb)) -}}
{{- end -}}
{{- if $_is_returning -}}
{{- break -}}
{{- end -}}
{{- range $_, $crb := (get (fromJson (include "operator.ClusterRoleBindings" (dict "a" (list $dot)))) "r") -}}
{{- $manifests = (concat (default (list) $manifests) (list $crb)) -}}
{{- end -}}
{{- if $_is_returning -}}
{{- break -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" $manifests) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

