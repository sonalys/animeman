package collections

import (
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/sonalys/animeman/internal/domain/shared"
)

type (
	MediaID struct{ uuid.UUID }

	Media struct {
		ID           MediaID
		CollectionID CollectionID

		Titles []Title

		MonitoringStatus MonitoringStatus
		MonitoredSince   time.Time

		Metadata         MediaMetadata
		QualityProfileID QualityProfileID

		Seasons []*Season

		CreatedAt time.Time
	}
)

func (m *Media) NewSeason(
	number int,
	airingStatus AiringStatus,
	metadata SeasonMetadata,
) *Season {
	season := &Season{
		ID:             shared.NewID[SeasonID](),
		MediaID:        m.ID,
		Number:         number,
		AiringStatus:   airingStatus,
		SeasonMetadata: metadata,
	}
	m.Seasons = append(m.Seasons, season)

	return season
}
