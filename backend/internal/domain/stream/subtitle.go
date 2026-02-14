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
		return "SRT"
	case SubtitleFormatASS:
		return "ASS"
	case SubtitleFormatSSA:
		return "SSA"
	case SubtitleFormatPGS:
		return "PGS"
	case SubtitleFormatVobSub:
		return "VOBSUB"
	default:
		return "UNKNOWN"
	}
}

func (f SubtitleFormat) IsValid() bool {
	return f > SubtitleFormatUnknown && f < subtitleFormatSentinel
}
