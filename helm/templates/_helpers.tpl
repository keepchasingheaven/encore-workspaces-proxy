{{/*
expandir o nome da chart
*/}}
{{- define "encore-workspaces-proxy.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
criar o nome de app totalmente qualificado.
truncar em 63 caracteres porque algumas fields de nome do kubernetes
são limitadas (pelo spec de nome de dns).
caso o nome de release contenha nome da chart isso será utilizado
como nome completo.
*/}}
{{- define "encore-workspaces-proxy.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "=" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
criar nome de chart e versão como utilizada na label da chart
*/}}
{{- define "encore-workspaces-proxy.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
labels comuns
*/}}
{{- define "encore-workspaces-proxy.labels" -}}
helm.sh/chart: {{ include "encore-workspaces-proxy.chart" . }}
{{ include "encore-workspaces-proxy.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- with .Values.labels }}
{{ toYaml . }}
{{- end }}
{{- end }}

{{/*
labels seletores
*/}}
*/}}
{{- define "encore-workspaces-proxy.selectorLabels" -}}
app.kubernetes.io/name: {{ include "encore-workspaces-proxy.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
criar o nome da conta de serviço a ser utilizada
*/}}
{{- define "encore-workspaces-proxy.serviceAccountName" -}}
{{- default (include "encore-workspaces-proxy.fullname" .) .Values.serviceAccount.name }}
{{- end }}
