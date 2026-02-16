package apperr

import (
	"errors"
	"fmt"

	"github.com/sonalys/animeman/internal/utils/sliceutils"
	"google.golang.org/grpc/codes"
)

type (
	CodedError interface {
		error
		Code() codes.Code
	}

	PublicError interface {
		error
		Details() string
	}

	Error struct {
		StatusCode    codes.Code
		Message       string
		PublicDetails string
		Cause         error
	}

	FieldError struct {
		Field   string `json:"field"`   // e.g., "userID".
		Message string `json:"message"` // e.g., "must be a valid UUID".
		Code    string `json:"code"`    // e.g., "invalid_format".
	}

	FormError struct {
		FieldErrors []FieldError
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

func (f *FormError) Add(field, code, message string) {
	f.FieldErrors = append(f.FieldErrors, FieldError{
		Field:   field,
		Code:    code,
		Message: message,
	})
}

func (f *FormError) Code() codes.Code {
	return codes.InvalidArgument
}

func (e FieldError) Error() string {
	return fmt.Sprintf("field '%s': %s", e.Field, e.Message)
}

func (e FieldError) Is(err error) bool {
	targetErr, ok := errors.AsType[FieldError](err)
	if !ok {
		return false
	}

	return e.Field == targetErr.Field && e.Code == targetErr.Code
}

func (e FormError) Error() string {
	return fmt.Sprintf("bad arguments: %v", e.FieldErrors)
}

func (e FormError) Unwrap() []error {
	return sliceutils.Map(e.FieldErrors, func(from FieldError) error { return from })
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

func Code(err error) codes.Code {
	if target, ok := errors.AsType[CodedError](err); ok {
		return target.Code()
	}

	return codes.Internal
}
