package stream

type (
	AudioCodec uint

	Audio struct {
		Language string
		Title    string
		Codec    AudioCodec
		Channels float32
		BitRate  uint
	}
)

const (
	AudioCodecUnknown  AudioCodec = iota
	AudioCodecAAC                 // Advanced Audio Coding (Standard)
	AudioCodecOpus                // Modern, high efficiency
	AudioCodecFLAC                // Free Lossless Audio Codec (High End)
	AudioCodecMP3                 // Legacy
	AudioCodecAC3                 // Dolby Digital
	AudioCodecDTS                 // Digital Theater Systems
	AudioCodecTrueHD              // Lossless Surround
	audioCodecSentinel            // Private ceiling
)

func (c AudioCodec) String() string {
	switch c {
	case AudioCodecAAC:
		return "aac"
	case AudioCodecOpus:
		return "opus"
	case AudioCodecFLAC:
		return "flac"
	case AudioCodecMP3:
		return "mp3"
	case AudioCodecAC3:
		return "ac3"
	case AudioCodecDTS:
		return "dts"
	case AudioCodecTrueHD:
		return "truehd"
	default:
		return "unknown"
	}
}

func (c AudioCodec) IsValid() bool {
	return c > AudioCodecUnknown && c < audioCodecSentinel
}

// IsLossless returns true if the codec preserves all audio data
func (c AudioCodec) IsLossless() bool {
	return c == AudioCodecFLAC || c == AudioCodecTrueHD
}
