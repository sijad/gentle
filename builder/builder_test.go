package builder

import (
	"go/types"
	"os"
	"testing"

	"golang.org/x/tools/go/packages"
)

func TestVarType(t *testing.T) {
	cfg := &packages.Config{Mode: packages.NeedTypesInfo | packages.NeedTypes}
	pkgs, err := packages.Load(cfg, "github.com/sijad/gentle/builder/testdata/simple")

	if err != nil {
		t.Error(err)
	}

	builder := NewGQLBuilder()
	pkg := pkgs[0]
	query := pkg.Types.Scope().Lookup("Query")
	typeName := query.(*types.TypeName)
	if _, err := builder.ImportType(typeName.Type()); err != nil {
		t.Error(err)
	}

	err = builder.Code(os.Stdout)
	if err != nil {
		t.Error(err)
	}
}
