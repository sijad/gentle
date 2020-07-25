package basic

import (
	"context"
	"time"
)

type AA struct {
	A  []int  // [Int!]!
	B  []bool // [Boolean!]!
	Aa *AA    // AA
}

func (aa AA) AA(args IArgsInput) AA {
	return aa
}

type MyStruct struct {
	A []int         // [Int!]!
	B []bool        // [Boolean!]!
	C []uint        // [Int!]!
	D int           // Int!
	E uint          // Int!
	F bool          // Boolean!
	G [][][]int     // [[[Int!]!]!]!
	H *[]*[]*[]*int // [[[Int]]]
	J AA
}

type IArgsInput struct {
	D int           // Int!
	E uint          // Int!
	F bool          // Boolean!
	G [][][]int     // [[[Int!]!]!]!
	H *[]*[]*[]*int // [[[Int]]]
}

func (ms *MyStruct) I(ctx context.Context, args IArgsInput) ([]string, error) {
	return []string{"I"}, nil
}

func (ms *MyStruct) II(args struct{ Name string }) (*MyStruct, error) {
	return nil, nil
}

func (ms *MyStruct) III(ctx context.Context, args IArgsInput) (MyStruct, error) {
	return MyStruct{}, nil
}

type MyTime time.Time

func (t MyTime) MarshalGQL() ([]byte, error) {
	return nil, nil
}

func (t MyTime) UnmarshalGQL(v interface{}) error {
	return nil
}

type KeyValue map[string]string

func (t KeyValue) MarshalGQL() ([]byte, error) {
	return nil, nil
}

func (t KeyValue) UnmarshalGQL(v interface{}) error {
	return nil
}

type MyScalars struct {
	KeyValue KeyValue
	Time     MyTime
}
