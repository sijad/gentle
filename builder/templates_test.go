package builder

import (
	"testing"
)

func TestTypeGo(t *testing.T) {
	typNameString := "String"
	typNameInt := "Int"
	typNameUint := "Uint32"
	typNameObj := "MyObjectName"
	typNameScalar := "MyScalar"

	stringRef := TypeRef{
		Kind: SCALAR,
		Name: &typNameString,
	}
	intRef := TypeRef{
		Kind: SCALAR,
		Name: &typNameInt,
	}
	uintRef := TypeRef{
		Kind: SCALAR,
		Name: &typNameUint,
	}
	objRef := TypeRef{
		Kind: OBJECT,
		Name: &typNameObj,
	}
	scalarRef := TypeRef{
		Kind: OBJECT,
		Name: &typNameScalar,
	}
	nonNullStringRef := TypeRef{
		Kind:   NONNULL,
		OfType: &stringRef,
	}
	nonNullIntRef := TypeRef{
		Kind:   NONNULL,
		OfType: &intRef,
	}
	nonNullUintRef := TypeRef{
		Kind:   NONNULL,
		OfType: &uintRef,
	}

	types := []struct {
		typ    TypeRef
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
			TypeRef{
				Kind:   NONNULL,
				OfType: &objRef,
			},
			"myPkg.MyObjectName",
		},
		{
			TypeRef{
				Kind: NONNULL,
				OfType: &TypeRef{
					Kind:   LIST,
					OfType: &nonNullIntRef,
				},
			},
			"[]int",
		},
		{
			TypeRef{
				Kind: NONNULL,
				OfType: &TypeRef{
					Kind:   LIST,
					OfType: &uintRef,
				},
			},
			"[]*uint32",
		},
		{
			TypeRef{
				Kind: NONNULL,
				OfType: &TypeRef{
					Kind: LIST,
					OfType: &TypeRef{
						Kind: NONNULL,
						OfType: &TypeRef{
							Kind: LIST,
							OfType: &TypeRef{
								Kind: NONNULL,
								OfType: &TypeRef{
									Kind:   LIST,
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
			TypeRef{
				Kind: NONNULL,
				OfType: &TypeRef{
					Kind: LIST,
					OfType: &TypeRef{
						Kind: NONNULL,
						OfType: &TypeRef{
							Kind: LIST,
							OfType: &TypeRef{
								Kind: NONNULL,
								OfType: &TypeRef{
									Kind:   LIST,
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
			TypeRef{
				Kind: LIST,
				OfType: &TypeRef{
					Kind: LIST,
					OfType: &TypeRef{
						Kind:   LIST,
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
