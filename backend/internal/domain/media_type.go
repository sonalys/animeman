package domain

type MediaType uint

const (
	MediaTypeUnknown MediaType = iota
	MediaTypeTV
	MediaTypeMovie
	MediaTypeOVA
	MediaTypeSpecial
)

func (t MediaType) String() string {
	switch t {
	case MediaTypeTV:
		return "TV"
	case MediaTypeMovie:
		return "MOVIE"
	case MediaTypeOVA:
		return "OVA"
	case MediaTypeSpecial:
		return "SPECIAL"
	default:
		return "UNKNOWN"
	}
}

func (t MediaType) IsValid() bool {
	return t > MediaTypeUnknown && t <= MediaTypeSpecial
}
