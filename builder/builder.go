package builder

import (
	"fmt"
	"go/types"
	"reflect"
	"strings"

	"github.com/sijad/gentle"
	"golang.org/x/tools/go/packages"
)

var scalarInterface = reflect.TypeOf(struct{ gentle.Scalar }{}).Field(0).Type

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

func (g *gqlBuilder) GetTypeRef(name string) *TypeRef {
	fullType := g.GetFullType(name)
	if fullType != nil {
		return &TypeRef{
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

func (g *gqlBuilder) ImportType(t types.Type) (*TypeRef, error) {
	switch x := t.(type) {
	case *types.Basic:
		name := basicTypeName(x.Kind())
		return nonNullAbleTypeRef(&TypeRef{
			Kind: SCALAR,
			Name: &name,
		}), nil
	case *types.Slice:
		ofType, err := g.ImportType(x.Elem())
		if err != nil {
			return nil, err
		}
		return nonNullAbleTypeRef(&TypeRef{
			Kind:   LIST,
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
		// fullType.Type = x
		fullType.Name = name

		if types.Implements(t, g.scalarInterface) {
			fullType.Kind = SCALAR
			g.AddFullType(fullType)
			return nonNullAbleTypeRef(&TypeRef{
				Kind: SCALAR,
				Name: &name,
			}), nil
		}

		// returns underlying element type and prevents infinite loop
		underlingType := func(t types.Type) (*TypeRef, error) {
			typ := t
			isPtr := false
			if ptr, ok := typ.(*types.Pointer); ok {
				isPtr = true
				typ = ptr.Elem()
			}

			if typ, ok := typ.(*types.Named); ok {
				if g.processingFullTypes[typ.String()] {
					ref := &TypeRef{
						Kind: SCALAR,
						Name: &name,
					}
					if isPtr {
						return ref, nil
					}
					return nonNullAbleTypeRef(ref), nil
				}
			}

			return g.ImportType(t)
		}

		var kind TypeKind = OBJECT
		if strings.HasSuffix(name, "Input") {
			kind = INPUTOBJECT
		}

		var fields []Field

		var addStrucFields func(namedTyp *types.Named) error
		addStrucFields = func(namedTyp *types.Named) error {
			strct, ok := namedTyp.Underlying().(*types.Struct)
			if !ok {
				return fmt.Errorf("only named structs are supported")
			}

			for i := 0; i < strct.NumFields(); i++ {
				typeField := strct.Field(i)

				if !typeField.Exported() {
					continue
				}

				if typeField.Embedded() {
					if t, ok := typeField.Type().(*types.Named); ok {
						if err := addStrucFields(t); err != nil {
							return err
						}
						continue
					}
				}

				fieldName := typeField.Name()
				field := Field{}
				field.Name = fieldName
				// TODO field.Description

				ftyp, err := underlingType(typeField.Type())
				if err != nil {
					return err
				}
				field.Type = *ftyp
				fields = append(fields, field)
			}

			for i := 0; i < namedTyp.NumMethods(); i++ {
				method := namedTyp.Method(i)

				if !method.Exported() {
					continue
				}

				if kind == INPUTOBJECT {
					return fmt.Errorf("Input types can not have resolvers")
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
							return fmt.Errorf("args can only be struct")
						}
						for i := 0; i < argsStruct.NumFields(); i++ {
							argField := argsStruct.Field(i)

							if !argField.Exported() {
								continue
							}

							atyp, err := underlingType(argField.Type())
							if err != nil {
								return err
							}

							field.Args = append(field.Args, InputValue{
								Name: argField.Name(),
								Type: *atyp,
								// TODO Description: "",
							})
						}
						field.HasArgs = true
					} else {
						if err := g.AddDependency(param); err != nil {
							return err
						}
					}
					field.Params = append(field.Params, param)
				}

				methodResults := methodSig.Results()

				switch methodResults.Len() {
				case 0:
					return fmt.Errorf("resolvers must return at least one result")
				case 1:
				case 2:
					if secTyp, ok := methodResults.At(1).Type().(*types.Named); !ok || secTyp.Obj().Id() != "_.error" {
						return fmt.Errorf("second resolvers result must be an error")
					}
					field.HasError = true
				default:
					return fmt.Errorf("resolvers must have exactly two results and second one should be an error")
				}

				resultTyp := methodResults.At(0)
				rtyp, err := underlingType(resultTyp.Type())
				if err != nil {
					return err
				}

				field.Type = *rtyp
				fields = append(fields, field)
			}

			return nil
		}

		if err := addStrucFields(x); err != nil {
			return nil, err
		}

		if len(fields) == 0 {
			return nil, fmt.Errorf("Named structs must have at aleast one exported property")
		}

		fullType.Kind = kind
		// TODO fullType.Description
		fullType.Fields = fields
		fullType.PackageName = x.Obj().Pkg().Name()
		fullType.PackagePath = x.Obj().Pkg().Path()

		g.AddFullType(fullType)
		return nonNullAbleTypeRef(&TypeRef{
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
	case types.Int:
		return "Int"
	case types.Uint:
		return "Uint"
	case types.Int8:
		return "Int8"
	case types.Uint8:
		return "Uint8"
	case types.Int16:
		return "Int16"
	case types.Uint16:
		return "Uint16"
	case types.Int32:
		return "Int32"
	case types.Uint32:
		return "Uint32"
	case types.Int64:
		return "Int64"
	case types.Uint64:
		return "Uint64"
	case types.Float32:
		return "Float"
	case types.Float64:
		return "Float64"
	case types.String:
		return "String"
	default:
		panic(fmt.Sprintf("Cannot determine type of %T", b))
	}
}

func nonNullAbleTypeRef(typ *TypeRef) *TypeRef {
	return &TypeRef{
		Kind:   NONNULL,
		OfType: typ,
	}
}
