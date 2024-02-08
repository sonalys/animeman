package animelist

import (
	"fmt"
	"net/url"
)

type ListStatus int
type AiringStatus int

const (
	ListStatusWatching    ListStatus = 1
	ListStatusCompleted   ListStatus = 2
	ListStatusOnHold      ListStatus = 3
	ListStatusDropped     ListStatus = 4
	ListStatusPlanToWatch ListStatus = 6
	ListStatusAll         ListStatus = 7
)

const (
	AiringStatusAired AiringStatus = iota
	AiringStatusAiring
)

type Entry struct {
	ListStatus   ListStatus
	Titles       []string
	AiringStatus AiringStatus
}

type AnimeListArg interface {
	ApplyList(url.Values)
}

func (s ListStatus) ApplyList(v url.Values) {
	v.Set("status", fmt.Sprint(s))
}

func (s AiringStatus) ApplyList(v url.Values) {
	v.Set("airing_status", fmt.Sprint(s))
}
