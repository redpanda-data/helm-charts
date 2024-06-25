{{- /* Generated from "post_install_upgrade_job.go" */ -}}

{{- define "redpanda.PostInstallUpgradeJob" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $values := $dot.Values.AsMap -}}
{{- if (not $values.post_install_job.enabled) -}}
{{- (dict "r" (coalesce nil)) | toJson -}}
{{- break -}}
{{- end -}}
{{- $job := (mustMergeOverwrite (dict "metadata" (dict "creationTimestamp" (coalesce nil) ) "spec" (dict "template" (dict "metadata" (dict "creationTimestamp" (coalesce nil) ) "spec" (dict "containers" (coalesce nil) ) ) ) "status" (dict ) ) (mustMergeOverwrite (dict ) (dict "apiVersion" "batch/v1" "kind" "Job" )) (dict "metadata" (mustMergeOverwrite (dict "creationTimestamp" (coalesce nil) ) (dict "name" (printf "%s-configuration" (get (fromJson (include "redpanda.Fullname" (dict "a" (list $dot) ))) "r")) "namespace" $dot.Release.Namespace "labels" (merge (dict ) (get (fromJson (include "redpanda.FullLabels" (dict "a" (list $dot) ))) "r") (default (dict ) $values.post_install_job.labels)) "annotations" (merge (dict ) (dict "helm.sh/hook" "post-install,post-upgrade" "helm.sh/hook-delete-policy" "before-hook-creation" "helm.sh/hook-weight" "-5" ) (default (dict ) $values.post_install_job.annotations)) )) "spec" (mustMergeOverwrite (dict "template" (dict "metadata" (dict "creationTimestamp" (coalesce nil) ) "spec" (dict "containers" (coalesce nil) ) ) ) (dict "template" (mustMergeOverwrite (dict "metadata" (dict "creationTimestamp" (coalesce nil) ) "spec" (dict "containers" (coalesce nil) ) ) (dict "metadata" (mustMergeOverwrite (dict "creationTimestamp" (coalesce nil) ) (dict "generateName" (printf "%s-post-" $dot.Release.Name) "labels" (merge (dict ) (dict "app.kubernetes.io/name" (get (fromJson (include "redpanda.Name" (dict "a" (list $dot) ))) "r") "app.kubernetes.io/instance" $dot.Release.Name "app.kubernetes.io/component" (printf "%.50s-post-install" (get (fromJson (include "redpanda.Name" (dict "a" (list $dot) ))) "r")) ) (default (dict ) $values.commonLabels)) )) "spec" (mustMergeOverwrite (dict "containers" (coalesce nil) ) (dict "nodeSelector" $values.nodeSelector "affinity" (get (fromJson (include "redpanda.postInstallJobAffinity" (dict "a" (list $dot) ))) "r") "tolerations" (get (fromJson (include "redpanda.tolerations" (dict "a" (list $dot) ))) "r") "restartPolicy" "Never" "securityContext" (get (fromJson (include "redpanda.PodSecurityContext" (dict "a" (list $dot) ))) "r") "imagePullSecrets" (get (fromJson (include "redpanda.pullSecrets" (dict "a" (list $dot) ))) "r") "containers" (list (mustMergeOverwrite (dict "name" "" "resources" (dict ) ) (dict "name" (printf "%s-post-install" (get (fromJson (include "redpanda.Name" (dict "a" (list $dot) ))) "r")) "image" (printf "%s:%s" $values.image.repository (get (fromJson (include "redpanda.Tag" (dict "a" (list $dot) ))) "r")) "env" (get (fromJson (include "redpanda.PostInstallUpgradeEnvironmentVariables" (dict "a" (list $dot) ))) "r") "command" (list "bash" "-c") "args" (list ) "resources" (get (fromJson (include "redpanda.postInstallJobResources" (dict "a" (list $values.post_install_job.resources) ))) "r") "securityContext" (merge (dict ) (default (mustMergeOverwrite (dict ) (dict )) $values.post_install_job.securityContext) (get (fromJson (include "redpanda.ContainerSecurityContext" (dict "a" (list $dot) ))) "r")) "volumeMounts" (get (fromJson (include "redpanda.DefaultMounts" (dict "a" (list $dot) ))) "r") ))) "volumes" (get (fromJson (include "redpanda.DefaultVolumes" (dict "a" (list $dot) ))) "r") "serviceAccountName" (get (fromJson (include "redpanda.ServiceAccountName" (dict "a" (list $dot) ))) "r") )) )) )) )) -}}
{{- $script := (coalesce nil) -}}
{{- $script = (concat (default (list ) $script) (list `set -e`)) -}}
{{- if (get (fromJson (include "redpanda.RedpandaAtLeast_22_2_0" (dict "a" (list $dot) ))) "r") -}}
{{- $script = (concat (default (list ) $script) (list `if [[ -n "$REDPANDA_LICENSE" ]] then` `  rpk cluster license set "$REDPANDA_LICENSE"` `fi`)) -}}
{{- end -}}
{{- $script = (concat (default (list ) $script) (list `` `` `` `` `rpk cluster config export -f /tmp/cfg.yml` `` `` `for KEY in "${!RPK_@}"; do` `  config="${KEY#*RPK_}"` `  rpk redpanda config set --config /tmp/cfg.yml "${config,,}" "${!KEY}"` `done` `` `` `rpk cluster config import -f /tmp/cfg.yml` ``)) -}}
{{- $_ := (set (index $job.spec.template.spec.containers (0 | int)) "args" (concat (default (list ) (index $job.spec.template.spec.containers (0 | int)).args) (list (get (fromJson (include "redpanda.unlines" (dict "a" (list $script) ))) "r")))) -}}
{{- (dict "r" $job) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.postInstallJobAffinity" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $values := $dot.Values.AsMap -}}
{{- $affinity := (dict ) -}}
{{- if (not (empty $values.post_install_job.affinity)) -}}
{{- $affinity = (merge (dict ) $values.post_install_job.affinity) -}}
{{- (dict "r" $affinity) | toJson -}}
{{- break -}}
{{- end -}}
{{- $_ := (set $affinity "nodeAffinity" (merge (dict ) (default (dict ) $values.affinity.nodeAffinity))) -}}
{{- $_ := (set $affinity "podAffinity" (merge (dict ) (default (dict ) $values.affinity.podAffinity))) -}}
{{- $_ := (set $affinity "podAntiAffinity" (merge (dict ) (default (dict ) $values.affinity.podAntiAffinity))) -}}
{{- (dict "r" $affinity) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.tolerations" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $values := $dot.Values.AsMap -}}
{{- $result := (coalesce nil) -}}
{{- range $_, $t := $values.tolerations -}}
{{- $result = (concat (default (list ) $result) (list (merge (dict ) $t))) -}}
{{- end -}}
{{- (dict "r" $result) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.postInstallJobResources" -}}
{{- $resources := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- (dict "r" (merge (dict ) $resources)) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.pullSecrets" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $values := $dot.Values.AsMap -}}
{{- $result := (coalesce nil) -}}
{{- range $_, $r := $values.imagePullSecrets -}}
{{- $result = (concat (default (list ) $result) (list (mustMergeOverwrite (dict ) (dict "name" $r )))) -}}
{{- end -}}
{{- (dict "r" $result) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

