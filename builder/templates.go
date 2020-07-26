package builder

import (
	"fmt"
	"strconv"
	"strings"
	"text/template"
	"unicode"
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

func gqlType(typ *TypeRef) string {
	switch typ.Kind {
	case LIST:
		return "[" + gqlType(typ.OfType) + "]"
	case NONNULL:
		return gqlType(typ.OfType) + "!"
	case SCALAR, INPUTOBJECT, OBJECT:
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

var basicTypesUnmarshalMap = map[string]string{
	"Boolean": "",
	"Int":     "",
	"Uint":    "",
	"Int8":    "",
	"Uint8":   "",
	"Int16":   "",
	"Uint16":  "",
	"Int32":   "",
	"Uint32":  "",
	"Int64":   "",
	"Uint64":  "",
	"Float":   "",
	"Float64": "",
	"String":  "graphql.MarshalString",
}

func typeGo(typ *TypeRef, fullTypes map[string]FullType) string {
	typName := "*"
	switch typ.Kind {
	case LIST:
		typName += "[]" + typeGo(typ.OfType, fullTypes)
	case NONNULL:
		typName = typeGo(typ.OfType, fullTypes)[1:]
	case SCALAR:
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
	case OBJECT:
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

func typeMarshalerMethodName(typ *TypeRef) string {
	return "Marshal" + _typeMarshalerMethodName(typ)
}

func _typeMarshalerMethodName(typ *TypeRef) string {
	switch typ.Kind {
	case LIST:
		return "List" + _typeMarshalerMethodName(typ.OfType)
	case NONNULL:
		return "NonNull" + _typeMarshalerMethodName(typ.OfType)
	case SCALAR:
		if val, ok := basicTypesGOMap[*typ.Name]; ok {
			return val
		}
		return *typ.Name
	case OBJECT:
		return "Object" + *typ.Name
	default:
		panic(fmt.Sprintf("cannot marshal %s", typ.Kind))
	}
}
