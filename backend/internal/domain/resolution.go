package domain

type Resolution uint

const (
	ResUnknown Resolution = iota
	ResSD
	Res720p
	Res1080p
	Res2160p
)

func (r Resolution) String() string {
	switch r {
	case ResSD:
		return "480p"
	case Res720p:
		return "720p"
	case Res1080p:
		return "1080p"
	case Res2160p:
		return "2160p"
	default:
		return "UNKNOWN"
	}
}

func (r Resolution) IsValid() bool {
	return r > ResUnknown && r <= Res2160p
}
