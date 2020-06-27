package builder

import (
	"fmt"
	"go/types"
	"log"
	"os"
	"text/template"
	"unicode"

	"github.com/jensneuse/graphql-go-tools/pkg/introspection"
)

func (g *gqlBuilder) SDL() string {
	const sdl = `
{{- range $t := .Types -}}
{{- if eq $t.Kind 0 -}}
scalar {{$t.Name}}
{{- end}}
{{- if or (eq $t.Kind 3) (eq $t.Kind 7) -}}
{{if eq $t.Kind 3}}type{{else}}input{{end}} {{$t.Name}} {
{{- range $f := $t.Fields}}
  {{$f.Name | lowerFirstRune}}
  {{- if gt (len $f.Args) 0 -}}
    (
      {{- range $i, $a := $f.Args}}
        {{- if $i}}{{", "}}{{end -}}
        {{- $a.Name | lowerFirstRune }}: {{$a.Type | gqlType -}}
      {{ end -}}
    )
  {{- end -}}
  : {{$f.Type | gqlType}}
{{- end}}
}
{{- end}}

{{end -}}
`
	funcMap := template.FuncMap{
		"gqlType":        gqlType,
		"lowerFirstRune": lowerFirstRune,
	}

	t := template.Must(template.New("sdl").Funcs(funcMap).Parse(sdl))

	type Data struct {
		Types []FullType
	}
	d := Data{g.FullTypes()}

	err := t.Execute(os.Stdout, d)
	if err != nil {
		log.Println("executing template:", err)
	}

	return ""
}

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

func nonNilAbleTypeRef(typ *introspection.TypeRef) *introspection.TypeRef {
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
	case introspection.SCALAR, introspection.INPUTOBJECT, introspection.OBJECT:
		return *typ.Name
	default:
		panic("not implimented")
	}
}
