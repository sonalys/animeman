package domain

type TitleType uint

const (
	TitleTypeUnknown TitleType = iota
	TitleTypeMain              // The primary title chosen by the user
	TitleTypeRomaji            // Romanized Japanese
	TitleTypeEnglish           // Official English
	TitleTypeNative            // Original Kanji/Kana
	TitleTypeSynonym           // Fan-names or abbreviations (e.g., "Danmachi")
	titleTypeSentinel
)

func (t TitleType) String() string {
	switch t {
	case TitleTypeMain:
		return "MAIN"
	case TitleTypeRomaji:
		return "ROMAJI"
	case TitleTypeEnglish:
		return "ENGLISH"
	case TitleTypeNative:
		return "NATIVE"
	case TitleTypeSynonym:
		return "SYNONYM"
	default:
		return "UNKNOWN"
	}
}

func (t TitleType) IsValid() bool {
	return t > TitleTypeUnknown && t < titleTypeSentinel
}

type AlternativeTitle struct {
	Title    string
	Language string // ISO 639-1 (en, ja, etc.)
	Type     TitleType
}
