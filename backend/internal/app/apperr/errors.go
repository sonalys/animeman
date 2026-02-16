package apperr

import (
	"errors"
	"fmt"
	"strings"

	"google.golang.org/grpc/codes"
)

type (
	codedError interface {
		error
		Code() codes.Code
	}

	Error struct {
		StatusCode codes.Code
		Message    string
		Cause      error
	}

	PublicError struct {
		cause   error
		Details string
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

func NewPublicError(cause error, mask string, args ...any) PublicError {
	return PublicError{
		cause:   cause,
		Details: fmt.Sprintf(mask, args...),
	}
}

func Code(err error) codes.Code {
	if target, ok := errors.AsType[codedError](err); ok {
		return target.Code()
	}

	return codes.Internal
}

func PublicDetails(err error) string {
	if target, ok := errors.AsType[PublicError](err); ok {
		return target.Details
	}

	return ""
}

func (e Error) Error() string {
	var b strings.Builder

	if e.Message != "" {
		b.WriteString(e.Message)
	}

	if e.Cause != nil {
		fmt.Fprintf(&b, ": %s", e.Cause)
	}

	return b.String()
}

func (e Error) Unwrap() error {
	return e.Cause
}

func (e Error) Code() codes.Code {
	return e.StatusCode
}

func (e PublicError) Error() string {
	return e.cause.Error()
}

func (e PublicError) Unwrap() error {
	return e.cause
}
