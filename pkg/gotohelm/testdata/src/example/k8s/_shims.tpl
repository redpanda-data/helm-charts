{{- /* Generated from "" */ -}}

{{- define "_shims.typetest" -}}
{{- $type := (index .a 0) -}}
{{- $value := (index .a 1) -}}
{{- dict "r" (list $value (typeIs $type $value)) | toJson -}}
{{- end -}}

{{- define "_shims.dicttest" -}}
{{- $dict := (index .a 0) -}}
{{- $key := (index .a 1) -}}
{{- if (hasKey $dict $key) -}}
{{- (dict "r" (list (index $dict $key) true)) | toJson -}}
{{- else -}}
{{- (dict "r" (list "" false)) | toJson -}}
{{- end -}}
{{- end -}}

{{- define "_shims.typeassertion" -}}
{{- $type := (index .a 0) -}}
{{- $value := (index .a 1) -}}
{{- if (not (typeIs $type $value)) -}}
{{- (fail "TODO MAKE THIS A NICE MESSAGE") -}}
{{- end -}}
{{- (dict "r" $value) | toJson -}}
{{- end -}}

{{- define "_shims.compact" -}}
{{- $out := (dict) -}}
{{- range $i, $e := (index .a 0) }}
{{- $_ := (set $out (printf "T%d" (add1 $i)) $e) -}}
{{- end -}}
{{- (dict "r" $out) | toJson -}}
{{- end -}}
