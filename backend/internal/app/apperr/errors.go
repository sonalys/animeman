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

func New(cause error, code codes.Code, msg string, args ...any) Error {
	return Error{
		StatusCode: code,
		Message:    fmt.Sprintf(msg, args...),
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
	var target interface{ Code() codes.Code }

	if errors.As(err, &target) {
		return target.Code()
	}

	return 0
}
