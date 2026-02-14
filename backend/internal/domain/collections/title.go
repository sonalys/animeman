package collections

type (
	TitleType uint

	Title struct {
		Value    string
		Language string
		Type     TitleType
	}
)

const (
	TitleTypeUnknown TitleType = iota
	TitleTypeRomaji
	TitleTypeEnglish
	TitleTypeNative
	TitleTypeSynonym
	titleTypeSentinel
)

func (t TitleType) String() string {
	switch t {
	case TitleTypeRomaji:
		return "romaji"
	case TitleTypeEnglish:
		return "english"
	case TitleTypeNative:
		return "native"
	case TitleTypeSynonym:
		return "synonym"
	default:
		return "unknown"
	}
}

func (t TitleType) IsValid() bool {
	return t > TitleTypeUnknown && t < titleTypeSentinel
}

func NewTitle(
	titleType TitleType,
	language string,
	value string,
) Title {
	return Title{
		Type:     titleType,
		Language: language,
		Value:    value,
	}
}
