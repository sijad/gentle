package builder

import (
	"bytes"
	"go/format"
	"go/types"
	"io"
)

type CodeData struct {
	PackageName         string
	Imports             map[string]string
	Types               map[string]FullType
	Dependencies        map[string]*types.Var
	DependenciesNameMap map[string]string
	Marshallers         []TypeRef
	Sdl                 string
}

func (g *gqlBuilder) Code(w io.Writer) error {
	var sdlBuf bytes.Buffer
	g.SDL(&sdlBuf)

	fullTypes := g.FullTypes()
	typesMap := make(map[string]FullType, len(fullTypes))
	imports := map[string]string{
		"context":                             "",
		"bytes":                               "",
		"github.com/99designs/gqlgen/graphql": "",
		// "github.com/99designs/gqlgen/graphql/introspection": "",
		"github.com/vektah/gqlparser/v2/ast":     "",
		"github.com/vektah/gqlparser/v2":         "gqlparser",
		"github.com/sijad/gentle/encoding/basic": "encodingBasic", // TODO check if we really need basic encoding
	}

	marshallers := []TypeRef{}
	marshallersMap := make(map[string]bool)

	for _, v := range fullTypes {
		typesMap[v.Name] = v

		if _, ok := imports[v.PackageName]; !ok {
			imports[v.PackagePath] = ""
		}

		for _, field := range v.Fields {
			typ := &field.Type
			for typ != nil {
				name := typeMarshalerMethodName(typ)
				if _, ok := marshallersMap[name]; !ok {
					marshallersMap[name] = true
					marshallers = append(marshallers, *typ)
				}
				typ = typ.OfType
			}
		}

	}

	d := CodeData{
		PackageName:         "graph",
		Imports:             imports,
		Dependencies:        g.dependencies,
		DependenciesNameMap: g.dependenciesNameMap,
		Types:               typesMap,
		Sdl:                 sdlBuf.String(),
		Marshallers:         marshallers,
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
