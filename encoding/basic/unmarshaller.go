package basic

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
)

const (
	maxUint = uint64(^uint(0))
	maxInt  = int64(maxUint >> 1)
	minInt  = -maxInt - 1
)

func UnmarshalString(v interface{}) (string, error) {
	switch v := v.(type) {
	case string:
		return v, nil
	default:
		return "", fmt.Errorf("%T is not a string", v)
	}
}

func UnmarshalBoolean(v interface{}) (bool, error) {
	switch v := v.(type) {
	case string:
		return strings.ToLower(v) == "true", nil
	case int:
		return v != 0, nil
	case bool:
		return v, nil
	default:
		return false, fmt.Errorf("%T is not a bool", v)
	}
}

type intOverflowError struct {
	ToType string
	Value  interface{}
}

func (e intOverflowError) Error() string {
	return fmt.Sprintf("%d (%T) overflows %s", e.Value, e.Value, e.ToType)
}

func UnmarshalInt(v interface{}) (int, error) {
	val, err := UnmarshalInt64(v)
	if err != nil {
		return 0, err
	}
	if val < minInt || val > maxInt {
		return 0, intOverflowError{ToType: "int", Value: val}
	}
	return int(val), nil
}

func UnmarshalInt8(v interface{}) (int8, error) {
	val, err := UnmarshalInt64(v)
	if err != nil {
		return 0, err
	}
	if val < math.MinInt8 || val > math.MaxInt8 {
		return 0, intOverflowError{ToType: "int8", Value: val}
	}
	return int8(val), nil
}

func UnmarshalInt16(v interface{}) (int16, error) {
	val, err := UnmarshalInt64(v)
	if err != nil {
		return 0, err
	}
	if val < math.MinInt16 || val > math.MaxInt16 {
		return 0, intOverflowError{ToType: "int16", Value: val}
	}
	return int16(val), nil
}

func UnmarshalInt32(v interface{}) (int32, error) {
	val, err := UnmarshalInt64(v)
	if err != nil {
		return 0, err
	}
	if val < math.MinInt32 || val > math.MaxInt32 {
		return 0, intOverflowError{ToType: "int32", Value: val}
	}
	return int32(val), nil
}

func UnmarshalInt64(v interface{}) (int64, error) {
	switch v := v.(type) {
	case string:
		return strconv.ParseInt(v, 10, 64)
	case int:
		return int64(v), nil
	case int64:
		return v, nil
	case json.Number:
		return strconv.ParseInt(string(v), 10, 64)
	default:
		return 0, fmt.Errorf("%T is not an int", v)
	}
}

func UnmarshalUint(v interface{}) (uint, error) {
	val, err := UnmarshalUint64(v)
	if err != nil {
		return 0, err
	}
	if val > maxUint {
		return 0, intOverflowError{ToType: "uint", Value: val}
	}
	return uint(val), nil
}

func UnmarshalUint8(v interface{}) (uint8, error) {
	val, err := UnmarshalUint64(v)
	if err != nil {
		return 0, err
	}
	if val > math.MaxUint8 {
		return 0, intOverflowError{ToType: "uint8", Value: val}
	}
	return uint8(val), nil
}

func UnmarshalUint16(v interface{}) (uint16, error) {
	val, err := UnmarshalUint64(v)
	if err != nil {
		return 0, err
	}
	if val > math.MaxUint16 {
		return 0, intOverflowError{ToType: "uint16", Value: val}
	}
	return uint16(val), nil
}

func UnmarshalUint32(v interface{}) (uint32, error) {
	val, err := UnmarshalUint64(v)
	if err != nil {
		return 0, err
	}
	if val > math.MaxUint32 {
		return 0, intOverflowError{ToType: "uint32", Value: val}
	}
	return uint32(val), nil
}

func UnmarshalUint64(v interface{}) (uint64, error) {
	val, err := UnmarshalInt64(v)
	if err != nil {
		return 0, err
	}
	return uint64(val), err
}
