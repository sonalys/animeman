package mappers

import (
	"github.com/sonalys/animeman/internal/adapters/postgres/sqlcgen"
	"github.com/sonalys/animeman/internal/domain/watchlists"
)

func NewWatchlistSource(from sqlcgen.WatchlistSource) watchlists.Source {
	switch from {
	case sqlcgen.WatchlistSourceAnilist:
		return watchlists.SourceAniList
	case sqlcgen.WatchlistSourceMal:
		return watchlists.SourceMyAnimeList
	case sqlcgen.WatchlistSourceLocal:
		return watchlists.SourceLocal
	default:
		return watchlists.SourceUnknown
	}
}

func NewWatchlistSourceModel(from watchlists.Source) sqlcgen.WatchlistSource {
	switch from {
	case watchlists.SourceAniList:
		return sqlcgen.WatchlistSourceAnilist
	case watchlists.SourceMyAnimeList:
		return sqlcgen.WatchlistSourceMal
	case watchlists.SourceLocal:
		return sqlcgen.WatchlistSourceLocal
	default:
		return sqlcgen.WatchlistSourceUnknown
	}
}
