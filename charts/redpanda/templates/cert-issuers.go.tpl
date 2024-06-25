{{- /* Generated from "cert_issuers.go" */ -}}

{{- define "redpanda.CertIssuers" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $tmp_tuple_1 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "redpanda.certIssuersAndCAs" (dict "a" (list $dot) ))) "r")) ))) "r") -}}
{{- $issuers := $tmp_tuple_1.T1 -}}
{{- (dict "r" $issuers) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.RootCAs" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $tmp_tuple_2 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "redpanda.certIssuersAndCAs" (dict "a" (list $dot) ))) "r")) ))) "r") -}}
{{- $cas := $tmp_tuple_2.T2 -}}
{{- (dict "r" $cas) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.certIssuersAndCAs" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $values := $dot.Values.AsMap -}}
{{- $issuers := (coalesce nil) -}}
{{- $certs := (coalesce nil) -}}
{{- if (not (get (fromJson (include "redpanda.TLSEnabled" (dict "a" (list $dot) ))) "r")) -}}
{{- (dict "r" (list $issuers $certs)) | toJson -}}
{{- break -}}
{{- end -}}
{{- range $name, $data := $values.tls.certs -}}
{{- if (ne $data.secretRef (coalesce nil)) -}}
{{- continue -}}
{{- end -}}
{{- if (eq $data.issuerRef (coalesce nil)) -}}
{{- $issuers = (concat (default (list ) $issuers) (list (mustMergeOverwrite (dict "metadata" (dict "creationTimestamp" (coalesce nil) ) "spec" (dict ) "status" (dict ) ) (mustMergeOverwrite (dict ) (dict "apiVersion" "cert-manager.io/v1" "kind" "Issuer" )) (dict "metadata" (mustMergeOverwrite (dict "creationTimestamp" (coalesce nil) ) (dict "name" (printf `%s-%s-selfsigned-issuer` (get (fromJson (include "redpanda.Fullname" (dict "a" (list $dot) ))) "r") $name) "namespace" $dot.Release.Namespace "labels" (get (fromJson (include "redpanda.FullLabels" (dict "a" (list $dot) ))) "r") )) "spec" (mustMergeOverwrite (dict ) (mustMergeOverwrite (dict ) (dict "selfSigned" (mustMergeOverwrite (dict ) (dict )) )) (dict )) )))) -}}
{{- end -}}
{{- $issuers = (concat (default (list ) $issuers) (list (mustMergeOverwrite (dict "metadata" (dict "creationTimestamp" (coalesce nil) ) "spec" (dict ) "status" (dict ) ) (mustMergeOverwrite (dict ) (dict "apiVersion" "cert-manager.io/v1" "kind" "Issuer" )) (dict "metadata" (mustMergeOverwrite (dict "creationTimestamp" (coalesce nil) ) (dict "name" (printf `%s-%s-root-issuer` (get (fromJson (include "redpanda.Fullname" (dict "a" (list $dot) ))) "r") $name) "namespace" $dot.Release.Namespace "labels" (get (fromJson (include "redpanda.FullLabels" (dict "a" (list $dot) ))) "r") )) "spec" (mustMergeOverwrite (dict ) (mustMergeOverwrite (dict ) (dict "ca" (mustMergeOverwrite (dict "secretName" "" ) (dict "secretName" (printf `%s-%s-root-certificate` (get (fromJson (include "redpanda.Fullname" (dict "a" (list $dot) ))) "r") $name) )) )) (dict )) )))) -}}
{{- $certs = (concat (default (list ) $certs) (list (mustMergeOverwrite (dict "metadata" (dict "creationTimestamp" (coalesce nil) ) "spec" (dict "secretName" "" "issuerRef" (dict "name" "" ) ) "status" (dict ) ) (mustMergeOverwrite (dict ) (dict "apiVersion" "cert-manager.io/v1" "kind" "Certificate" )) (dict "metadata" (mustMergeOverwrite (dict "creationTimestamp" (coalesce nil) ) (dict "name" (printf `%s-%s-root-certificate` (get (fromJson (include "redpanda.Fullname" (dict "a" (list $dot) ))) "r") $name) "namespace" $dot.Release.Namespace "labels" (get (fromJson (include "redpanda.FullLabels" (dict "a" (list $dot) ))) "r") )) "spec" (mustMergeOverwrite (dict "secretName" "" "issuerRef" (dict "name" "" ) ) (dict "duration" (default "43800h" $data.duration) "isCA" true "commonName" (printf `%s-%s-root-certificate` (get (fromJson (include "redpanda.Fullname" (dict "a" (list $dot) ))) "r") $name) "secretName" (printf `%s-%s-root-certificate` (get (fromJson (include "redpanda.Fullname" (dict "a" (list $dot) ))) "r") $name) "privateKey" (mustMergeOverwrite (dict ) (dict "algorithm" "ECDSA" "size" (256 | int) )) "issuerRef" (mustMergeOverwrite (dict "name" "" ) (dict "name" (printf `%s-%s-selfsigned-issuer` (get (fromJson (include "redpanda.Fullname" (dict "a" (list $dot) ))) "r") $name) "kind" "Issuer" "group" "cert-manager.io" )) )) )))) -}}
{{- end -}}
{{- (dict "r" (list $issuers $certs)) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

