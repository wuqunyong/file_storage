package errs

import "fmt"

type CodeError interface {
	Code() int32
	Msg() string
	error
}

func NewCodeError(err error, code ...int32) CodeError {
	if err == nil {
		return &codeError{
			code: CODE_OK,
		}
	}

	var iCode int32
	if len(code) > 0 {
		iCode = code[0]
	} else {
		iCode = CODE_ServerInternalError
	}

	return &codeError{
		code: iCode,
		msg:  err.Error(),
	}
}

type codeError struct {
	code int32
	msg  string
}

func (e *codeError) Code() int32 {
	return e.code
}

func (e *codeError) Msg() string {
	return e.msg
}

func (e *codeError) Error() string {
	sError := fmt.Sprintf("{code:%d,msg:%q}", e.code, e.msg)
	return sError
}
