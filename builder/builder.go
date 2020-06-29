package builder

import (
	"fmt"
	"go/types"
	"reflect"
	"strings"

	"github.com/jensneuse/graphql-go-tools/pkg/introspection"
	"github.com/sijad/gentle"
	"golang.org/x/tools/go/packages"
)

var scalarInterface = reflect.TypeOf(struct{ gentle.Scalar }{}).Field(0).Type

type Field struct {
	introspection.Field
	IsMethod bool
	HasError bool
	Params   []*types.Var
}

type FullType struct {
	introspection.FullType
	Id     string
	Type   *types.Named
	Fields []Field
}

type gqlBuilder struct {
	types               []FullType
	typeMap             map[string]int
	processingFullTypes map[string]bool
	scalarInterface     *types.Interface
	dependencies        map[string]*types.Var
	dependenciesNameMap map[string]string
}

func (g *gqlBuilder) GetFullType(name string) *FullType {
	index, ok := g.typeMap[name]
	if !ok {
		return nil
	}
	return &g.types[index]
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

func (g *gqlBuilder) AddFullType(typ FullType) error {
	if pos, ok := g.typeMap[typ.Name]; ok {
		if g.types[pos].Id != typ.Id {
			return fmt.Errorf("Same type %s exists with different id (%s)", typ.Id, g.types[pos].Id)
		}
		return nil
	}

	g.types = append(g.types, typ)
	g.typeMap[typ.Name] = len(g.types) - 1

	return nil
}

func (g *gqlBuilder) AddDependency(dep *types.Var) error {
	key := dep.Type().String()
	depArgName := dep.Name()

	if key == "context.Context" {
		if depArgName != "ctx" {
			return fmt.Errorf("context.Context name input need to be ctx got %s", depArgName)
		}
		return nil
	}

	if d, ok := g.dependencies[key]; ok {
		depArgName = d.Name()
	} else {
		g.dependencies[key] = dep
	}

	g.dependenciesNameMap[dep.Name()] = depArgName

	// TODO if a dependency with same name but different
	// type exists return an error

	return nil
}

func (g *gqlBuilder) ImportType(t types.Type) (*introspection.TypeRef, error) {
	switch x := t.(type) {
	case *types.Basic:
		name := basicTypeName(x.Kind())
		return nonNilAbleTypeRef(&introspection.TypeRef{
			Kind: introspection.SCALAR,
			Name: &name,
		}), nil
	case *types.Slice:
		ofType, err := g.ImportType(x.Elem())
		if err != nil {
			return nil, err
		}
		return nonNilAbleTypeRef(&introspection.TypeRef{
			Kind:   introspection.LIST,
			OfType: ofType,
		}), nil
	case *types.Pointer:
		if _, ok := x.Elem().(*types.Pointer); ok {
			return nil, fmt.Errorf("Multiple indirection (*pointer) is not supported")
		}
		if ref, err := g.ImportType(x.Elem()); err != nil {
			return nil, err
		} else {
			return ref.OfType, nil
		}
	case *types.Named:
		name := x.Obj().Name()
		id := x.String()
		g.processingFullTypes[id] = true

		fullType := FullType{}
		fullType.Id = id
		fullType.Type = x
		fullType.Name = name

		if types.Implements(t, g.scalarInterface) {
			fullType.Kind = introspection.SCALAR
			g.AddFullType(fullType)
			return nonNilAbleTypeRef(&introspection.TypeRef{
				Kind: introspection.SCALAR,
				Name: &name,
			}), nil
		}
		strct, ok := x.Underlying().(*types.Struct)
		if !ok {
			return nil, fmt.Errorf("only named structs are supported")
		}

		// returns underlying element type and prevents infinite loop
		underlingType := func(t types.Type) (*introspection.TypeRef, error) {
			typ := t
			isPtr := false
			if ptr, ok := typ.(*types.Pointer); ok {
				isPtr = true
				typ = ptr.Elem()
			}

			if typ, ok := typ.(*types.Named); ok {
				if g.processingFullTypes[typ.String()] {
					ref := &introspection.TypeRef{
						Kind: introspection.SCALAR,
						Name: &name,
					}
					if isPtr {
						return ref, nil
					}
					return nonNilAbleTypeRef(ref), nil
				}
			}

			return g.ImportType(t)
		}

		var fields []Field
		for i := 0; i < strct.NumFields(); i++ {
			typeField := strct.Field(i)

			if !typeField.Exported() {
				continue
			}

			fieldName := typeField.Name()
			field := Field{}
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
			field := Field{}
			field.Name = fieldName
			field.IsMethod = true
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
					if err := g.AddDependency(param); err != nil {
						return nil, err
					}
				}
				field.Params = append(field.Params, param)
			}

			methodResults := methodSig.Results()

			switch methodResults.Len() {
			case 0:
				return nil, fmt.Errorf("resolvers must return at least one result")
			case 1:
			case 2:
				if secTyp, ok := methodResults.At(1).Type().(*types.Named); !ok || secTyp.Obj().Id() != "_.error" {
					return nil, fmt.Errorf("second resolvers result must be an error")
				}
				field.HasError = true
			default:
				return nil, fmt.Errorf("resolvers must have exactly two results and second one should be an error")
			}

			resultTyp := methodResults.At(0)
			rtyp, err := underlingType(resultTyp.Type())
			if err != nil {
				return nil, err
			}

			field.Type = *rtyp
			fields = append(fields, field)
		}

		if len(fields) == 0 {
			return nil, fmt.Errorf("Named structs must have at aleast one exported property")
		}

		fullType.Kind = kind
		// TODO fullType.Description
		fullType.Fields = fields

		g.AddFullType(fullType)
		return nonNilAbleTypeRef(&introspection.TypeRef{
			Kind: kind,
			Name: &name,
		}), nil
	default:
		return nil, fmt.Errorf("not implimented")
	}
}

func (g *gqlBuilder) FullTypes() (types []FullType) {
	return g.types
}

func NewGQLBuilder() *gqlBuilder {
	b := &gqlBuilder{}
	b.typeMap = make(map[string]int)
	b.processingFullTypes = make(map[string]bool)
	b.dependencies = make(map[string]*types.Var)
	b.dependenciesNameMap = make(map[string]string)

	pkgs, _ := packages.Load(&packages.Config{Mode: packages.LoadSyntax}, scalarInterface.PkgPath())
	b.scalarInterface = pkgs[0].Types.Scope().Lookup(scalarInterface.Name()).Type().Underlying().(*types.Interface)

	return b
}

func basicTypeName(b types.BasicKind) string {
	switch b {
	case types.Bool:
		return "Boolean"
	case types.Int, types.Int8, types.Int16, types.Int32, types.Int64,
		types.Uint, types.Uint8, types.Uint16, types.Uint32, types.Uint64:
		return "Int"
	case types.Float32, types.Float64:
		return "Float"
	case types.String:
		return "String"
	default:
		panic(fmt.Sprintf("Cannot determine type of %T", b))
	}
}

func nonNilAbleTypeRef(typ *introspection.TypeRef) *introspection.TypeRef {
	return &introspection.TypeRef{
		Kind:   introspection.NONNULL,
		OfType: typ,
	}
}
