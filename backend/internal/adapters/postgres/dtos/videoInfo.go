package dtos

type VideoInfo struct {
	Codec      string `json:"codec,omitzero"`
	Resolution string `json:"resolution,omitzero"`
	BitDepth   uint   `json:"bit_depth,omitzero"`
	BitRate    uint   `json:"bit_rate,omitzero"`
	Width      uint   `json:"width,omitzero"`
	Height     uint   `json:"height,omitzero"`
}
