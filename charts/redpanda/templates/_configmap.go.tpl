{{- /* Generated from "configmap.tpl.go" */ -}}

{{- define "redpanda.RedpandaAdditionalStartFlags" -}}
{{- $dot := (index .a 0) -}}
{{- $smp := (index .a 1) -}}
{{- $memory := (index .a 2) -}}
{{- $reserveMemory := (index .a 3) -}}
{{- range $_ := (list 1) -}}
{{- $values := $dot.Values.AsMap -}}
{{- $chartFlags := (dict "smp" $smp "memory" (printf "%sM" $memory) "reserve-memory" (printf "%sM" $reserveMemory) "default-log-level" $values.logging.logLevel ) -}}
{{- if (eq (index $values.config.node "developer_mode") true) -}}
{{- $_ := (unset $chartFlags "reserve-memory") -}}
{{- end -}}
{{- range $flag, $_ := $chartFlags -}}
{{- range $_, $userFlag := $values.statefulset.additionalRedpandaCmdFlags -}}
{{- if (regexMatch (printf "^--%s" $flag) $userFlag) -}}
{{- $_ := (unset $chartFlags $flag) -}}
{{- end -}}
{{- end -}}
{{- end -}}
{{- $keys := (keys $chartFlags) -}}
{{- $_ := (sortAlpha $keys) -}}
{{- $flags := (list ) -}}
{{- range $_, $key := $keys -}}
{{- $flags = (mustAppend $flags (printf "--%s=%s" $key (index $chartFlags $key))) -}}
{{- end -}}
{{- (dict "r" (concat $flags $values.statefulset.additionalRedpandaCmdFlags)) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

