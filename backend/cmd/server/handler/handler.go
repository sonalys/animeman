package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/sonalys/animeman/cmd/server/middlewares"
	"github.com/sonalys/animeman/cmd/server/ogen"
	"github.com/sonalys/animeman/internal/app/apperr"
	"github.com/sonalys/animeman/internal/utils/otel"
	"github.com/sonalys/animeman/internal/utils/sliceutils"
	"go.opentelemetry.io/otel/trace"
)

type Handler struct{}

func (h *Handler) AuthenticationWhoAmI(ctx context.Context) (*ogen.AuthenticationWhoAmIOK, error) {
	panic("unimplemented")
}

func (h *Handler) NewError(ctx context.Context, err error) *ogen.ErrorResponseStatusCode {
	span := trace.SpanFromContext(ctx)
	statusCode := GRPCCodeToHTTP(apperr.Code(err))

	var fieldErrors []apperr.FieldError
	var details string

	if publicErr, ok := errors.AsType[apperr.PublicError](err); ok {
		details = publicErr.Details()
	}

	if formErr, ok := errors.AsType[apperr.FormError](err); ok {
		fieldErrors = formErr.FieldErrors
	}

	return &ogen.ErrorResponseStatusCode{
		StatusCode: statusCode,
		Response: ogen.ErrorResponse{
			TraceID: uuid.UUID(span.SpanContext().TraceID()),
			Details: details,
			FieldErrors: sliceutils.Map(fieldErrors, func(from apperr.FieldError) ogen.FieldError {
				return ogen.FieldError{
					Field:   from.Field,
					Code:    from.Code,
					Message: from.Message,
				}
			}),
		},
	}
}

func (h *Handler) RegisterUser(ctx context.Context, req *ogen.UserRegistration) (ogen.RegisterUserRes, error) {
	panic("unimplemented")
}

func New() (http.Handler, error) {
	handler := &Handler{}
	return setupHandler(nil, handler)
}

func setupHandler(securityHandler ogen.SecurityHandler, client ogen.Handler) (http.Handler, error) {
	return ogen.NewServer(client, securityHandler,
		ogen.WithPathPrefix("/api/v1"),
		ogen.WithTracerProvider(otel.Provider.TracerProvider()),
		ogen.WithMiddleware(
			middlewares.Recoverer,
			middlewares.Logger,
		),
		// ogen.WithErrorHandler(errorHandler(client)),
	)
}

// func errorHandler(client ogen.Handler) func(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
// 	return func(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
// 		statusCodeResponse := client.NewError(ctx, err)

// 		w.WriteHeader(statusCodeResponse.StatusCode)
// 		w.Header().Set("Content-Type", "application/json")

// 		if err := json.NewEncoder(w).Encode(statusCodeResponse); err != nil {
// 			log.Error().Ctx(ctx).Err(err).Msg("failed to encode error response")
// 		}
// 	}
// }
