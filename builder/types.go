package builder

import (
	"go/types"
)

type TypeKind string

const (
	SCALAR      TypeKind = "SCALAR"
	LIST                 = "LIST"
	NONNULL              = "NONNULL"
	OBJECT               = "OBJECT"
	ENUM                 = "ENUM"
	INTERFACE            = "INTERFACE"
	UNION                = "UNION"
	INPUTOBJECT          = "INPUTOBJECT"
)

type Field struct {
	Name              string
	Description       string
	Args              []InputValue
	Type              TypeRef
	IsDeprecated      bool
	DeprecationReason *string
	IsMethod          bool
	HasError          bool
	HasArgs           bool
	Params            []*types.Var
}

type InputValue struct {
	Name        string
	Description string
	Type        TypeRef
}

type EnumValue struct {
	Name              string
	Description       string
	IsDeprecated      bool
	DeprecationReason *string
}

type FullType struct {
	Kind          TypeKind
	Name          string
	Description   string
	Fields        []Field
	InputFields   []InputValue
	Interfaces    []TypeRef
	EnumValues    []EnumValue
	PossibleTypes []TypeRef
	Id            string
	PackageName   string
	PackagePath   string
}

type TypeRef struct {
	Kind   TypeKind
	Name   *string
	OfType *TypeRef
}
