package builder

import (
	"encoding/json"
	"fmt"
	"go/types"
	"testing"

	"github.com/jensneuse/graphql-go-tools/pkg/introspection"
	"golang.org/x/tools/go/packages"
)

func TestVarType(t *testing.T) {
	cfg := &packages.Config{Mode: packages.NeedTypesInfo | packages.NeedTypes}
	pkgs, err := packages.Load(cfg, "github.com/sijad/gentle/builder/testdata/basic")

	if err != nil {
		t.Error(err)
	}

	builder := NewGQLBuilder()
	generator := introspection.NewGenerator()

	for _, pkg := range pkgs {
		for _, v := range pkg.TypesInfo.Defs {
			if v == nil {
				continue
			}

			typeName, ok := v.(*types.TypeName)

			if !ok {
				continue
			}

			if _, err := builder.ImportType(typeName.Type(), false); err != nil {
				t.Error(err)
			}
		}
	}

	schema := introspection.NewSchema()
	schema.QueryType = &introspection.TypeName{Name: "MyStruct"}
	var types []introspection.FullType
	for _, t := range builder.types {
		types = append(types, t.Typ)
	}

	schema.Types = []introspection.FullType(types)
	generator.Data = &introspection.Data{Schema: schema}
	json.Marshal(generator)

	if err != nil {
		t.Error(err)
	}

	fmt.Println(builder.SDL())
}
