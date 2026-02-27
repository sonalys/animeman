package dtos

type (
	Subtitles []Subtitle

	Subtitle struct {
		Language string `json:"language,omitzero"`
		Title    string `json:"title,omitzero"`
		Format   string `json:"format,omitzero"`
	}
)
