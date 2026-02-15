package mappers

import (
	"github.com/sonalys/animeman/internal/adapters/postgres/sqlcgen"
	"github.com/sonalys/animeman/internal/domain/collections"
)

func NewMediaTypeModel(from collections.MediaType) sqlcgen.MediaType {
	switch from {
	case collections.MediaTypeMovie:
		return sqlcgen.MediaTypeMovie
	case collections.MediaTypeOVA:
		return sqlcgen.MediaTypeOva
	case collections.MediaTypeSpecial:
		return sqlcgen.MediaTypeSpecial
	case collections.MediaTypeTV:
		return sqlcgen.MediaTypeTv
	default:
		return sqlcgen.MediaTypeUnknown
	}
}

func NewMediaType(from sqlcgen.MediaType) collections.MediaType {
	switch from {
	case sqlcgen.MediaTypeMovie:
		return collections.MediaTypeMovie
	case sqlcgen.MediaTypeOva:
		return collections.MediaTypeOVA
	case sqlcgen.MediaTypeSpecial:
		return collections.MediaTypeSpecial
	case sqlcgen.MediaTypeTv:
		return collections.MediaTypeTV
	default:
		return collections.MediaTypeUnknown
	}
}
