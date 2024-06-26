{{- /* Generated from "rbac.go" */ -}}

{{- define "redpanda.ClusterRoles" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $values := $dot.Values.AsMap -}}
{{- $crs := (coalesce nil) -}}
{{- $cr_1 := (get (fromJson (include "redpanda.SidecarControllersClusterRole" (dict "a" (list $dot) ))) "r") -}}
{{- if (ne $cr_1 (coalesce nil)) -}}
{{- $crs = (concat (default (list ) $crs) (list $cr_1)) -}}
{{- end -}}
{{- if (not $values.rbac.enabled) -}}
{{- $_is_returning = true -}}
{{- (dict "r" $crs) | toJson -}}
{{- break -}}
{{- end -}}
{{- $rpkBundleName := (printf "%s-rpk-bundle" (get (fromJson (include "redpanda.Fullname" (dict "a" (list $dot) ))) "r")) -}}
{{- $crs = (concat (default (list ) $crs) (default (list ) (list (mustMergeOverwrite (dict "metadata" (dict "creationTimestamp" (coalesce nil) ) "rules" (coalesce nil) ) (mustMergeOverwrite (dict ) (dict "apiVersion" "rbac.authorization.k8s.io/v1" "kind" "ClusterRole" )) (dict "metadata" (mustMergeOverwrite (dict "creationTimestamp" (coalesce nil) ) (dict "name" (get (fromJson (include "redpanda.Fullname" (dict "a" (list $dot) ))) "r") "labels" (get (fromJson (include "redpanda.FullLabels" (dict "a" (list $dot) ))) "r") "annotations" $values.serviceAccount.annotations )) "rules" (list (mustMergeOverwrite (dict "verbs" (coalesce nil) ) (dict "apiGroups" (list "") "resources" (list "nodes") "verbs" (list "get" "list") ))) )) (mustMergeOverwrite (dict "metadata" (dict "creationTimestamp" (coalesce nil) ) "rules" (coalesce nil) ) (mustMergeOverwrite (dict ) (dict "apiVersion" "rbac.authorization.k8s.io/v1" "kind" "ClusterRole" )) (dict "metadata" (mustMergeOverwrite (dict "creationTimestamp" (coalesce nil) ) (dict "name" $rpkBundleName "labels" (get (fromJson (include "redpanda.FullLabels" (dict "a" (list $dot) ))) "r") "annotations" $values.serviceAccount.annotations )) "rules" (list (mustMergeOverwrite (dict "verbs" (coalesce nil) ) (dict "apiGroups" (list "") "resources" (list "configmaps" "endpoints" "events" "limitranges" "persistentvolumeclaims" "pods" "pods/log" "replicationcontrollers" "resourcequotas" "serviceaccounts" "services") "verbs" (list "get" "list") ))) ))))) -}}
{{- $_is_returning = true -}}
{{- (dict "r" $crs) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.ClusterRoleBindings" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $values := $dot.Values.AsMap -}}
{{- $crbs := (coalesce nil) -}}
{{- $crb_2 := (get (fromJson (include "redpanda.SidecarControllersClusterRoleBinding" (dict "a" (list $dot) ))) "r") -}}
{{- if (ne $crb_2 (coalesce nil)) -}}
{{- $crbs = (concat (default (list ) $crbs) (list $crb_2)) -}}
{{- end -}}
{{- if (not $values.rbac.enabled) -}}
{{- $_is_returning = true -}}
{{- (dict "r" $crbs) | toJson -}}
{{- break -}}
{{- end -}}
{{- $rpkBundleName := (printf "%s-rpk-bundle" (get (fromJson (include "redpanda.Fullname" (dict "a" (list $dot) ))) "r")) -}}
{{- $crbs = (concat (default (list ) $crbs) (default (list ) (list (mustMergeOverwrite (dict "metadata" (dict "creationTimestamp" (coalesce nil) ) "roleRef" (dict "apiGroup" "" "kind" "" "name" "" ) ) (mustMergeOverwrite (dict ) (dict "apiVersion" "rbac.authorization.k8s.io/v1" "kind" "ClusterRoleBinding" )) (dict "metadata" (mustMergeOverwrite (dict "creationTimestamp" (coalesce nil) ) (dict "name" (get (fromJson (include "redpanda.Fullname" (dict "a" (list $dot) ))) "r") "labels" (get (fromJson (include "redpanda.FullLabels" (dict "a" (list $dot) ))) "r") "annotations" $values.serviceAccount.annotations )) "roleRef" (mustMergeOverwrite (dict "apiGroup" "" "kind" "" "name" "" ) (dict "apiGroup" "rbac.authorization.k8s.io" "kind" "ClusterRole" "name" (get (fromJson (include "redpanda.Fullname" (dict "a" (list $dot) ))) "r") )) "subjects" (list (mustMergeOverwrite (dict "kind" "" "name" "" ) (dict "kind" "ServiceAccount" "name" (get (fromJson (include "redpanda.ServiceAccountName" (dict "a" (list $dot) ))) "r") "namespace" $dot.Release.Namespace ))) )) (mustMergeOverwrite (dict "metadata" (dict "creationTimestamp" (coalesce nil) ) "roleRef" (dict "apiGroup" "" "kind" "" "name" "" ) ) (mustMergeOverwrite (dict ) (dict "apiVersion" "rbac.authorization.k8s.io/v1" "kind" "ClusterRoleBinding" )) (dict "metadata" (mustMergeOverwrite (dict "creationTimestamp" (coalesce nil) ) (dict "name" $rpkBundleName "labels" (get (fromJson (include "redpanda.FullLabels" (dict "a" (list $dot) ))) "r") "annotations" $values.serviceAccount.annotations )) "roleRef" (mustMergeOverwrite (dict "apiGroup" "" "kind" "" "name" "" ) (dict "apiGroup" "rbac.authorization.k8s.io" "kind" "ClusterRole" "name" $rpkBundleName )) "subjects" (list (mustMergeOverwrite (dict "kind" "" "name" "" ) (dict "kind" "ServiceAccount" "name" (get (fromJson (include "redpanda.ServiceAccountName" (dict "a" (list $dot) ))) "r") "namespace" $dot.Release.Namespace ))) ))))) -}}
{{- $_is_returning = true -}}
{{- (dict "r" $crbs) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.SidecarControllersClusterRole" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $values := $dot.Values.AsMap -}}
{{- if (or (not $values.statefulset.sideCars.controllers.enabled) (not $values.statefulset.sideCars.controllers.createRBAC)) -}}
{{- $_is_returning = true -}}
{{- (dict "r" (coalesce nil)) | toJson -}}
{{- break -}}
{{- end -}}
{{- $sidecarControllerName := (printf "%s-sidecar-controllers" (get (fromJson (include "redpanda.Fullname" (dict "a" (list $dot) ))) "r")) -}}
{{- $_is_returning = true -}}
{{- (dict "r" (mustMergeOverwrite (dict "metadata" (dict "creationTimestamp" (coalesce nil) ) "rules" (coalesce nil) ) (mustMergeOverwrite (dict ) (dict "apiVersion" "rbac.authorization.k8s.io/v1" "kind" "ClusterRole" )) (dict "metadata" (mustMergeOverwrite (dict "creationTimestamp" (coalesce nil) ) (dict "name" $sidecarControllerName "labels" (get (fromJson (include "redpanda.FullLabels" (dict "a" (list $dot) ))) "r") "annotations" $values.serviceAccount.annotations )) "rules" (list (mustMergeOverwrite (dict "verbs" (coalesce nil) ) (dict "apiGroups" (list "") "resources" (list "nodes") "verbs" (list "get" "list" "watch") )) (mustMergeOverwrite (dict "verbs" (coalesce nil) ) (dict "apiGroups" (list "") "resources" (list "persistentvolumes") "verbs" (list "delete" "get" "list" "patch" "update" "watch") ))) ))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.SidecarControllersClusterRoleBinding" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $values := $dot.Values.AsMap -}}
{{- if (or (not $values.statefulset.sideCars.controllers.enabled) (not $values.statefulset.sideCars.controllers.createRBAC)) -}}
{{- $_is_returning = true -}}
{{- (dict "r" (coalesce nil)) | toJson -}}
{{- break -}}
{{- end -}}
{{- $sidecarControllerName := (printf "%s-sidecar-controllers" (get (fromJson (include "redpanda.Fullname" (dict "a" (list $dot) ))) "r")) -}}
{{- $_is_returning = true -}}
{{- (dict "r" (mustMergeOverwrite (dict "metadata" (dict "creationTimestamp" (coalesce nil) ) "roleRef" (dict "apiGroup" "" "kind" "" "name" "" ) ) (mustMergeOverwrite (dict ) (dict "apiVersion" "rbac.authorization.k8s.io/v1" "kind" "ClusterRoleBinding" )) (dict "metadata" (mustMergeOverwrite (dict "creationTimestamp" (coalesce nil) ) (dict "name" $sidecarControllerName "labels" (get (fromJson (include "redpanda.FullLabels" (dict "a" (list $dot) ))) "r") "annotations" $values.serviceAccount.annotations )) "roleRef" (mustMergeOverwrite (dict "apiGroup" "" "kind" "" "name" "" ) (dict "apiGroup" "rbac.authorization.k8s.io" "kind" "ClusterRole" "name" $sidecarControllerName )) "subjects" (list (mustMergeOverwrite (dict "kind" "" "name" "" ) (dict "kind" "ServiceAccount" "name" (get (fromJson (include "redpanda.ServiceAccountName" (dict "a" (list $dot) ))) "r") "namespace" $dot.Release.Namespace ))) ))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.SidecarControllersRole" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $values := $dot.Values.AsMap -}}
{{- if (or (not $values.statefulset.sideCars.controllers.enabled) (not $values.statefulset.sideCars.controllers.createRBAC)) -}}
{{- $_is_returning = true -}}
{{- (dict "r" (coalesce nil)) | toJson -}}
{{- break -}}
{{- end -}}
{{- $sidecarControllerName := (printf "%s-sidecar-controllers" (get (fromJson (include "redpanda.Fullname" (dict "a" (list $dot) ))) "r")) -}}
{{- $_is_returning = true -}}
{{- (dict "r" (mustMergeOverwrite (dict "metadata" (dict "creationTimestamp" (coalesce nil) ) "rules" (coalesce nil) ) (mustMergeOverwrite (dict ) (dict "apiVersion" "rbac.authorization.k8s.io/v1" "kind" "Role" )) (dict "metadata" (mustMergeOverwrite (dict "creationTimestamp" (coalesce nil) ) (dict "name" $sidecarControllerName "namespace" $dot.Release.Namespace "labels" (get (fromJson (include "redpanda.FullLabels" (dict "a" (list $dot) ))) "r") "annotations" $values.serviceAccount.annotations )) "rules" (list (mustMergeOverwrite (dict "verbs" (coalesce nil) ) (dict "apiGroups" (list "apps") "resources" (list "statefulsets/status") "verbs" (list "patch" "update") )) (mustMergeOverwrite (dict "verbs" (coalesce nil) ) (dict "apiGroups" (list "") "resources" (list "secrets" "pods") "verbs" (list "get" "list" "watch") )) (mustMergeOverwrite (dict "verbs" (coalesce nil) ) (dict "apiGroups" (list "apps") "resources" (list "statefulsets") "verbs" (list "get" "patch" "update" "list" "watch") )) (mustMergeOverwrite (dict "verbs" (coalesce nil) ) (dict "apiGroups" (list "") "resources" (list "persistentvolumeclaims") "verbs" (list "delete" "get" "list" "patch" "update" "watch") ))) ))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

{{- define "redpanda.SidecarControllersRoleBinding" -}}
{{- $dot := (index .a 0) -}}
{{- range $_ := (list 1) -}}
{{- $_is_returning := false -}}
{{- $values := $dot.Values.AsMap -}}
{{- if (or (not $values.statefulset.sideCars.controllers.enabled) (not $values.statefulset.sideCars.controllers.createRBAC)) -}}
{{- $_is_returning = true -}}
{{- (dict "r" (coalesce nil)) | toJson -}}
{{- break -}}
{{- end -}}
{{- $sidecarControllerName := (printf "%s-sidecar-controllers" (get (fromJson (include "redpanda.Fullname" (dict "a" (list $dot) ))) "r")) -}}
{{- $_is_returning = true -}}
{{- (dict "r" (mustMergeOverwrite (dict "metadata" (dict "creationTimestamp" (coalesce nil) ) "roleRef" (dict "apiGroup" "" "kind" "" "name" "" ) ) (mustMergeOverwrite (dict ) (dict "apiVersion" "rbac.authorization.k8s.io/v1" "kind" "RoleBinding" )) (dict "metadata" (mustMergeOverwrite (dict "creationTimestamp" (coalesce nil) ) (dict "name" $sidecarControllerName "namespace" $dot.Release.Namespace "labels" (get (fromJson (include "redpanda.FullLabels" (dict "a" (list $dot) ))) "r") "annotations" $values.serviceAccount.annotations )) "roleRef" (mustMergeOverwrite (dict "apiGroup" "" "kind" "" "name" "" ) (dict "apiGroup" "rbac.authorization.k8s.io" "kind" "Role" "name" $sidecarControllerName )) "subjects" (list (mustMergeOverwrite (dict "kind" "" "name" "" ) (dict "kind" "ServiceAccount" "name" (get (fromJson (include "redpanda.ServiceAccountName" (dict "a" (list $dot) ))) "r") "namespace" $dot.Release.Namespace ))) ))) | toJson -}}
{{- break -}}
{{- end -}}
{{- end -}}

