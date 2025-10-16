{{/*
Expand the name of the chart.
*/}}
{{- define "demo-app.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "demo-app.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "demo-app.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "demo-app.labels" -}}
helm.sh/chart: {{ include "demo-app.chart" . }}
{{ include "demo-app.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- with .Values.global.labels }}
{{ toYaml . }}
{{- end }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "demo-app.selectorLabels" -}}
app.kubernetes.io/name: {{ include "demo-app.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Generator component labels
*/}}
{{- define "demo-app.generator.labels" -}}
{{ include "demo-app.labels" . }}
app.kubernetes.io/component: generator
{{- end }}

{{/*
Generator selector labels
*/}}
{{- define "demo-app.generator.selectorLabels" -}}
{{ include "demo-app.selectorLabels" . }}
app.kubernetes.io/component: generator
{{- end }}

{{/*
Backend component labels
*/}}
{{- define "demo-app.backend.labels" -}}
{{ include "demo-app.labels" . }}
app.kubernetes.io/component: backend
{{- end }}

{{/*
Backend selector labels
*/}}
{{- define "demo-app.backend.selectorLabels" -}}
{{ include "demo-app.selectorLabels" . }}
app.kubernetes.io/component: backend
{{- end }}

{{/*
Frontend component labels
*/}}
{{- define "demo-app.frontend.labels" -}}
{{ include "demo-app.labels" . }}
app.kubernetes.io/component: frontend
{{- end }}

{{/*
Frontend selector labels
*/}}
{{- define "demo-app.frontend.selectorLabels" -}}
{{ include "demo-app.selectorLabels" . }}
app.kubernetes.io/component: frontend
{{- end }}

{{/*
Image pull policy
*/}}
{{- define "demo-app.imagePullPolicy" -}}
{{- if . }}
{{- . }}
{{- else }}
{{- $.Values.global.imagePullPolicy }}
{{- end }}
{{- end }}

{{/*
Full image name for generator
*/}}
{{- define "demo-app.generator.image" -}}
{{- $registry := .Values.global.imageRegistry }}
{{- $repository := .Values.generator.image.repository }}
{{- $tag := .Values.generator.image.tag | default .Chart.AppVersion }}
{{- printf "%s/%s:%s" $registry $repository $tag }}
{{- end }}

{{/*
Full image name for backend
*/}}
{{- define "demo-app.backend.image" -}}
{{- $registry := .Values.global.imageRegistry }}
{{- $repository := .Values.backend.image.repository }}
{{- $tag := .Values.backend.image.tag | default .Chart.AppVersion }}
{{- printf "%s/%s:%s" $registry $repository $tag }}
{{- end }}

{{/*
Full image name for frontend
*/}}
{{- define "demo-app.frontend.image" -}}
{{- $registry := .Values.global.imageRegistry }}
{{- $repository := .Values.frontend.image.repository }}
{{- $tag := .Values.frontend.image.tag | default .Chart.AppVersion }}
{{- printf "%s/%s:%s" $registry $repository $tag }}
{{- end }}

{{/*
RabbitMQ connection URL
*/}}
{{- define "demo-app.rabbitmq.url" -}}
{{- if .Values.rabbitmq.url }}
{{- .Values.rabbitmq.url }}
{{- else }}
{{- printf "amqp://%s:%s@%s:%d" .Values.rabbitmq.user .Values.rabbitmq.password .Values.rabbitmq.host (int .Values.rabbitmq.port) }}
{{- end }}
{{- end }}

{{/*
PostgreSQL connection DSN
*/}}
{{- define "demo-app.postgresql.dsn" -}}
{{- printf "host=%s port=%d user=%s password=%s dbname=%s sslmode=%s" .Values.postgresql.host (int .Values.postgresql.port) .Values.postgresql.user .Values.postgresql.password .Values.postgresql.database .Values.postgresql.sslmode }}
{{- end }}
