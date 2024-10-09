apiVersion: v1
kind: ConfigMap
metadata:
  creationTimestamp: null
  name: dependency
data:
  values: |
    {{- toYaml .Values | nindent 4 }}