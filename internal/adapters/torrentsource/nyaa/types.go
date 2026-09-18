package nyaa

type (
	SourceType int
	Query      string
	QueryOr    []string
	Category   string
	User       string
)

const (
	SourceTypeAll SourceType = iota
	SourceTypeNoRemake
	SourceTypeTrusted
)

const (
	CategoryAnime                  Category = "1_0"
	CategoryAnimeEnglishTranslated Category = "1_2"
)
