package interactive

import (
	"fmt"
	"html/template"
	"time"
)

func age(date time.Time) string {
	d := time.Since(date)

	if d <= time.Minute {
		return "1m"
	} else if d < time.Hour {
		return fmt.Sprintf("%dm", d/time.Minute)
	} else if d < 24*time.Hour {
		return fmt.Sprintf("%dh", d/time.Hour)
	} else if d < 365*24*time.Hour {
		return fmt.Sprintf("%dd", d/(24*time.Hour))
	} else {
		return fmt.Sprintf("%dy", d/(365*24*time.Hour))
	}
}

var diskDetails = template.Must(template.New("").
	Funcs(template.FuncMap{
		"header":  headerStyle.Render,
		"age":     age,
		"rfc3339": func(t time.Time) string { return t.Format(time.RFC3339) },
	}).
	Parse(
		`{{ block "disk" . }}
{{ .Name | header }}

{{ header "Created:" }} {{.CreatedAt | rfc3339 }} ({{ .CreatedAt | age }})

{{- block "meta" .Meta }}
{{ header "Metadata" }}
{{ range $k, $v := . }}{{ $k }}: {{ $v }}
{{ end }}
{{ end -}}

{{- with .Provider }}
{{ header "Provider:" }} {{ .Name }}
{{ template "meta" .Meta }}
{{ end -}}
{{ end}}
`))
