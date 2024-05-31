{{- /* Generated from "bootstrap.go" */ -}}

{{- define "_shims.isIntLikeFloat" -}}
{{- $value := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- (dict "r" (and (typeIs "float64" $value) (eq (((subf (float64 $value) (floor $value)) | float64)) (0.0 | float64)))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

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
{{- $_ := (set $out (printf "T%d" ((add (1 | int) $i) | int)) $e) -}}
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

{{- define "_shims._len" -}}
{{- $m := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- if (eq $m (coalesce nil)) -}}
{{- (dict "r" (0 | int)) | toJson -}}
{{- break -}}
{{- end -}}
{{- (dict "r" (len $m)) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "_shims.ptr_Deref" -}}
{{- $ptr := (index .a 0) -}}
{{- $def := (index .a 1) -}}
{{- range $_ := (list 1) -}}
{{- if (ne $ptr (coalesce nil)) -}}
{{- (dict "r" $ptr) | toJson -}}
{{- break -}}
{{- end -}}
{{- (dict "r" $def) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "_shims.ptr_Equal" -}}
{{- $a := (index .a 0) -}}
{{- $b := (index .a 1) -}}
{{- range $_ := (list 1) -}}
{{- if (and (eq $a (coalesce nil)) (eq $b (coalesce nil))) -}}
{{- (dict "r" true) | toJson -}}
{{- break -}}
{{- end -}}
{{- (dict "r" (eq $a $b)) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "_shims.lookup" -}}
{{- $apiVersion := (index .a 0) -}}
{{- $kind := (index .a 1) -}}
{{- $namespace := (index .a 2) -}}
{{- $name := (index .a 3) -}}
{{- range $_ := (list 1) -}}
{{- $result := (lookup $apiVersion $kind $namespace $name) -}}
{{- if (empty $result) -}}
{{- (dict "r" (list (coalesce nil) false)) | toJson -}}
{{- break -}}
{{- end -}}
{{- (dict "r" (list $result true)) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "_shims.asnumeric" -}}
{{- $value := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- if (typeIs "float64" $value) -}}
{{- (dict "r" (list $value true)) | toJson -}}
{{- break -}}
{{- end -}}
{{- if (typeIs "int64" $value) -}}
{{- (dict "r" (list $value true)) | toJson -}}
{{- break -}}
{{- end -}}
{{- if (typeIs "int" $value) -}}
{{- (dict "r" (list $value true)) | toJson -}}
{{- break -}}
{{- end -}}
{{- (dict "r" (list (0 | int) false)) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "_shims.asintegral" -}}
{{- $value := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- if (or (typeIs "int64" $value) (typeIs "int" $value)) -}}
{{- (dict "r" (list $value true)) | toJson -}}
{{- break -}}
{{- end -}}
{{- if (and (typeIs "float64" $value) (eq (floor $value) $value)) -}}
{{- (dict "r" (list $value true)) | toJson -}}
{{- break -}}
{{- end -}}
{{- (dict "r" (list (0 | int) false)) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "_shims.parseResource" -}}
{{- $repr := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- if (typeIs "float64" $repr) -}}
{{- (dict "r" (list (float64 $repr) 1.0)) | toJson -}}
{{- break -}}
{{- end -}}
{{- if (not (typeIs "string" $repr)) -}}
{{- $_ := (fail (printf "invalid Quantity expected string or float64 got: %T" $repr)) -}}
{{- end -}}
{{- if (not (regexMatch `^[0-9]+(\.[0-9]{0,6})?(k|m|M|G|T|P|Ki|Mi|Gi|Ti|Pi)?$` $repr)) -}}
{{- $_ := (fail (printf "invalid Quantity: %q" $repr)) -}}
{{- end -}}
{{- $reprStr := (toString $repr) -}}
{{- $unit := (regexFind "(k|m|M|G|T|P|Ki|Mi|Gi|Ti|Pi)$" $repr) -}}
{{- $numeric := (float64 (substr (0 | int) ((sub ((get (fromJson (include "_shims._len" (dict "a" (list $reprStr) ))) "r") | int) ((get (fromJson (include "_shims._len" (dict "a" (list $unit) ))) "r") | int)) | int) $reprStr)) -}}
{{- $scale := (float64 0) -}}
{{- if (eq $unit "") -}}
{{- $scale = 1.0 -}}
{{- else -}}{{- if (eq $unit "m") -}}
{{- $scale = 0.001 -}}
{{- else -}}{{- if (eq $unit "k") -}}
{{- $scale = ((1000 | int) | float64) -}}
{{- else -}}{{- if (eq $unit "M") -}}
{{- $scale = ((1000000 | int) | float64) -}}
{{- else -}}{{- if (eq $unit "G") -}}
{{- $scale = ((1000000000 | int) | float64) -}}
{{- else -}}{{- if (eq $unit "T") -}}
{{- $scale = ((1000000000000 | int) | float64) -}}
{{- else -}}{{- if (eq $unit "P") -}}
{{- $scale = ((1000000000000000 | int) | float64) -}}
{{- else -}}{{- if (eq $unit "Ki") -}}
{{- $scale = ((1024 | int) | float64) -}}
{{- else -}}{{- if (eq $unit "Mi") -}}
{{- $scale = ((1048576 | int) | float64) -}}
{{- else -}}{{- if (eq $unit "Gi") -}}
{{- $scale = ((1073741824 | int) | float64) -}}
{{- else -}}{{- if (eq $unit "Ti") -}}
{{- $scale = ((1099511627776 | int) | float64) -}}
{{- else -}}{{- if (eq $unit "Pi") -}}
{{- $scale = ((1125899906842624 | int) | float64) -}}
{{- else -}}
{{- $_ := (fail (printf "unknown unit: %q" $unit)) -}}
{{- end -}}
{{- end -}}
{{- end -}}
{{- end -}}
{{- end -}}
{{- end -}}
{{- end -}}
{{- end -}}
{{- end -}}
{{- end -}}
{{- end -}}
{{- end -}}
{{- (dict "r" (list $numeric $scale)) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "_shims.resource_MustParse" -}}
{{- $repr := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $tmp_tuple_1 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "_shims.parseResource" (dict "a" (list $repr) ))) "r")) ))) "r") -}}
{{- $scale := ($tmp_tuple_1.T2 | float64) -}}
{{- $numeric := ($tmp_tuple_1.T1 | float64) -}}
{{- $scales := (list 1.0 0.001 (1000 | int) (1000000 | int) (1000000000 | int) (1000000000000 | int) (1000000000000000 | int) (1024 | int) (1048576 | int) (1073741824 | int) (1099511627776 | int) (1125899906842624 | int)) -}}
{{- $strs := (list "" "m" "K" "M" "G" "T" "P" "Ki" "Mi" "Gi" "Ti" "Pi") -}}
{{- $idx := -1 -}}
{{- range $i, $s := $scales -}}
{{- if (eq ($s | float64) ($scale | float64)) -}}
{{- $idx = $i -}}
{{- break -}}
{{- end -}}
{{- end -}}
{{- if (eq $idx -1) -}}
{{- $_ := (fail (printf "unknown scale: %v" $scale)) -}}
{{- end -}}
{{- (dict "r" (printf "%s%s" (toString $numeric) (index $strs $idx))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "_shims.resource_Value" -}}
{{- $repr := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $tmp_tuple_2 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "_shims.parseResource" (dict "a" (list $repr) ))) "r")) ))) "r") -}}
{{- $scale := ($tmp_tuple_2.T2 | float64) -}}
{{- $numeric := ($tmp_tuple_2.T1 | float64) -}}
{{- (dict "r" (int64 (ceil ((mulf $numeric $scale) | float64)))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "_shims.resource_AsInt64" -}}
{{- $repr := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $tmp_tuple_3 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "_shims.parseResource" (dict "a" (list $repr) ))) "r")) ))) "r") -}}
{{- $scale := ($tmp_tuple_3.T2 | float64) -}}
{{- $numeric := ($tmp_tuple_3.T1 | float64) -}}
{{- $tmp_tuple_4 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "_shims.asintegral" (dict "a" (list ((mulf $numeric $scale) | float64)) ))) "r")) ))) "r") -}}
{{- $ok := $tmp_tuple_4.T2 -}}
{{- $asInt := $tmp_tuple_4.T1 -}}
{{- (dict "r" (list $asInt $ok)) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

