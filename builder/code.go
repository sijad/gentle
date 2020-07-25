package builder

import (
	"bytes"
	"go/format"
	"go/types"
	"io"
	"text/template"
)

func (g *gqlBuilder) Code(w io.Writer) error {
	t := template.Must(template.New("code.tmpl").Funcs(funcMap).ParseFiles("code.tmpl"))

	type Data struct {
		PackageName         string
		Imports             []string
		Types               []FullType
		Dependencies        map[string]*types.Var
		DependenciesNameMap map[string]string
		Sdl                 string
	}

	var sdlBuf bytes.Buffer
	g.SDL(&sdlBuf)

	d := Data{
		PackageName:         "generated",
		Imports:             []string{"context"},
		Dependencies:        g.dependencies,
		DependenciesNameMap: g.dependenciesNameMap,
		Types:               g.FullTypes(),
		Sdl:                 sdlBuf.String(),
	}

	source := &bytes.Buffer{}

	if err := t.Execute(source, d); err != nil {
		return err
	}

	formatted, err := format.Source(source.Bytes())
	if err != nil {
		w.Write(source.Bytes())
		return err
	}

	if _, err := w.Write(formatted); err != nil {
		return err
	}

	return nil
}
