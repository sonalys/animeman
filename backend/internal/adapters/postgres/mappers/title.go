package mappers

import (
	"github.com/sonalys/animeman/internal/adapters/postgres/dtos"
	"github.com/sonalys/animeman/internal/domain/collections"
)

func NewTitleModel(from collections.Title) dtos.Title {
	return dtos.Title{
		TitleValue: from.Value,
		Language:   from.Language,
		Type:       from.Type.String(),
	}
}

func NewTitle(from dtos.Title) collections.Title {
	return collections.Title{
		Value:    from.TitleValue,
		Language: from.Language,
		Type: func() collections.TitleType {
			switch from.Type {
			case collections.TitleTypeNative.String():
				return collections.TitleTypeNative
			case collections.TitleTypeEnglish.String():
				return collections.TitleTypeEnglish
			case collections.TitleTypeRomaji.String():
				return collections.TitleTypeRomaji
			case collections.TitleTypeSynonym.String():
				return collections.TitleTypeSynonym
			default:
				return collections.TitleTypeUnknown
			}
		}(),
	}
}
