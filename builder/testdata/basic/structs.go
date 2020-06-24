package basic

type AA struct {
	A  []int  // [Int!]!
	B  []bool // [Boolean!]!
	Aa *AA    // AA
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

func (ms *MyStruct) I(args IArgsInput) []string {
	return []string{"I"}
}

func (ms *MyStruct) II(args IArgsInput) *MyStruct {
	return nil
}

func (ms *MyStruct) III(args IArgsInput) MyStruct {
	return MyStruct{}
}
