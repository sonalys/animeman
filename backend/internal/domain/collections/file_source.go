package collections

type FileSource uint

const (
	FileSourceUnknown FileSource = iota
	FileSourceTV
	FileSourceWEB
	FileSourceDVD
	FileSourceBluRay
	fileSourceSentinel
)

func (c FileSource) String() string {
	switch c {
	case FileSourceTV:
		return "tv"
	case FileSourceWEB:
		return "web"
	case FileSourceDVD:
		return "dvd"
	case FileSourceBluRay:
		return "br"
	default:
		return "unknown"
	}
}

func (s FileSource) IsValid() bool {
	return s >= FileSourceUnknown && s < fileSourceSentinel
}
