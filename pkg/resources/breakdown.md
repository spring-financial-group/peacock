[Peacock Validation] Successfully parsed {{ len .messages }} message(s).
---
{{ range $idx, $val := .messages -}}
Message {{ inc $idx }} will be sent to: {{ commaSep $val.TeamNames }}
<details open>
<summary>Message Breakdown</summary>
{{ $val.Content }}
</details>
{{- end }}
