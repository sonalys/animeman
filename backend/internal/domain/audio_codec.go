package domain

type AudioCodec uint

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
		return "AAC"
	case AudioCodecOpus:
		return "OPUS"
	case AudioCodecFLAC:
		return "FLAC"
	case AudioCodecMP3:
		return "MP3"
	case AudioCodecAC3:
		return "AC3"
	case AudioCodecDTS:
		return "DTS"
	case AudioCodecTrueHD:
		return "TRUEHD"
	default:
		return "UNKNOWN"
	}
}

func (c AudioCodec) IsValid() bool {
	return c > AudioCodecUnknown && c < audioCodecSentinel
}

// IsLossless returns true if the codec preserves all audio data
func (c AudioCodec) IsLossless() bool {
	return c == AudioCodecFLAC || c == AudioCodecTrueHD
}
