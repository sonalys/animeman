package orchestration

type (
	TaskStatus uint
)

const (
	TaskStatusUnknown TaskStatus = iota
	TaskStatusPending
	TaskStatusRunning
	TaskStatusCompleted
	TaskStatusFailed
	TaskStatusRetrying
	taskStatusSentinel
)

func (s TaskStatus) IsValid() bool {
	return s > TaskStatusUnknown && s < taskStatusSentinel
}

func (s TaskStatus) String() string {
	switch s {
	case TaskStatusPending:
		return "pending"
	case TaskStatusRunning:
		return "running"
	case TaskStatusCompleted:
		return "completed"
	case TaskStatusFailed:
		return "failed"
	case TaskStatusRetrying:
		return "retrying"
	default:
		return "unknown"
	}
}
