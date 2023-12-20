{{- define "helper.manageMountedSecrets" }}
{{- $ := .ctx }}
{{- if and $.Values $.Values.secrets $.Values.secrets.secretManagement $.Values.secrets.secretManagement.externalSecretsOperator}}
{{- $secretNamePrefix := (printf "%s-ext-secret" $.Chart.Name) }} 
{{- $esoSecretName := (printf "%s-%d" $secretNamePrefix 0) }}

{{- $isExternalPopulated := "true"}}
{{- range $item := .items }}
{{- if eq $isExternalPopulated "false"}}
{{- break}}
{{- end}}
{{- if eq (include "harnesscommon.secrets.hasESOSecret" (dict "variableName" $item.key "esoSecretCtxs" (list (dict "secretCtxIdentifier" "local" "secretCtx" $.Values.secrets.secretManagement.externalSecretsOperator)))) "false" }}
{{- $isExternalPopulated = "false"}}
{{- end}}
{{- end}}
- name: {{ print .secretName }}
  secret:
    defaultMode: 420
{{- if eq $isExternalPopulated "true"}}
    secretName: {{ printf "%s" $esoSecretName }}
{{- else}}
    secretName: {{ printf "%s" .defaultKubernetesSecretName }}
{{- end}}
    items:
{{- range $item := .items }}
    - key: {{ printf  $item.key }}
      path: {{ printf "%s" $item.path }}
{{- end }}
{{- end }}
{{- end }}