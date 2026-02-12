package apperr

import (
	"errors"
	"fmt"

	"google.golang.org/grpc/codes"
)

type Error struct {
	StatusCode codes.Code
	Message    string
	Cause      error
}

func New(cause error, code codes.Code, msgAndArgs ...any) Error {
	var message string

	if len(msgAndArgs) > 0 {
		if mask, ok := msgAndArgs[0].(string); ok {
			message = fmt.Sprintf(mask, msgAndArgs[1:]...)
		}
	}

	return Error{
		StatusCode: code,
		Message:    message,
		Cause:      cause,
	}
}

func (e Error) Error() string {
	return fmt.Sprintf("[%s] %s: %s", e.StatusCode, e.Message, e.Cause)
}

func (e Error) Unwrap() error {
	return e.Cause
}

func (e Error) Code() codes.Code {
	return e.StatusCode
}

func Code(err error) codes.Code {
	type codedError interface {
		error
		Code() codes.Code
	}

	if target, ok := errors.AsType[codedError](err); ok {
		return target.Code()
	}

	return codes.Internal
}
