package animelist

import (
	"context"
	"slices"
	"sort"
	"time"

	"github.com/sonalys/animeman/internal/pkg/sliceutils"
	"github.com/sonalys/animeman/internal/pkg/stringutils"
)

type (
	// AnimeListSource is the port implemented by every anime list adapter (anilist, myanimelist, ...).
	AnimeListSource interface {
		GetCurrentlyWatching(ctx context.Context) ([]Entry, error)
	}

	// AnilistIDResolver is an optional port for anime list adapters that can
	// resolve an AniList id from a MAL id, so entries always carry an AniList id.
	AnilistIDResolver interface {
		ResolveMAL(ctx context.Context, malID int) (anilistID int, err error)
	}

	ListStatus   int
	AiringStatus int

	EpisodeSchedule struct {
		Number  int
		AirDate time.Time
	}

	Entry struct {
		ListStatus      ListStatus
		Titles          []string
		AiringStatus    AiringStatus
		StartDate       time.Time
		EndDate         time.Time
		NumEpisodes     int
		EpisodeSchedule []EpisodeSchedule
		AnilistID       int
		MALID           int
	}
)

const (
	ListStatusUnknown ListStatus = iota
	ListStatusWatching
	ListStatusCompleted
	ListStatusOnHold
	ListStatusDropped
	ListStatusPlanToWatch
	ListStatusAll
)

const (
	AiringStatusUnknown AiringStatus = iota
	AiringStatusAired
	AiringStatusAiring
)

func NewEntry(
	titles []string,
	listStatus ListStatus,
	airingStatus AiringStatus,
	startDate time.Time,
	endDate time.Time,
	numEpisodes int,
	episodeSchedule []EpisodeSchedule,
) Entry {
	titles = sliceutils.Filter(titles, func(s string) bool { return len(s) > 0 })
	slices.Sort(titles)
	titles = slices.Compact(titles)

	return Entry{
		Titles:          titles,
		ListStatus:      listStatus,
		AiringStatus:    airingStatus,
		StartDate:       startDate,
		EndDate:         endDate,
		NumEpisodes:     numEpisodes,
		EpisodeSchedule: episodeSchedule,
	}
}

func (e *Entry) WithAnilistID(anilistID int) *Entry {
	e.AnilistID = anilistID
	return e
}

func (e *Entry) WithMALID(malID int) *Entry {
	e.MALID = malID
	return e
}

// selectIdealTitle avoids kanji titles for example, preferring english ones.
func (e Entry) GetBestTitle() string {
	titles := e.Titles

	if len(titles) == 0 {
		return ""
	}

	viableCandidates := make([]string, 0, len(titles))

	for _, t := range titles {
		if stringutils.IsASCII(t) {
			viableCandidates = append(viableCandidates, t)
		}
	}

	// Prefer the shortest title for the tags.
	sort.Slice(viableCandidates, func(i, j int) bool {
		return len(viableCandidates[i]) < len(viableCandidates[j])
	})

	if len(viableCandidates) > 0 {
		return viableCandidates[0]
	}

	// Fallback to first element if no ASCII title is found
	return titles[0]
}
