apiVersion: v1
kind: ConfigMap
metadata:
  creationTimestamp: null
  name: values-overwrite
data:
  values: |
    {{- toYaml .Values | nindent 4 }}