package builder

import (
	"fmt"
	"go/types"

	"github.com/jensneuse/graphql-go-tools/pkg/introspection"
)

type GQLType struct {
	Id  string
	Typ introspection.FullType
}

type gqlBuilder struct {
	types     []GQLType
	typeMap   map[string]int
	typeIdMap map[string]int
}

func (g *gqlBuilder) GetFullType(name string) *introspection.FullType {
	return nil
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
		ofType, err := g.ImportType(nil, x.Elem(), nilAble)
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
		return g.ImportType(nil, x.Underlying(), true)
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

			fieldName := firstLowerRune(typeField.Name())
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
	case *types.Struct:
		fmt.Println(x)
		return nil, nil
	default:
		panic(fmt.Sprintf("%T not implimented", x))
	}
}

func (g *gqlBuilder) ImportQueryType(typ types.Type) error {
	_, err := g.ImportType(nil, typ, false)
	// g.schema.QueryType = &introspection.TypeName{string(name)}
	// TODO check if imported type is object, panic if not
	return err
}

func NewGQLBuilder() *gqlBuilder {
	return &gqlBuilder{}
}
