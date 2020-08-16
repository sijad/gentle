package gentle

type (
	Scalar interface {
		UnmarshalGQL(v interface{}) error
		MarshalGQL() []byte
	}
)
