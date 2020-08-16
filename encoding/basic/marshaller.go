package basic

import (
	"fmt"
	"io"
	"strconv"

	"github.com/99designs/gqlgen/graphql"
	"github.com/sijad/gentle"
)

// MarshalString returns graphql for string
func MarshalString(s string) graphql.Marshaler {
	return graphql.MarshalString(s)
}

// MarshalBoolean returns graphql for boolean
func MarshalBoolean(b bool) graphql.Marshaler {
	return graphql.MarshalBoolean(b)
}

// MarshalInt returns graphql for int
func MarshalInt(i int) graphql.Marshaler {
	return MarshalInt64(int64(i))
}

// MarshalInt8 returns graphql for int8
func MarshalInt8(i int8) graphql.Marshaler {
	return MarshalInt64(int64(i))
}

// MarshalInt16 returns graphql for int16
func MarshalInt16(i int16) graphql.Marshaler {
	return MarshalInt64(int64(i))
}

// MarshalInt32 returns graphql for int32
func MarshalInt32(i int32) graphql.Marshaler {
	return MarshalInt64(int64(i))
}

// MarshalInt64 returns graphql for int64
func MarshalInt64(i int64) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		io.WriteString(w, strconv.FormatInt(i, 10))
	})
}

// MarshalUInt returns graphql for uint
func MarshalUInt(i uint) graphql.Marshaler {
	return MarshalUInt64(uint64(i))
}

// MarshalUInt8 returns graphql for uint8
func MarshalUInt8(i uint8) graphql.Marshaler {
	return MarshalUInt64(uint64(i))
}

// MarshalUInt16 returns graphql for uint16
func MarshalUInt16(i uint16) graphql.Marshaler {
	return MarshalUInt64(uint64(i))
}

// MarshalUInt32 returns graphql for uint32
func MarshalUInt32(i uint32) graphql.Marshaler {
	return MarshalUInt64(uint64(i))
}

// MarshalUInt64 returns graphql for uint64
func MarshalUInt64(i uint64) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		io.WriteString(w, strconv.FormatUint(i, 10))
	})
}

// MarshalFloat32 returns graphql for float32
func MarshalFloat32(f float32) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		io.WriteString(w, fmt.Sprintf("%g", f))
	})
}

// MarshalFloat32 returns graphql for float64
func MarshalFloat64(f float64) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		io.WriteString(w, fmt.Sprintf("%g", f))
	})
}

// MarshalScalar returns graphql marshaler for a custom scalar
func MarshalScalar(obj gentle.Scalar) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		if _, err := w.Write(obj.MarshalGQL()); err != nil {
			panic(err)
		}
	})
}
