package orchestration

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/sonalys/animeman/internal/domain/orchestration"
	"github.com/sonalys/animeman/internal/ports"
	"go.opentelemetry.io/otel/trace"
)

// injectTelemetry takes the hex strings from the DB and returns a new context
// that OpenTelemetry and Zerolog will recognize as the active trace.
func injectTelemetry(ctx context.Context, tidStr, sidStr string) context.Context {
	if tidStr == "" || sidStr == "" {
		return ctx
	}

	// 1. Parse strings back into OTel types
	tid, errT := trace.TraceIDFromHex(tidStr)
	sid, errS := trace.SpanIDFromHex(sidStr)

	if errT != nil || errS != nil {
		return ctx // Fallback to original context if strings are malformed
	}

	// 2. Create a SpanContext
	// We mark it as 'Remote' because it originated from outside this specific execution
	spanContext := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: tid,
		SpanID:  sid,
		Remote:  true,
	})

	// 3. Inject into context
	return trace.ContextWithRemoteSpanContext(ctx, spanContext)
}

type orchestrationHook struct {
	store ports.TaskRepository
	task  *orchestration.Task
}

func newLoggerOrchestrationHook(store ports.TaskRepository, task *orchestration.Task) zerolog.Hook {
	return orchestrationHook{
		store: store,
		task:  task,
	}
}

func (h orchestrationHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	ctx := e.GetCtx()

	e.Stringer("taskID", h.task.ID)

	var taskLogLevel orchestration.LogLevel = func() orchestration.LogLevel {
		switch level {
		case zerolog.DebugLevel:
			return orchestration.LogLevelDebug
		case zerolog.WarnLevel:
			return orchestration.LogLevelWarn
		case zerolog.ErrorLevel:
			return orchestration.LogLevelError
		case zerolog.FatalLevel:
			return orchestration.LogLevelFatal
		default:
			return orchestration.LogLevelInfo
		}
	}()

	go func() {
		_ = h.store.AddLog(context.Background(), h.task.NewLog(ctx, taskLogLevel, msg))
	}()
}
