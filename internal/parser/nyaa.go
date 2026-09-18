package parser

import (
	"github.com/sonalys/animeman/internal/ports/torrentsource"
	"github.com/sonalys/animeman/internal/ports/animelist"
)

// ParsedNyaa holds a parsed entry from a torrent source.
// Used for smart episode detection.
type ParsedNyaa struct {
	// Metadata parsed from title.
	ExtractedMetadata Metadata
	// Torrent entry.
	NyaaTorrent torrentsource.Torrent
}

func NewParsedNyaa(
	animeListEntry animelist.Entry,
	entry torrentsource.Torrent,
	sources []string,
) ParsedNyaa {
	fallbackSeason := 1

	for _, title := range animeListEntry.Titles {
		if season := ParseSeason(title); season > 0 {
			fallbackSeason = season
			break
		}
	}

	meta := Parse(entry.Title, fallbackSeason, sources)
	return ParsedNyaa{
		ExtractedMetadata: meta,
		NyaaTorrent:       entry,
	}
}
