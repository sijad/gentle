package builder

import (
	"io"
)

func (g *gqlBuilder) SDL(wr io.Writer) error {
	type Data struct {
		Types []FullType
	}
	d := Data{g.FullTypes()}

	return templates.ExecuteTemplate(wr, "template/sdl.tmpl", d)
}
