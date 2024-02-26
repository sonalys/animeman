package parser

import "github.com/sonalys/animeman/integrations/nyaa"

// ParsedNyaa holds a parsed entry from Nyaa.
// Used for smart episode detection.
type ParsedNyaa struct {
	// Metadata parsed from title.
	Meta Metadata
	// Example: S02E03.
	SeasonEpisodeTag string
	// Nyaa entry.
	Entry nyaa.Entry
}

func NewParsedNyaa(entry nyaa.Entry) ParsedNyaa {
	meta := TitleParse(entry.Title)
	return ParsedNyaa{
		Meta:             meta,
		SeasonEpisodeTag: meta.TagBuildSeasonEpisode(),
		Entry:            entry,
	}
}
