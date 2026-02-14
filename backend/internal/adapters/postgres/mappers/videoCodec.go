package mappers

import (
	"github.com/sonalys/animeman/internal/adapters/postgres/sqlcgen"
	"github.com/sonalys/animeman/internal/domain/stream"
)

func NewVideoCodecModel(from stream.VideoCodec) sqlcgen.VideoCodec {
	switch from {
	case stream.VideoCodecAV1:
		return sqlcgen.VideoCodecAv1
	case stream.VideoCodecX264:
		return sqlcgen.VideoCodecX264
	case stream.VideoCodecX265:
		return sqlcgen.VideoCodecX265
	default:
		return sqlcgen.VideoCodecUnknown
	}
}

func NewVideoCodec(from sqlcgen.VideoCodec) stream.VideoCodec {
	switch from {
	case sqlcgen.VideoCodecAv1:
		return stream.VideoCodecAV1
	case sqlcgen.VideoCodecX264:
		return stream.VideoCodecX264
	case sqlcgen.VideoCodecX265:
		return stream.VideoCodecX265
	default:
		return stream.VideoCodecUnknown
	}
}
