{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "cert-manager-webhook-inwx.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "cert-manager-webhook-inwx.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "cert-manager-webhook-inwx.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "cert-manager-webhook-inwx.selfSignedIssuer" -}}
{{ printf "%s-selfsign" (include "cert-manager-webhook-inwx.fullname" .) }}
{{- end -}}

{{- define "cert-manager-webhook-inwx.rootCAIssuer" -}}
{{ printf "%s-ca" (include "cert-manager-webhook-inwx.fullname" .) }}
{{- end -}}

{{- define "cert-manager-webhook-inwx.rootCACertificate" -}}
{{ printf "%s-ca" (include "cert-manager-webhook-inwx.fullname" .) }}
{{- end -}}

{{- define "cert-manager-webhook-inwx.servingCertificate" -}}
{{ printf "%s-webhook-tls" (include "cert-manager-webhook-inwx.fullname" .) }}
{{- end -}}

{{- define "cert-manager-webhook-inwx.secretName" -}}
{{- if .Values.inwx.existingSecret -}}
{{- .Values.inwx.existingSecret -}}
{{- else -}}
{{ include "cert-manager-webhook-inwx.fullname" . }}-inwx-credentials
{{- end -}}
{{- end -}}
