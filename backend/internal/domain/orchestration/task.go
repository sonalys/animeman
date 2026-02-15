package orchestration

import (
	"context"
	"encoding/json"
	"math"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/sonalys/animeman/internal/domain/shared"
	"go.opentelemetry.io/otel/trace"
)

type (
	TaskID struct{ uuid.UUID }

	Task struct {
		ID          TaskID
		Type        string
		Payload     json.RawMessage
		Status      TaskStatus
		RetryCount  int
		MaxRetries  int
		TraceID     string
		SpanID      string
		NextRetryAt time.Time
		CreatedAt   time.Time
		UpdatedAt   time.Time
	}
)

func NewTask(
	ctx context.Context,
	taskType string,
	payload json.RawMessage,
	maxRetries int,
) *Task {
	now := time.Now()

	return &Task{
		ID:          shared.NewID[TaskID](),
		Type:        taskType,
		Payload:     payload,
		Status:      TaskStatusPending,
		RetryCount:  0,
		MaxRetries:  maxRetries,
		NextRetryAt: time.Time{},
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func (t *Task) CalculateBackoff(baseDelay time.Duration) (time.Time, bool) {
	if t.RetryCount == t.MaxRetries {
		return time.Time{}, false
	}
	exponent := math.Pow(2, float64(t.RetryCount))
	return time.Now().Add(time.Duration(float64(baseDelay) * exponent)), true
}

func (t *Task) NewLog(
	ctx context.Context,
	level LogLevel,
	message string,
) *TaskLog {
	traceID, spanID := extractTelemetry(ctx)
	return &TaskLog{
		ID:        shared.NewID[TaskLogID](),
		TaskID:    t.ID,
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		TraceID:   traceID,
		SpanID:    spanID,
	}
}

func extractTelemetry(ctx context.Context) (traceID, spanID string) {
	span := trace.SpanFromContext(ctx)
	if !span.SpanContext().IsValid() {
		return "", ""
	}

	return span.SpanContext().TraceID().String(),
		span.SpanContext().SpanID().String()
}
