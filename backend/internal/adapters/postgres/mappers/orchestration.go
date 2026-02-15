package mappers

import (
	"github.com/sonalys/animeman/internal/adapters/postgres/sqlcgen"
	"github.com/sonalys/animeman/internal/domain/orchestration"
)

func NewTask(from sqlcgen.OrchestrationTask) orchestration.Task {
	return orchestration.Task{
		ID:          from.ID,
		Type:        from.TaskType,
		Payload:     from.Payload,
		Status:      NewTaskStatus(from.Status),
		RetryCount:  int(from.RetryCount),
		MaxRetries:  int(from.MaxRetries),
		TraceID:     from.TraceID.String,
		SpanID:      from.SpanID.String,
		NextRetryAt: from.NextRetryAt.Time,
		CreatedAt:   from.CreatedAt.Time,
		UpdatedAt:   from.UpdatedAt.Time,
	}
}

func NewTaskLog(from sqlcgen.TaskLog) orchestration.TaskLog {
	return orchestration.TaskLog{
		ID:        from.ID,
		TaskID:    from.TaskID,
		Level:     NewLogLevel(from.Level),
		Message:   from.Message,
		Timestamp: from.CreatedAt.Time,
		TraceID:   from.TraceID.String,
		SpanID:    from.SpanID.String,
	}
}

func NewTaskStatusModel(from orchestration.TaskStatus) string {
	switch from {
	case orchestration.TaskStatusPending:
		return "pending"
	case orchestration.TaskStatusRunning:
		return "running"
	case orchestration.TaskStatusCompleted:
		return "completed"
	case orchestration.TaskStatusFailed:
		return "failed"
	case orchestration.TaskStatusRetrying:
		return "retrying"
	default:
		return "unknown"
	}
}

func NewTaskStatus(from sqlcgen.TaskStatus) orchestration.TaskStatus {
	switch from {
	case "pending":
		return orchestration.TaskStatusPending
	case "running":
		return orchestration.TaskStatusRunning
	case "completed":
		return orchestration.TaskStatusCompleted
	case "failed":
		return orchestration.TaskStatusFailed
	case "retrying":
		return orchestration.TaskStatusRetrying
	default:
		return orchestration.TaskStatusUnknown
	}
}

func NewLogLevelModel(from orchestration.LogLevel) sqlcgen.LogLevel {
	switch from {
	case orchestration.LogLevelDebug:
		return sqlcgen.LogLevelDebug
	case orchestration.LogLevelInfo:
		return sqlcgen.LogLevelInfo
	case orchestration.LogLevelWarn:
		return sqlcgen.LogLevelWarn
	case orchestration.LogLevelError:
		return sqlcgen.LogLevelError
	case orchestration.LogLevelFatal:
		return sqlcgen.LogLevelFatal
	default:
		return "unknown"
	}
}

func NewLogLevel(from sqlcgen.LogLevel) orchestration.LogLevel {
	switch from {
	case sqlcgen.LogLevelDebug:
		return orchestration.LogLevelDebug
	case sqlcgen.LogLevelInfo:
		return orchestration.LogLevelInfo
	case sqlcgen.LogLevelWarn:
		return orchestration.LogLevelWarn
	case sqlcgen.LogLevelError:
		return orchestration.LogLevelError
	case sqlcgen.LogLevelFatal:
		return orchestration.LogLevelFatal
	default:
		return orchestration.LogLevelUnknown
	}
}
