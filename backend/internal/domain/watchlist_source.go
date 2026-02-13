package domain

// WatchlistSource identifies where the item originated
type WatchlistSource uint

const (
	WatchlistSourceUnknown WatchlistSource = iota
	WatchlistSourceLocal
	WatchlistSourceAniList
	WatchlistSourceMyAnimeList
	watchlistSourceSentinel
)

func (s WatchlistSource) IsValid() bool {
	return s > WatchlistSourceUnknown && s < watchlistSourceSentinel
}

func (s WatchlistSource) String() string {
	switch s {
	case WatchlistSourceLocal:
		return "LOCAL"
	case WatchlistSourceAniList:
		return "ANILIST"
	case WatchlistSourceMyAnimeList:
		return "MAL"
	default:
		return "UNKNOWN"
	}
}
