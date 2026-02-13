package domain

import "slices"

type QualityProfile struct {
	ID   int64
	Name string

	MinResolution Resolution
	MaxResolution Resolution

	// Represents a list from least to most preferrable codec.
	CodecPreference []Codec
	// Represents a list from least to most preferrable release group.
	ReleaseGroupPreference []string
}

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
