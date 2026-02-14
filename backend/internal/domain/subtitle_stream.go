package domain

type (
	SubtitleStream struct {
		Language  string
		Title     string
		Format    SubtitleFormat
		IsDefault bool
		IsForced  bool
	}
)
