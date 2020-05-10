package simple

import (
	"context"
	"errors"
	"fmt"
	"io"
)

type resolvers struct {
	hello int
}

type Query struct {
	*resolvers
	hello string
	aaaa  uint32
}

// YesNo helps
type YesNo bool

func (y *YesNo) UnmarshalGQL(v interface{}) error {
	yes, ok := v.(string)
	if !ok {
		return fmt.Errorf("points must be strings")
	}
	if yes == "yes" {
		*y = true
	} else {
		*y = false
	}
	return nil
}

func (y YesNo) MarshalGQL(w io.Writer) {
	if y {
		w.Write([]byte(`"yes"`))
	} else {
		w.Write([]byte(`"no"`))
	}
}

type echoPayload struct {
	result string
}

func (r *Query) Echo(txt string) echoPayload {
	return echoPayload{txt}
}

func (r *Query) RandomNumber(ctx context.Context, id int) (*int, error) {
	fmt.Println(r.hello + "121")
	return nil, errors.New("not found")
}

func (r Query) RandomYesOrNo(ctx context.Context, id int) []YesNo {
	return nil
}

func RandomYesOrNo(ctx context.Context, id int) [][]YesNo {
	return nil
}
