package watchlists

type Status uint

const (
	StatusUnknown Status = iota
	StatusWatching
	StatusCompleted
	StatusDropped
	StatusPlanToWatch
	statusSentinel
)

func (s Status) String() string {
	switch s {
	case StatusWatching:
		return "watching"
	case StatusCompleted:
		return "completed"
	case StatusDropped:
		return "dropped"
	case StatusPlanToWatch:
		return "planToWatch"
	default:
		return "unknown"
	}
}

func (s Status) IsValid() bool {
	return s > StatusUnknown && s < statusSentinel
}
