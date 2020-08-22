package builder

import (
	"os"
	"testing"
)

func TestVarType(t *testing.T) {
	builder := NewGQLBuilder()
	schemaPath := "./testdata/simple"

	if err := builder.ImportPackage(schemaPath); err != nil {
		t.Error(err)
	}

	err := builder.Code(os.Stdout)
	if err != nil {
		t.Error(err)
	}
}
