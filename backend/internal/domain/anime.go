package domain

import (
	"slices"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/sonalys/animeman/internal/utils/sliceutils"
)

type (
	AnimeListSource uint
	ListStatus      uint
	AiringStatus    uint

	Image struct {
		MimeType string
		URL      string
		Width    int
		Height   int
	}

	AnimeListEntry struct {
		ExternalID   string
		Progress     int
		ListStatus   ListStatus
		Titles       []string
		AiringStatus AiringStatus
		StartDate    time.Time
		NumEpisodes  int
		Images       []Image
	}

	AnimeListID struct{ uuid.UUID }

	AnimeList struct {
		ID             AnimeListID
		OwnerID        UserID
		RemoteUsername string
		Source         AnimeListSource
		Entries        []AnimeListEntry
	}
)

const (
	ListStatusUnset ListStatus = iota
	ListStatusWatching
	ListStatusCompleted
	ListStatusOnHold
	ListStatusDropped
	ListStatusPlanToWatch
	ListStatusAll
	_listStatusCeiling
)

const (
	AiringStatusUnset AiringStatus = iota
	AiringStatusAired
	AiringStatusAiring
	_airingStatusCeiling
)

const (
	AnimeListSourceUnset AnimeListSource = iota
	AnimeListSourceMAL
	AnimeListSourceAnilist
	_animeListSourceCeiling
)

func NewEntry(
	titles []string,
	listStatus ListStatus,
	airingStatus AiringStatus,
	startDate time.Time,
	numEpisodes int,
) AnimeListEntry {
	titles = sliceutils.Filter(titles, func(s string) bool { return len(s) > 0 })
	titles = slices.Compact(titles)

	return AnimeListEntry{
		Titles:       titles,
		ListStatus:   listStatus,
		AiringStatus: airingStatus,
		StartDate:    startDate,
		NumEpisodes:  numEpisodes,
	}
}
