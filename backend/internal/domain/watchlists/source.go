package watchlists

// Source identifies where the item originated
type Source uint

const (
	WatchlistSourceUnknown Source = iota
	WatchlistSourceLocal
	WatchlistSourceAniList
	WatchlistSourceMyAnimeList
	watchlistSourceSentinel
)

func (s Source) IsValid() bool {
	return s > WatchlistSourceUnknown && s < watchlistSourceSentinel
}

func (s Source) String() string {
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
