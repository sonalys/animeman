package orchestration

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type (
	TaskLogID struct{ uuid.UUID }

	TaskLog struct {
		ID        TaskLogID
		TaskID    TaskID
		Level     LogLevel
		Message   string
		Timestamp time.Time
		TraceID   string
		SpanID    string
	}
)
