package discovery

import (
	"github.com/sonalys/animeman/internal/parser"
	"github.com/sonalys/animeman/internal/ports/torrentclient"
	"github.com/sonalys/animeman/internal/tags"
)

// getLatestTag is a pure function implementation for fetching the latest tag from a list of torrent entries.
func getLatestTag(torrents []torrentclient.Torrent) tags.Tag {
	if len(torrents) == 0 {
		return tags.Tag{}
	}

	var latestTag tags.Tag

	for _, torrent := range torrents {
		tags := torrent.Tags
		seasonEpisodeTag := tags[len(tags)-1]
		meta := parser.Parse(seasonEpisodeTag, 1, nil)
		tag := meta.Tag

		if latestTag.IsZero() || tag.Compare(latestTag) > 0 {
			latestTag = tag
		}
	}

	return latestTag
}
