{{/* template para horizontalpodautoscaler */}}

{{/*
retorna a apiversion apropriada para horizontalpodautoscaler
*/}}
{{- define "hpa.apiVersion" -}}
{{-     if .Capabilities.APIVersions.Has "autoscaling/v2" -}}
{{-         "autoscaling/v2" -}}
{{-     else if .Capabilities.APIVersions.Has "autoscaling/v2beta2" -}}
{{-         "autoscaling/v2beta2" -}}
{{-     else -}}
{{-         "autoscaling/v2beta1" -}}
{{-     end -}}
{{- end -}}

{{/*
checa se as métricas de autoscaling/v2 é suportado
*/}}
{{- define "hpa.supportsV2MetricsSpec" -}}
{{-     $apiVersion := include "hpa.apiVersion" . -}}
{{-     if or (eq $apiVersion "autoscaling/v2") (eq $apiVersion "autoscaling/v2beta2") -}}
true
{{-     end -}}
{{- end -}}
