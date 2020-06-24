package builder

import (
	"fmt"
	"go/types"
	"log"
	"os"
	"text/template"

	"github.com/jensneuse/graphql-go-tools/pkg/introspection"
)

type GQLType struct {
	Id  string
	Typ introspection.FullType
}

type gqlBuilder struct {
	types   []GQLType
	typeMap map[string]int
}

func (g *gqlBuilder) GetFullType(name string) *introspection.FullType {
	index, ok := g.typeMap[name]
	fmt.Println(index)
	if !ok {
		return nil
	}
	return &g.types[index].Typ
}

func (g *gqlBuilder) GetTypeRef(name string) *introspection.TypeRef {
	return nil
}

func (g *gqlBuilder) AddFullType(id string, typ introspection.FullType) error {
	if pos, ok := g.typeMap[typ.Name]; ok {
		if g.types[pos].Id != id {
			return fmt.Errorf("Same type %s exists with different id (%s)", id, g.types[pos].Id)
		}
		return nil
	}

	g.types = append(g.types, GQLType{id, typ})
	g.typeMap[typ.Name] = len(g.types) - 1

	return nil
}

func (g *gqlBuilder) ImportType(name *string, t types.Type, nilAble bool) (*introspection.TypeRef, error) {
	switch x := t.(type) {
	case *types.Basic:
		name := basicTypeName(x.Kind())
		return nilAbleTypeRef(&introspection.TypeRef{
			Kind: introspection.SCALAR,
			Name: &name,
		}, nilAble), nil
	case *types.Slice:
		ofType, err := g.ImportType(nil, x.Elem(), false)
		if err != nil {
			return nil, err
		}
		return nilAbleTypeRef(&introspection.TypeRef{
			Kind:   introspection.LIST,
			OfType: ofType,
		}, nilAble), nil
	case *types.Pointer:
		if nilAble {
			return nil, fmt.Errorf("Multiple indirection (*pointer) is not supported")
		}
		return g.ImportType(nil, x.Elem(), true)
	case *types.Named:
		name := x.Obj().Name()
		id := x.Obj().Id()
		strct := x.Underlying().(*types.Struct)

		var fields []introspection.Field
		for i := 0; i < strct.NumFields(); i++ {
			typeField := strct.Field(i)

			if !typeField.Exported() {
				continue
			}

			fieldName := typeField.Name()
			field := introspection.NewField()
			field.Name = fieldName
			// TODO field.Description

			ftyp, err := g.ImportType(nil, typeField.Type(), false)
			if err != nil {
				return nil, err
			}

			field.Type = *ftyp
			fields = append(fields, field)
		}

		fullType := introspection.NewFullType()
		fullType.Kind = introspection.OBJECT
		fullType.Name = name
		// TODO fullType.Description
		fullType.Fields = fields

		// TODO ADD resolvers

		g.AddFullType(id, fullType)
		return nilAbleTypeRef(&introspection.TypeRef{
			Kind: introspection.SCALAR,
			Name: &name,
		}, nilAble), nil
	default:
		return nil, errors.New("not implimented")
	}
}

func (g *gqlBuilder) ImportQueryType(typ types.Type) error {
	_, err := g.ImportType(nil, typ, false)
	// g.schema.QueryType = &introspection.TypeName{string(name)}
	// TODO check if imported type is object, panic if not
	return err
}

func (g *gqlBuilder) SDL() string {
	const sdl = `
{{- range $t := .Types -}}
type {{$t.Name}} {
{{- range $f := $t.Fields}}
  {{$f.Name | lowerFirstRune}}: {{$f.Type | gqlType}}
{{- end}}
}
{{- end -}}
`
	funcMap := template.FuncMap{
		"gqlType":        gqlType,
		"lowerFirstRune": lowerFirstRune,
	}

	t := template.Must(template.New("sdl").Funcs(funcMap).Parse(sdl))

	type Data struct {
		Types []introspection.FullType
	}
	var types []introspection.FullType
	for _, t := range g.types {
		types = append(types, t.Typ)
	}
	d := Data{types}

	err := t.Execute(os.Stdout, d)
	if err != nil {
		log.Println("executing template:", err)
	}

	return ""
}

func NewGQLBuilder() *gqlBuilder {
	b := &gqlBuilder{}
	b.typeMap = make(map[string]int)
	return b
}
