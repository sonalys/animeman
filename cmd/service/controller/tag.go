package controller

import (
	"slices"
	"strings"

	"github.com/sonalys/animeman/internal/pkg/metadata"
	"github.com/sonalys/animeman/internal/ports/torrentclient"
)

// getLatestTag is a pure function implementation for fetching the latest tag from a list of torrent entries.
func getLatestTag(torrents []torrentclient.Torrent) metadata.Tag {
	if len(torrents) == 0 {
		return metadata.Tag{}
	}

	var latestTag metadata.Tag

	for _, torrent := range torrents {
		seasonEpisodeTag, ok := findSeasonEpisodeTag(torrent.Tags)
		if !ok {
			continue
		}

		tags := metadata.ParseTags(seasonEpisodeTag)
		if len(tags) == 0 {
			continue
		}

		slices.SortFunc(tags, func(a, b metadata.Tag) int { return a.Compare(b) })

		if latestTag.IsZero() || tags[0].Compare(latestTag) > 0 {
			latestTag = tags[0]
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
