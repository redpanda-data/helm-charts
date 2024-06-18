{{- /* Generated from "post_install_upgrade_job.tpl.go" */ -}}

{{- define "redpanda.PostInstallUpgradeEnvironmentVariables" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
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
{{- $tmp_tuple_3 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "_shims.typetest" (dict "a" (list "string" $v "") ))) "r")) ))) "r") -}}
{{- $isStr_4 := $tmp_tuple_3.T2 -}}
{{- $asStr_3 := $tmp_tuple_3.T1 -}}
{{- if (and (and (eq $k "cloud_storage_cache_size") $isStr_4) (ne $asStr_3 "")) -}}
{{- $envars = (concat (default (list ) $envars) (list (mustMergeOverwrite (dict "name" "" ) (dict "name" (printf "RPK_%s" (upper $k)) "value" (toJson ((get (fromJson (include "redpanda.SIToBytes" (dict "a" (list (get (fromJson (include "_shims.typeassertion" (dict "a" (list "string" $v) ))) "r")) ))) "r") | int)) )))) -}}
{{- continue -}}
{{- end -}}
{{- $tmp_tuple_4 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "_shims.typetest" (dict "a" (list "string" $v "") ))) "r")) ))) "r") -}}
{{- $ok_6 := $tmp_tuple_4.T2 -}}
{{- $str_5 := $tmp_tuple_4.T1 -}}
{{- if $ok_6 -}}
{{- $envars = (concat (default (list ) $envars) (list (mustMergeOverwrite (dict "name" "" ) (dict "name" (printf "RPK_%s" (upper $k)) "value" $str_5 )))) -}}
{{- else -}}
{{- $envars = (concat (default (list ) $envars) (list (mustMergeOverwrite (dict "name" "" ) (dict "name" (printf "RPK_%s" (upper $k)) "value" (mustToJson $v) )))) -}}
{{- end -}}
{{- end -}}
{{- (dict "r" $envars) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.addCloudStorageAccessKey" -}}
{{- $tieredStorageConfig := (index .a 0) -}}
{{- $values := (index .a 1) -}}
{{- range $_ := (list 1) -}}
{{- $tmp_tuple_5 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "_shims.dicttest" (dict "a" (list $tieredStorageConfig "cloud_storage_access_key" (coalesce nil)) ))) "r")) ))) "r") -}}
{{- $ok_8 := $tmp_tuple_5.T2 -}}
{{- $v_7 := $tmp_tuple_5.T1 -}}
{{- $ak_9 := $values.storage.tiered.credentialsSecretRef.accessKey -}}
{{- if (and $ok_8 (ne $v_7 "")) -}}
{{- (dict "r" (list (mustMergeOverwrite (dict "name" "" ) (dict "name" "RPK_CLOUD_STORAGE_ACCESS_KEY" "value" (get (fromJson (include "_shims.typeassertion" (dict "a" (list "string" $v_7) ))) "r") )))) | toJson -}}
{{- break -}}
{{- else -}}{{- if (get (fromJson (include "redpanda.SecretRef.IsValid" (dict "a" (list $ak_9) ))) "r") -}}
{{- (dict "r" (list (mustMergeOverwrite (dict "name" "" ) (dict "name" "RPK_CLOUD_STORAGE_ACCESS_KEY" "valueFrom" (mustMergeOverwrite (dict ) (dict "secretKeyRef" (mustMergeOverwrite (dict "key" "" ) (mustMergeOverwrite (dict ) (dict "name" $ak_9.name )) (dict "key" $ak_9.key )) )) )))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}
{{- (dict "r" (list )) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.addCloudStorageSecretKey" -}}
{{- $tieredStorageConfig := (index .a 0) -}}
{{- $values := (index .a 1) -}}
{{- range $_ := (list 1) -}}
{{- $tmp_tuple_6 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "_shims.dicttest" (dict "a" (list $tieredStorageConfig "cloud_storage_secret_key" (coalesce nil)) ))) "r")) ))) "r") -}}
{{- $ok_11 := $tmp_tuple_6.T2 -}}
{{- $v_10 := $tmp_tuple_6.T1 -}}
{{- $sk_12 := $values.storage.tiered.credentialsSecretRef.secretKey -}}
{{- if (and $ok_11 (ne $v_10 "")) -}}
{{- (dict "r" (list (mustMergeOverwrite (dict "name" "" ) (dict "name" "RPK_CLOUD_STORAGE_SECRET_KEY" "value" (get (fromJson (include "_shims.typeassertion" (dict "a" (list "string" $v_10) ))) "r") )))) | toJson -}}
{{- break -}}
{{- else -}}{{- if (get (fromJson (include "redpanda.SecretRef.IsValid" (dict "a" (list $sk_12) ))) "r") -}}
{{- (dict "r" (list (mustMergeOverwrite (dict "name" "" ) (dict "name" "RPK_CLOUD_STORAGE_SECRET_KEY" "valueFrom" (mustMergeOverwrite (dict ) (dict "secretKeyRef" (mustMergeOverwrite (dict "key" "" ) (mustMergeOverwrite (dict ) (dict "name" $sk_12.name )) (dict "key" $sk_12.key )) )) )))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}
{{- (dict "r" (list )) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.addAzureSharedKey" -}}
{{- $tieredStorageConfig := (index .a 0) -}}
{{- $values := (index .a 1) -}}
{{- range $_ := (list 1) -}}
{{- $tmp_tuple_7 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "_shims.dicttest" (dict "a" (list $tieredStorageConfig "cloud_storage_azure_shared_key" (coalesce nil)) ))) "r")) ))) "r") -}}
{{- $ok_14 := $tmp_tuple_7.T2 -}}
{{- $v_13 := $tmp_tuple_7.T1 -}}
{{- $sk_15 := $values.storage.tiered.credentialsSecretRef.secretKey -}}
{{- if (and $ok_14 (ne $v_13 "")) -}}
{{- (dict "r" (list (mustMergeOverwrite (dict "name" "" ) (dict "name" "RPK_CLOUD_STORAGE_AZURE_SHARED_KEY" "value" (get (fromJson (include "_shims.typeassertion" (dict "a" (list "string" $v_13) ))) "r") )))) | toJson -}}
{{- break -}}
{{- else -}}{{- if (get (fromJson (include "redpanda.SecretRef.IsValid" (dict "a" (list $sk_15) ))) "r") -}}
{{- (dict "r" (list (mustMergeOverwrite (dict "name" "" ) (dict "name" "RPK_CLOUD_STORAGE_AZURE_SHARED_KEY" "valueFrom" (mustMergeOverwrite (dict ) (dict "secretKeyRef" (mustMergeOverwrite (dict "key" "" ) (mustMergeOverwrite (dict ) (dict "name" $sk_15.name )) (dict "key" $sk_15.key )) )) )))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}
{{- (dict "r" (list )) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.GetLicenseLiteral" -}}
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

{{- define "redpanda.GetLicenseSecretReference" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $values := $dot.Values.AsMap -}}
{{- if (not (empty $values.enterprise.licenseSecretRef)) -}}
{{- (dict "r" (mustMergeOverwrite (dict "key" "" ) (mustMergeOverwrite (dict ) (dict "name" $values.enterprise.licenseSecretRef.name )) (dict "key" $values.enterprise.licenseSecretRef.key ))) | toJson -}}
{{- break -}}
{{- else -}}{{- if (not (empty $values.license_secret_ref)) -}}
{{- (dict "r" (mustMergeOverwrite (dict "key" "" ) (mustMergeOverwrite (dict ) (dict "name" $values.license_secret_ref.secret_name )) (dict "key" $values.license_secret_ref.secret_key ))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}
{{- (dict "r" (coalesce nil)) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

