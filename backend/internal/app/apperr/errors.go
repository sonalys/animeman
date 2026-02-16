package apperr

import (
	"errors"
	"fmt"

	"google.golang.org/grpc/codes"
)

type (
	codedError interface {
		error
		Code() codes.Code
	}

	Error struct {
		StatusCode    codes.Code
		Message       string
		PublicDetails string
		Cause         error
	}
)

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

func Code(err error) codes.Code {
	if target, ok := errors.AsType[codedError](err); ok {
		return target.Code()
	}

	return codes.Internal
}

func PublicDetails(err error) string {
	if target, ok := errors.AsType[Error](err); ok {
		return target.PublicDetails
	}

	return ""
}

func (e Error) WithPublicDetails(details string) Error {
	e.PublicDetails = details
	return e
}

func (e Error) Details() string {
	return e.PublicDetails
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
