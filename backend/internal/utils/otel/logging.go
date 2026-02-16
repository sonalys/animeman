package otel

import (
	"context"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/trace"
)

func newLogExporter(ctx context.Context, endpoint string) (log.Exporter, error) {
	return otlploggrpc.New(ctx,
		otlploggrpc.WithInsecure(),
		otlploggrpc.WithEndpoint(endpoint),
	)
}

func newLoggerProvider(ctx context.Context, endpoint string, res *resource.Resource) (*log.LoggerProvider, error) {
	exporter, err := newLogExporter(ctx, endpoint)
	if err != nil {
		return nil, err
	}
	processor := log.NewBatchProcessor(exporter)
	provider := log.NewLoggerProvider(
		log.WithResource(res),
		log.WithProcessor(processor),
	)

	return provider, nil
}

type OTelHook struct{}

func (h OTelHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	ctx := e.GetCtx()
	spanContext := trace.SpanContextFromContext(ctx)

	if spanContext.IsValid() {
		e.Str("traceID", spanContext.TraceID().String())
		e.Str("spanID", spanContext.SpanID().String())
	}
}
