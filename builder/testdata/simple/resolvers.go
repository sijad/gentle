package simple

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"
)

type WeekDay int

const (
	// defines comment comment ...
	Sunday WeekDay = iota
	// defines comment comment ...
	MONDAY
	Tuesday
	Wednesday
	Thursday
	Friday
	Saturday
)

// Query defines comment comment ...
type Query struct {
	// Hello defines comment comment ...
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
	Txt    string
	Things *[]*[]int
	Test2  *[]*[]TestInput
}

type TestInput struct {
	A string
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

func (r Query) EchoWeekDay(ctx context.Context, args struct{ Wd WeekDay }) WeekDay {
	return args.Wd
}

type MyQuery struct {
	Q []*Query
}

func (r *Query) TestQuery(ctx context.Context) []MyQuery {
	q := []*Query{r}
	return []MyQuery{q}
}
