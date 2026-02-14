package stream

type (
	SubtitleFormat uint

	Subtitle struct {
		Language string
		Title    string
		Format   SubtitleFormat
	}
)

const (
	SubtitleFormatUnknown  SubtitleFormat = iota
	SubtitleFormatSRT                     // SubRip (Basic text)
	SubtitleFormatASS                     // Advanced Substation Alpha (Styled/Typeset)
	SubtitleFormatSSA                     // Substation Alpha (Older styled)
	SubtitleFormatPGS                     // Presentation Graphic Stream (Blu-ray image-based)
	SubtitleFormatVobSub                  // DVD image-based (idx/sub)
	subtitleFormatSentinel                // Private ceiling
)

func (f SubtitleFormat) String() string {
	switch f {
	case SubtitleFormatSRT:
		return "srt"
	case SubtitleFormatASS:
		return "ass"
	case SubtitleFormatSSA:
		return "ssa"
	case SubtitleFormatPGS:
		return "pgs"
	case SubtitleFormatVobSub:
		return "vobsub"
	default:
		return "unknown"
	}
}

func (f SubtitleFormat) IsValid() bool {
	return f > SubtitleFormatUnknown && f < subtitleFormatSentinel
}
