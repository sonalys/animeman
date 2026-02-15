package mappers

import (
	"github.com/sonalys/animeman/internal/adapters/postgres/sqlcgen"
	"github.com/sonalys/animeman/internal/domain/watchlists"
)

func NewWatchlistStatus(from sqlcgen.WatchlistStatus) watchlists.Status {
	switch from {
	case sqlcgen.WatchlistStatusCompleted:
		return watchlists.StatusCompleted
	case sqlcgen.WatchlistStatusDropped:
		return watchlists.StatusDropped
	case sqlcgen.WatchlistStatusPlanToWatch:
		return watchlists.StatusPlanToWatch
	case sqlcgen.WatchlistStatusWatching:
		return watchlists.StatusWatching
	default:
		return watchlists.StatusUnknown
	}
}

func NewWatchlistStatusModel(from watchlists.Status) sqlcgen.WatchlistStatus {
	switch from {
	case watchlists.StatusCompleted:
		return sqlcgen.WatchlistStatusCompleted
	case watchlists.StatusDropped:
		return sqlcgen.WatchlistStatusDropped
	case watchlists.StatusPlanToWatch:
		return sqlcgen.WatchlistStatusPlanToWatch
	case watchlists.StatusWatching:
		return sqlcgen.WatchlistStatusWatching
	default:
		return sqlcgen.WatchlistStatusUnknown
	}
}
