package builder

import (
	"fmt"
	"go/types"
	"log"
	"os"
	"reflect"
	"strings"
	"text/template"

	"github.com/jensneuse/graphql-go-tools/pkg/introspection"
	"github.com/sijad/gentle"
	"golang.org/x/tools/go/packages"
)

var scalarInterface = reflect.TypeOf(struct{ gentle.Scalar }{}).Field(0).Type

type GQLType struct {
	Id  string
	Typ introspection.FullType
}

type gqlBuilder struct {
	types               []GQLType
	typeMap             map[string]int
	processingFullTypes map[string]bool
	scalarInterface     *types.Interface
}

func (g *gqlBuilder) GetFullType(name string) *introspection.FullType {
	index, ok := g.typeMap[name]
	if !ok {
		return nil
	}
	return &g.types[index].Typ
}

func (g *gqlBuilder) GetTypeRef(name string) *introspection.TypeRef {
	fullType := g.GetFullType(name)
	if fullType != nil {
		return &introspection.TypeRef{
			Kind: fullType.Kind,
			Name: &fullType.Name,
		}
	}
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

func (g *gqlBuilder) ImportType(t types.Type, nilAble bool) (*introspection.TypeRef, error) {
	switch x := t.(type) {
	case *types.Basic:
		name := basicTypeName(x.Kind())
		return nilAbleTypeRef(&introspection.TypeRef{
			Kind: introspection.SCALAR,
			Name: &name,
		}, nilAble), nil
	case *types.Slice:
		ofType, err := g.ImportType(x.Elem(), false)
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
		return g.ImportType(x.Elem(), true)
	case *types.Named:
		name := x.Obj().Name()
		id := x.Obj().Id()
		g.processingFullTypes[id] = true

		if types.Implements(t, g.scalarInterface) {
			fullType := introspection.NewFullType()
			fullType.Kind = introspection.SCALAR
			fullType.Name = name

			g.AddFullType(id, fullType)
			return nilAbleTypeRef(&introspection.TypeRef{
				Kind: introspection.SCALAR,
				Name: &name,
			}, nilAble), nil
		}
		strct, ok := x.Underlying().(*types.Struct)
		if !ok {
			return nil, fmt.Errorf("only named structs are supported")
		}

		// returns underlying element type and prevents infinite loop
		underlingType := func(t types.Type) (*introspection.TypeRef, error) {
			typ := t
			if ptr, ok := typ.(*types.Pointer); ok {
				typ = ptr.Elem()
			}

			if typ, ok := typ.(*types.Named); ok {
				if ref := g.GetTypeRef(typ.Obj().Name()); ref != nil {
					return ref, nil
				}
				if g.processingFullTypes[typ.Obj().Id()] {
					return &introspection.TypeRef{
						Kind: introspection.SCALAR,
						Name: &name,
					}, nil
				}
			}

			return g.ImportType(typ, false)
		}

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

			ftyp, err := underlingType(typeField.Type())
			if err != nil {
				return nil, err
			}
			field.Type = *ftyp
			fields = append(fields, field)
		}

		kind := introspection.OBJECT
		if strings.HasSuffix(name, "Input") {
			kind = introspection.INPUTOBJECT
		}
		for i := 0; i < x.NumMethods(); i++ {
			method := x.Method(i)

			if !method.Exported() {
				continue
			}

			if kind == introspection.INPUTOBJECT {
				return nil, fmt.Errorf("Input types can not have resolvers")
			}

			methodSig := method.Type().(*types.Signature)

			fieldName := method.Name()
			field := introspection.NewField()
			field.Name = fieldName
			// TODO field.Description

			params := methodSig.Params()
			for i := 0; i < params.Len(); i++ {
				param := params.At(i)
				if param.Name() == "args" {
					var argsStruct *types.Struct
					switch xx := param.Type().(type) {
					case *types.Named:
						argsStruct = xx.Underlying().(*types.Struct)
					case *types.Struct:
						argsStruct = xx
					default:
						return nil, fmt.Errorf("args can only be struct")
					}
					for i := 0; i < argsStruct.NumFields(); i++ {
						argField := argsStruct.Field(i)

						if !argField.Exported() {
							continue
						}

						atyp, err := underlingType(argField.Type())
						if err != nil {
							return nil, err
						}

						field.Args = append(field.Args, introspection.InputValue{
							Name: argField.Name(),
							Type: *atyp,
							// TODO Description: "",
						})
					}
				} else {
					// TODO add needed injectables
					panic("injectables are not implimented yet")
				}
			}

			methodResults := methodSig.Results()

			switch methodResults.Len() {
			case 0:
				return nil, fmt.Errorf("resolvers must return at least one result")
			// TODO case 1: // we need let field know if it has only one result somhow
			// maybe we have to use somthing else instead if introspection
			case 2:
				if secTyp, ok := methodResults.At(1).Type().(*types.Named); !ok || secTyp.Obj().Id() != "_.error" {
					return nil, fmt.Errorf("second resolvers result must be an error")
				}
			default:
				return nil, fmt.Errorf("resolvers must have exactly two results and second one should be an error")
			}

			var rtyp *introspection.TypeRef
			resultTyp := methodResults.At(0)
			rtyp, err := underlingType(resultTyp.Type())
			if err != nil {
				return nil, err
			}

			// TODO let field knows if it returs error

			field.Type = *rtyp
			fields = append(fields, field)
		}

		if len(fields) == 0 {
			return nil, fmt.Errorf("Named structs must have at aleast one exported property")
		}

		fullType := introspection.NewFullType()
		fullType.Kind = kind
		fullType.Name = name
		// TODO fullType.Description
		fullType.Fields = fields

		g.AddFullType(id, fullType)
		return nilAbleTypeRef(&introspection.TypeRef{
			Kind: kind,
			Name: &name,
		}, nilAble), nil
	default:
		return nil, fmt.Errorf("not implimented")
	}
}

func (g *gqlBuilder) SDL() string {
	const sdl = `
{{- range $t := .Types -}}
{{- if eq $t.Kind 0 -}}
scalar {{$t.Name}}
{{- end}}
{{- if or (eq $t.Kind 3) (eq $t.Kind 7) -}}
{{if eq $t.Kind 3}}type{{else}}input{{end}} {{$t.Name}} {
{{- range $f := $t.Fields}}
  {{$f.Name | lowerFirstRune}}
  {{- if gt (len $f.Args) 0 -}}
    (
      {{- range $i, $a := $f.Args}}
        {{- if $i}}{{", "}}{{end -}}
        {{- $a.Name | lowerFirstRune }}: {{$a.Type | gqlType -}}
      {{ end -}}
    )
  {{- end -}}
  : {{$f.Type | gqlType}}
{{- end}}
}
{{- end}}

{{end -}}
`
	funcMap := template.FuncMap{
		"gqlType":        gqlType,
		"lowerFirstRune": lowerFirstRune,
	}

	t := template.Must(template.New("sdl").Funcs(funcMap).Parse(sdl))

	type Data struct {
		Types []introspection.FullType
	}
	d := Data{g.FullTypes()}

	err := t.Execute(os.Stdout, d)
	if err != nil {
		log.Println("executing template:", err)
	}

	return ""
}

func (g *gqlBuilder) FullTypes() (types []introspection.FullType) {
	for _, t := range g.types {
		types = append(types, t.Typ)
	}
	return
}

func NewGQLBuilder() *gqlBuilder {
	b := &gqlBuilder{}
	b.typeMap = make(map[string]int)
	b.processingFullTypes = make(map[string]bool)

	pkgs, _ := packages.Load(&packages.Config{Mode: packages.LoadSyntax}, scalarInterface.PkgPath())
	b.scalarInterface = pkgs[0].Types.Scope().Lookup(scalarInterface.Name()).Type().Underlying().(*types.Interface)

	return b
}
