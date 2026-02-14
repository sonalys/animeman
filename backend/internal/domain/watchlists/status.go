package watchlists

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
		return "watching"
	case WatchlistStatusCompleted:
		return "completed"
	case WatchlistStatusDropped:
		return "dropped"
	case WatchlistStatusPlanToWatch:
		return "planToWatch"
	default:
		return "unknown"
	}
}

func (s WatchlistStatus) IsValid() bool {
	return s > WatchlistStatusUnknown && s < watchlistStatusSentinel
}
