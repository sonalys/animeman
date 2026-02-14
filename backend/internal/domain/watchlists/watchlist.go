package watchlists

import (
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/sonalys/animeman/internal/domain/collections"
	"github.com/sonalys/animeman/internal/domain/shared"
)

type (
	WatchlistID      struct{ uuid.UUID }
	WatchlistEntryID struct{ uuid.UUID }

	Watchlist struct {
		ID    WatchlistID
		Owner shared.UserID

		Source     WatchlistSource
		ExternalID string

		SyncFrequency time.Duration
		LastSyncedAt  time.Time

		CreatedAt time.Time
	}

	WatchlistEntry struct {
		ID          WatchlistEntryID
		WatchlistID WatchlistID

		MediaID     collections.MediaID
		SeasonID    collections.SeasonID
		LastWatched collections.EpisodeID

		Status WatchlistStatus

		CreatedAt time.Time
		UpdatedAt time.Time
	}
)

func (l Watchlist) NewEntry(
	mediaID collections.MediaID,
	seasonID collections.SeasonID,
	status WatchlistStatus,
) *WatchlistEntry {
	now := time.Now()

	return &WatchlistEntry{
		ID:          shared.NewID[WatchlistEntryID](),
		WatchlistID: l.ID,
		MediaID:     mediaID,
		SeasonID:    seasonID,
		Status:      status,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func (e *WatchlistEntry) SetLastWatchedEpisode(id collections.EpisodeID) {
	e.LastWatched = id
}
