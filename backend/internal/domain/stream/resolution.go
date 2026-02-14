package stream

type Resolution uint

const (
	ResolutionUnknown Resolution = iota
	ResolutionSD
	Resolution720p
	Resolution1080p
	Resolution2160p
	resolutionSentinel
)

func (r Resolution) String() string {
	switch r {
	case ResolutionSD:
		return "480p"
	case Resolution720p:
		return "720p"
	case Resolution1080p:
		return "1080p"
	case Resolution2160p:
		return "2160p"
	default:
		return "unknown"
	}
}

func (r Resolution) IsValid() bool {
	return r > ResolutionUnknown && r < resolutionSentinel
}
