package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/sonalys/animeman/cmd/server/middlewares"
	"github.com/sonalys/animeman/cmd/server/ogen"
	"github.com/sonalys/animeman/internal/utils/otel"
	"go.opentelemetry.io/otel/trace"
)

type Handler struct{}

func New() (http.Handler, error) {
	return setupHandler(nil, nil)
}

func setupHandler(securityHandler ogen.SecurityHandler, client ogen.Handler) (http.Handler, error) {
	return ogen.NewServer(client, securityHandler,
		ogen.WithPathPrefix("/api/v1"),
		ogen.WithTracerProvider(otel.Provider.TracerProvider()),
		ogen.WithMiddleware(
			middlewares.Recoverer,
			middlewares.Logger,
		),
		ogen.WithErrorHandler(errorHandler(client)),
	)
}

func errorHandler(client ogen.Handler) func(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
		statusCodeResponse := client.NewError(ctx, err)

		w.WriteHeader(statusCodeResponse.StatusCode)
		w.Header().Set("Content-Type", "application/json")

		span := trace.SpanFromContext(ctx)
		statusCodeResponse.Response.SetTraceID(uuid.UUID(span.SpanContext().TraceID()))

		if err := json.NewEncoder(w).Encode(statusCodeResponse); err != nil {
			log.Error().Ctx(ctx).Err(err).Msg("failed to encode error response")
		}
	}
}
