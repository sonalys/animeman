package mappers

import (
	"github.com/sonalys/animeman/internal/adapters/postgres/sqlcgen"
	"github.com/sonalys/animeman/internal/domain/stream"
)

func NewResolutionModel(from stream.Resolution) sqlcgen.Resolution {
	switch from {
	case stream.Resolution720p:
		return sqlcgen.Resolution720p
	case stream.Resolution1080p:
		return sqlcgen.Resolution1080p
	case stream.Resolution2160p:
		return sqlcgen.Resolution2160p
	default:
		return sqlcgen.ResolutionUnknown
	}
}

func NewResolution(from sqlcgen.Resolution) stream.Resolution {
	switch from {
	case sqlcgen.Resolution720p:
		return stream.Resolution720p
	case sqlcgen.Resolution1080p:
		return stream.Resolution1080p
	case sqlcgen.Resolution2160p:
		return stream.Resolution2160p
	default:
		return stream.ResolutionUnknown
	}
}
