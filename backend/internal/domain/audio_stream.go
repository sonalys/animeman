package domain

type (
	AudioStream struct {
		Language  string
		Title     string
		Codec     AudioCodec
		Channels  float32
		BitRate   uint
		IsDefault bool
	}
)
