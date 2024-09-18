{{- /* Generated from "console.tpl.go" */ -}}

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
{{- if (eq $values.console.console (coalesce nil)) -}}
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

