{{/*
Expand the name of the chart.
*/}}
{{- define "rss-to-bluesky.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name. CronJob names are limited to 52
characters: the controller appends an 11-character suffix to Job names.
*/}}
{{- define "rss-to-bluesky.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 52 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 52 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 52 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{- define "rss-to-bluesky.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "rss-to-bluesky.labels" -}}
helm.sh/chart: {{ include "rss-to-bluesky.chart" . }}
{{ include "rss-to-bluesky.selectorLabels" . }}
app.kubernetes.io/version: {{ .Values.image.tag | default .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{- define "rss-to-bluesky.selectorLabels" -}}
app.kubernetes.io/name: {{ include "rss-to-bluesky.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{- define "rss-to-bluesky.claimName" -}}
{{- .Values.persistence.existingClaim | default (include "rss-to-bluesky.fullname" .) }}
{{- end }}

{{/*
Fail early on settings rss-to-bluesky cannot run with.
*/}}
{{- define "rss-to-bluesky.validate" -}}
{{- if not .Values.existingSecret }}
{{- fail "existingSecret is required: a Secret with the bluesky-user and bluesky-password keys" }}
{{- end }}
{{- end }}
