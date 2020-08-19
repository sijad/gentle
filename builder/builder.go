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
		name, err := basicTypeName(x.Kind())
		if err != nil {
			return nil, fmt.Errorf("Cannot determine type of %x", x)
		}
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
		fullType.Name = name
		fullType.PackageName = x.Obj().Pkg().Name()
		fullType.PackagePath = x.Obj().Pkg().Path()

		if types.Implements(t, g.scalarInterface) || types.Implements(types.NewPointer(t), g.scalarInterface) {
			fullType.Kind = SCALAR
			if err := g.AddFullType(fullType); err != nil {
				return nil, err
			}
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
			var strct *types.Struct
			switch t := namedTyp.Underlying().(type) {
			case *types.Struct:
				strct = t
			default:
				return fmt.Errorf("only named structs are supported, got %s", t.String())
			}

			for i := 0; i < strct.NumFields(); i++ {
				typeField := strct.Field(i)

				if !typeField.Exported() {
					continue
				}

				if typeField.Embedded() {
					if t, ok := typeField.Type().(*types.Named); ok {
						// TODO panic on duplicate fields
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
						if argsStruct.NumFields() == 0 {
							return fmt.Errorf("args must have at least one field")
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
						field.ArgsType = argsStruct
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

		if err := g.AddFullType(fullType); err != nil {
			return nil, err
		}
		return nonNullAbleTypeRef(&TypeRef{
			Kind: kind,
			Name: &name,
		}), nil
	default:
		return nil, fmt.Errorf("not implemented")
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

	pkgs, _ := packages.Load(&packages.Config{Mode: packages.NeedTypes | packages.NeedTypesInfo}, scalarInterface.PkgPath())
	b.scalarInterface = pkgs[0].Types.Scope().Lookup(scalarInterface.Name()).Type().Underlying().(*types.Interface)

	return b
}

func basicTypeName(b types.BasicKind) (string, error) {
	var name string
	switch b {
	case types.Bool:
		name = "Boolean"
	case types.Int:
		name = "Int"
	case types.Uint:
		name = "Uint"
	case types.Int8:
		name = "Int8"
	case types.Uint8:
		name = "Uint8"
	case types.Int16:
		name = "Int16"
	case types.Uint16:
		name = "Uint16"
	case types.Int32:
		name = "Int32"
	case types.Uint32:
		name = "Uint32"
	case types.Int64:
		name = "Int64"
	case types.Uint64:
		name = "Uint64"
	case types.Float32:
		name = "Float"
	case types.Float64:
		name = "Float64"
	case types.String:
		name = "String"
	case types.Invalid:
		return "", fmt.Errorf("basic type is invalid")
	}

	if name != "" {
		return name, nil
	}

	return "", fmt.Errorf("basic type %v is not supported", b)
}

func nonNullAbleTypeRef(typ *TypeRef) *TypeRef {
	return &TypeRef{
		Kind:   NONNULL,
		OfType: typ,
	}
}
