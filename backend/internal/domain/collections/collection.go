package collections

import (
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/sonalys/animeman/internal/domain/shared"
)

type CollectionID struct{ uuid.UUID }

type Collection struct {
	ID    CollectionID
	Owner shared.UserID

	Name     string
	BasePath string
	Tags     []string

	Monitored bool

	CreatedAt time.Time
}

func (c Collection) NewMedia(
	titles []Title,
	monitoringStatus MonitoringStatus,
	metadata MediaMetadata,
	qualityProfileID QualityProfileID,
) *Media {
	var monitoredSince time.Time

	now := time.Now()

	if monitoringStatus != MonitoringStatusNone {
		monitoredSince = now
	}

	return &Media{
		ID:               shared.NewID[MediaID](),
		CollectionID:     c.ID,
		Titles:           titles,
		MonitoringStatus: monitoringStatus,
		MonitoredSince:   monitoredSince,
		Metadata:         metadata,
		QualityProfileID: qualityProfileID,
		CreatedAt:        now,
	}
}
