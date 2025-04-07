{{- /* Generated from "secret.go" */ -}}

{{- define "console.Secret" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $values := $dot.Values.AsMap -}}
{{- if (not $values.secret.create) -}}
{{- $_is_returning = true -}}
{{- (dict "r" (coalesce nil)) | toJson -}}
{{- break -}}
{{- end -}}
{{- $jwtSigningKey := $values.secret.authentication.jwtSigningKey -}}
{{- if (eq $jwtSigningKey "") -}}
{{- $jwtSigningKey = (randAlphaNum (32 | int)) -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" (mustMergeOverwrite (dict "metadata" (dict "creationTimestamp" (coalesce nil) ) ) (mustMergeOverwrite (dict ) (dict "apiVersion" "v1" "kind" "Secret" )) (dict "metadata" (mustMergeOverwrite (dict "creationTimestamp" (coalesce nil) ) (dict "name" (get (fromJson (include "console.Fullname" (dict "a" (list $dot) ))) "r") "labels" (get (fromJson (include "console.Labels" (dict "a" (list $dot) ))) "r") "namespace" $dot.Release.Namespace )) "type" "Opaque" "stringData" (dict "kafka-sasl-password" (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $values.secret.kafka.saslPassword "") ))) "r") "kafka-sasl-aws-msk-iam-secret-key" (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $values.secret.kafka.awsMskIamSecretKey "") ))) "r") "kafka-tls-ca" (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $values.secret.kafka.tlsCa "") ))) "r") "kafka-tls-cert" (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $values.secret.kafka.tlsCert "") ))) "r") "kafka-tls-key" (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $values.secret.kafka.tlsKey "") ))) "r") "schema-registry-password" (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $values.secret.schemaRegistry.password "") ))) "r") "schemaregistry-tls-ca" (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $values.secret.schemaRegistry.tlsCa "") ))) "r") "schemaregistry-tls-cert" (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $values.secret.schemaRegistry.tlsCert "") ))) "r") "schemaregistry-tls-key" (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $values.secret.schemaRegistry.tlsKey "") ))) "r") "authentication-jwt-signingkey" $jwtSigningKey "authentication-oidc-client-secret" (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $values.secret.authentication.oidc.clientSecret "") ))) "r") "license" $values.secret.license "redpanda-admin-api-password" (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $values.secret.redpanda.adminApi.password "") ))) "r") "redpanda-admin-api-tls-ca" (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $values.secret.redpanda.adminApi.tlsCa "") ))) "r") "redpanda-admin-api-tls-cert" (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $values.secret.redpanda.adminApi.tlsCert "") ))) "r") "redpanda-admin-api-tls-key" (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $values.secret.redpanda.adminApi.tlsKey "") ))) "r") "serde-protobuf-git-basicauth-password" (get (fromJson (include "_shims.ptr_Deref" (dict "a" (list $values.secret.serde.protobufGitBasicAuthPassword "") ))) "r") ) ))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

