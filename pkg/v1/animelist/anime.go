package animelist

import (
	"time"

	"github.com/sonalys/animeman/internal/utils"
)

type ListStatus int
type AiringStatus int

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
	ListStatus   ListStatus
	Titles       []string
	AiringStatus AiringStatus
	StartDate    time.Time
	NumEpisodes  int
}

func NewEntry(
	titles []string,
	listStatus ListStatus,
	airingStatus AiringStatus,
	startDate time.Time,
	numEpisodes int,
) Entry {
	titles = utils.Filter(titles, func(s string) bool { return len(s) > 0 })
	titles = utils.Deduplicate(titles)

	return Entry{
		Titles:       titles,
		ListStatus:   listStatus,
		AiringStatus: airingStatus,
		StartDate:    startDate,
		NumEpisodes:  numEpisodes,
	}
}
