{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
*/}}
{{- define "fullname" -}}
{{- $name := default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create the mongodb connection string if the internal instance is enabled and the the overide is empty.
*/}}
{{- define "mongodb.connectionString" -}}
    {{ if (and (eq .Values.mongodb.connectionStringOverride "") .Values.mongodb.useInternalInstance ) }}
        {{- $mongoSrvName := include "mongodb.service.nameOverride" . -}}
        {{- printf "mongodb://%s:%s@%s/%s" ( index .Values.mongodb.auth.usernames 0 ) ( index .Values.mongodb.auth.passwords 0 )  $mongoSrvName ( index .Values.mongodb.auth.databases 0 ) -}}
    {{- else }}
        {{- .Values.mongodb.connectionStringOverride -}}
    {{- end -}}
{{- end -}}
