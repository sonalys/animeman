package myanimelist

import "fmt"

type (
	ListStatus int
	Title      string

	AnimeListEntry struct {
		Status   ListStatus `json:"status"`
		Title    any        `json:"anime_title"`
		TitleEng string     `json:"anime_title_eng"`
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

func (e *AnimeListEntry) GetTitle() string {
	if e.TitleEng != "" {
		return e.TitleEng
	}
	return fmt.Sprint(e.Title)
}
