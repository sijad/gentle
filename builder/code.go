package builder

import (
	"bytes"
	"go/format"
	"go/types"
	"io"
)

func (g *gqlBuilder) Code(w io.Writer) error {
	type Data struct {
		PackageName         string
		Imports             []string
		Types               map[string]FullType
		Dependencies        map[string]*types.Var
		DependenciesNameMap map[string]string
		Sdl                 string
	}

	var sdlBuf bytes.Buffer
	g.SDL(&sdlBuf)

	fullTypes := g.FullTypes()
	typesMap := make(map[string]FullType, len(fullTypes))
	for _, v := range fullTypes {
		typesMap[v.Name] = v
	}

	d := Data{
		PackageName:         "graph",
		Imports:             []string{"context"},
		Dependencies:        g.dependencies,
		DependenciesNameMap: g.dependenciesNameMap,
		Types:               typesMap,
		Sdl:                 sdlBuf.String(),
	}

	source := &bytes.Buffer{}

	if err := templates.ExecuteTemplate(source, "template/code.tmpl", d); err != nil {
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
