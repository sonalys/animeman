package parser

import (
	"github.com/sonalys/animeman/internal/integrations/nyaa"
	"github.com/sonalys/animeman/pkg/v1/animelist"
)

// ParsedNyaa holds a parsed entry from Nyaa.
// Used for smart episode detection.
type ParsedNyaa struct {
	// Metadata parsed from title.
	ExtractedMetadata Metadata
	// Nyaa entry.
	NyaaTorrent nyaa.Item
}

func NewParsedNyaa(animeListEntry animelist.Entry, entry nyaa.Item) ParsedNyaa {
	fallbackSeason := 1

	for _, title := range animeListEntry.Titles {
		if season := ParseSeason(title); season > 0 {
			fallbackSeason = season
			break
		}
	}

	meta := Parse(entry.Title, fallbackSeason)
	return ParsedNyaa{
		ExtractedMetadata: meta,
		NyaaTorrent:       entry,
	}
}
