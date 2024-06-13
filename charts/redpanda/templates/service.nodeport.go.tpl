{{- /* Generated from "service.nodeport.go" */ -}}

{{- define "redpanda.NodePortService" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $values := $dot.Values.AsMap -}}
{{- if (or (not $values.external.enabled) (ne $values.external.type "NodePort")) -}}
{{- (dict "r" (coalesce nil)) | toJson -}}
{{- break -}}
{{- end -}}
{{- if (or (eq $values.external.service (coalesce nil)) (not $values.external.service.enabled)) -}}
{{- (dict "r" (coalesce nil)) | toJson -}}
{{- break -}}
{{- end -}}
{{- $ports := (list ) -}}
{{- range $name, $listener := $values.listeners.admin.external -}}
{{- if (and (ne $listener.enabled (coalesce nil)) (not $listener.enabled)) -}}
{{- continue -}}
{{- end -}}
{{- $nodePort := ($listener.port | int) -}}
{{- if (gt ((get (fromJson (include "_shims.len" (dict "a" (list $listener.advertisedPorts) ))) "r") | int) (0 | int)) -}}
{{- $nodePort = (index $listener.advertisedPorts (0 | int)) -}}
{{- end -}}
{{- $ports = (mustAppend $ports (mustMergeOverwrite (dict "port" 0 "targetPort" 0 ) (dict "name" (printf "admin-%s" $name) "protocol" "TCP" "port" ($listener.port | int) "nodePort" $nodePort ))) -}}
{{- end -}}
{{- range $name, $listener := $values.listeners.kafka.external -}}
{{- if (and (ne $listener.enabled (coalesce nil)) (not $listener.enabled)) -}}
{{- continue -}}
{{- end -}}
{{- $nodePort := ($listener.port | int) -}}
{{- if (gt ((get (fromJson (include "_shims.len" (dict "a" (list $listener.advertisedPorts) ))) "r") | int) (0 | int)) -}}
{{- $nodePort = (index $listener.advertisedPorts (0 | int)) -}}
{{- end -}}
{{- $ports = (mustAppend $ports (mustMergeOverwrite (dict "port" 0 "targetPort" 0 ) (dict "name" (printf "kafka-%s" $name) "protocol" "TCP" "port" ($listener.port | int) "nodePort" $nodePort ))) -}}
{{- end -}}
{{- range $name, $listener := $values.listeners.http.external -}}
{{- if (and (ne $listener.enabled (coalesce nil)) (not $listener.enabled)) -}}
{{- continue -}}
{{- end -}}
{{- $nodePort := ($listener.port | int) -}}
{{- if (gt ((get (fromJson (include "_shims.len" (dict "a" (list $listener.advertisedPorts) ))) "r") | int) (0 | int)) -}}
{{- $nodePort = (index $listener.advertisedPorts (0 | int)) -}}
{{- end -}}
{{- $ports = (mustAppend $ports (mustMergeOverwrite (dict "port" 0 "targetPort" 0 ) (dict "name" (printf "http-%s" $name) "protocol" "TCP" "port" ($listener.port | int) "nodePort" $nodePort ))) -}}
{{- end -}}
{{- range $name, $listener := $values.listeners.schemaRegistry.external -}}
{{- if (and (ne $listener.enabled (coalesce nil)) (not $listener.enabled)) -}}
{{- continue -}}
{{- end -}}
{{- $nodePort := ($listener.port | int) -}}
{{- if (gt ((get (fromJson (include "_shims.len" (dict "a" (list $listener.advertisedPorts) ))) "r") | int) (0 | int)) -}}
{{- $nodePort = (index $listener.advertisedPorts (0 | int)) -}}
{{- end -}}
{{- $ports = (mustAppend $ports (mustMergeOverwrite (dict "port" 0 "targetPort" 0 ) (dict "name" (printf "schema-%s" $name) "protocol" "TCP" "port" ($listener.port | int) "nodePort" $nodePort ))) -}}
{{- end -}}
{{- $annotations := $values.external.annotations -}}
{{- if (eq $annotations (coalesce nil)) -}}
{{- $annotations = (dict ) -}}
{{- end -}}
{{- (dict "r" (mustMergeOverwrite (dict "metadata" (dict "creationTimestamp" (coalesce nil) ) "spec" (dict ) "status" (dict "loadBalancer" (dict ) ) ) (mustMergeOverwrite (dict ) (dict "apiVersion" "v1" "kind" "Service" )) (dict "metadata" (mustMergeOverwrite (dict "creationTimestamp" (coalesce nil) ) (dict "name" (printf "%s-external" (get (fromJson (include "redpanda.ServiceName" (dict "a" (list $dot) ))) "r")) "namespace" $dot.Release.Namespace "labels" (get (fromJson (include "redpanda.FullLabels" (dict "a" (list $dot) ))) "r") "annotations" $annotations )) "spec" (mustMergeOverwrite (dict ) (dict "externalTrafficPolicy" "Local" "ports" $ports "publishNotReadyAddresses" true "selector" (get (fromJson (include "redpanda.StatefulSetPodLabelsSelector" (dict "a" (list $dot) ))) "r") "sessionAffinity" "None" "type" "NodePort" )) ))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

