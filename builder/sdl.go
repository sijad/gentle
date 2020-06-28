package builder

import (
	"io"
	"text/template"
)

func (g *gqlBuilder) SDL(wr io.Writer) error {
	t := template.Must(template.New("sdl.tmpl").Funcs(funcMap).ParseFiles("sdl.tmpl"))

	type Data struct {
		Types []FullType
	}
	d := Data{g.FullTypes()}

	return t.Execute(wr, d)
}
