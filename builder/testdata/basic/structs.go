package basic

type MyStruct struct {
	A []int     // [Int!]!
	B []bool    // [Int]!
	C []uint    // ?
	D int       // Int!
	E uint      // Int
	F bool      // ?
	G [][][]int // [[[Int!]!]!]!
}
