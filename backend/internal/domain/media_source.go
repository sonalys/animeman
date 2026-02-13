package domain

type MediaSource uint

const (
	MediaSourceUnknown MediaSource = iota
	MediaSourceTV
	MediaSourceWEB
	MediaSourceDVD
	MediaSourceBluRay
)

func (c MediaSource) String() string {
	switch c {
	case MediaSourceTV:
		return "TV"
	case MediaSourceWEB:
		return "WEB"
	case MediaSourceDVD:
		return "DVD"
	case MediaSourceBluRay:
		return "BR"
	default:
		return "UNKNOWN"
	}
}

func (s MediaSource) IsValid() bool {
	return s > MediaSourceUnknown && s <= MediaSourceBluRay
}
