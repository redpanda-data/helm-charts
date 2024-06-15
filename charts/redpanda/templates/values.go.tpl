{{- /* Generated from "values.go" */ -}}

{{- define "redpanda.Storage.IsTieredStorageEnabled" -}}
{{- $s := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $conf := (get (fromJson (include "redpanda.Storage.GetTieredStorageConfig" (dict "a" (list $s) ))) "r") -}}
{{- $tmp_tuple_3 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "_shims.dicttest" (dict "a" (list $conf "cloud_storage_enabled" (coalesce nil)) ))) "r")) ))) "r") -}}
{{- $ok := $tmp_tuple_3.T2 -}}
{{- $b := $tmp_tuple_3.T1 -}}
{{- (dict "r" (and $ok (get (fromJson (include "_shims.typeassertion" (dict "a" (list "bool" $b) ))) "r"))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.Storage.GetTieredStorageConfig" -}}
{{- $s := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- if (gt ((get (fromJson (include "_shims.len" (dict "a" (list $s.tieredConfig) ))) "r") | int) (0 | int)) -}}
{{- (dict "r" $s.tieredConfig) | toJson -}}
{{- break -}}
{{- end -}}
{{- (dict "r" $s.tiered.config) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

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

