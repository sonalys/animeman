package watchlists

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
		return "local"
	case WatchlistSourceAniList:
		return "anilist"
	case WatchlistSourceMyAnimeList:
		return "mal"
	default:
		return "unknown"
	}
}
