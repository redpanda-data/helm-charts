apiVersion: v1
kind: ConfigMap
metadata:
  creationTimestamp: null
  name: root-chart-test
data:
  values: |
    {{- toYaml .Values | nindent 4 }}