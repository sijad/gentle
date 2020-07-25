package builder

import (
	"testing"

	"github.com/jensneuse/graphql-go-tools/pkg/introspection"
)

func TestTypeGo(t *testing.T) {
	typNameString := "String"
	typNameInt := "Int"
	typNameUint := "Uint32"
	typNameObj := "MyObjectName"
	typNameScalar := "MyScalar"

	stringRef := introspection.TypeRef{
		Kind: introspection.SCALAR,
		Name: &typNameString,
	}
	intRef := introspection.TypeRef{
		Kind: introspection.SCALAR,
		Name: &typNameInt,
	}
	uintRef := introspection.TypeRef{
		Kind: introspection.SCALAR,
		Name: &typNameUint,
	}
	objRef := introspection.TypeRef{
		Kind: introspection.OBJECT,
		Name: &typNameObj,
	}
	scalarRef := introspection.TypeRef{
		Kind: introspection.OBJECT,
		Name: &typNameScalar,
	}
	nonNullStringRef := introspection.TypeRef{
		Kind:   introspection.NONNULL,
		OfType: &stringRef,
	}
	nonNullIntRef := introspection.TypeRef{
		Kind:   introspection.NONNULL,
		OfType: &intRef,
	}
	nonNullUintRef := introspection.TypeRef{
		Kind:   introspection.NONNULL,
		OfType: &uintRef,
	}

	types := []struct {
		typ    introspection.TypeRef
		expect string
	}{
		{
			stringRef,
			"*string",
		},
		{
			intRef,
			"*int",
		},
		{
			uintRef,
			"*uint32",
		},
		{
			nonNullStringRef,
			"string",
		},
		{
			nonNullIntRef,
			"int",
		},
		{
			nonNullUintRef,
			"uint32",
		},
		{
			scalarRef,
			"*myPkg.MyScalar",
		},
		{
			introspection.TypeRef{
				Kind:   introspection.NONNULL,
				OfType: &objRef,
			},
			"myPkg.MyObjectName",
		},
		{
			introspection.TypeRef{
				Kind: introspection.NONNULL,
				OfType: &introspection.TypeRef{
					Kind:   introspection.LIST,
					OfType: &nonNullIntRef,
				},
			},
			"[]int",
		},
		{
			introspection.TypeRef{
				Kind: introspection.NONNULL,
				OfType: &introspection.TypeRef{
					Kind:   introspection.LIST,
					OfType: &uintRef,
				},
			},
			"[]*uint32",
		},
		{
			introspection.TypeRef{
				Kind: introspection.NONNULL,
				OfType: &introspection.TypeRef{
					Kind: introspection.LIST,
					OfType: &introspection.TypeRef{
						Kind: introspection.NONNULL,
						OfType: &introspection.TypeRef{
							Kind: introspection.LIST,
							OfType: &introspection.TypeRef{
								Kind: introspection.NONNULL,
								OfType: &introspection.TypeRef{
									Kind:   introspection.LIST,
									OfType: &nonNullIntRef,
								},
							},
						},
					},
				},
			},
			"[][][]int",
		},
		{
			introspection.TypeRef{
				Kind: introspection.NONNULL,
				OfType: &introspection.TypeRef{
					Kind: introspection.LIST,
					OfType: &introspection.TypeRef{
						Kind: introspection.NONNULL,
						OfType: &introspection.TypeRef{
							Kind: introspection.LIST,
							OfType: &introspection.TypeRef{
								Kind: introspection.NONNULL,
								OfType: &introspection.TypeRef{
									Kind:   introspection.LIST,
									OfType: &objRef,
								},
							},
						},
					},
				},
			},
			"[][][]*myPkg.MyObjectName",
		},
		{
			introspection.TypeRef{
				Kind: introspection.LIST,
				OfType: &introspection.TypeRef{
					Kind: introspection.LIST,
					OfType: &introspection.TypeRef{
						Kind:   introspection.LIST,
						OfType: &objRef,
					},
				},
			},
			"*[]*[]*[]*myPkg.MyObjectName",
		},
	}

	fullTypes := map[string]FullType{
		"MyObjectName": {
			PackageName: "myPkg",
		},
		"MyScalar": {
			PackageName: "myPkg",
		},
	}

	for _, typ := range types {
		got := typeGo(&typ.typ, fullTypes)
		if got != typ.expect {
			t.Errorf("Marshaler Name (%v) was incorrect, got: %s, want: %s", typ.typ, got, typ.expect)
		}
	}
}
