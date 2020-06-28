package builder

import (
	"bytes"
	"go/format"
	"go/types"
	"io"
	"text/template"

	"github.com/dave/jennifer/jen"
)

func (g *gqlBuilder) Code(w io.Writer) error {
	t := template.Must(template.New("code.tmpl").Funcs(funcMap).ParseFiles("code.tmpl"))

	type Data struct {
		PackageName         string
		Imports             []string
		Types               []FullType
		Dependencies        map[string]*types.Var
		DependenciesNameMap map[string]string
	}

	d := Data{
		PackageName:         "generated",
		Imports:             []string{"context"},
		Dependencies:        g.dependencies,
		DependenciesNameMap: g.dependenciesNameMap,
		Types:               g.FullTypes(),
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

func assertType(id, index string, typ types.Type) jen.Code {
	typPkg, typName := typePath(typ)
	return jen.Id(id).Index(jen.Lit(index)).Assert(jen.Do(func(s *jen.Statement) {
		if typPkg != "" {
			s.Qual(typPkg, typName)
		} else {
			s.Id(typName)
		}
	}))
}

func typePath(typ types.Type) (string, string) {
	switch x := typ.(type) {
	case *types.Basic:
		return "", x.Name()
	case *types.Named:
		obj := x.Obj()
		return obj.Pkg().Path(), obj.Name()
	default:
		return "", x.String()
	}
}

func typeCode(typ types.Type) jen.Code {
	return nil
}
