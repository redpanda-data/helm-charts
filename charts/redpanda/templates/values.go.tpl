{{- /* Generated from "values.go" */ -}}

{{- define "redpanda.TieredStorageCredentials.IsAccessKeyReferenceValid" -}}
{{- $tsc := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- (dict "r" (and (and (ne $tsc.accessKey (coalesce nil)) (ne $tsc.accessKey.name "")) (ne $tsc.accessKey.key ""))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.TieredStorageCredentials.IsSecretKeyReferenceValid" -}}
{{- $tsc := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- (dict "r" (and (and (ne $tsc.secretKey (coalesce nil)) (ne $tsc.secretKey.name "")) (ne $tsc.secretKey.key ""))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.TieredStorageConfig.IsTieredStorageEnabled" -}}
{{- $tsc := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $tmp_tuple_8 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "_shims.dicttest" (dict "a" (list $tsc "cloud_storage_enabled" (coalesce nil)) ))) "r")) ))) "r") -}}
{{- $ok_2 := $tmp_tuple_8.T2 -}}
{{- $b_1 := $tmp_tuple_8.T1 -}}
{{- if (and $ok_2 (get (fromJson (include "_shims.typeassertion" (dict "a" (list "bool" $b_1) ))) "r")) -}}
{{- (dict "r" true) | toJson -}}
{{- break -}}
{{- end -}}
{{- (dict "r" false) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

