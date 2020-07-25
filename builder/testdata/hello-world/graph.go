package helloWorld

type Query struct{}

func (q *Query) Hello() string {
	return "world"
}
