{{- /* Generated from "console.tpl.go" */ -}}

{{- define "redpanda.renderConsole" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $v := $dot.Values -}}
{{- if (not (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $v.console.enabled true) ))) "r")) -}}
{{- $_is_returning = true -}}
{{- (dict "r" (coalesce nil)) | toJson -}}
{{- break -}}
{{- end -}}
{{- $consoleDot := (index $dot.Subcharts "console") -}}
{{- $consoleValue := $consoleDot.Values -}}
{{- if (not (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $v.console.secret.create false) ))) "r")) -}}
{{- $_ := (set $consoleValue.secret "create" true) -}}
{{- $license_1 := (get (fromJson (include "redpanda.GetLicenseLiteral" (dict "a" (list $dot) ))) "r") -}}
{{- if (ne $license_1 "") -}}
{{- $_ := (set $consoleValue.secret "enterprise" (mustMergeOverwrite (dict ) (dict "license" $license_1 ))) -}}
{{- end -}}
{{- end -}}
{{- if (not (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $v.console.configmap.create false) ))) "r")) -}}
{{- $_ := (set $consoleValue.configmap "create" true) -}}
{{- $_ := (set $consoleValue.console "config" (get (fromJson (include "redpanda.ConsoleConfig" (dict "a" (list $dot) ))) "r")) -}}
{{- end -}}
{{- if (not (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $v.console.deployment.create false) ))) "r")) -}}
{{- $_ := (set $consoleValue.deployment "create" true) -}}
{{- $extraVolumes := (list ) -}}
{{- $extraVolumeMounts := (list ) -}}
{{- $extraEnvVars := (list ) -}}
{{- if (get (fromJson (include "redpanda.Auth.IsSASLEnabled" (dict "a" (list $v.auth) ))) "r") -}}
{{- $command := (concat (default (list ) (list )) (list "sh" "-c" (printf "%s%s" (printf "%s%s" (printf "%s%s" (printf "%s%s" (printf "%s%s" (printf "%s%s" (printf "%s%s" "set -e; IFS=':' read -r KAFKA_SASL_USERNAME KAFKA_SASL_PASSWORD KAFKA_SASL_MECHANISM < <(grep \"\" $(find /mnt/users/* -print));" (printf " KAFKA_SASL_MECHANISM=${KAFKA_SASL_MECHANISM:-%s};" (get (fromJson (include "redpanda.SASLMechanism" (dict "a" (list $dot) ))) "r"))) " export KAFKA_SASL_USERNAME KAFKA_SASL_PASSWORD KAFKA_SASL_MECHANISM;") " export KAFKA_SCHEMAREGISTRY_USERNAME=$KAFKA_SASL_USERNAME;") " export KAFKA_SCHEMAREGISTRY_PASSWORD=$KAFKA_SASL_PASSWORD;") " export REDPANDA_ADMINAPI_USERNAME=$KAFKA_SASL_USERNAME;") " export REDPANDA_ADMINAPI_PASSWORD=$KAFKA_SASL_PASSWORD;") " /app/console $@") " --")) -}}
{{- $_ := (set $consoleValue.deployment "command" $command) -}}
{{- $extraVolumes = (concat (default (list ) $extraVolumes) (list (mustMergeOverwrite (dict "name" "" ) (mustMergeOverwrite (dict ) (dict "secret" (mustMergeOverwrite (dict ) (dict "secretName" $v.auth.sasl.secretRef )) )) (dict "name" (printf "%s-users" (get (fromJson (include "redpanda.Fullname" (dict "a" (list $dot) ))) "r")) )))) -}}
{{- $extraVolumeMounts = (concat (default (list ) $extraVolumeMounts) (list (mustMergeOverwrite (dict "name" "" "mountPath" "" ) (dict "name" (printf "%s-users" (get (fromJson (include "redpanda.Fullname" (dict "a" (list $dot) ))) "r")) "mountPath" "/mnt/users" "readOnly" true )))) -}}
{{- end -}}
{{- if (get (fromJson (include "redpanda.InternalTLS.IsEnabled" (dict "a" (list $v.listeners.kafka.tls $v.tls) ))) "r") -}}
{{- $certName := $v.listeners.kafka.tls.cert -}}
{{- $cert := (get (fromJson (include "redpanda.TLSCertMap.MustGet" (dict "a" (list (deepCopy $v.tls.certs) $certName) ))) "r") -}}
{{- $secretName := (printf "%s-%s-cert" (get (fromJson (include "redpanda.Fullname" (dict "a" (list $dot) ))) "r") $certName) -}}
{{- if (ne (toJson $cert.secretRef) "null") -}}
{{- $secretName = $cert.secretRef.name -}}
{{- end -}}
{{- if $cert.caEnabled -}}
{{- $extraEnvVars = (concat (default (list ) $extraEnvVars) (list (mustMergeOverwrite (dict "name" "" ) (dict "name" "KAFKA_TLS_CAFILEPATH" "value" (printf "/mnt/cert/kafka/%s/ca.crt" $certName) )))) -}}
{{- $extraVolumes = (concat (default (list ) $extraVolumes) (list (mustMergeOverwrite (dict "name" "" ) (mustMergeOverwrite (dict ) (dict "secret" (mustMergeOverwrite (dict ) (dict "defaultMode" (0o420 | int) "secretName" $secretName )) )) (dict "name" (printf "kafka-%s-cert" $certName) )))) -}}
{{- $extraVolumeMounts = (concat (default (list ) $extraVolumeMounts) (list (mustMergeOverwrite (dict "name" "" "mountPath" "" ) (dict "name" (printf "kafka-%s-cert" $certName) "mountPath" (printf "/mnt/cert/kafka/%s" $certName) "readOnly" true )))) -}}
{{- end -}}
{{- end -}}
{{- if (get (fromJson (include "redpanda.InternalTLS.IsEnabled" (dict "a" (list $v.listeners.schemaRegistry.tls $v.tls) ))) "r") -}}
{{- $certName := $v.listeners.schemaRegistry.tls.cert -}}
{{- $cert := (get (fromJson (include "redpanda.TLSCertMap.MustGet" (dict "a" (list (deepCopy $v.tls.certs) $certName) ))) "r") -}}
{{- $secretName := (printf "%s-%s-cert" (get (fromJson (include "redpanda.Fullname" (dict "a" (list $dot) ))) "r") $certName) -}}
{{- if (ne (toJson $cert.secretRef) "null") -}}
{{- $secretName = $cert.secretRef.name -}}
{{- end -}}
{{- if $cert.caEnabled -}}
{{- $extraEnvVars = (concat (default (list ) $extraEnvVars) (list (mustMergeOverwrite (dict "name" "" ) (dict "name" "KAFKA_SCHEMAREGISTRY_TLS_CAFILEPATH" "value" (printf "/mnt/cert/schemaregistry/%s/ca.crt" $certName) )))) -}}
{{- $extraVolumes = (concat (default (list ) $extraVolumes) (list (mustMergeOverwrite (dict "name" "" ) (mustMergeOverwrite (dict ) (dict "secret" (mustMergeOverwrite (dict ) (dict "defaultMode" (0o420 | int) "secretName" $secretName )) )) (dict "name" (printf "schemaregistry-%s-cert" $certName) )))) -}}
{{- $extraVolumeMounts = (concat (default (list ) $extraVolumeMounts) (list (mustMergeOverwrite (dict "name" "" "mountPath" "" ) (dict "name" (printf "schemaregistry-%s-cert" $certName) "mountPath" (printf "/mnt/cert/schemaregistry/%s" $certName) "readOnly" true )))) -}}
{{- end -}}
{{- end -}}
{{- if (get (fromJson (include "redpanda.InternalTLS.IsEnabled" (dict "a" (list $v.listeners.admin.tls $v.tls) ))) "r") -}}
{{- $certName := $v.listeners.admin.tls.cert -}}
{{- $cert := (get (fromJson (include "redpanda.TLSCertMap.MustGet" (dict "a" (list (deepCopy $v.tls.certs) $certName) ))) "r") -}}
{{- $secretName := (printf "%s-%s-cert" (get (fromJson (include "redpanda.Fullname" (dict "a" (list $dot) ))) "r") $certName) -}}
{{- if (ne (toJson $cert.secretRef) "null") -}}
{{- $secretName = $cert.secretRef.name -}}
{{- end -}}
{{- if $cert.caEnabled -}}
{{- $extraVolumes = (concat (default (list ) $extraVolumes) (list (mustMergeOverwrite (dict "name" "" ) (mustMergeOverwrite (dict ) (dict "secret" (mustMergeOverwrite (dict ) (dict "defaultMode" (0o420 | int) "secretName" $secretName )) )) (dict "name" (printf "adminapi-%s-cert" $certName) )))) -}}
{{- $extraVolumeMounts = (concat (default (list ) $extraVolumeMounts) (list (mustMergeOverwrite (dict "name" "" "mountPath" "" ) (dict "name" (printf "adminapi-%s-cert" $certName) "mountPath" (printf "/mnt/cert/adminapi/%s" $certName) "readOnly" true )))) -}}
{{- end -}}
{{- end -}}
{{- $secret_2 := (get (fromJson (include "redpanda.GetLicenseSecretReference" (dict "a" (list $dot) ))) "r") -}}
{{- if (ne (toJson $secret_2) "null") -}}
{{- $_ := (set $consoleValue "enterprise" (mustMergeOverwrite (dict "licenseSecretRef" (dict "name" "" "key" "" ) ) (dict "licenseSecretRef" (mustMergeOverwrite (dict "name" "" "key" "" ) (dict "name" $secret_2.name "key" $secret_2.key )) ))) -}}
{{- end -}}
{{- $_ := (set $consoleValue "extraEnv" $extraEnvVars) -}}
{{- $_ := (set $consoleValue "extraVolumes" $extraVolumes) -}}
{{- $_ := (set $consoleValue "extraVolumeMounts" $extraVolumeMounts) -}}
{{- $_ := (set $consoleDot "Values" $consoleValue) -}}
{{- $cfg := (get (fromJson (include "console.ConfigMap" (dict "a" (list $consoleDot) ))) "r") -}}
{{- if (eq (toJson $consoleValue.podAnnotations) "null") -}}
{{- $_ := (set $consoleValue "podAnnotations" (dict )) -}}
{{- end -}}
{{- $_ := (set $consoleValue.podAnnotations "checksum-redpanda-chart/config" (sha256sum (toYaml $cfg))) -}}
{{- end -}}
{{- $_ := (set $consoleDot "Values" $consoleValue) -}}
{{- $manifests := (list (get (fromJson (include "console.Secret" (dict "a" (list $consoleDot) ))) "r") (get (fromJson (include "console.ConfigMap" (dict "a" (list $consoleDot) ))) "r") (get (fromJson (include "console.Deployment" (dict "a" (list $consoleDot) ))) "r")) -}}
{{- $_is_returning = true -}}
{{- (dict "r" $manifests) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.ConsoleConfig" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $values := $dot.Values.AsMap -}}
{{- $schemaURLs := (coalesce nil) -}}
{{- if $values.listeners.schemaRegistry.enabled -}}
{{- $schema := "http" -}}
{{- if (get (fromJson (include "redpanda.InternalTLS.IsEnabled" (dict "a" (list $values.listeners.schemaRegistry.tls $values.tls) ))) "r") -}}
{{- $schema = "https" -}}
{{- end -}}
{{- range $_, $i := untilStep (((0 | int) | int)|int) (($values.statefulset.replicas | int)|int) (1|int) -}}
{{- $schemaURLs = (concat (default (list ) $schemaURLs) (list (printf "%s://%s-%d.%s:%d" $schema (get (fromJson (include "redpanda.Fullname" (dict "a" (list $dot) ))) "r") $i (get (fromJson (include "redpanda.InternalDomain" (dict "a" (list $dot) ))) "r") ($values.listeners.schemaRegistry.port | int)))) -}}
{{- end -}}
{{- if $_is_returning -}}
{{- break -}}
{{- end -}}
{{- end -}}
{{- $schema := "http" -}}
{{- if (get (fromJson (include "redpanda.InternalTLS.IsEnabled" (dict "a" (list $values.listeners.admin.tls $values.tls) ))) "r") -}}
{{- $schema = "https" -}}
{{- end -}}
{{- $c := (dict "kafka" (dict "brokers" (get (fromJson (include "redpanda.BrokerList" (dict "a" (list $dot ($values.statefulset.replicas | int) ($values.listeners.kafka.port | int)) ))) "r") "sasl" (dict "enabled" (get (fromJson (include "redpanda.Auth.IsSASLEnabled" (dict "a" (list $values.auth) ))) "r") ) "tls" (get (fromJson (include "redpanda.KafkaListeners.ConsoleTLS" (dict "a" (list $values.listeners.kafka $values.tls) ))) "r") "schemaRegistry" (dict "enabled" $values.listeners.schemaRegistry.enabled "urls" $schemaURLs "tls" (get (fromJson (include "redpanda.SchemaRegistryListeners.ConsoleTLS" (dict "a" (list $values.listeners.schemaRegistry $values.tls) ))) "r") ) ) "redpanda" (dict "adminApi" (dict "enabled" true "urls" (list (printf "%s://%s:%d" $schema (get (fromJson (include "redpanda.InternalDomain" (dict "a" (list $dot) ))) "r") ($values.listeners.admin.port | int))) "tls" (get (fromJson (include "redpanda.AdminListeners.ConsoleTLS" (dict "a" (list $values.listeners.admin $values.tls) ))) "r") ) ) ) -}}
{{- if $values.connectors.enabled -}}
{{- $port := (dig "connectors" "connectors" "restPort" (8083 | int) $dot.Values.AsMap) -}}
{{- $tmp_tuple_1 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "_shims.asintegral" (dict "a" (list $port) ))) "r")) ))) "r") -}}
{{- $ok := $tmp_tuple_1.T2 -}}
{{- $p := ($tmp_tuple_1.T1 | int) -}}
{{- if (not $ok) -}}
{{- $_is_returning = true -}}
{{- (dict "r" $c) | toJson -}}
{{- break -}}
{{- end -}}
{{- $connectorsURL := (printf "http://%s.%s.svc.%s:%d" (get (fromJson (include "redpanda.ConnectorsFullName" (dict "a" (list $dot) ))) "r") $dot.Release.Namespace (trimSuffix "." $values.clusterDomain) $p) -}}
{{- $_ := (set $c "connect" (mustMergeOverwrite (dict "enabled" false "clusters" (coalesce nil) "connectTimeout" 0 "readTimeout" 0 "requestTimeout" 0 ) (dict "enabled" $values.connectors.enabled "clusters" (list (mustMergeOverwrite (dict "name" "" "url" "" "tls" (dict "enabled" false "caFilepath" "" "certFilepath" "" "keyFilepath" "" "insecureSkipTlsVerify" false ) "username" "" "password" "" "token" "" ) (dict "name" "connectors" "url" $connectorsURL "tls" (mustMergeOverwrite (dict "enabled" false "caFilepath" "" "certFilepath" "" "keyFilepath" "" "insecureSkipTlsVerify" false ) (dict "enabled" false "caFilepath" "" "certFilepath" "" "keyFilepath" "" "insecureSkipTlsVerify" false )) "username" "" "password" "" "token" "" ))) ))) -}}
{{- end -}}
{{- if (eq (toJson $values.console.console) "null") -}}
{{- $_ := (set $values.console "console" (mustMergeOverwrite (dict ) (dict "config" (dict ) ))) -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" (merge (dict ) $values.console.console.config $c)) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.ConnectorsFullName" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $values := $dot.Values.AsMap -}}
{{- if (ne (dig "connectors" "connectors" "fullnameOverwrite" "" $dot.Values.AsMap) "") -}}
{{- $_is_returning = true -}}
{{- (dict "r" (get (fromJson (include "redpanda.cleanForK8s" (dict "a" (list $values.connectors.connectors.fullnameOverwrite) ))) "r")) | toJson -}}
{{- break -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" (get (fromJson (include "redpanda.cleanForK8s" (dict "a" (list (printf "%s-connectors" $dot.Release.Name)) ))) "r")) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

