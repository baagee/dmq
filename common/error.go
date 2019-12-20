package common

const (
	ErrorCodeDefault           = 100
	ErrorCodeParseParamsFailed = 110
	ErrorCodeValidateFailed    = 120
	ErrorCodeUnknowProduct     = 130
	ErrorCodeUnknowCommand     = 140
	ErrorCodeJsonMarshal       = 150
	ErrorCodeRedisSave         = 160
)

type Throw interface {
	error
	Code() int
}

type Notice struct {
	CodeInt int
	Err     error
}

func (n Notice) Error() string {
	return n.Err.Error()
}

func (n Notice) Code() int {
	return n.CodeInt
}
