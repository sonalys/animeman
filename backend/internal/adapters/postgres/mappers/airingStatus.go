package mappers

import (
	"github.com/sonalys/animeman/internal/adapters/postgres/sqlcgen"
	"github.com/sonalys/animeman/internal/domain/collections"
)

func NewAiringStatusModel(from collections.AiringStatus) sqlcgen.AiringStatus {
	switch from {
	case collections.AiringStatusAiring:
		return sqlcgen.AiringStatusAiring
	case collections.AiringStatusFinished:
		return sqlcgen.AiringStatusFinished
	case collections.AiringStatusUpcoming:
		return sqlcgen.AiringStatusUpcoming
	default:
		return sqlcgen.AiringStatusUnknown
	}
}

func NewAiringStatus(from sqlcgen.AiringStatus) collections.AiringStatus {
	switch from {
	case sqlcgen.AiringStatusAiring:
		return collections.AiringStatusAiring
	case sqlcgen.AiringStatusFinished:
		return collections.AiringStatusFinished
	case sqlcgen.AiringStatusUpcoming:
		return collections.AiringStatusUpcoming
	default:
		return collections.AiringStatusUnknown
	}
}
