{{- /* Generated from "_shims.go" */ -}}

{{- define "_shims.typetest" -}}
{{- $typ := (index .a 0) -}}
{{- $value := (index .a 1) -}}
{{- $zero := (index .a 2) -}}
{{- range $_ := (list 1) -}}
{{- if (typeIs $typ $value) -}}
{{- (dict "r" (list $value true)) | toJson -}}
{{- break -}}
{{- end -}}
{{- (dict "r" (list $zero false)) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "_shims.typeassertion" -}}
{{- $typ := (index .a 0) -}}
{{- $value := (index .a 1) -}}
{{- range $_ := (list 1) -}}
{{- if (not (typeIs $typ $value)) -}}
{{- $_ := (fail (printf "expected type of %q got: %T" $typ $value)) -}}
{{- end -}}
{{- (dict "r" $value) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "_shims.dicttest" -}}
{{- $m := (index .a 0) -}}
{{- $key := (index .a 1) -}}
{{- $zero := (index .a 2) -}}
{{- range $_ := (list 1) -}}
{{- if (hasKey $m $key) -}}
{{- (dict "r" (list (index $m $key) true)) | toJson -}}
{{- break -}}
{{- end -}}
{{- (dict "r" (list $zero false)) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "_shims.compact" -}}
{{- $args := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $out := (dict ) -}}
{{- range $i, $e := $args -}}
{{- $_ := (set $out (printf "T%d" (int (add 1 $i))) $e) -}}
{{- end -}}
{{- (dict "r" $out) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "_shims.deref" -}}
{{- $ptr := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- if (eq $ptr (coalesce nil)) -}}
{{- $_ := (fail "nil dereference") -}}
{{- end -}}
{{- (dict "r" $ptr) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "_shims.len" -}}
{{- $m := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- if (eq $m (coalesce nil)) -}}
{{- (dict "r" 0) | toJson -}}
{{- break -}}
{{- end -}}
{{- (dict "r" (len $m)) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

