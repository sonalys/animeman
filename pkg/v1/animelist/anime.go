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
	NumEpisodes     int
	EpisodeSchedule []EpisodeSchedule
}

func NewEntry(
	titles []string,
	listStatus ListStatus,
	airingStatus AiringStatus,
	startDate time.Time,
	numEpisodes int,
	episodeSchedule []EpisodeSchedule,
) Entry {
	titles = utils.Filter(titles, func(s string) bool { return len(s) > 0 })
	titles = slices.Compact(titles)
	slices.Sort(titles)

	return Entry{
		Titles:          titles,
		ListStatus:      listStatus,
		AiringStatus:    airingStatus,
		StartDate:       startDate,
		NumEpisodes:     numEpisodes,
		EpisodeSchedule: episodeSchedule,
	}
}
