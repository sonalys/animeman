package collections

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
		return "tv"
	case MediaTypeMovie:
		return "movie"
	case MediaTypeOVA:
		return "ova"
	case MediaTypeSpecial:
		return "special"
	default:
		return "unknown"
	}
}

func (t MediaType) IsValid() bool {
	return t > MediaTypeUnknown && t <= MediaTypeSpecial
}
