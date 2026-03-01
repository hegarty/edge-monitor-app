{{- define "jitter-probe.name" -}}
jitter-probe
{{- end -}}

{{- define "jitter-probe.fullname" -}}
{{ include "jitter-probe.name" . }}
{{- end -}}
