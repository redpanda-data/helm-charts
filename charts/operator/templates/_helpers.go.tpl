{{- /* Generated from "helpers.go" */ -}}

{{- define "operator.Name" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $values := $dot.Values.AsMap -}}
{{- $name := (default $dot.Chart.Name $values.nameOverride) -}}
{{- $_is_returning = true -}}
{{- (dict "r" (get (fromJson (include "operator.cleanForK8s" (dict "a" (list $name)))) "r")) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "operator.Fullname" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $values := $dot.Values.AsMap -}}
{{- if (ne $values.fullnameOverride "") -}}
{{- $_is_returning = true -}}
{{- (dict "r" (get (fromJson (include "operator.cleanForK8s" (dict "a" (list $values.fullnameOverride)))) "r")) | toJson -}}
{{- break -}}
{{- end -}}
{{- $name := (default $dot.Chart.Name $values.nameOverride) -}}
{{- if (contains $name $dot.Release.Name) -}}
{{- $_is_returning = true -}}
{{- (dict "r" (get (fromJson (include "operator.cleanForK8s" (dict "a" (list $dot.Release.Name)))) "r")) | toJson -}}
{{- break -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" (get (fromJson (include "operator.cleanForK8s" (dict "a" (list (printf "%s-%s" $dot.Release.Name $name))))) "r")) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "operator.ChartName" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $chart := (printf "%s-%s" $dot.Chart.Name $dot.Chart.Version) -}}
{{- $_is_returning = true -}}
{{- (dict "r" (get (fromJson (include "operator.cleanForK8s" (dict "a" (list (replace "+" "_" $chart))))) "r")) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "operator.Labels" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $values := $dot.Values.AsMap -}}
{{- $labels := (dict "helm.sh/chart" (get (fromJson (include "operator.ChartName" (dict "a" (list $dot)))) "r") "app.kubernetes.io/managed-by" $dot.Release.Service) -}}
{{- if (ne $dot.Chart.AppVersion "") -}}
{{- $_ := (set $labels "app.kubernetes.io/version" $dot.Chart.AppVersion) -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" (merge (dict) $labels (get (fromJson (include "operator.SelectorLabels" (dict "a" (list $dot)))) "r") $values.commonLabels)) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "operator.SelectorLabels" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $_is_returning = true -}}
{{- (dict "r" (dict "app.kubernetes.io/name" (get (fromJson (include "operator.Name" (dict "a" (list $dot)))) "r") "app.kubernetes.io/instance" $dot.Release.Name)) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "operator.cleanForK8s" -}}
{{- $s := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $_is_returning = true -}}
{{- (dict "r" (trimSuffix "-" (trunc (63 | int) $s))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "operator.cleanForK8sWithSuffix" -}}
{{- $s := (index .a 0) -}}
{{- $suffix := (index .a 1) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $lengthToTruncate := ((sub (((add ((get (fromJson (include "_shims.len" (dict "a" (list $s)))) "r") | int) ((get (fromJson (include "_shims.len" (dict "a" (list $suffix)))) "r") | int)) | int)) (63 | int)) | int) -}}
{{- if (gt $lengthToTruncate (0 | int)) -}}
{{- $s = (trunc $lengthToTruncate $s) -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" (printf "%s-%s" $s $suffix)) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "operator.StrategicMergePatch" -}}
{{- $overrides := (index .a 0) -}}
{{- $original := (index .a 1) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- if (ne (toJson $overrides.metadata.labels) "null") -}}
{{- $_ := (set $original.metadata "labels" (merge (dict) $overrides.metadata.labels (default (dict) $original.metadata.labels))) -}}
{{- end -}}
{{- if (ne (toJson $overrides.metadata.annotations) "null") -}}
{{- $_ := (set $original.metadata "annotations" (merge (dict) $overrides.metadata.annotations (default (dict) $original.metadata.annotations))) -}}
{{- end -}}
{{- if (ne (toJson $overrides.spec.securityContext) "null") -}}
{{- $_ := (set $original.spec "securityContext" (merge (dict) $overrides.spec.securityContext (default (mustMergeOverwrite (dict) (dict)) $original.spec.securityContext))) -}}
{{- end -}}
{{- if (not (empty $overrides.spec.automountServiceAccountToken)) -}}
{{- $_ := (set $original.spec "automountServiceAccountToken" $overrides.spec.automountServiceAccountToken) -}}
{{- end -}}
{{- if (and (ne (toJson $overrides.spec.imagePullSecrets) "null") (gt ((get (fromJson (include "_shims.len" (dict "a" (list $overrides.spec.imagePullSecrets)))) "r") | int) (0 | int))) -}}
{{- $_ := (set $original.spec "imagePullSecrets" $overrides.spec.imagePullSecrets) -}}
{{- end -}}
{{- if (not (empty $overrides.spec.serviceAccountName)) -}}
{{- $_ := (set $original.spec "serviceAccountName" $overrides.spec.serviceAccountName) -}}
{{- end -}}
{{- if (not (empty $overrides.spec.nodeSelector)) -}}
{{- $_ := (set $original.spec "nodeSelector" (merge (dict) $overrides.spec.nodeSelector (default (dict) $original.spec.nodeSelector))) -}}
{{- end -}}
{{- if (ne (toJson $overrides.spec.affinity) "null") -}}
{{- $_ := (set $original.spec "affinity" (merge (dict) $overrides.spec.affinity (default (mustMergeOverwrite (dict) (dict)) $original.spec.affinity))) -}}
{{- end -}}
{{- if (and (ne (toJson $overrides.spec.topologySpreadConstraints) "null") (gt ((get (fromJson (include "_shims.len" (dict "a" (list $overrides.spec.topologySpreadConstraints)))) "r") | int) (0 | int))) -}}
{{- $_ := (set $original.spec "topologySpreadConstraints" $overrides.spec.topologySpreadConstraints) -}}
{{- end -}}
{{- if (and (ne (toJson $overrides.spec.volumes) "null") (gt ((get (fromJson (include "_shims.len" (dict "a" (list $overrides.spec.volumes)))) "r") | int) (0 | int))) -}}
{{- $newVolumes := (list) -}}
{{- $overrideVolumes := (dict) -}}
{{- range $i, $_ := $overrides.spec.volumes -}}
{{- $vol := (index $overrides.spec.volumes $i) -}}
{{- $_ := (set $overrideVolumes $vol.name $vol) -}}
{{- end -}}
{{- if $_is_returning -}}
{{- break -}}
{{- end -}}
{{- range $_, $vol := $original.spec.volumes -}}
{{- $_169_overrideVol_1_ok_2 := (get (fromJson (include "_shims.dicttest" (dict "a" (list $overrideVolumes $vol.name (dict "name" ""))))) "r") -}}
{{- $overrideVol_1 := (index $_169_overrideVol_1_ok_2 0) -}}
{{- $ok_2 := (index $_169_overrideVol_1_ok_2 1) -}}
{{- if $ok_2 -}}
{{- $newVolumes = (concat (default (list) $newVolumes) (list $overrideVol_1)) -}}
{{- $_ := (unset $overrideVolumes $vol.name) -}}
{{- continue -}}
{{- end -}}
{{- $newVolumes = (concat (default (list) $newVolumes) (list $vol)) -}}
{{- end -}}
{{- if $_is_returning -}}
{{- break -}}
{{- end -}}
{{- range $_, $vol := $overrideVolumes -}}
{{- $newVolumes = (concat (default (list) $newVolumes) (list $vol)) -}}
{{- end -}}
{{- if $_is_returning -}}
{{- break -}}
{{- end -}}
{{- $_ := (set $original.spec "volumes" $newVolumes) -}}
{{- end -}}
{{- $overrideContainers := (dict) -}}
{{- range $i, $_ := $overrides.spec.containers -}}
{{- $container := (index $overrides.spec.containers $i) -}}
{{- $_ := (set $overrideContainers (toString $container.name) $container) -}}
{{- end -}}
{{- if $_is_returning -}}
{{- break -}}
{{- end -}}
{{- if (not (empty $overrides.spec.restartPolicy)) -}}
{{- $_ := (set $original.spec "restartPolicy" $overrides.spec.restartPolicy) -}}
{{- end -}}
{{- if (ne (toJson $overrides.spec.terminationGracePeriodSeconds) "null") -}}
{{- $_ := (set $original.spec "terminationGracePeriodSeconds" $overrides.spec.terminationGracePeriodSeconds) -}}
{{- end -}}
{{- if (ne (toJson $overrides.spec.activeDeadlineSeconds) "null") -}}
{{- $_ := (set $original.spec "activeDeadlineSeconds" $overrides.spec.activeDeadlineSeconds) -}}
{{- end -}}
{{- if (not (empty $overrides.spec.dnsPolicy)) -}}
{{- $_ := (set $original.spec "dnsPolicy" $overrides.spec.dnsPolicy) -}}
{{- end -}}
{{- if (not (empty $overrides.spec.nodeName)) -}}
{{- $_ := (set $original.spec "nodeName" $overrides.spec.nodeName) -}}
{{- end -}}
{{- if (not (empty $overrides.spec.hostNetwork)) -}}
{{- $_ := (set $original.spec "hostNetwork" $overrides.spec.hostNetwork) -}}
{{- end -}}
{{- if (not (empty $overrides.spec.hostPID)) -}}
{{- $_ := (set $original.spec "hostPID" $overrides.spec.hostPID) -}}
{{- end -}}
{{- if (not (empty $overrides.spec.hostIPC)) -}}
{{- $_ := (set $original.spec "hostIPC" $overrides.spec.hostIPC) -}}
{{- end -}}
{{- if (not (empty $overrides.spec.shareProcessNamespace)) -}}
{{- $_ := (set $original.spec "shareProcessNamespace" $overrides.spec.shareProcessNamespace) -}}
{{- end -}}
{{- if (not (empty $overrides.spec.hostname)) -}}
{{- $_ := (set $original.spec "hostname" $overrides.spec.hostname) -}}
{{- end -}}
{{- if (not (empty $overrides.spec.subdomain)) -}}
{{- $_ := (set $original.spec "subdomain" $overrides.spec.subdomain) -}}
{{- end -}}
{{- if (not (empty $overrides.spec.schedulerName)) -}}
{{- $_ := (set $original.spec "schedulerName" $overrides.spec.schedulerName) -}}
{{- end -}}
{{- if (and (ne (toJson $overrides.spec.tolerations) "null") (gt ((get (fromJson (include "_shims.len" (dict "a" (list $overrides.spec.tolerations)))) "r") | int) (0 | int))) -}}
{{- $_ := (set $original.spec "tolerations" $overrides.spec.tolerations) -}}
{{- end -}}
{{- if (and (ne (toJson $overrides.spec.hostAliases) "null") (gt ((get (fromJson (include "_shims.len" (dict "a" (list $overrides.spec.hostAliases)))) "r") | int) (0 | int))) -}}
{{- $_ := (set $original.spec "hostAliases" $overrides.spec.hostAliases) -}}
{{- end -}}
{{- if (not (empty $overrides.spec.priorityClassName)) -}}
{{- $_ := (set $original.spec "priorityClassName" $overrides.spec.priorityClassName) -}}
{{- end -}}
{{- if (not (empty $overrides.spec.priority)) -}}
{{- $_ := (set $original.spec "priority" $overrides.spec.priority) -}}
{{- end -}}
{{- if (ne (toJson $overrides.spec.dnsConfig) "null") -}}
{{- $_ := (set $original.spec "dnsConfig" (merge (dict) $overrides.spec.dnsConfig (default (mustMergeOverwrite (dict) (dict)) $original.spec.dnsConfig))) -}}
{{- end -}}
{{- if (and (ne (toJson $overrides.spec.readinessGates) "null") (gt ((get (fromJson (include "_shims.len" (dict "a" (list $overrides.spec.readinessGates)))) "r") | int) (0 | int))) -}}
{{- $_ := (set $original.spec "readinessGates" $overrides.spec.readinessGates) -}}
{{- end -}}
{{- if (not (empty $overrides.spec.runtimeClassName)) -}}
{{- $_ := (set $original.spec "runtimeClassName" $overrides.spec.runtimeClassName) -}}
{{- end -}}
{{- if (not (empty $overrides.spec.enableServiceLinks)) -}}
{{- $_ := (set $original.spec "enableServiceLinks" $overrides.spec.enableServiceLinks) -}}
{{- end -}}
{{- if (ne (toJson $overrides.spec.preemptionPolicy) "null") -}}
{{- $_ := (set $original.spec "preemptionPolicy" $overrides.spec.preemptionPolicy) -}}
{{- end -}}
{{- if (ne (toJson $overrides.spec.setHostnameAsFQDN) "null") -}}
{{- $_ := (set $original.spec "setHostnameAsFQDN" $overrides.spec.setHostnameAsFQDN) -}}
{{- end -}}
{{- if (ne (toJson $overrides.spec.hostUsers) "null") -}}
{{- $_ := (set $original.spec "hostUsers" $overrides.spec.hostUsers) -}}
{{- end -}}
{{- if (and (ne (toJson $overrides.spec.schedulingGates) "null") (gt ((get (fromJson (include "_shims.len" (dict "a" (list $overrides.spec.schedulingGates)))) "r") | int) (0 | int))) -}}
{{- $_ := (set $original.spec "schedulingGates" $overrides.spec.schedulingGates) -}}
{{- end -}}
{{- if (and (ne (toJson $overrides.spec.resourceClaims) "null") (gt ((get (fromJson (include "_shims.len" (dict "a" (list $overrides.spec.resourceClaims)))) "r") | int) (0 | int))) -}}
{{- $_ := (set $original.spec "resourceClaims" $overrides.spec.resourceClaims) -}}
{{- end -}}
{{- $merged := (coalesce nil) -}}
{{- range $_, $container := $original.spec.containers -}}
{{- $_308_override_3_ok_4 := (get (fromJson (include "_shims.dicttest" (dict "a" (list $overrideContainers $container.name (coalesce nil))))) "r") -}}
{{- $override_3 := (index $_308_override_3_ok_4 0) -}}
{{- $ok_4 := (index $_308_override_3_ok_4 1) -}}
{{- if $ok_4 -}}
{{- $env := (concat (default (list) $container.env) (default (list) $override_3.env)) -}}
{{- $container = (merge (dict) $override_3 $container) -}}
{{- $_ := (set $container "env" $env) -}}
{{- end -}}
{{- if (eq (toJson $container.env) "null") -}}
{{- $_ := (set $container "env" (list)) -}}
{{- end -}}
{{- $merged = (concat (default (list) $merged) (list $container)) -}}
{{- end -}}
{{- if $_is_returning -}}
{{- break -}}
{{- end -}}
{{- $_ := (set $original.spec "containers" $merged) -}}
{{- $overrideContainers = (dict) -}}
{{- range $i, $_ := $overrides.spec.initContainers -}}
{{- $container := (index $overrides.spec.initContainers $i) -}}
{{- $_ := (set $overrideContainers (toString $container.name) $container) -}}
{{- end -}}
{{- if $_is_returning -}}
{{- break -}}
{{- end -}}
{{- $merged = (list) -}}
{{- range $_, $container := $original.spec.initContainers -}}
{{- $_339_override_5_ok_6 := (get (fromJson (include "_shims.dicttest" (dict "a" (list $overrideContainers $container.name (coalesce nil))))) "r") -}}
{{- $override_5 := (index $_339_override_5_ok_6 0) -}}
{{- $ok_6 := (index $_339_override_5_ok_6 1) -}}
{{- if $ok_6 -}}
{{- $env := (concat (default (list) $container.env) (default (list) $override_5.env)) -}}
{{- $container = (merge (dict) $override_5 $container) -}}
{{- $_ := (set $container "env" $env) -}}
{{- end -}}
{{- if (eq (toJson $container.env) "null") -}}
{{- $_ := (set $container "env" (list)) -}}
{{- end -}}
{{- $merged = (concat (default (list) $merged) (list $container)) -}}
{{- end -}}
{{- if $_is_returning -}}
{{- break -}}
{{- end -}}
{{- $_ := (set $original.spec "initContainers" $merged) -}}
{{- $overrideEphemeralContainers := (dict) -}}
{{- range $i, $_ := $overrides.spec.ephemeralContainers -}}
{{- $container := (index $overrides.spec.ephemeralContainers $i) -}}
{{- $_ := (set $overrideEphemeralContainers (toString $container.name) $container) -}}
{{- end -}}
{{- if $_is_returning -}}
{{- break -}}
{{- end -}}
{{- $mergedEphemeralContainers := (coalesce nil) -}}
{{- range $_, $container := $original.spec.ephemeralContainers -}}
{{- $_370_override_7_ok_8 := (get (fromJson (include "_shims.dicttest" (dict "a" (list $overrideEphemeralContainers $container.name (coalesce nil))))) "r") -}}
{{- $override_7 := (index $_370_override_7_ok_8 0) -}}
{{- $ok_8 := (index $_370_override_7_ok_8 1) -}}
{{- if $ok_8 -}}
{{- $env := (concat (default (list) $container.env) (default (list) $override_7.env)) -}}
{{- $container = (merge (dict) $override_7 $container) -}}
{{- $_ := (set $container "env" $env) -}}
{{- end -}}
{{- if (eq (toJson $container.env) "null") -}}
{{- $_ := (set $container "env" (list)) -}}
{{- end -}}
{{- $mergedEphemeralContainers = (concat (default (list) $mergedEphemeralContainers) (list $container)) -}}
{{- end -}}
{{- if $_is_returning -}}
{{- break -}}
{{- end -}}
{{- $_ := (set $original.spec "ephemeralContainers" $mergedEphemeralContainers) -}}
{{- $_is_returning = true -}}
{{- (dict "r" $original) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

