package builder

import (
	"go/types"
	"strconv"
	"text/template"
	"unicode"

	"github.com/jensneuse/graphql-go-tools/pkg/introspection"
)

var funcMap = template.FuncMap{
	"gqlType":        gqlType,
	"lowerCaseFirst": lowerCaseFirst,
	"upperCaseFirst": upperCaseFirst,
	"quote":          strconv.Quote,
	"typeGo":         typeGo,
}

func lowerCaseFirst(s string) string {
	r := []rune(s)
	r[0] = unicode.ToLower(r[0])
	return string(r)
}

func upperCaseFirst(s string) string {
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

func gqlType(typ *introspection.TypeRef) string {
	switch typ.Kind {
	case introspection.LIST:
		return "[" + gqlType(typ.OfType) + "]"
	case introspection.NONNULL:
		return gqlType(typ.OfType) + "!"
	case introspection.SCALAR, introspection.INPUTOBJECT, introspection.OBJECT:
		return *typ.Name
	default:
		panic("not implimented")
	}
}

func typeGo(typ types.Type) string {
	switch x := typ.(type) {
	case *types.Named:
		obj := x.Obj()
		return obj.Pkg().Name() + "." + obj.Name()
	default:
		return x.String()
	}
}
