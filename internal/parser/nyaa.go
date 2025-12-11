package parser

import "github.com/sonalys/animeman/integrations/nyaa"

// ParsedNyaa holds a parsed entry from Nyaa.
// Used for smart episode detection.
type ParsedNyaa struct {
	// Metadata parsed from title.
	Meta Metadata
	// Nyaa entry.
	Entry nyaa.Entry
}

func NewParsedNyaa(entry nyaa.Entry) ParsedNyaa {
	meta := Parse(entry.Title)
	return ParsedNyaa{
		Meta:  meta,
		Entry: entry,
	}
}
