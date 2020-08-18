package builder

import (
	"fmt"
	"strconv"
	"strings"
	"text/template"
	"unicode"

	"github.com/sijad/gentle/builder/internal"
)

//go:generate go run github.com/go-bindata/go-bindata/go-bindata -o=internal/bindata.go -pkg=internal -modtime=1 ./template/...

func init() {
	templates.Funcs(funcMap)
	for _, asset := range internal.AssetNames() {
		templates = template.Must(templates.New(asset).Parse(string(internal.MustAsset(asset))))
	}
}

var templates = template.New("templates")

var funcMap = template.FuncMap{
	"gqlType":                   gqlType,
	"lowerCaseFirst":            lowerCaseFirst,
	"upperCaseFirst":            upperCaseFirst,
	"quote":                     strconv.Quote,
	"typeGo":                    typeGo,
	"rawQuote":                  rawQuote,
	"typeMarshalerMethodName":   typeMarshalerMethodName,
	"typeUnmarshalerMethodName": typeUnmarshalerMethodName,
	"isBasicScalar":             isBasicScalar,
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
		panic("not implemented")
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
		return *typ.Name
	case OBJECT:
		return "Object" + *typ.Name
	default:
		panic(fmt.Sprintf("cannot marshal %s", typ.Kind))
	}
}

func typeUnmarshalerMethodName(typ *TypeRef) string {
	return "Unmarshal" + _typeMarshalerMethodName(typ)
}

func isBasicScalar(typ *TypeRef) bool {
	if typ.Kind == SCALAR {
		if _, exists := basicTypesGOMap[*typ.Name]; exists {
			return true
		}
	}
	return false
}
