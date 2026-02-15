package mappers

import (
	"github.com/sonalys/animeman/internal/adapters/postgres/sqlcgen"
	"github.com/sonalys/animeman/internal/domain/watchlists"
)

func NewWatchlistSource(from sqlcgen.WatchlistSource) watchlists.Source {
	switch from {
	case sqlcgen.WatchlistSourceAnilist:
		return watchlists.WatchlistSourceAniList
	case sqlcgen.WatchlistSourceMal:
		return watchlists.WatchlistSourceMyAnimeList
	case sqlcgen.WatchlistSourceLocal:
		return watchlists.WatchlistSourceLocal
	default:
		return watchlists.WatchlistSourceUnknown
	}
}

func NewWatchlistSourceModel(from watchlists.Source) sqlcgen.WatchlistSource {
	switch from {
	case watchlists.WatchlistSourceAniList:
		return sqlcgen.WatchlistSourceAnilist
	case watchlists.WatchlistSourceMyAnimeList:
		return sqlcgen.WatchlistSourceMal
	case watchlists.WatchlistSourceLocal:
		return sqlcgen.WatchlistSourceLocal
	default:
		return sqlcgen.WatchlistSourceUnknown
	}
}
