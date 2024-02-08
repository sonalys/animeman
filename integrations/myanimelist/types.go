package myanimelist

import (
	"fmt"
	"net/url"
)

type (
	ListStatus   int
	AiringStatus int

	AnimeListEntry struct {
		Status       ListStatus   `json:"status"`
		Title        string       `json:"anime_title"`
		TitleEng     string       `json:"anime_title_eng"`
		AiringStatus AiringStatus `json:"anime_airing_status"`
	}
)

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

var statusNames = []string{
	"Unknown",
	"Watching",
	"Completed",
	"OnHold",
	"Dropped",
	"WontWatch",
	"PlanToWatch",
	"All",
}

func (s ListStatus) Name() string {
	return statusNames[s]
}

func (s ListStatus) ApplyList(v url.Values) {
	v.Set("status", fmt.Sprint(s))
}

func (s AiringStatus) ApplyList(v url.Values) {
	v.Set("airing_status", fmt.Sprint(s))
}
