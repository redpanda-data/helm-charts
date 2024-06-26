{{- /* Generated from "post_install_upgrade_job.tpl.go" */ -}}

{{- define "redpanda.PostInstallUpgradeEnvironmentVariables" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $values := $dot.Values.AsMap -}}
{{- $envars := (list ) -}}
{{- $license_1 := (get (fromJson (include "redpanda.GetLicenseLiteral" (dict "a" (list $dot) ))) "r") -}}
{{- $secretReference_2 := (get (fromJson (include "redpanda.GetLicenseSecretReference" (dict "a" (list $dot) ))) "r") -}}
{{- if (ne $license_1 "") -}}
{{- $envars = (concat (default (list ) $envars) (list (mustMergeOverwrite (dict "name" "" ) (dict "name" "REDPANDA_LICENSE" "value" $license_1 )))) -}}
{{- else -}}{{- if (ne $secretReference_2 (coalesce nil)) -}}
{{- $envars = (concat (default (list ) $envars) (list (mustMergeOverwrite (dict "name" "" ) (dict "name" "REDPANDA_LICENSE" "valueFrom" (mustMergeOverwrite (dict ) (dict "secretKeyRef" $secretReference_2 )) )))) -}}
{{- end -}}
{{- end -}}
{{- if (not (get (fromJson (include "redpanda.Storage.IsTieredStorageEnabled" (dict "a" (list $values.storage) ))) "r")) -}}
{{- $_is_returning = true -}}
{{- (dict "r" $envars) | toJson -}}
{{- break -}}
{{- end -}}
{{- $tieredStorageConfig := (get (fromJson (include "redpanda.Storage.GetTieredStorageConfig" (dict "a" (list $values.storage) ))) "r") -}}
{{- $tmp_tuple_1 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "_shims.dicttest" (dict "a" (list $tieredStorageConfig "cloud_storage_azure_container" (coalesce nil)) ))) "r")) ))) "r") -}}
{{- $azureContainerExists := $tmp_tuple_1.T2 -}}
{{- $ac := $tmp_tuple_1.T1 -}}
{{- $tmp_tuple_2 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "_shims.dicttest" (dict "a" (list $tieredStorageConfig "cloud_storage_azure_storage_account" (coalesce nil)) ))) "r")) ))) "r") -}}
{{- $azureStorageAccountExists := $tmp_tuple_2.T2 -}}
{{- $asa := $tmp_tuple_2.T1 -}}
{{- if (and (and (and $azureContainerExists (ne $ac (coalesce nil))) $azureStorageAccountExists) (ne $asa (coalesce nil))) -}}
{{- $envars = (concat (default (list ) $envars) (default (list ) (get (fromJson (include "redpanda.addAzureSharedKey" (dict "a" (list $tieredStorageConfig $values) ))) "r"))) -}}
{{- else -}}
{{- $envars = (concat (default (list ) $envars) (default (list ) (get (fromJson (include "redpanda.addCloudStorageSecretKey" (dict "a" (list $tieredStorageConfig $values) ))) "r"))) -}}
{{- end -}}
{{- $envars = (concat (default (list ) $envars) (default (list ) (get (fromJson (include "redpanda.addCloudStorageAccessKey" (dict "a" (list $tieredStorageConfig $values) ))) "r"))) -}}
{{- range $k, $v := $tieredStorageConfig -}}
{{- if (or (or (eq $k "cloud_storage_access_key") (eq $k "cloud_storage_secret_key")) (eq $k "cloud_storage_azure_shared_key")) -}}
{{- continue -}}
{{- end -}}
{{- if (or (eq $v (coalesce nil)) (empty $v)) -}}
{{- continue -}}
{{- end -}}
{{- if (and (eq $k "cloud_storage_cache_size") (ne $v (coalesce nil))) -}}
{{- $envars = (concat (default (list ) $envars) (list (mustMergeOverwrite (dict "name" "" ) (dict "name" (printf "RPK_%s" (upper $k)) "value" (toJson ((get (fromJson (include "_shims.resource_Value" (dict "a" (list $v) ))) "r") | int64)) )))) -}}
{{- continue -}}
{{- end -}}
{{- $tmp_tuple_3 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "_shims.typetest" (dict "a" (list "string" $v "") ))) "r")) ))) "r") -}}
{{- $ok_4 := $tmp_tuple_3.T2 -}}
{{- $str_3 := $tmp_tuple_3.T1 -}}
{{- if $ok_4 -}}
{{- $envars = (concat (default (list ) $envars) (list (mustMergeOverwrite (dict "name" "" ) (dict "name" (printf "RPK_%s" (upper $k)) "value" $str_3 )))) -}}
{{- else -}}
{{- $envars = (concat (default (list ) $envars) (list (mustMergeOverwrite (dict "name" "" ) (dict "name" (printf "RPK_%s" (upper $k)) "value" (mustToJson $v) )))) -}}
{{- end -}}
{{- end -}}
{{- if $_is_returning -}}
{{- break -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" $envars) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.addCloudStorageAccessKey" -}}
{{- $tieredStorageConfig := (index .a 0) -}}
{{- $values := (index .a 1) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $tmp_tuple_4 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "_shims.dicttest" (dict "a" (list $tieredStorageConfig "cloud_storage_access_key" (coalesce nil)) ))) "r")) ))) "r") -}}
{{- $ok_6 := $tmp_tuple_4.T2 -}}
{{- $v_5 := $tmp_tuple_4.T1 -}}
{{- $ak_7 := $values.storage.tiered.credentialsSecretRef.accessKey -}}
{{- if (and $ok_6 (ne $v_5 "")) -}}
{{- $_is_returning = true -}}
{{- (dict "r" (list (mustMergeOverwrite (dict "name" "" ) (dict "name" "RPK_CLOUD_STORAGE_ACCESS_KEY" "value" (get (fromJson (include "_shims.typeassertion" (dict "a" (list "string" $v_5) ))) "r") )))) | toJson -}}
{{- break -}}
{{- else -}}{{- if (get (fromJson (include "redpanda.SecretRef.IsValid" (dict "a" (list $ak_7) ))) "r") -}}
{{- $_is_returning = true -}}
{{- (dict "r" (list (mustMergeOverwrite (dict "name" "" ) (dict "name" "RPK_CLOUD_STORAGE_ACCESS_KEY" "valueFrom" (mustMergeOverwrite (dict ) (dict "secretKeyRef" (mustMergeOverwrite (dict "key" "" ) (mustMergeOverwrite (dict ) (dict "name" $ak_7.name )) (dict "key" $ak_7.key )) )) )))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" (list )) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.addCloudStorageSecretKey" -}}
{{- $tieredStorageConfig := (index .a 0) -}}
{{- $values := (index .a 1) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $tmp_tuple_5 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "_shims.dicttest" (dict "a" (list $tieredStorageConfig "cloud_storage_secret_key" (coalesce nil)) ))) "r")) ))) "r") -}}
{{- $ok_9 := $tmp_tuple_5.T2 -}}
{{- $v_8 := $tmp_tuple_5.T1 -}}
{{- $sk_10 := $values.storage.tiered.credentialsSecretRef.secretKey -}}
{{- if (and $ok_9 (ne $v_8 "")) -}}
{{- $_is_returning = true -}}
{{- (dict "r" (list (mustMergeOverwrite (dict "name" "" ) (dict "name" "RPK_CLOUD_STORAGE_SECRET_KEY" "value" (get (fromJson (include "_shims.typeassertion" (dict "a" (list "string" $v_8) ))) "r") )))) | toJson -}}
{{- break -}}
{{- else -}}{{- if (get (fromJson (include "redpanda.SecretRef.IsValid" (dict "a" (list $sk_10) ))) "r") -}}
{{- $_is_returning = true -}}
{{- (dict "r" (list (mustMergeOverwrite (dict "name" "" ) (dict "name" "RPK_CLOUD_STORAGE_SECRET_KEY" "valueFrom" (mustMergeOverwrite (dict ) (dict "secretKeyRef" (mustMergeOverwrite (dict "key" "" ) (mustMergeOverwrite (dict ) (dict "name" $sk_10.name )) (dict "key" $sk_10.key )) )) )))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" (list )) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.addAzureSharedKey" -}}
{{- $tieredStorageConfig := (index .a 0) -}}
{{- $values := (index .a 1) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $tmp_tuple_6 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "_shims.dicttest" (dict "a" (list $tieredStorageConfig "cloud_storage_azure_shared_key" (coalesce nil)) ))) "r")) ))) "r") -}}
{{- $ok_12 := $tmp_tuple_6.T2 -}}
{{- $v_11 := $tmp_tuple_6.T1 -}}
{{- $sk_13 := $values.storage.tiered.credentialsSecretRef.secretKey -}}
{{- if (and $ok_12 (ne $v_11 "")) -}}
{{- $_is_returning = true -}}
{{- (dict "r" (list (mustMergeOverwrite (dict "name" "" ) (dict "name" "RPK_CLOUD_STORAGE_AZURE_SHARED_KEY" "value" (get (fromJson (include "_shims.typeassertion" (dict "a" (list "string" $v_11) ))) "r") )))) | toJson -}}
{{- break -}}
{{- else -}}{{- if (get (fromJson (include "redpanda.SecretRef.IsValid" (dict "a" (list $sk_13) ))) "r") -}}
{{- $_is_returning = true -}}
{{- (dict "r" (list (mustMergeOverwrite (dict "name" "" ) (dict "name" "RPK_CLOUD_STORAGE_AZURE_SHARED_KEY" "valueFrom" (mustMergeOverwrite (dict ) (dict "secretKeyRef" (mustMergeOverwrite (dict "key" "" ) (mustMergeOverwrite (dict ) (dict "name" $sk_13.name )) (dict "key" $sk_13.key )) )) )))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" (list )) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.GetLicenseLiteral" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $values := $dot.Values.AsMap -}}
{{- if (ne $values.enterprise.license "") -}}
{{- $_is_returning = true -}}
{{- (dict "r" $values.enterprise.license) | toJson -}}
{{- break -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" $values.license_key) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.GetLicenseSecretReference" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $values := $dot.Values.AsMap -}}
{{- if (not (empty $values.enterprise.licenseSecretRef)) -}}
{{- $_is_returning = true -}}
{{- (dict "r" (mustMergeOverwrite (dict "key" "" ) (mustMergeOverwrite (dict ) (dict "name" $values.enterprise.licenseSecretRef.name )) (dict "key" $values.enterprise.licenseSecretRef.key ))) | toJson -}}
{{- break -}}
{{- else -}}{{- if (not (empty $values.license_secret_ref)) -}}
{{- $_is_returning = true -}}
{{- (dict "r" (mustMergeOverwrite (dict "key" "" ) (mustMergeOverwrite (dict ) (dict "name" $values.license_secret_ref.secret_name )) (dict "key" $values.license_secret_ref.secret_key ))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" (coalesce nil)) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

