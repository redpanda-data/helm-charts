{{- /* Generated from "post_install_upgrade_job.go" */ -}}

{{- define "redpanda.bootstrapYamlTemplater" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $values := $dot.Values.AsMap -}}
{{- $env := (get (fromJson (include "redpanda.TieredStorageCredentials.AsEnvVars" (dict "a" (list $values.storage.tiered.credentialsSecretRef (get (fromJson (include "redpanda.Storage.GetTieredStorageConfig" (dict "a" (list $values.storage) ))) "r")) ))) "r") -}}
{{- $image := (printf `%s:%s` $values.statefulset.sideCars.controllers.image.repository $values.statefulset.sideCars.controllers.image.tag) -}}
{{- $_is_returning = true -}}
{{- (dict "r" (mustMergeOverwrite (dict "name" "" "resources" (dict ) ) (dict "name" "bootstrap-yaml-envsubst" "image" $image "command" (list "/redpanda-operator" "envsubst" "/tmp/base-config/bootstrap.yaml" "--output" "/tmp/config/.bootstrap.yaml") "env" $env "resources" (mustMergeOverwrite (dict ) (dict "limits" (dict "cpu" (get (fromJson (include "_shims.resource_MustParse" (dict "a" (list "100m") ))) "r") "memory" (get (fromJson (include "_shims.resource_MustParse" (dict "a" (list "125Mi") ))) "r") ) "requests" (dict "cpu" (get (fromJson (include "_shims.resource_MustParse" (dict "a" (list "100m") ))) "r") "memory" (get (fromJson (include "_shims.resource_MustParse" (dict "a" (list "125Mi") ))) "r") ) )) "securityContext" (mustMergeOverwrite (dict ) (dict "allowPrivilegeEscalation" false "readOnlyRootFilesystem" true "runAsNonRoot" true )) "volumeMounts" (list (mustMergeOverwrite (dict "name" "" "mountPath" "" ) (dict "name" "config" "mountPath" "/tmp/config/" )) (mustMergeOverwrite (dict "name" "" "mountPath" "" ) (dict "name" "base-config" "mountPath" "/tmp/base-config/" ))) ))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.PostInstallUpgradeJob" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $values := $dot.Values.AsMap -}}
{{- if (not $values.post_install_job.enabled) -}}
{{- $_is_returning = true -}}
{{- (dict "r" (coalesce nil)) | toJson -}}
{{- break -}}
{{- end -}}
{{- $image := (printf `%s:%s` $values.statefulset.sideCars.controllers.image.repository $values.statefulset.sideCars.controllers.image.tag) -}}
{{- $job := (mustMergeOverwrite (dict "metadata" (dict "creationTimestamp" (coalesce nil) ) "spec" (dict "template" (dict "metadata" (dict "creationTimestamp" (coalesce nil) ) "spec" (dict "containers" (coalesce nil) ) ) ) "status" (dict ) ) (mustMergeOverwrite (dict ) (dict "apiVersion" "batch/v1" "kind" "Job" )) (dict "metadata" (mustMergeOverwrite (dict "creationTimestamp" (coalesce nil) ) (dict "name" (printf "%s-configuration" (get (fromJson (include "redpanda.Fullname" (dict "a" (list $dot) ))) "r")) "namespace" $dot.Release.Namespace "labels" (merge (dict ) (default (dict ) $values.post_install_job.labels) (get (fromJson (include "redpanda.FullLabels" (dict "a" (list $dot) ))) "r")) "annotations" (merge (dict ) (default (dict ) $values.post_install_job.annotations) (dict "helm.sh/hook" "post-install,post-upgrade" "helm.sh/hook-delete-policy" "before-hook-creation" "helm.sh/hook-weight" "-5" )) )) "spec" (mustMergeOverwrite (dict "template" (dict "metadata" (dict "creationTimestamp" (coalesce nil) ) "spec" (dict "containers" (coalesce nil) ) ) ) (dict "template" (get (fromJson (include "redpanda.StrategicMergePatch" (dict "a" (list (get (fromJson (include "redpanda.StructuredTpl" (dict "a" (list $dot $values.post_install_job.podTemplate) ))) "r") (get (fromJson (include "redpanda.StrategicMergePatch" (dict "a" (list (get (fromJson (include "redpanda.StructuredTpl" (dict "a" (list $dot $values.podTemplate) ))) "r") (mustMergeOverwrite (dict "metadata" (dict "creationTimestamp" (coalesce nil) ) "spec" (dict "containers" (coalesce nil) ) ) (dict "metadata" (mustMergeOverwrite (dict "creationTimestamp" (coalesce nil) ) (dict "generateName" (printf "%s-post-" $dot.Release.Name) "labels" (merge (dict ) (dict "app.kubernetes.io/name" (get (fromJson (include "redpanda.Name" (dict "a" (list $dot) ))) "r") "app.kubernetes.io/instance" $dot.Release.Name "app.kubernetes.io/component" (printf "%.50s-post-install" (get (fromJson (include "redpanda.Name" (dict "a" (list $dot) ))) "r")) ) (default (dict ) $values.commonLabels)) )) "spec" (mustMergeOverwrite (dict "containers" (coalesce nil) ) (dict "restartPolicy" "Never" "initContainers" (list (get (fromJson (include "redpanda.bootstrapYamlTemplater" (dict "a" (list $dot) ))) "r")) "automountServiceAccountToken" false "containers" (list (mustMergeOverwrite (dict "name" "" "resources" (dict ) ) (dict "name" "post-install" "image" $image "env" (get (fromJson (include "redpanda.PostInstallUpgradeEnvironmentVariables" (dict "a" (list $dot) ))) "r") "command" (list "/redpanda-operator" "sync-cluster-config" "--users-directory" "/etc/secrets/users" "--redpanda-yaml" "/tmp/base-config/redpanda.yaml" "--bootstrap-yaml" "/tmp/config/.bootstrap.yaml") "volumeMounts" (concat (default (list ) (get (fromJson (include "redpanda.CommonMounts" (dict "a" (list $dot) ))) "r")) (list (mustMergeOverwrite (dict "name" "" "mountPath" "" ) (dict "name" "config" "mountPath" "/tmp/config" )) (mustMergeOverwrite (dict "name" "" "mountPath" "" ) (dict "name" "base-config" "mountPath" "/tmp/base-config" )))) ))) "volumes" (concat (default (list ) (get (fromJson (include "redpanda.CommonVolumes" (dict "a" (list $dot) ))) "r")) (list (mustMergeOverwrite (dict "name" "" ) (mustMergeOverwrite (dict ) (dict "configMap" (mustMergeOverwrite (dict ) (mustMergeOverwrite (dict ) (dict "name" (get (fromJson (include "redpanda.Fullname" (dict "a" (list $dot) ))) "r") )) (dict )) )) (dict "name" "base-config" )) (mustMergeOverwrite (dict "name" "" ) (mustMergeOverwrite (dict ) (dict "emptyDir" (mustMergeOverwrite (dict ) (dict )) )) (dict "name" "config" )))) "serviceAccountName" (get (fromJson (include "redpanda.ServiceAccountName" (dict "a" (list $dot) ))) "r") )) ))) ))) "r")) ))) "r") )) )) -}}
{{- $_is_returning = true -}}
{{- (dict "r" $job) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.PostInstallUpgradeEnvironmentVariables" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $envars := (list ) -}}
{{- $values := $dot.Values.AsMap -}}
{{- $license_1 := $values.enterprise.license -}}
{{- $secretReference_2 := $values.enterprise.licenseSecretRef -}}
{{- if (ne $license_1 "") -}}
{{- $envars = (concat (default (list ) $envars) (list (mustMergeOverwrite (dict "name" "" ) (dict "name" "REDPANDA_LICENSE" "value" $license_1 )))) -}}
{{- else -}}{{- if (ne (toJson $secretReference_2) "null") -}}
{{- $envars = (concat (default (list ) $envars) (list (mustMergeOverwrite (dict "name" "" ) (dict "name" "REDPANDA_LICENSE" "valueFrom" (mustMergeOverwrite (dict ) (dict "secretKeyRef" $secretReference_2 )) )))) -}}
{{- end -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" (get (fromJson (include "redpanda.bootstrapEnvVars" (dict "a" (list $dot $envars) ))) "r")) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

