package animelist

import "time"

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
