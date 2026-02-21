package usecases

import (
	"context"
	"errors"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/sonalys/animeman/internal/app/apperr"
	"go.opentelemetry.io/otel/attribute"
	otelCodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
)

func logError(ctx context.Context, err error, mask string, args ...any) {
	span := trace.SpanFromContext(ctx)
	level := zerolog.ErrorLevel

	appErr, ok := errors.AsType[apperr.Error](err)
	if ok {
		if appErr.Code() != codes.Internal {
			level = zerolog.InfoLevel
		} else {
			span.SetStatus(otelCodes.Error, "")
			span.SetAttributes(
				attribute.Stringer("code", appErr.Code()),
			)
		}

		log.
			WithLevel(level).
			Ctx(ctx).
			Stringer("code", appErr.Code()).
			Str("message", appErr.Message).
			Err(appErr.Cause).
			Msgf(mask, args...)
		return
	}

	log.
		WithLevel(level).
		Ctx(ctx).
		Err(err).
		Msgf(mask, args...)
}
