package simple

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"
)

type Query struct {
	Hello string
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

func (y YesNo) MarshalGQL() []byte {
	if *y {
		return []byte(`"yes"`)
	} else {
		return []byte(`"no"`)
	}
}

type EchoPayload struct {
	Echo string
}

func (r *Query) Echo(args struct{ Txt string }) *EchoPayload {
	return &EchoPayload{args.Txt}
}

type SimpleEchoInputArgs struct {
	Txt string
}

func (r *Query) SimpleEcho(args SimpleEchoInputArgs) *string {
	return &args.Txt
}

func (r *Query) RandomNumber(ctx context.Context) (*int, error) {
	return nil, errors.New("not found")
}

func (r Query) RandomYesOrNo(ctx context.Context) []YesNo {
	rand.Seed(time.Now().UnixNano())
	if rand.Intn(2) == 1 {
		return []YesNo{true}
	} else {
		return []YesNo{false}
	}
}

func RandomYesOrNo(ctx context.Context) [][]YesNo {
	return nil
}
