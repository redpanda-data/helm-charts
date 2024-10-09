apiVersion: v1
kind: ConfigMap
metadata:
  creationTimestamp: null
  name: dependency-included-by-default
data:
  values: |
    {{- toYaml .Values | nindent 4 }}