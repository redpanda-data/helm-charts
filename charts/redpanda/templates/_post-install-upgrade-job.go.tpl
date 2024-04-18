{{- /* Generated from "post_install_upgrade_job.tpl.go" */ -}}

{{- define "redpanda.RedpandaEnvironmentVariables" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $values := $dot.Values.AsMap -}}
{{- $tieredStorageConfig := $values.storage.tiered.config -}}
{{- if (ne $values.storage.tieredConfig (coalesce nil)) -}}
{{- if (gt (len $values.storage.tieredConfig) 0) -}}
{{- $tieredStorageConfig = $values.storage.tieredConfig -}}
{{- end -}}
{{- end -}}
{{- $envars := (list ) -}}
{{- $license_1 := (get (fromJson (include "redpanda.GetLicense" (dict "a" (list $dot) ))) "r") -}}
{{- $secretReference_2 := (get (fromJson (include "redpanda.EnterpriseSecretNameReference" (dict "a" (list $dot) ))) "r") -}}
{{- if (ne $license_1 "") -}}
{{- $envars = (mustAppend $envars (mustMergeOverwrite (dict "name" "" ) (dict "name" "REDPANDA_LICENSE" "value" $license_1 ))) -}}
{{- else -}}{{- if (ne $secretReference_2 (coalesce nil)) -}}
{{- $envars = (mustAppend $envars (mustMergeOverwrite (dict "name" "" ) (dict "name" "REDPANDA_LICENSE" "valueFrom" (mustMergeOverwrite (dict ) (dict "secretKeyRef" $secretReference_2 )) ))) -}}
{{- end -}}
{{- end -}}
{{- if (not (get (fromJson (include "redpanda.IsTieredStorageEnabled" (dict "a" (list $tieredStorageConfig) ))) "r")) -}}
{{- (dict "r" $envars) | toJson -}}
{{- break -}}
{{- end -}}
{{- if (not (empty (index $tieredStorageConfig "cloud_storage_secret_key"))) -}}
{{- $envars = (mustAppend $envars (mustMergeOverwrite (dict "name" "" ) (dict "name" "RPK_CLOUD_STORAGE_SECRET_KEY" "value" (index $tieredStorageConfig "cloud_storage_secret_key") ))) -}}
{{- else -}}{{- if (and (and (ne $values.storage.tiered.credentialsSecretRef.secretKey (coalesce nil)) (not (empty $values.storage.tiered.credentialsSecretRef.secretKey.name))) (not (empty $values.storage.tiered.credentialsSecretRef.secretKey.key))) -}}
{{- $envars = (mustAppend $envars (mustMergeOverwrite (dict "name" "" ) (dict "name" "RPK_CLOUD_STORAGE_SECRET_KEY" "valueFrom" (mustMergeOverwrite (dict ) (dict "secretKeyRef" (mustMergeOverwrite (mustMergeOverwrite (dict ) (dict "key" "" )) (mustMergeOverwrite (dict ) (dict "name" $values.storage.tiered.credentialsSecretRef.secretKey.name )) (dict "key" $values.storage.tiered.credentialsSecretRef.secretKey.key )) )) ))) -}}
{{- end -}}
{{- end -}}
{{- if (not (empty (index $tieredStorageConfig "cloud_storage_access_key"))) -}}
{{- $envars = (mustAppend $envars (mustMergeOverwrite (dict "name" "" ) (dict "name" "RPK_CLOUD_STORAGE_ACCESS_KEY" "value" (index $tieredStorageConfig "cloud_storage_access_key") ))) -}}
{{- else -}}{{- if (and (and (ne $values.storage.tiered.credentialsSecretRef.secretKey (coalesce nil)) (not (empty $values.storage.tiered.credentialsSecretRef.accessKey.name))) (not (empty $values.storage.tiered.credentialsSecretRef.accessKey.key))) -}}
{{- $envars = (mustAppend $envars (mustMergeOverwrite (dict "name" "" ) (dict "name" "RPK_CLOUD_STORAGE_ACCESS_KEY" "valueFrom" (mustMergeOverwrite (dict ) (dict "secretKeyRef" (mustMergeOverwrite (mustMergeOverwrite (dict ) (dict "key" "" )) (mustMergeOverwrite (dict ) (dict "name" $values.storage.tiered.credentialsSecretRef.accessKey.name )) (dict "key" $values.storage.tiered.credentialsSecretRef.accessKey.key )) )) ))) -}}
{{- end -}}
{{- end -}}
{{- range $k, $v := $tieredStorageConfig -}}
{{- if (or (eq $k "cloud_storage_access_key") (eq $k "cloud_storage_secret_key")) -}}
{{- continue -}}
{{- end -}}
{{- if (or (eq $v (coalesce nil)) (empty $v)) -}}
{{- continue -}}
{{- end -}}
{{- if (eq $k "cloud_storage_cache_size") -}}
{{- $envars = (mustAppend $envars (mustMergeOverwrite (dict "name" "" ) (dict "name" (printf "RPK_%s" (upper $k)) "value" (toJson (int64 (include "_shims.sitobytes" $v))) ))) -}}
{{- continue -}}
{{- end -}}
{{- $envars = (mustAppend $envars (mustMergeOverwrite (dict "name" "" ) (dict "name" (printf "RPK_%s" (upper $k)) "value" (mustToJson $v) ))) -}}
{{- end -}}
{{- (dict "r" $envars) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.GetLicense" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $values := $dot.Values.AsMap -}}
{{- if (ne $values.enterprise.license "") -}}
{{- (dict "r" $values.enterprise.license) | toJson -}}
{{- break -}}
{{- end -}}
{{- (dict "r" $values.license_key) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.EnterpriseSecretNameReference" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $values := $dot.Values.AsMap -}}
{{- if (not (empty $values.enterprise.licenseSecretRef)) -}}
{{- (dict "r" (mustMergeOverwrite (mustMergeOverwrite (dict ) (dict "key" "" )) (mustMergeOverwrite (dict ) (dict "name" $values.enterprise.licenseSecretRef.name )) (dict "key" $values.enterprise.licenseSecretRef.key ))) | toJson -}}
{{- break -}}
{{- else -}}{{- if (not (empty $values.license_secret_ref)) -}}
{{- (dict "r" (mustMergeOverwrite (mustMergeOverwrite (dict ) (dict "key" "" )) (mustMergeOverwrite (dict ) (dict "name" $values.license_secret_ref.secret_name )) (dict "key" $values.license_secret_ref.secret_key ))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}
{{- (dict "r" (coalesce nil)) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.IsTieredStorageEnabled" -}}
{{- $tieredStorageConfig := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $tmp_tuple_1 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "_shims.dicttest" (dict "a" (list $tieredStorageConfig "cloud_storage_enabled") ))) "r")) ))) "r") -}}
{{- $ok_4 := $tmp_tuple_1.T2 -}}
{{- $b_3 := $tmp_tuple_1.T1 -}}
{{- if (and $ok_4 $b_3) -}}
{{- (dict "r" true) | toJson -}}
{{- break -}}
{{- end -}}
{{- (dict "r" false) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

