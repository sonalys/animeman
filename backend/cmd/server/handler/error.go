package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/ogen-go/ogen/ogenerrors"
	"github.com/ogen-go/ogen/validate"
	"github.com/rs/zerolog/log"
	"github.com/sonalys/animeman/cmd/server/ogen"
	"github.com/sonalys/animeman/internal/app/apperr"
	"github.com/sonalys/animeman/internal/utils/sliceutils"
	"go.opentelemetry.io/otel/trace"
)

func (h *Handler) NewError(ctx context.Context, err error) (resp *ogen.ErrorResponseStatusCode) {
	defer func() {
		span := trace.SpanFromContext(ctx)
		if traceID := span.SpanContext().TraceID(); traceID.IsValid() {
			resp.Response.SetTraceID(ogen.NewOptUUID(uuid.UUID(traceID)))
		}

		if !resp.Response.Details.IsSet() {
			if len(resp.Response.FieldErrors) == 0 {
				resp.Response.Details = ogen.NewOptString("internal server error")
				return
			}
			resp.Response.Details = ogen.NewOptString("")
		}
	}()

	if target := new(ogenerrors.SecurityError); errors.As(err, &target) {
		return &ogen.ErrorResponseStatusCode{
			StatusCode: http.StatusForbidden,
			Response: ogen.ErrorResponse{
				Details: ogen.NewOptString("security requirements not satisfied"),
			},
		}
	}

	if target := new(validate.Error); errors.As(err, &target) {
		errs := make([]ogen.FieldError, 0, len(target.Fields))

		for _, fieldErr := range target.Fields {
			errs = append(errs, ogen.FieldError{
				Message: fieldErr.Error.Error(),
				Field:   fieldErr.Name,
				Code: func() ogen.FieldErrorCode {
					if errors.Is(fieldErr.Error, validate.ErrFieldRequired) {
						return ogen.FieldErrorCodeRequired
					}

					if _, ok := errors.AsType[*validate.MinLengthError](fieldErr.Error); ok {
						return ogen.FieldErrorCodeMinLength
					}

					if _, ok := errors.AsType[*validate.MaxLengthError](fieldErr.Error); ok {
						return ogen.FieldErrorCodeMaxLength
					}

					if _, ok := errors.AsType[*validate.NoRegexMatchError](fieldErr.Error); ok {
						return ogen.FieldErrorCodeInvalidFormat
					}

					return ogen.FieldErrorCodeUnknown
				}(),
			})
		}

		return &ogen.ErrorResponseStatusCode{
			StatusCode: http.StatusBadRequest,
			Response: ogen.ErrorResponse{
				FieldErrors: errs,
			},
		}
	}

	statusCode := GRPCCodeToHTTP(apperr.Code(err))
	fieldErrors := apperr.FieldErrors(err)
	details := apperr.PublicDetails(err)

	return &ogen.ErrorResponseStatusCode{
		StatusCode: statusCode,
		Response: ogen.ErrorResponse{
			Details: ogen.OptString{
				Value: details,
				Set:   true,
			},
			FieldErrors: sliceutils.Map(fieldErrors, func(from apperr.FieldError) ogen.FieldError {
				return ogen.FieldError{
					Field:   from.Field,
					Code:    ogen.FieldErrorCode(from.ErrorCode),
					Message: from.Message,
				}
			}),
		},
	}
}

func validationErrorHandler(client ogen.Handler) func(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
		statusCodeResponse := client.NewError(ctx, err)

		w.WriteHeader(statusCodeResponse.StatusCode)
		w.Header().Set("Content-Type", "application/json")

		log.Error().
			Ctx(ctx).
			Err(err).
			Msg("Request validation failed")

		if err := json.NewEncoder(w).Encode(statusCodeResponse.Response); err != nil {
			log.Error().
				Ctx(ctx).
				Err(err).
				Msg("Failed to encode error response")
		}
	}
}
