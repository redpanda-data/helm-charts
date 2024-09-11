{{- /* Generated from "helpers.go" */ -}}

{{- define "redpanda.ChartLabel" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $_is_returning = true -}}
{{- (dict "r" (get (fromJson (include "redpanda.cleanForK8s" (dict "a" (list (replace "+" "_" (printf "%s-%s" $dot.Chart.Name $dot.Chart.Version))) ))) "r")) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.Name" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $tmp_tuple_1 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "_shims.typetest" (dict "a" (list "string" (index $dot.Values "nameOverride") "") ))) "r")) ))) "r") -}}
{{- $ok_2 := $tmp_tuple_1.T2 -}}
{{- $override_1 := $tmp_tuple_1.T1 -}}
{{- if (and $ok_2 (ne $override_1 "")) -}}
{{- $_is_returning = true -}}
{{- (dict "r" (get (fromJson (include "redpanda.cleanForK8s" (dict "a" (list $override_1) ))) "r")) | toJson -}}
{{- break -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" (get (fromJson (include "redpanda.cleanForK8s" (dict "a" (list $dot.Chart.Name) ))) "r")) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.Fullname" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $tmp_tuple_2 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "_shims.typetest" (dict "a" (list "string" (index $dot.Values "fullnameOverride") "") ))) "r")) ))) "r") -}}
{{- $ok_4 := $tmp_tuple_2.T2 -}}
{{- $override_3 := $tmp_tuple_2.T1 -}}
{{- if (and $ok_4 (ne $override_3 "")) -}}
{{- $_is_returning = true -}}
{{- (dict "r" (get (fromJson (include "redpanda.cleanForK8s" (dict "a" (list $override_3) ))) "r")) | toJson -}}
{{- break -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" (get (fromJson (include "redpanda.cleanForK8s" (dict "a" (list $dot.Release.Name) ))) "r")) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.FullLabels" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $values := $dot.Values.AsMap -}}
{{- $labels := (dict ) -}}
{{- if (ne $values.commonLabels (coalesce nil)) -}}
{{- $labels = $values.commonLabels -}}
{{- end -}}
{{- $defaults := (dict "helm.sh/chart" (get (fromJson (include "redpanda.ChartLabel" (dict "a" (list $dot) ))) "r") "app.kubernetes.io/name" (get (fromJson (include "redpanda.Name" (dict "a" (list $dot) ))) "r") "app.kubernetes.io/instance" $dot.Release.Name "app.kubernetes.io/managed-by" $dot.Release.Service "app.kubernetes.io/component" (get (fromJson (include "redpanda.Name" (dict "a" (list $dot) ))) "r") ) -}}
{{- $_is_returning = true -}}
{{- (dict "r" (merge (dict ) $labels $defaults)) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.ServiceAccountName" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $values := $dot.Values.AsMap -}}
{{- $serviceAccount := $values.serviceAccount -}}
{{- if (and $serviceAccount.create (ne $serviceAccount.name "")) -}}
{{- $_is_returning = true -}}
{{- (dict "r" $serviceAccount.name) | toJson -}}
{{- break -}}
{{- else -}}{{- if $serviceAccount.create -}}
{{- $_is_returning = true -}}
{{- (dict "r" (get (fromJson (include "redpanda.Fullname" (dict "a" (list $dot) ))) "r")) | toJson -}}
{{- break -}}
{{- else -}}{{- if (ne $serviceAccount.name "") -}}
{{- $_is_returning = true -}}
{{- (dict "r" $serviceAccount.name) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" "default") | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.Tag" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $values := $dot.Values.AsMap -}}
{{- $tag := (toString $values.image.tag) -}}
{{- if (eq $tag "") -}}
{{- $tag = $dot.Chart.AppVersion -}}
{{- end -}}
{{- $pattern := "^v(0|[1-9]\\d*)\\.(0|[1-9]\\d*)\\.(0|[1-9]\\d*)(?:-((?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\\.(?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\\+([0-9a-zA-Z-]+(?:\\.[0-9a-zA-Z-]+)*))?$" -}}
{{- if (not (regexMatch $pattern $tag)) -}}
{{- $_ := (fail "image.tag must start with a 'v' and be a valid semver") -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" $tag) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.ServiceName" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $values := $dot.Values.AsMap -}}
{{- if (and (ne $values.service (coalesce nil)) (ne $values.service.name (coalesce nil))) -}}
{{- $_is_returning = true -}}
{{- (dict "r" (get (fromJson (include "redpanda.cleanForK8s" (dict "a" (list $values.service.name) ))) "r")) | toJson -}}
{{- break -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" (get (fromJson (include "redpanda.Fullname" (dict "a" (list $dot) ))) "r")) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.InternalDomain" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $values := $dot.Values.AsMap -}}
{{- $service := (get (fromJson (include "redpanda.ServiceName" (dict "a" (list $dot) ))) "r") -}}
{{- $ns := $dot.Release.Namespace -}}
{{- $domain := (trimSuffix "." $values.clusterDomain) -}}
{{- $_is_returning = true -}}
{{- (dict "r" (printf "%s.%s.svc.%s." $service $ns $domain)) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.TLSEnabled" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $values := $dot.Values.AsMap -}}
{{- if $values.tls.enabled -}}
{{- $_is_returning = true -}}
{{- (dict "r" true) | toJson -}}
{{- break -}}
{{- end -}}
{{- $listeners := (list "kafka" "admin" "schemaRegistry" "rpc" "http") -}}
{{- range $_, $listener := $listeners -}}
{{- $tlsCert := (dig "listeners" $listener "tls" "cert" false $dot.Values.AsMap) -}}
{{- $tlsEnabled := (dig "listeners" $listener "tls" "enabled" false $dot.Values.AsMap) -}}
{{- if (and (not (empty $tlsEnabled)) (not (empty $tlsCert))) -}}
{{- $_is_returning = true -}}
{{- (dict "r" true) | toJson -}}
{{- break -}}
{{- end -}}
{{- $external := (dig "listeners" $listener "external" false $dot.Values.AsMap) -}}
{{- if (empty $external) -}}
{{- continue -}}
{{- end -}}
{{- $keys := (keys (get (fromJson (include "_shims.typeassertion" (dict "a" (list (printf "map[%s]%s" "string" "interface {}") $external) ))) "r")) -}}
{{- range $_, $key := $keys -}}
{{- $enabled := (dig "listeners" $listener "external" $key "enabled" false $dot.Values.AsMap) -}}
{{- $tlsCert := (dig "listeners" $listener "external" $key "tls" "cert" false $dot.Values.AsMap) -}}
{{- $tlsEnabled := (dig "listeners" $listener "external" $key "tls" "enabled" false $dot.Values.AsMap) -}}
{{- if (and (and (not (empty $enabled)) (not (empty $tlsCert))) (not (empty $tlsEnabled))) -}}
{{- $_is_returning = true -}}
{{- (dict "r" true) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}
{{- if $_is_returning -}}
{{- break -}}
{{- end -}}
{{- end -}}
{{- if $_is_returning -}}
{{- break -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" false) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.ClientAuthRequired" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $listeners := (list "kafka" "admin" "schemaRegistry" "rpc" "http") -}}
{{- range $_, $listener := $listeners -}}
{{- $required := (dig "listeners" $listener "tls" "requireClientAuth" false $dot.Values.AsMap) -}}
{{- if (not (empty $required)) -}}
{{- $_is_returning = true -}}
{{- (dict "r" true) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}
{{- if $_is_returning -}}
{{- break -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" false) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.DefaultMounts" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $_is_returning = true -}}
{{- (dict "r" (concat (default (list ) (list (mustMergeOverwrite (dict "name" "" "mountPath" "" ) (dict "name" "base-config" "mountPath" "/etc/redpanda" )))) (default (list ) (get (fromJson (include "redpanda.CommonMounts" (dict "a" (list $dot) ))) "r")))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.CommonMounts" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $values := $dot.Values.AsMap -}}
{{- $mounts := (list ) -}}
{{- $sasl_5 := $values.auth.sasl -}}
{{- if (and $sasl_5.enabled (ne $sasl_5.secretRef "")) -}}
{{- $mounts = (concat (default (list ) $mounts) (list (mustMergeOverwrite (dict "name" "" "mountPath" "" ) (dict "name" "users" "mountPath" "/etc/secrets/users" "readOnly" true )))) -}}
{{- end -}}
{{- if (get (fromJson (include "redpanda.TLSEnabled" (dict "a" (list $dot) ))) "r") -}}
{{- $certNames := (keys $values.tls.certs) -}}
{{- $_ := (sortAlpha $certNames) -}}
{{- range $_, $name := $certNames -}}
{{- $cert := (index $values.tls.certs $name) -}}
{{- if (not (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $cert.enabled true) ))) "r")) -}}
{{- continue -}}
{{- end -}}
{{- $mounts = (concat (default (list ) $mounts) (list (mustMergeOverwrite (dict "name" "" "mountPath" "" ) (dict "name" (printf "redpanda-%s-cert" $name) "mountPath" (printf "/etc/tls/certs/%s" $name) )))) -}}
{{- end -}}
{{- if $_is_returning -}}
{{- break -}}
{{- end -}}
{{- $adminTLS := $values.listeners.admin.tls -}}
{{- if $adminTLS.requireClientAuth -}}
{{- $mounts = (concat (default (list ) $mounts) (list (mustMergeOverwrite (dict "name" "" "mountPath" "" ) (dict "name" "mtls-client" "mountPath" (printf "/etc/tls/certs/%s-client" (get (fromJson (include "redpanda.Fullname" (dict "a" (list $dot) ))) "r")) )))) -}}
{{- end -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" $mounts) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.DefaultVolumes" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $_is_returning = true -}}
{{- (dict "r" (concat (default (list ) (list (mustMergeOverwrite (dict "name" "" ) (mustMergeOverwrite (dict ) (dict "configMap" (mustMergeOverwrite (dict ) (mustMergeOverwrite (dict ) (dict "name" (get (fromJson (include "redpanda.Fullname" (dict "a" (list $dot) ))) "r") )) (dict )) )) (dict "name" "base-config" )))) (default (list ) (get (fromJson (include "redpanda.CommonVolumes" (dict "a" (list $dot) ))) "r")))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.CommonVolumes" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $volumes := (list ) -}}
{{- $values := $dot.Values.AsMap -}}
{{- if (get (fromJson (include "redpanda.TLSEnabled" (dict "a" (list $dot) ))) "r") -}}
{{- $certNames := (keys $values.tls.certs) -}}
{{- $_ := (sortAlpha $certNames) -}}
{{- range $_, $name := $certNames -}}
{{- $cert := (index $values.tls.certs $name) -}}
{{- if (not (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $cert.enabled true) ))) "r")) -}}
{{- continue -}}
{{- end -}}
{{- $volumes = (concat (default (list ) $volumes) (list (mustMergeOverwrite (dict "name" "" ) (mustMergeOverwrite (dict ) (dict "secret" (mustMergeOverwrite (dict ) (dict "secretName" (get (fromJson (include "redpanda.CertSecretName" (dict "a" (list $dot $name $cert) ))) "r") "defaultMode" (0o440 | int) )) )) (dict "name" (printf "redpanda-%s-cert" $name) )))) -}}
{{- end -}}
{{- if $_is_returning -}}
{{- break -}}
{{- end -}}
{{- $adminTLS := $values.listeners.admin.tls -}}
{{- $cert := (index $values.tls.certs $adminTLS.cert) -}}
{{- if $adminTLS.requireClientAuth -}}
{{- $secretName := (printf "%s-client" (get (fromJson (include "redpanda.Fullname" (dict "a" (list $dot) ))) "r")) -}}
{{- if (ne $cert.clientSecretRef (coalesce nil)) -}}
{{- $secretName = $cert.clientSecretRef.name -}}
{{- end -}}
{{- $volumes = (concat (default (list ) $volumes) (list (mustMergeOverwrite (dict "name" "" ) (mustMergeOverwrite (dict ) (dict "secret" (mustMergeOverwrite (dict ) (dict "secretName" $secretName "defaultMode" (0o440 | int) )) )) (dict "name" "mtls-client" )))) -}}
{{- end -}}
{{- end -}}
{{- $sasl_6 := $values.auth.sasl -}}
{{- if (and $sasl_6.enabled (ne $sasl_6.secretRef "")) -}}
{{- $volumes = (concat (default (list ) $volumes) (list (mustMergeOverwrite (dict "name" "" ) (mustMergeOverwrite (dict ) (dict "secret" (mustMergeOverwrite (dict ) (dict "secretName" $sasl_6.secretRef )) )) (dict "name" "users" )))) -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" $volumes) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.CertSecretName" -}}
{{- $dot := (index .a 0) -}}
{{- $certName := (index .a 1) -}}
{{- $cert := (index .a 2) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- if (ne $cert.secretRef (coalesce nil)) -}}
{{- $_is_returning = true -}}
{{- (dict "r" $cert.secretRef.name) | toJson -}}
{{- break -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" (printf "%s-%s-cert" (get (fromJson (include "redpanda.Fullname" (dict "a" (list $dot) ))) "r") $certName)) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.PodSecurityContext" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $values := $dot.Values.AsMap -}}
{{- $sc := (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $values.statefulset.podSecurityContext $values.statefulset.securityContext) ))) "r") -}}
{{- $_is_returning = true -}}
{{- (dict "r" (mustMergeOverwrite (dict ) (dict "fsGroup" $sc.fsGroup "fsGroupChangePolicy" $sc.fsGroupChangePolicy ))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.ContainerSecurityContext" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $values := $dot.Values.AsMap -}}
{{- $sc := (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $values.statefulset.podSecurityContext $values.statefulset.securityContext) ))) "r") -}}
{{- $_is_returning = true -}}
{{- (dict "r" (mustMergeOverwrite (dict ) (dict "runAsUser" $sc.runAsUser "runAsGroup" (get (fromJson (include "redpanda.coalesce" (dict "a" (list (list $sc.runAsGroup $sc.fsGroup)) ))) "r") "allowPrivilegeEscalation" (get (fromJson (include "redpanda.coalesce" (dict "a" (list (list $sc.allowPrivilegeEscalation $sc.allowPriviledgeEscalation)) ))) "r") "runAsNonRoot" $sc.runAsNonRoot ))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.RedpandaAtLeast_22_2_0" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $_is_returning = true -}}
{{- (dict "r" (get (fromJson (include "redpanda.redpandaAtLeast" (dict "a" (list $dot ">=22.2.0-0 || <0.0.1-0") ))) "r")) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.RedpandaAtLeast_22_3_0" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $_is_returning = true -}}
{{- (dict "r" (get (fromJson (include "redpanda.redpandaAtLeast" (dict "a" (list $dot ">=22.3.0-0 || <0.0.1-0") ))) "r")) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.RedpandaAtLeast_23_1_1" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $_is_returning = true -}}
{{- (dict "r" (get (fromJson (include "redpanda.redpandaAtLeast" (dict "a" (list $dot ">=23.1.1-0 || <0.0.1-0") ))) "r")) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.RedpandaAtLeast_23_1_2" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $_is_returning = true -}}
{{- (dict "r" (get (fromJson (include "redpanda.redpandaAtLeast" (dict "a" (list $dot ">=23.1.2-0 || <0.0.1-0") ))) "r")) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.RedpandaAtLeast_22_3_atleast_22_3_13" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $_is_returning = true -}}
{{- (dict "r" (get (fromJson (include "redpanda.redpandaAtLeast" (dict "a" (list $dot ">=22.3.13-0,<22.4") ))) "r")) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.RedpandaAtLeast_22_2_atleast_22_2_10" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $_is_returning = true -}}
{{- (dict "r" (get (fromJson (include "redpanda.redpandaAtLeast" (dict "a" (list $dot ">=22.2.10-0,<22.3") ))) "r")) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.RedpandaAtLeast_23_2_1" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $_is_returning = true -}}
{{- (dict "r" (get (fromJson (include "redpanda.redpandaAtLeast" (dict "a" (list $dot ">=23.2.1-0 || <0.0.1-0") ))) "r")) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.RedpandaAtLeast_23_3_0" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $_is_returning = true -}}
{{- (dict "r" (get (fromJson (include "redpanda.redpandaAtLeast" (dict "a" (list $dot ">=23.3.0-0 || <0.0.1-0") ))) "r")) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.redpandaAtLeast" -}}
{{- $dot := (index .a 0) -}}
{{- $constraint := (index .a 1) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $version := (trimPrefix "v" (get (fromJson (include "redpanda.Tag" (dict "a" (list $dot) ))) "r")) -}}
{{- $tmp_tuple_3 := (get (fromJson (include "_shims.compact" (dict "a" (list (list (semverCompare $constraint $version) nil)) ))) "r") -}}
{{- $err := $tmp_tuple_3.T2 -}}
{{- $result := $tmp_tuple_3.T1 -}}
{{- if (ne $err (coalesce nil)) -}}
{{- $_ := (fail $err) -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" $result) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.cleanForK8s" -}}
{{- $in := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $_is_returning = true -}}
{{- (dict "r" (trimSuffix "-" (trunc (63 | int) $in))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.RedpandaSMP" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $values := $dot.Values.AsMap -}}
{{- $coresInMillies := ((get (fromJson (include "_shims.resource_MilliValue" (dict "a" (list $values.resources.cpu.cores) ))) "r") | int64) -}}
{{- if (lt $coresInMillies (1000 | int64)) -}}
{{- $_is_returning = true -}}
{{- (dict "r" (1 | int64)) | toJson -}}
{{- break -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" ((get (fromJson (include "_shims.resource_Value" (dict "a" (list $values.resources.cpu.cores) ))) "r") | int64)) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.coalesce" -}}
{{- $values := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- range $_, $v := $values -}}
{{- if (ne $v (coalesce nil)) -}}
{{- $_is_returning = true -}}
{{- (dict "r" $v) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}
{{- if $_is_returning -}}
{{- break -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" (coalesce nil)) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.StrategicMergePatch" -}}
{{- $overrides := (index .a 0) -}}
{{- $original := (index .a 1) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- if (ne $overrides.labels (coalesce nil)) -}}
{{- $_ := (set $original.metadata "labels" (merge (dict ) $overrides.labels (default (dict ) $original.metadata.labels))) -}}
{{- end -}}
{{- if (ne $overrides.annotations (coalesce nil)) -}}
{{- $_ := (set $original.metadata "annotations" (merge (dict ) $overrides.annotations (default (dict ) $original.metadata.annotations))) -}}
{{- end -}}
{{- if (ne $overrides.spec.securityContext (coalesce nil)) -}}
{{- $_ := (set $original.spec "securityContext" (merge (dict ) $overrides.spec.securityContext (default (mustMergeOverwrite (dict ) (dict )) $original.spec.securityContext))) -}}
{{- end -}}
{{- $overrideContainers := (dict ) -}}
{{- range $i, $_ := $overrides.spec.containers -}}
{{- $container := (index $overrides.spec.containers $i) -}}
{{- $_ := (set $overrideContainers (toString $container.name) $container) -}}
{{- end -}}
{{- if $_is_returning -}}
{{- break -}}
{{- end -}}
{{- $merged := (coalesce nil) -}}
{{- range $_, $container := $original.spec.containers -}}
{{- $tmp_tuple_4 := (get (fromJson (include "_shims.compact" (dict "a" (list (get (fromJson (include "_shims.dicttest" (dict "a" (list $overrideContainers $container.name (coalesce nil)) ))) "r")) ))) "r") -}}
{{- $ok_8 := $tmp_tuple_4.T2 -}}
{{- $override_7 := $tmp_tuple_4.T1 -}}
{{- if $ok_8 -}}
{{- $env := (concat (default (list ) $container.env) (default (list ) $override_7.env)) -}}
{{- $container = (merge (dict ) $override_7 $container) -}}
{{- $_ := (set $container "env" $env) -}}
{{- end -}}
{{- if (eq $container.env (coalesce nil)) -}}
{{- $_ := (set $container "env" (list )) -}}
{{- end -}}
{{- $merged = (concat (default (list ) $merged) (list $container)) -}}
{{- end -}}
{{- if $_is_returning -}}
{{- break -}}
{{- end -}}
{{- $_ := (set $original.spec "containers" $merged) -}}
{{- $_is_returning = true -}}
{{- (dict "r" $original) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

