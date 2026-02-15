package watchlists

type Source uint

const (
	SourceUnknown Source = iota
	SourceLocal
	SourceAniList
	SourceMyAnimeList
	sourceSentinel
)

func (s Source) IsValid() bool {
	return s > SourceUnknown && s < sourceSentinel
}

func (s Source) String() string {
	switch s {
	case SourceLocal:
		return "local"
	case SourceAniList:
		return "anilist"
	case SourceMyAnimeList:
		return "mal"
	default:
		return "unknown"
	}
}
