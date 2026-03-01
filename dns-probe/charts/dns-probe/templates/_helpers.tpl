{{- define "dns-probe.name" -}}
dns-probe
{{- end -}}

{{- define "dns-probe.fullname" -}}
{{ include "dns-probe.name" . }}
{{- end -}}
