apiVersion: v1
kind: ConfigMap
metadata:
  creationTimestamp: null
  name: dep-excluded
data:
  values: |
    {{- toYaml .Values | nindent 4 }}