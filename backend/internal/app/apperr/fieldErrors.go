package apperr

import (
	"errors"
	"fmt"

	"github.com/sonalys/animeman/internal/utils/sliceutils"
	"google.golang.org/grpc/codes"
)

type (
	FieldError struct {
		Field   string
		Message string
		Code    string
	}

	fieldErrors []FieldError

	FormValidation struct {
		FieldErrors []FieldError
	}
)

var (
	_ codedError = &fieldErrors{}
)

func FieldErrors(err error) []FieldError {
	if target, ok := errors.AsType[fieldErrors](err); ok {
		return target
	}

	return nil
}

func (f fieldErrors) Error() string {
	return fmt.Sprintf("bad arguments: %v", []FieldError(f))
}

func (f *fieldErrors) Code() codes.Code {
	return codes.InvalidArgument
}

func (e fieldErrors) Unwrap() []error {
	return sliceutils.Map(e, func(from FieldError) error { return from })
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

func (f *FormValidation) Add(field, code, message string) {
	f.FieldErrors = append(f.FieldErrors, FieldError{
		Field:   field,
		Code:    code,
		Message: message,
	})
}

func (f *FormValidation) Validate() error {
	if len(f.FieldErrors) == 0 {
		return nil
	}

	return fieldErrors(f.FieldErrors)
}
