package builder

import (
	"fmt"
	"go/types"
	"unicode"

	"github.com/jensneuse/graphql-go-tools/pkg/introspection"
)

func lowerFirstRune(s string) string {
	r := []rune(s)
	r[0] = unicode.ToLower(r[0])
	return string(r)
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

func nilAbleTypeRef(typ *introspection.TypeRef, nilAble bool) *introspection.TypeRef {
	if nilAble {
		return typ
	}

	return &introspection.TypeRef{
		Kind:   introspection.NONNULL,
		OfType: typ,
	}
}

func gqlType(typ *introspection.TypeRef) string {
	switch typ.Kind {
	case introspection.LIST:
		return "[" + gqlType(typ.OfType) + "]"
	case introspection.NONNULL:
		return gqlType(typ.OfType) + "!"
	case introspection.SCALAR:
		return *typ.Name
	default:
		panic("not implimented")
	}
}
