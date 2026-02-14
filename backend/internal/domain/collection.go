package domain

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type CollectionID struct{ uuid.UUID }

type Collection struct {
	ID    CollectionID
	Owner UserID

	Name     string
	BasePath string
	Tags     []string

	Monitored bool

	CreatedAt time.Time
}

func (c Collection) NewMedia(
	titles []AlternativeTitle,
	monitoringStatus MonitoringStatus,
	metadata MediaMetadata,
	qualityProfile QualityProfile,
) *Media {
	var monitoredSince time.Time

	now := time.Now()

	if monitoringStatus != MonitoringStatusNone {
		monitoredSince = now
	}

	return &Media{
		ID:               NewID[MediaID](),
		CollectionID:     c.ID,
		Titles:           titles,
		MonitoringStatus: monitoringStatus,
		MonitoredSince:   monitoredSince,
		Metadata:         metadata,
		QualityProfile:   qualityProfile,
		CreatedAt:        now,
	}
}
