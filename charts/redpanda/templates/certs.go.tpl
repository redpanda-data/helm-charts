{{- /* Generated from "certs.go" */ -}}

{{- define "redpanda.ClientCerts" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- if (not (get (fromJson (include "redpanda.TLSEnabled" (dict "a" (list $dot) ))) "r")) -}}
{{- (dict "r" (list )) | toJson -}}
{{- break -}}
{{- end -}}
{{- $values := $dot.Values.AsMap -}}
{{- $fullname := (get (fromJson (include "redpanda.Fullname" (dict "a" (list $dot) ))) "r") -}}
{{- $service := (get (fromJson (include "redpanda.ServiceName" (dict "a" (list $dot) ))) "r") -}}
{{- $ns := $dot.Release.Namespace -}}
{{- $domain := (trimSuffix "." $values.clusterDomain) -}}
{{- $certs := (list ) -}}
{{- range $name, $data := $values.tls.certs -}}
{{- if (not (empty $data.secretRef)) -}}
{{- continue -}}
{{- end -}}
{{- $names := (list ) -}}
{{- if (or (eq $data.issuerRef (coalesce nil)) (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $data.applyInternalDNSNames false) ))) "r")) -}}
{{- $names = (mustAppend $names (printf "%s-cluster.%s.%s.svc.%s" $fullname $service $ns $domain)) -}}
{{- $names = (mustAppend $names (printf "%s-cluster.%s.%s.svc" $fullname $service $ns)) -}}
{{- $names = (mustAppend $names (printf "%s-cluster.%s.%s" $fullname $service $ns)) -}}
{{- $names = (mustAppend $names (printf "*.%s-cluster.%s.%s.svc.%s" $fullname $service $ns $domain)) -}}
{{- $names = (mustAppend $names (printf "*.%s-cluster.%s.%s.svc" $fullname $service $ns)) -}}
{{- $names = (mustAppend $names (printf "*.%s-cluster.%s.%s" $fullname $service $ns)) -}}
{{- $names = (mustAppend $names (printf "%s.%s.svc.%s" $service $ns $domain)) -}}
{{- $names = (mustAppend $names (printf "%s.%s.svc" $service $ns)) -}}
{{- $names = (mustAppend $names (printf "%s.%s" $service $ns)) -}}
{{- $names = (mustAppend $names (printf "*.%s.%s.svc.%s" $service $ns $domain)) -}}
{{- $names = (mustAppend $names (printf "*.%s.%s.svc" $service $ns)) -}}
{{- $names = (mustAppend $names (printf "*.%s.%s" $service $ns)) -}}
{{- end -}}
{{- if (ne $values.external.domain (coalesce nil)) -}}
{{- $names = (mustAppend $names (tpl $values.external.domain $dot)) -}}
{{- $names = (mustAppend $names (tpl (printf "*.%s" $values.external.domain) $dot)) -}}
{{- end -}}
{{- $duration := (default "43800h" $data.duration) -}}
{{- $issuerRef := (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $data.issuerRef (mustMergeOverwrite (dict "name" "" ) (dict "kind" "Issuer" "group" "cert-manager.io" "name" (printf "%s-%s-root-issuer" $fullname $name) ))) ))) "r") -}}
{{- $certs = (mustAppend $certs (mustMergeOverwrite (mustMergeOverwrite (dict ) (dict "metadata" (dict "creationTimestamp" (coalesce nil) ) "spec" (dict "secretName" "" "issuerRef" (dict "name" "" ) ) "status" (dict ) )) (mustMergeOverwrite (dict ) (dict "apiVersion" "cert-manager.io/v1" "kind" "Certificate" )) (dict "metadata" (mustMergeOverwrite (dict "creationTimestamp" (coalesce nil) ) (dict "name" (printf "%s-%s-cert" $fullname $name) "labels" (get (fromJson (include "redpanda.FullLabels" (dict "a" (list $dot) ))) "r") "namespace" $dot.Release.Namespace )) "spec" (mustMergeOverwrite (dict "secretName" "" "issuerRef" (dict "name" "" ) ) (dict "dnsNames" $names "duration" $duration "isCA" false "issuerRef" $issuerRef "secretName" (printf "%s-%s-cert" $fullname $name) "privateKey" (mustMergeOverwrite (dict ) (dict "algorithm" "ECDSA" "size" 256 )) )) ))) -}}
{{- end -}}
{{- $name := $values.listeners.kafka.tls.cert -}}
{{- $tmp_tuple_1 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "_shims.dicttest" (dict "a" (list $values.tls.certs $name (coalesce nil)) ))) "r")) ))) "r") -}}
{{- $ok := $tmp_tuple_1.T2 -}}
{{- $data := $tmp_tuple_1.T1 -}}
{{- if (or (not $ok) ((and (empty $data.secretRef) (not (get (fromJson (include "redpanda.ClientAuthRequired" (dict "a" (list $dot) ))) "r"))))) -}}
{{- (dict "r" $certs) | toJson -}}
{{- break -}}
{{- end -}}
{{- $issuerRef := (mustMergeOverwrite (dict "name" "" ) (dict "group" "cert-manager.io" "kind" "Issuer" "name" (printf "%s-%s-root-issuer" $fullname $name) )) -}}
{{- if (ne $data.issuerRef (coalesce nil)) -}}
{{- $issuerRef = $data.issuerRef -}}
{{- $_ := (set $issuerRef "group" "cert-manager.io") -}}
{{- end -}}
{{- $duration := (default "43800h" $data.duration) -}}
{{- (dict "r" (mustAppend $certs (mustMergeOverwrite (mustMergeOverwrite (dict ) (dict "metadata" (dict "creationTimestamp" (coalesce nil) ) "spec" (dict "secretName" "" "issuerRef" (dict "name" "" ) ) "status" (dict ) )) (mustMergeOverwrite (dict ) (dict "apiVersion" "cert-manager.io/v1" "kind" "Certificate" )) (dict "metadata" (mustMergeOverwrite (dict "creationTimestamp" (coalesce nil) ) (dict "name" (printf "%s-client" $fullname) "labels" (get (fromJson (include "redpanda.FullLabels" (dict "a" (list $dot) ))) "r") )) "spec" (mustMergeOverwrite (dict "secretName" "" "issuerRef" (dict "name" "" ) ) (dict "commonName" (printf "%s-client" $fullname) "duration" $duration "isCA" false "secretName" (printf "%s-client" $fullname) "privateKey" (mustMergeOverwrite (dict ) (dict "algorithm" "ECDSA" "size" 256 )) "issuerRef" $issuerRef )) )))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

