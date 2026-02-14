package domain

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type (
	MediaID struct{ uuid.UUID }

	Media struct {
		ID           MediaID
		CollectionID CollectionID

		Titles []AlternativeTitle

		MonitoringStatus MonitoringStatus
		MonitoredSince   time.Time

		Metadata       MediaMetadata
		QualityProfile QualityProfile

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
		ID:             NewID[SeasonID](),
		MediaID:        m.ID,
		Number:         number,
		AiringStatus:   airingStatus,
		SeasonMetadata: metadata,
	}
	m.Seasons = append(m.Seasons, season)

	return season
}
