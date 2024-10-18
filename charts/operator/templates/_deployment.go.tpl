{{- /* Generated from "deployment.go" */ -}}

{{- define "operator.Deployment" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $values := $dot.Values.AsMap -}}
{{- $dep := (mustMergeOverwrite (dict "metadata" (dict "creationTimestamp" (coalesce nil) ) "spec" (dict "selector" (coalesce nil) "template" (dict "metadata" (dict "creationTimestamp" (coalesce nil) ) "spec" (dict "containers" (coalesce nil) ) ) "strategy" (dict ) ) "status" (dict ) ) (mustMergeOverwrite (dict ) (dict "apiVersion" "apps/v1" "kind" "Deployment" )) (dict "metadata" (mustMergeOverwrite (dict "creationTimestamp" (coalesce nil) ) (dict "name" (get (fromJson (include "operator.Fullname" (dict "a" (list $dot) ))) "r") "labels" (get (fromJson (include "operator.Labels" (dict "a" (list $dot) ))) "r") "namespace" $dot.Release.Namespace "annotations" $values.annotations )) "spec" (mustMergeOverwrite (dict "selector" (coalesce nil) "template" (dict "metadata" (dict "creationTimestamp" (coalesce nil) ) "spec" (dict "containers" (coalesce nil) ) ) "strategy" (dict ) ) (dict "replicas" ($values.replicaCount | int) "selector" (mustMergeOverwrite (dict ) (dict "matchLabels" (get (fromJson (include "operator.SelectorLabels" (dict "a" (list $dot) ))) "r") )) "strategy" $values.strategy "template" (get (fromJson (include "operator.StrategicMergePatch" (dict "a" (list (mustMergeOverwrite (dict "metadata" (dict "creationTimestamp" (coalesce nil) ) "spec" (dict "containers" (coalesce nil) ) ) (dict "metadata" (mustMergeOverwrite (dict "creationTimestamp" (coalesce nil) ) (dict "labels" $values.podTemplate.metadata.labels "annotations" $values.podTemplate.metadata.annotations )) "spec" $values.podTemplate.spec )) (mustMergeOverwrite (dict "metadata" (dict "creationTimestamp" (coalesce nil) ) "spec" (dict "containers" (coalesce nil) ) ) (dict "metadata" (mustMergeOverwrite (dict "creationTimestamp" (coalesce nil) ) (dict "annotations" $values.podAnnotations "labels" (merge (dict ) (get (fromJson (include "operator.SelectorLabels" (dict "a" (list $dot) ))) "r") $values.podLabels) )) "spec" (mustMergeOverwrite (dict "containers" (coalesce nil) ) (dict "terminationGracePeriodSeconds" ((10 | int64) | int64) "imagePullSecrets" $values.imagePullSecrets "serviceAccountName" (get (fromJson (include "operator.ServiceAccountName" (dict "a" (list $dot) ))) "r") "nodeSelector" $values.nodeSelector "tolerations" $values.tolerations "volumes" (get (fromJson (include "operator.operatorPodVolumes" (dict "a" (list $dot) ))) "r") "containers" (get (fromJson (include "operator.operatorContainers" (dict "a" (list $dot (coalesce nil)) ))) "r") )) ))) ))) "r") )) )) -}}
{{- if (not (empty $values.affinity)) -}}
{{- $_ := (set $dep.spec.template.spec "affinity" $values.affinity) -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" $dep) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "operator.operatorContainers" -}}
{{- $dot := (index .a 0) -}}
{{- $podTerminationGracePeriodSeconds := (index .a 1) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $values := $dot.Values.AsMap -}}
{{- $_is_returning = true -}}
{{- (dict "r" (list (mustMergeOverwrite (dict "name" "" "resources" (dict ) ) (dict "name" "manager" "image" (get (fromJson (include "operator.containerImage" (dict "a" (list $dot) ))) "r") "imagePullPolicy" $values.image.pullPolicy "command" (list "/manager") "args" (get (fromJson (include "operator.operatorArguments" (dict "a" (list $dot) ))) "r") "securityContext" (mustMergeOverwrite (dict ) (dict "allowPrivilegeEscalation" false )) "ports" (list (mustMergeOverwrite (dict "containerPort" 0 ) (dict "name" "webhook-server" "containerPort" (9443 | int) "protocol" "TCP" ))) "volumeMounts" (get (fromJson (include "operator.operatorPodVolumesMounts" (dict "a" (list $dot) ))) "r") "livenessProbe" (get (fromJson (include "operator.livenessProbe" (dict "a" (list $dot $podTerminationGracePeriodSeconds) ))) "r") "readinessProbe" (get (fromJson (include "operator.readinessProbe" (dict "a" (list $dot $podTerminationGracePeriodSeconds) ))) "r") "resources" $values.resources )) (mustMergeOverwrite (dict "name" "" "resources" (dict ) ) (dict "name" "kube-rbac-proxy" "args" (list "--secure-listen-address=0.0.0.0:8443" "--upstream=http://127.0.0.1:8080/" "--logtostderr=true" (printf "--v=%d" ($values.kubeRbacProxy.logLevel | int))) "image" (printf "%s:%s" $values.kubeRbacProxy.image.repository $values.kubeRbacProxy.image.tag) "imagePullPolicy" $values.kubeRbacProxy.image.pullPolicy "ports" (list (mustMergeOverwrite (dict "containerPort" 0 ) (dict "containerPort" (8443 | int) "name" "https" ))) )))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "operator.livenessProbe" -}}
{{- $dot := (index .a 0) -}}
{{- $podTerminationGracePeriodSeconds := (index .a 1) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $values := $dot.Values.AsMap -}}
{{- if (ne (toJson $values.livenessProbe) "null") -}}
{{- $_is_returning = true -}}
{{- (dict "r" (mustMergeOverwrite (dict ) (mustMergeOverwrite (dict ) (dict "httpGet" (mustMergeOverwrite (dict "port" 0 ) (dict "path" "/healthz/" "port" (8081 | int) )) )) (dict "initialDelaySeconds" (default (15 | int) ($values.livenessProbe.initialDelaySeconds | int)) "periodSeconds" (default (20 | int) ($values.livenessProbe.periodSeconds | int)) "timeoutSeconds" ($values.livenessProbe.timeoutSeconds | int) "successThreshold" ($values.livenessProbe.successThreshold | int) "failureThreshold" ($values.livenessProbe.failureThreshold | int) "terminationGracePeriodSeconds" (default $podTerminationGracePeriodSeconds $values.livenessProbe.terminationGracePeriodSeconds) ))) | toJson -}}
{{- break -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" (mustMergeOverwrite (dict ) (mustMergeOverwrite (dict ) (dict "httpGet" (mustMergeOverwrite (dict "port" 0 ) (dict "path" "/healthz/" "port" (8081 | int) )) )) (dict "initialDelaySeconds" (15 | int) "periodSeconds" (20 | int) ))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "operator.readinessProbe" -}}
{{- $dot := (index .a 0) -}}
{{- $podTerminationGracePeriodSeconds := (index .a 1) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $values := $dot.Values.AsMap -}}
{{- if (ne (toJson $values.livenessProbe) "null") -}}
{{- $_is_returning = true -}}
{{- (dict "r" (mustMergeOverwrite (dict ) (mustMergeOverwrite (dict ) (dict "httpGet" (mustMergeOverwrite (dict "port" 0 ) (dict "path" "/readyz" "port" (8081 | int) )) )) (dict "initialDelaySeconds" (default (5 | int) ($values.readinessProbe.initialDelaySeconds | int)) "periodSeconds" (default (10 | int) ($values.readinessProbe.periodSeconds | int)) "timeoutSeconds" ($values.readinessProbe.timeoutSeconds | int) "successThreshold" ($values.readinessProbe.successThreshold | int) "failureThreshold" ($values.readinessProbe.failureThreshold | int) "terminationGracePeriodSeconds" (default $podTerminationGracePeriodSeconds $values.readinessProbe.terminationGracePeriodSeconds) ))) | toJson -}}
{{- break -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" (mustMergeOverwrite (dict ) (mustMergeOverwrite (dict ) (dict "httpGet" (mustMergeOverwrite (dict "port" 0 ) (dict "path" "/readyz" "port" (8081 | int) )) )) (dict "initialDelaySeconds" (5 | int) "periodSeconds" (10 | int) ))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "operator.containerImage" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $values := $dot.Values.AsMap -}}
{{- $tag := $dot.Chart.AppVersion -}}
{{- if (not (empty $values.image.tag)) -}}
{{- $tag = $values.image.tag -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" (printf "%s:%s" $values.image.repository $tag)) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "operator.configuratorTag" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $values := $dot.Values.AsMap -}}
{{- if (not (empty $values.configurator.tag)) -}}
{{- $_is_returning = true -}}
{{- (dict "r" $values.configurator.tag) | toJson -}}
{{- break -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" $dot.Chart.AppVersion) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "operator.isWebhookEnabled" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $values := $dot.Values.AsMap -}}
{{- $_is_returning = true -}}
{{- (dict "r" (and $values.webhook.enabled (eq $values.scope "Cluster"))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "operator.operatorPodVolumes" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $values := $dot.Values.AsMap -}}
{{- if (not $values.webhook.enabled) -}}
{{- $_is_returning = true -}}
{{- (dict "r" (list )) | toJson -}}
{{- break -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" (list (mustMergeOverwrite (dict "name" "" ) (mustMergeOverwrite (dict ) (dict "secret" (mustMergeOverwrite (dict ) (dict "defaultMode" ((420 | int) | int) "secretName" $values.webhookSecretName )) )) (dict "name" "cert" )))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "operator.operatorPodVolumesMounts" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $values := $dot.Values.AsMap -}}
{{- if (not $values.webhook.enabled) -}}
{{- $_is_returning = true -}}
{{- (dict "r" (list )) | toJson -}}
{{- break -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" (list (mustMergeOverwrite (dict "name" "" "mountPath" "" ) (dict "name" "cert" "mountPath" "/tmp/k8s-webhook-server/serving-certs" "readOnly" true )))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "operator.operatorArguments" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $values := $dot.Values.AsMap -}}
{{- $args := (list "--health-probe-bind-address=:8081" "--metrics-bind-address=127.0.0.1:8080" "--leader-elect" (printf "--configurator-tag=%s" (get (fromJson (include "operator.configuratorTag" (dict "a" (list $dot) ))) "r")) (printf "--configurator-base-image=%s" $values.configurator.repository) (printf "--webhook-enabled=%t" (get (fromJson (include "operator.isWebhookEnabled" (dict "a" (list $dot) ))) "r"))) -}}
{{- if (eq $values.scope "Namespace") -}}
{{- $args = (concat (default (list ) $args) (list (printf "--namespace=%s" $dot.Release.Namespace) (printf "--log-level=%s" $values.logLevel))) -}}
{{- end -}}
{{- $_is_returning = true -}}
{{- (dict "r" (concat (default (list ) $args) (default (list ) $values.additionalCmdFlags))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

