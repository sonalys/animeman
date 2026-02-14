package domain

type (
	VideoInfo struct {
		Codec      VideoCodec
		Resolution Resolution
		BitDepth   uint
		BitRate    uint
		Width      uint
		Height     uint
	}
)
