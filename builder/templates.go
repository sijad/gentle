package builder

import (
	"fmt"
	"strconv"
	"strings"
	"text/template"
	"unicode"

	"github.com/jensneuse/graphql-go-tools/pkg/introspection"
)

var funcMap = template.FuncMap{
	"gqlType":                 gqlType,
	"lowerCaseFirst":          lowerCaseFirst,
	"upperCaseFirst":          upperCaseFirst,
	"quote":                   strconv.Quote,
	"typeGo":                  typeGo,
	"rawQuote":                rawQuote,
	"typeMarshalerMethodName": typeMarshalerMethodName,
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

func rawQuote(s string) string {
	return "`" + strings.Replace(s, "`", "`+\"`\"+`", -1) + "`"
}

var basicTypesGOMap = map[string]string{
	"Boolean": "bool",
	"Int":     "int",
	"Uint":    "uint",
	"Int8":    "int8",
	"Uint8":   "uint8",
	"Int16":   "int16",
	"Uint16":  "uint16",
	"Int32":   "int32",
	"Uint32":  "uint32",
	"Int64":   "int64",
	"Uint64":  "uint64",
	"Float":   "float32",
	"Float64": "float64",
	"String":  "string",
}

func typeGo(typ *introspection.TypeRef, fullTypes map[string]FullType) string {
	typName := "*"
	switch typ.Kind {
	case introspection.LIST:
		typName += "[]" + typeGo(typ.OfType, fullTypes)
	case introspection.NONNULL:
		typName = typeGo(typ.OfType, fullTypes)[1:]
	case introspection.SCALAR:
		if val, ok := basicTypesGOMap[*typ.Name]; ok {
			typName += val
		} else {
			name := *typ.Name
			fullType, ok := fullTypes[name]
			if !ok {
				panic(fmt.Sprintf("cannot fild full type %s", name))
			}
			typName += fullType.PackageName + "." + name
		}
	case introspection.OBJECT:
		name := *typ.Name
		fullType, ok := fullTypes[name]
		if !ok {
			panic(fmt.Sprintf("cannot fild full type %s", name))
		}
		typName += fullType.PackageName + "." + name
	default:
		panic(fmt.Sprintf("cannot convert %s to golang type", typ.Kind))
	}
	return typName
}

func typeMarshalerMethodName(typ *introspection.TypeRef) string {
	return "Marshal" + _typeMarshalerMethodName(typ)
}

func _typeMarshalerMethodName(typ *introspection.TypeRef) string {
	switch typ.Kind {
	case introspection.LIST:
		return "List" + _typeMarshalerMethodName(typ.OfType)
	case introspection.NONNULL:
		return "NonNull" + _typeMarshalerMethodName(typ.OfType)
	case introspection.SCALAR:
		return "TODOScalar" + *typ.Name
	case introspection.OBJECT:
		return "Object" + *typ.Name
	default:
		panic(fmt.Sprintf("cannot marshal %s", typ.Kind))
	}
}
