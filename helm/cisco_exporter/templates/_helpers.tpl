{{- define "cisco-exporter.fullname" -}}
{{- default .Chart.Name .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- end }}

{{- define "cisco-exporter.name" -}}
{{- .Chart.Name -}}
{{- end }}
