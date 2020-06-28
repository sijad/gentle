package builder

import (
	"text/template"
	"unicode"

	"github.com/jensneuse/graphql-go-tools/pkg/introspection"
)

var funcMap = template.FuncMap{
	"gqlType":        gqlType,
	"lowerCaseFirst": lowerCaseFirst,
	"upperCaseFirst": upperCaseFirst,
}

func lowerCaseFirst(s string) string {
	r := []rune(s)
	r[0] = unicode.ToLower(r[0])
	return string(r)
}

func upperCaseFirst(s string) string {
	r := []rune(s)
	r[0] = unicode.ToLower(r[0])
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
