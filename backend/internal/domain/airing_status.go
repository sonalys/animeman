package domain

type AiringStatus uint

const (
	AiringStatusUnknown AiringStatus = iota
	AiringStatusUpcoming
	AiringStatusAiring
	AiringStatusFinished
	_AiringStatusCeiling
)

func (s AiringStatus) String() string {
	switch s {
	case AiringStatusUpcoming:
		return "UPCOMING"
	case AiringStatusAiring:
		return "AIRING"
	case AiringStatusFinished:
		return "FINISHED"
	default:
		return "UNKNOWN"
	}
}

func (s AiringStatus) IsValid() bool {
	return s > AiringStatusUnknown && s < _AiringStatusCeiling
}
