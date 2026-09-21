package parser

import (
	"github.com/sonalys/animeman/internal/ports/animelist"
	"github.com/sonalys/animeman/internal/ports/torrentsource"
)

// TorrentMetadata holds a parsed entry from a torrent source.
// Used for smart episode detection.
type TorrentMetadata struct {
	Metadata Metadata
	Torrent  torrentsource.Torrent
}

func ParseTorrentMetadata(
	animeListEntry animelist.Entry,
	entry torrentsource.Torrent,
	sources []string,
) TorrentMetadata {
	fallbackSeason := 1

	for _, title := range animeListEntry.Titles {
		if season := parseSeason(title); season > 0 {
			fallbackSeason = season
			break
		}
	}

	meta := Parse(entry.Title, fallbackSeason, sources)
	return TorrentMetadata{
		Metadata: meta,
		Torrent:  entry,
	}
}
