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

		Titles       []AlternativeTitle
		AiringStatus AiringStatus

		MonitoringStatus MonitoringStatus
		MonitoredSince   time.Time

		Metadata       MediaMetadata
		QualityProfile QualityProfile

		Seasons []Season

		CreatedAt time.Time
	}
)
