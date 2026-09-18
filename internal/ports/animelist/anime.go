package animelist

import (
	"slices"
	"time"

	"github.com/sonalys/animeman/internal/utils"
)

type ListStatus int
type AiringStatus int

type EpisodeSchedule struct {
	Number  int
	AirDate time.Time
}

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

type Entry struct {
	ListStatus      ListStatus
	Titles          []string
	AiringStatus    AiringStatus
	StartDate       time.Time
	EndDate         time.Time
	NumEpisodes     int
	EpisodeSchedule []EpisodeSchedule
	// AnilistID is the AniList database id, 0 when unknown.
	AnilistID int
	// MALID is the MyAnimeList database id, 0 when unknown.
	MALID int
}

func NewEntry(
	titles []string,
	listStatus ListStatus,
	airingStatus AiringStatus,
	startDate time.Time,
	endDate time.Time,
	numEpisodes int,
	episodeSchedule []EpisodeSchedule,
) Entry {
	titles = utils.Filter(titles, func(s string) bool { return len(s) > 0 })
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

// WithIDs returns a copy of the entry with the given tracker ids set.
func (e Entry) WithIDs(anilistID, malID int) Entry {
	e.AnilistID = anilistID
	e.MALID = malID
	return e
}
