package domain

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type (
	WatchlistID      struct{ uuid.UUID }
	WatchlistEntryID struct{ uuid.UUID }

	Watchlist struct {
		ID    WatchlistID
		Owner UserID

		Source     WatchlistSource
		ExternalID string

		SyncFrequency time.Duration
		LastSyncedAt  time.Time

		CreatedAt time.Time
	}

	WatchlistEntry struct {
		ID          WatchlistEntryID
		WatchlistID WatchlistID

		MediaID     MediaID
		SeasonID    SeasonID
		LastWatched EpisodeID

		Status WatchlistStatus

		CreatedAt time.Time
		UpdatedAt time.Time
	}
)

func (l Watchlist) NewEntry(
	mediaID MediaID,
	seasonID SeasonID,
	status WatchlistStatus,
) *WatchlistEntry {
	now := time.Now()

	return &WatchlistEntry{
		ID:          NewID[WatchlistEntryID](),
		WatchlistID: l.ID,
		MediaID:     mediaID,
		SeasonID:    seasonID,
		Status:      status,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func (e *WatchlistEntry) SetLastWatchedEpisode(id EpisodeID) {
	e.LastWatched = id
}
