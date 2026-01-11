package parser

import "github.com/sonalys/animeman/internal/integrations/nyaa"

// ParsedNyaa holds a parsed entry from Nyaa.
// Used for smart episode detection.
type ParsedNyaa struct {
	// Metadata parsed from title.
	ExtractedMetadata Metadata
	// Nyaa entry.
	NyaaTorrent nyaa.Entry
}

func NewParsedNyaa(entry nyaa.Entry) ParsedNyaa {
	meta := Parse(entry.Title)
	return ParsedNyaa{
		ExtractedMetadata: meta,
		NyaaTorrent:       entry,
	}
}
