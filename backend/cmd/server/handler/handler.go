package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/ogen-go/ogen/ogenerrors"
	"github.com/ogen-go/ogen/validate"
	"github.com/rs/zerolog/log"
	"github.com/sonalys/animeman/cmd/server/middlewares"
	"github.com/sonalys/animeman/cmd/server/ogen"
	"github.com/sonalys/animeman/cmd/server/security"
	"github.com/sonalys/animeman/internal/app/apperr"
	"github.com/sonalys/animeman/internal/app/jwt"
	"github.com/sonalys/animeman/internal/app/usecases"
	"github.com/sonalys/animeman/internal/domain/users"
	"github.com/sonalys/animeman/internal/utils/otel"
	"github.com/sonalys/animeman/internal/utils/sliceutils"
	"go.opentelemetry.io/otel/trace"
)

type Handler struct {
	JWTClient jwt.Client
	Usecases  usecases.Usecases
}

func (h *Handler) AuthenticationLogin(ctx context.Context, req *ogen.AuthenticationLoginReq) (*ogen.AuthenticationLoginOK, error) {
	userID, err := h.Usecases.Login(ctx, req.Username, []byte(req.Password))
	if err != nil {
		return nil, err
	}

	stringifiedToken, err := h.JWTClient.Encode(&jwt.Token{
		UserID: *userID,
		Exp:    time.Now().Add(24 * time.Hour).Unix(),
	})
	if err != nil {
		return nil, err
	}

	return &ogen.AuthenticationLoginOK{
		SetCookie: stringifiedToken,
	}, nil
}

func (h *Handler) AuthenticationWhoAmI(ctx context.Context) (*ogen.AuthenticationWhoAmIOK, error) {
	userID, err := security.GetIdentity(ctx)
	if err != nil {
		return nil, err
	}

	return &ogen.AuthenticationWhoAmIOK{
		UserID: uuid.UUID(userID.UUID),
	}, nil
}

func (h *Handler) RegisterUser(ctx context.Context, req *ogen.UserRegistration) (ogen.RegisterUserRes, error) {
	user, err := h.Usecases.RegisterUser(ctx, req.Username, req.Password)
	if err != nil {
		if errors.Is(err, users.ErrUniqueUsername) {
			return nil, apperr.FieldError{
				Field:   "username",
				Message: err.Error(),
				Code:    "alreadyExists",
			}
		}

		return nil, err
	}

	token := &jwt.Token{
		UserID: user.ID,
		Exp:    time.Now().Add(24 * time.Hour).Unix(),
	}

	stringifiedToken, err := h.JWTClient.Encode(token)
	if err != nil {
		return nil, err
	}

	return &ogen.RegisterUserCreatedHeaders{
		SetCookie: stringifiedToken,
		Response: ogen.RegisterUserCreated{
			ID: uuid.UUID(user.ID.UUID),
		},
	}, nil
}

func (h *Handler) NewError(ctx context.Context, err error) (resp *ogen.ErrorResponseStatusCode) {
	defer func() {
		span := trace.SpanFromContext(ctx)
		if traceID := span.SpanContext().TraceID(); traceID.IsValid() {
			resp.Response.SetTraceID(ogen.NewOptUUID(uuid.UUID(traceID)))
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
				Code: func() string {
					if errors.Is(fieldErr.Error, validate.ErrFieldRequired) {
						return "required"
					}

					if _, ok := errors.AsType[*validate.MinLengthError](fieldErr.Error); ok {
						return "minLength"
					}

					if _, ok := errors.AsType[*validate.MaxLengthError](fieldErr.Error); ok {
						return "maxLength"
					}

					if _, ok := errors.AsType[*validate.NoRegexMatchError](fieldErr.Error); ok {
						return "invalidFormat"
					}

					return "unknown"
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
				Set:   details != "",
			},
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

func New(
	jwtClient *jwt.Client,
	usecases usecases.Usecases,
) (http.Handler, error) {
	securityHandler := security.NewHandler(jwtClient)

	handler := &Handler{
		Usecases: usecases,
	}

	return ogen.NewServer(handler, securityHandler,
		ogen.WithPathPrefix("/api/v1"),
		ogen.WithTracerProvider(otel.Provider.TracerProvider()),
		ogen.WithErrorHandler(validationErrorHandler(handler)),
		ogen.WithNotFound(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)

			response := ogen.ErrorResponse{
				Details:     ogen.NewOptString("not found"),
				FieldErrors: nil,
			}

			if err := json.NewEncoder(w).Encode(response); err != nil {
				log.Error().Ctx(r.Context()).Err(err).Msg("failed to encode error response")
			}
		}),
		ogen.WithMiddleware(
			middlewares.Logger,
			middlewares.Recoverer,
		),
	)
}

func validationErrorHandler(client ogen.Handler) func(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
		statusCodeResponse := client.NewError(ctx, err)

		w.WriteHeader(statusCodeResponse.StatusCode)
		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(statusCodeResponse); err != nil {
			log.Error().Ctx(ctx).Err(err).Msg("failed to encode error response")
		}
	}
}
