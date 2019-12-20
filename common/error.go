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

type ThrowAble interface {
	error
	Code() int
}

type Notice struct {
	code int
	err  error
}

func ThrowNotice(code int, err error) Notice {
	return Notice{
		code: code,
		err:  err,
	}
}

func (n Notice) Error() string {
	return n.err.Error()
}

func (n Notice) Code() int {
	return n.code
}
