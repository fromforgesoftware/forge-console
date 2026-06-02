{{- define "forge.name" -}}{{ default .Chart.Name .Values.nameOverride }}{{- end -}}
{{- define "forge.fullname" -}}{{- if .Values.fullnameOverride -}}{{ .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}{{- else -}}{{ printf "%s-%s" .Release.Name (include "forge.name" .) | trunc 63 | trimSuffix "-" }}{{- end -}}{{- end -}}

{{- define "forge.selectorLabels" -}}
app.kubernetes.io/name: {{ include "forge.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{- define "forge.labels" -}}
{{ include "forge.selectorLabels" . }}
app.kubernetes.io/part-of: forge
{{- end -}}

{{- define "forge.image" -}}{{ .Values.image.repository }}:{{ .Values.image.tag | default .Chart.AppVersion }}{{- end -}}

{{- define "forge.serviceAccountName" -}}
{{- default (include "forge.fullname" .) .Values.serviceAccount.name -}}
{{- end -}}

{{- define "forge.dbSecretName" -}}
{{- if .Values.database.existingSecret -}}{{ .Values.database.existingSecret }}{{- else -}}{{ printf "%s-db" (include "forge.fullname" .) }}{{- end -}}
{{- end -}}

{{/* Shared env for server + migrator. */}}
{{- define "forge.env" -}}
- name: SVC_NAME
  value: {{ include "forge.name" . | quote }}
- name: REST_ADDRESS
  value: ":{{ .Values.ports.http }}"
- name: HTTP_ADDRESS
  value: ":{{ .Values.ports.http }}"
- name: FOUNDRY_NAMESPACE
  valueFrom:
    fieldRef:
      fieldPath: metadata.namespace
- name: DB_HOST
  value: {{ .Values.database.host | quote }}
- name: DB_PORT
  value: {{ .Values.database.port | quote }}
- name: DB_NAME
  value: {{ .Values.database.name | quote }}
- name: DB_SCHEMA
  value: {{ .Values.database.schema | quote }}
- name: DB_SSL
  value: {{ .Values.database.ssl | quote }}
- name: DB_LOG_LEVEL
  value: {{ .Values.database.logLevel | default "warn" | quote }}
- name: DB_USER
  valueFrom:
    secretKeyRef:
      name: {{ include "forge.dbSecretName" . }}
      key: DB_USER
- name: DB_PASSWORD
  valueFrom:
    secretKeyRef:
      name: {{ include "forge.dbSecretName" . }}
      key: DB_PASSWORD
{{- if .Values.bootstrap.enabled }}
- name: FOUNDRY_BOOTSTRAP_ADMIN_EMAIL
  value: {{ .Values.bootstrap.adminEmail | quote }}
- name: FOUNDRY_BOOTSTRAP_ADMIN_PASSWORD
  value: {{ .Values.bootstrap.adminPassword | quote }}
- name: FOUNDRY_BOOTSTRAP_ADMIN_NAME
  value: {{ .Values.bootstrap.adminName | quote }}
- name: FOUNDRY_APPS
  value: {{ .Values.bootstrap.apps | quote }}
{{- end }}
- name: FOUNDRY_OIDC_PROVIDERS
  value: {{ .Values.oidcProviders | default "" | quote }}
{{- if .Values.gatewaySecret }}
- name: FORGE_GATEWAY_SECRET
  value: {{ .Values.gatewaySecret | quote }}
{{- end }}
{{- end -}}
