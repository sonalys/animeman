package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/sonalys/animeman/internal/utils/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type tracer struct{}

var (
	_ pgx.QueryTracer    = tracer{}
	_ pgx.PrepareTracer  = tracer{}
	_ pgx.CopyFromTracer = tracer{}
	_ pgx.ConnectTracer  = tracer{}
	_ pgx.BatchTracer    = tracer{}
)

func endContextSpan(ctx context.Context, err error) {
	span := trace.SpanFromContext(ctx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
	}
	span.End()
}

func (t tracer) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	//nolint:spancheck // span is closed.
	ctx, _ = otel.Tracer.Start(ctx, data.SQL)

	return ctx
}

func (t tracer) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
	endContextSpan(ctx, data.Err)
}

func (t tracer) TracePrepareStart(ctx context.Context, conn *pgx.Conn, data pgx.TracePrepareStartData) context.Context {
	//nolint:spancheck // span is closed.
	ctx, _ = otel.Tracer.Start(ctx, data.Name)

	return ctx
}

func (t tracer) TracePrepareEnd(ctx context.Context, conn *pgx.Conn, data pgx.TracePrepareEndData) {
	endContextSpan(ctx, data.Err)
}

func (t tracer) TraceCopyFromStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceCopyFromStartData) context.Context {
	//nolint:spancheck // span is closed.
	ctx, _ = otel.Tracer.Start(ctx, data.TableName.Sanitize())

	return ctx
}

func (t tracer) TraceCopyFromEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceCopyFromEndData) {
	endContextSpan(ctx, data.Err)
}

func (t tracer) TraceConnectStart(ctx context.Context, data pgx.TraceConnectStartData) context.Context {
	//nolint:spancheck // span is closed.
	ctx, _ = otel.Tracer.Start(ctx, "connect")

	return ctx
}

func (t tracer) TraceConnectEnd(ctx context.Context, data pgx.TraceConnectEndData) {
	endContextSpan(ctx, data.Err)
}

func (t tracer) TraceBatchQuery(ctx context.Context, conn *pgx.Conn, data pgx.TraceBatchQueryData) {
	span := trace.SpanFromContext(ctx)
	span.AddEvent(data.SQL)
}

func (t tracer) TraceBatchStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceBatchStartData) context.Context {
	//nolint:spancheck // span is closed.
	ctx, _ = otel.Tracer.Start(ctx, "batch")

	return ctx
}

func (t tracer) TraceBatchEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceBatchEndData) {
	endContextSpan(ctx, data.Err)
}
