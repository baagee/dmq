package common

const (
	ErrorCodeDefault           = 100
	ErrorCodeParseParamsFailed = ErrorCodeDefault + iota
	ErrorCodeValidateFailed
	ErrorCodeUnknownProduct
	ErrorCodeUnknownCommand
	ErrorCodeJsonMarshal
	ErrorCodeRedisSave
	ErrorCodeFoundBucketsFailed
	ErrorCodeFoundPointFailed
	ErrorCodeRemoveBucketsFailed
	ErrorCodeResponseCodeNot200
	ErrorCodePreRequestFailed
	ErrorCodeRequestFailed
	ErrorCodeGetStatusFailed
	ErrorCodeGetMessageFailed
	ErrorCodeRedisLoadLuaFailed
	ErrorCodeGetPendingFailed
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
