{{- define "gateway-monitor.name" -}}
gateway-monitor
{{- end -}}

{{- define "gateway-monitor.fullname" -}}
{{ include "gateway-monitor.name" . }}
{{- end -}}
