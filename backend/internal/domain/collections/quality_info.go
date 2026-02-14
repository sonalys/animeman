package collections

import "github.com/sonalys/animeman/internal/domain/stream"

type QualityInfo struct {
	Resolution stream.Resolution
	Codec      stream.VideoCodec
}
