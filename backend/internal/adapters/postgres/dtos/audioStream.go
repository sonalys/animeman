package dtos

type (
	AudioStreams []AudioStream

	AudioStream struct {
		Language string  `json:"language,omitzero"`
		Title    string  `json:"title,omitzero"`
		Codec    string  `json:"codec,omitzero"`
		Channels float32 `json:"channels,omitzero"`
		BitRate  uint    `json:"bit_rate,omitzero"`
	}
)
