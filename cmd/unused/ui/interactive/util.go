package interactive

import (
	"html/template"
	"time"

	"github.com/grafana/unused/cli"
)

var diskDetails = template.Must(template.New("").
	Funcs(template.FuncMap{
		"header":  headerStyle.Render,
		"age":     cli.Age,
		"rfc3339": func(t time.Time) string { return t.Format(time.RFC3339) },
	}).
	Parse(
		`{{ block "disk" . }}
{{- .Name | header }}

{{ header "Created:" }} {{.CreatedAt | rfc3339 }} ({{ .CreatedAt | age }})
{{ header "Last used:" }} {{ if .LastUsedAt.IsZero }}n/a{{ else }}{{.CreatedAt | rfc3339 }} ({{ .CreatedAt | age }}){{ end }}

{{- block "meta" .Meta }}
{{ header "Metadata" }}
{{ range $k, $v := . }}{{ $k }}: {{ $v }}
{{ end }}
{{ end -}}

{{- with .Provider }}
{{ header "Provider:" }} {{ .Name }}
{{- template "meta" .Meta }}
{{ end -}}
{{ end}}
`))

const outputTpl = `{{ range . }}
{{ if .Error }}{{ "𐄂" | error -}}
{{ else if .Done }}✔
{{- else if .Deleting }}#
{{- end -}}
  {{ .Disk.Name }} ({{ .Disk.Provider.Name }} {{ .Disk.Provider.Meta }}) {{ if .Error }}
{{ .Error.Error | wrap | error }}{{ end -}}
{{ end }}
`
