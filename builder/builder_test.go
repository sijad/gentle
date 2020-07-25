package builder

import (
	"fmt"
	"go/types"
	"os"
	"testing"

	"github.com/jensneuse/graphql-go-tools/pkg/introspection"
	"golang.org/x/tools/go/packages"
)

func TestVarType(t *testing.T) {
	cfg := &packages.Config{Mode: packages.NeedTypesInfo | packages.NeedTypes}
	pkgs, err := packages.Load(cfg, "github.com/sijad/gentle/builder/testdata/hello-world")

	if err != nil {
		t.Error(err)
	}

	builder := NewGQLBuilder()

	for _, pkg := range pkgs {
		for _, v := range pkg.TypesInfo.Defs {
			if v == nil {
				continue
			}

			typeName, ok := v.(*types.TypeName)

			if !ok {
				continue
			}

			if _, err := builder.ImportType(typeName.Type()); err != nil {
				t.Error(err)
			}
		}
	}

	schema := introspection.NewSchema()
	schema.QueryType = &introspection.TypeName{Name: "Query"}

	err = builder.Code(os.Stdout)
	if err != nil {
		fmt.Println(err)
	}
}
