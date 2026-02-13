package domain

type Codec uint

const (
	CodecUnknown Codec = iota
	CodecX264
	CodecX265
	CodecAV1
)

func (c Codec) String() string {
	switch c {
	case CodecX264:
		return "x264"
	case CodecX265:
		return "x265"
	case CodecAV1:
		return "AV1"
	default:
		return "UNKNOWN"
	}
}

func (s Codec) IsValid() bool {
	return s > CodecUnknown && s <= CodecAV1
}
