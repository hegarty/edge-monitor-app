{{- define "alert-receiver.name" -}}
alert-receiver
{{- end -}}

{{- define "alert-receiver.fullname" -}}
{{ include "alert-receiver.name" . }}
{{- end -}}
