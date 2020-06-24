package basic

type AA struct {
	A []int  // [Int!]!
	B []bool // [Boolean!]!
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
