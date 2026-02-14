package collections

import (
	"slices"

	"github.com/gofrs/uuid/v5"
	"github.com/sonalys/animeman/internal/domain/shared"
	"github.com/sonalys/animeman/internal/domain/stream"
)

type (
	QualityProfileID struct{ uuid.UUID }

	QualityProfile struct {
		ID   QualityProfileID
		Name string

		MinResolution stream.Resolution
		MaxResolution stream.Resolution

		// Represents a list from least to most preferrable codec.
		CodecPreference []stream.VideoCodec
		// Represents a list from least to most preferrable release group.
		ReleaseGroupPreference []string
	}
)

// Compare compares two quality infos, and returns 1 if incoming is better, 0 if equivalent, and -1 if worse.
func (p *QualityProfile) Compare(incoming, existing QualityInfo) int {
	if incoming.Resolution > existing.Resolution {
		return 1
	}

	if incoming.Resolution < existing.Resolution {
		return -1
	}

	existingCodecValue := slices.Index(p.CodecPreference, existing.Codec)
	incomingCodecValue := slices.Index(p.CodecPreference, incoming.Codec)

	if incomingCodecValue > existingCodecValue {
		return 1
	}

	if incomingCodecValue < existingCodecValue {
		return -1
	}

	return 0
}

func NewQualityProfile(
	name string,
	minResolution stream.Resolution,
	maxResolution stream.Resolution,
	codecPreference []stream.VideoCodec,
	releaseGroupPreference []string,
) QualityProfile {
	return QualityProfile{
		ID:                     shared.NewID[QualityProfileID](),
		Name:                   name,
		MinResolution:          minResolution,
		MaxResolution:          maxResolution,
		CodecPreference:        codecPreference,
		ReleaseGroupPreference: releaseGroupPreference,
	}
}
