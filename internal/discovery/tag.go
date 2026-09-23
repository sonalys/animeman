package discovery

import (
	"strings"

	"github.com/sonalys/animeman/internal/ports/torrentclient"
	"github.com/sonalys/animeman/internal/tags"
	"github.com/sonalys/animeman/internal/utils/parser"
)

// getLatestTag is a pure function implementation for fetching the latest tag from a list of torrent entries.
func getLatestTag(torrents []torrentclient.Torrent) tags.Tag {
	if len(torrents) == 0 {
		return tags.Tag{}
	}

	var latestTag tags.Tag

	for _, torrent := range torrents {
		seasonEpisodeTag, ok := findSeasonEpisodeTag(torrent.Tags)
		if !ok {
			continue
		}
		tag := parser.Parse(seasonEpisodeTag, 1, nil).Tag

		if latestTag.IsZero() || tag.Compare(latestTag) > 0 {
			latestTag = tag
		}
	}

	return latestTag
}

// findSeasonEpisodeTag returns the first tag starting with "S", which is
// the season/episode tag convention used by this app.
func findSeasonEpisodeTag(torrentTags []string) (string, bool) {
	for _, torrentTag := range torrentTags {
		if strings.HasPrefix(torrentTag, "S") {
			return torrentTag, true
		}
	}
	return "", false
}
