package otel

import (
	"context"
	"errors"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	traceSDK "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

const name = "github.com/sonalys/animeman"

var (
	Tracer = otel.Tracer(name)
	Meter  = otel.Meter(name)
	Logger = global.Logger(name)
)

type provider struct{}

var Provider = provider{}

func (tp provider) TracerProvider() trace.TracerProvider {
	return otel.GetTracerProvider()
}

func (tp provider) MeterProvider() metric.MeterProvider {
	return otel.GetMeterProvider()
}

func (tp provider) TextMapPropagator() propagation.TextMapPropagator {
	return otel.GetTextMapPropagator()
}

// Initialize bootstraps the OpenTelemetry pipeline.
// If it does not return an error, make sure to call shutdown for proper cleanup.
func Initialize(ctx context.Context, endpoint, version string) (shutdown func(context.Context) error, err error) {
	var shutdownFuncs []func(context.Context) error

	shutdown = func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil

		return err
	}

	handleErr := func(inErr error) {
		err = errors.Join(inErr, shutdown(ctx))
	}

	res, err := newResource(version)
	if err != nil {
		handleErr(err)

		return
	}

	loggerProvider, err := newLoggerProvider(ctx, endpoint, res)
	if err != nil {
		handleErr(err)

		return
	}
	global.SetLoggerProvider(loggerProvider)

	shutdownFuncs = append(shutdownFuncs, loggerProvider.Shutdown)

	prop := newPropagator()
	otel.SetTextMapPropagator(prop)

	tracerProvider, err := newTraceProvider(ctx, endpoint, res)
	if err != nil {
		handleErr(err)

		return
	}
	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
	otel.SetTracerProvider(tracerProvider)

	return shutdown, err
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func newTraceProvider(ctx context.Context, endpoint string, res *resource.Resource) (*traceSDK.TracerProvider, error) {
	traceExporter, err := newTraceExporter(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	traceProvider := traceSDK.NewTracerProvider(
		traceSDK.WithBatcher(traceExporter),
		traceSDK.WithResource(res),
	)

	return traceProvider, nil
}

func newTraceExporter(ctx context.Context, endpoint string) (traceSDK.SpanExporter, error) {
	return otlptracegrpc.New(ctx,
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(endpoint),
	)
}

func newResource(version string) (*resource.Resource, error) {
	return resource.Merge(resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL,
			semconv.ServiceName(name),
			semconv.ServiceVersion(version),
		))
}
