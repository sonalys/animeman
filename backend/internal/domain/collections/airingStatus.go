package collections

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
		return "upcoming"
	case AiringStatusAiring:
		return "airing"
	case AiringStatusFinished:
		return "finished"
	default:
		return "unknown"
	}
}

func (s AiringStatus) IsValid() bool {
	return s > AiringStatusUnknown && s < _AiringStatusCeiling
}
