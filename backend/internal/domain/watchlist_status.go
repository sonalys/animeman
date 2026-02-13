package domain

type WatchlistStatus uint

const (
	WatchlistStatusUnknown WatchlistStatus = iota
	WatchlistStatusWatching
	WatchlistStatusCompleted
	WatchlistStatusDropped
	WatchlistStatusPlanToWatch
	watchlistStatusSentinel
)

func (s WatchlistStatus) String() string {
	switch s {
	case WatchlistStatusWatching:
		return "WATCHING"
	case WatchlistStatusCompleted:
		return "COMPLETED"
	case WatchlistStatusDropped:
		return "DROPPED"
	case WatchlistStatusPlanToWatch:
		return "PLAN_TO_WATCH"
	default:
		return "UNKNOWN"
	}
}

func (s WatchlistStatus) IsValid() bool {
	return s > WatchlistStatusUnknown && s < watchlistStatusSentinel
}
