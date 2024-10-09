{{- /* Generated from "chart.go" */ -}}

{{- define "connectors.render" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $manifests := (list ) -}}
{{- $_is_returning = true -}}
{{- (dict "r" $manifests) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

