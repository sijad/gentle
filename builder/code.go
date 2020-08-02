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
		Imports             map[string]string
		Types               map[string]FullType
		Dependencies        map[string]*types.Var
		DependenciesNameMap map[string]string
		Sdl                 string
	}

	var sdlBuf bytes.Buffer
	g.SDL(&sdlBuf)

	fullTypes := g.FullTypes()
	typesMap := make(map[string]FullType, len(fullTypes))
	imports := map[string]string{
		"context":                             "",
		"bytes":                               "",
		"github.com/99designs/gqlgen/graphql": "",
		// "github.com/99designs/gqlgen/graphql/introspection": "",
		"github.com/vektah/gqlparser/v2/ast": "",
		"github.com/vektah/gqlparser/v2":     "gqlparser",
	}

	for _, v := range fullTypes {
		typesMap[v.Name] = v

		if _, ok := imports[v.PackageName]; !ok {
			imports[v.PackagePath] = ""
		}
	}

	d := Data{
		PackageName:         "graph",
		Imports:             imports,
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
