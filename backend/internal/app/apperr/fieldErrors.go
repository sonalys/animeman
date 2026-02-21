package apperr

import (
	"errors"
	"fmt"

	"github.com/sonalys/animeman/internal/utils/sliceutils"
	"google.golang.org/grpc/codes"
)

type (
	FieldErrorCode string

	FieldError struct {
		Field     string
		Message   string
		ErrorCode FieldErrorCode
	}

	fieldErrors []FieldError

	FormValidation struct {
		FieldErrors []FieldError
	}
)

const (
	FieldErrorCodeAlreadyExists FieldErrorCode = "alreadyExists"
	FieldErrorCodeMinLength     FieldErrorCode = "minLength"
	FieldErrorCodeMaxLength     FieldErrorCode = "maxLength"
	FieldErrorCodeRequired      FieldErrorCode = "required"
	FieldErrorCodeInvalidFormat FieldErrorCode = "invalidFormat"
	FieldErrorCodeInvalid       FieldErrorCode = "invalid"
	FieldErrorCodeUnknown       FieldErrorCode = "unknown"
)

var (
	_ codedError = &fieldErrors{}
)

func NewFieldError(code FieldErrorCode, field string, maskAndArgs ...any) FieldError {
	var message string

	if len(maskAndArgs) > 0 {
		mask, ok := maskAndArgs[0].(string)
		if ok {
			message = fmt.Sprintf(mask, maskAndArgs[1:]...)
		}
	}

	return FieldError{
		ErrorCode: code,
		Field:     field,
		Message:   message,
	}
}

func NewFormValidation(errs ...FieldError) *FormValidation {
	var fv FormValidation
	return fv.Add(errs...)
}

func (c FieldErrorCode) String() string {
	return string(c)
}

func FieldErrors(err error) []FieldError {
	if target, ok := errors.AsType[fieldErrors](err); ok {
		return target
	}

	if target, ok := errors.AsType[FieldError](err); ok {
		return []FieldError{target}
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

	return e.Field == targetErr.Field && e.ErrorCode == targetErr.ErrorCode
}

func (f *FormValidation) Add(errs ...FieldError) *FormValidation {
	f.FieldErrors = append(f.FieldErrors, errs...)
	return f
}

func (f *FormValidation) Validate() error {
	if len(f.FieldErrors) == 0 {
		return nil
	}

	return fieldErrors(f.FieldErrors)
}

func (f FieldError) Code() codes.Code {
	return codes.InvalidArgument
}
