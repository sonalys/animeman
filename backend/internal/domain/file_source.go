package domain

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
		return "TV"
	case FileSourceWEB:
		return "WEB"
	case FileSourceDVD:
		return "DVD"
	case FileSourceBluRay:
		return "BR"
	default:
		return "UNKNOWN"
	}
}

func (s FileSource) IsValid() bool {
	return s >= FileSourceUnknown && s < fileSourceSentinel
}
