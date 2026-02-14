package domain

type VideoCodec uint

const (
	VideoCodecUnknown VideoCodec = iota
	VideoCodecX264
	VideoCodecX265
	VideoCodecAV1
	videoCodecSentinel
)

func (c VideoCodec) String() string {
	switch c {
	case VideoCodecX264:
		return "x264"
	case VideoCodecX265:
		return "x265"
	case VideoCodecAV1:
		return "AV1"
	default:
		return "UNKNOWN"
	}
}

func (s VideoCodec) IsValid() bool {
	return s >= VideoCodecUnknown && s < videoCodecSentinel
}
