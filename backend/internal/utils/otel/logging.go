package otel

import (
	"context"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
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
