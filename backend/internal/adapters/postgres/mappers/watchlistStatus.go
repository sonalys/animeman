package mappers

import (
	"github.com/sonalys/animeman/internal/adapters/postgres/sqlcgen"
	"github.com/sonalys/animeman/internal/domain/watchlists"
)

func NewWatchlistStatus(from sqlcgen.WatchlistStatus) watchlists.WatchlistStatus {
	switch from {
	case sqlcgen.WatchlistStatusCompleted:
		return watchlists.WatchlistStatusCompleted
	case sqlcgen.WatchlistStatusDropped:
		return watchlists.WatchlistStatusDropped
	case sqlcgen.WatchlistStatusPlanToWatch:
		return watchlists.WatchlistStatusPlanToWatch
	case sqlcgen.WatchlistStatusWatching:
		return watchlists.WatchlistStatusWatching
	default:
		return watchlists.WatchlistStatusUnknown
	}
}

func NewWatchlistStatusModel(from watchlists.WatchlistStatus) sqlcgen.WatchlistStatus {
	switch from {
	case watchlists.WatchlistStatusCompleted:
		return sqlcgen.WatchlistStatusCompleted
	case watchlists.WatchlistStatusDropped:
		return sqlcgen.WatchlistStatusDropped
	case watchlists.WatchlistStatusPlanToWatch:
		return sqlcgen.WatchlistStatusPlanToWatch
	case watchlists.WatchlistStatusWatching:
		return sqlcgen.WatchlistStatusWatching
	default:
		return sqlcgen.WatchlistStatusUnknown
	}
}
