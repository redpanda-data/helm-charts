{{- /* Generated from "configmap.tpl.go" */ -}}

{{- define "redpanda.RedpandaAdditionalStartFlags" -}}
{{- $dot := (index .a 0) -}}
{{- $smp := (index .a 1) -}}
{{- range $_ := (list 1) -}}
{{- $values := $dot.Values.AsMap -}}
{{- $chartFlags := (dict "smp" $smp "memory" (printf "%dM" (int (get (fromJson (include "redpanda.RedpandaMemory" (dict "a" (list $dot) ))) "r"))) "reserve-memory" (printf "%dM" (int (get (fromJson (include "redpanda.RedpandaReserveMemory" (dict "a" (list $dot) ))) "r"))) "default-log-level" $values.logging.logLevel ) -}}
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

