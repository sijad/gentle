package builder

import (
	"log"
	"os"
	"text/template"
)

func (g *gqlBuilder) SDL() string {
	const sdl = `
{{- range $t := .Types -}}
{{- if eq $t.Kind 0 -}}
scalar {{$t.Name}}
{{- end}}
{{- if or (eq $t.Kind 3) (eq $t.Kind 7) -}}
{{if eq $t.Kind 3}}type{{else}}input{{end}} {{$t.Name}} {
{{- range $f := $t.Fields}}
  {{$f.Name | lowerCaseFirst}}
  {{- if gt (len $f.Args) 0 -}}
    (
      {{- range $i, $a := $f.Args}}
        {{- if $i}}{{", "}}{{end -}}
        {{- $a.Name | lowerCaseFirst }}: {{$a.Type | gqlType -}}
      {{ end -}}
    )
  {{- end -}}
  : {{$f.Type | gqlType}}
{{- end}}
}
{{- end}}

{{end -}}
`

	t := template.Must(template.New("sdl").Funcs(funcMap).Parse(sdl))

	type Data struct {
		Types []FullType
	}
	d := Data{g.FullTypes()}

	err := t.Execute(os.Stdout, d)
	if err != nil {
		log.Println("executing template:", err)
	}

	return ""
}
