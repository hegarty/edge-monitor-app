{{- define "wifi-probe.name" -}}
wifi-probe
{{- end -}}

{{- define "wifi-probe.fullname" -}}
{{ include "wifi-probe.name" . }}
{{- end -}}
