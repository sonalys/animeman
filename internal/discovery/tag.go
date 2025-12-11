package discovery

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/sonalys/animeman/internal/parser"
	"github.com/sonalys/animeman/internal/utils"
	"github.com/sonalys/animeman/pkg/v1/torrentclient"
)

// Regexp for detecting numbers.
var numberExpr = regexp.MustCompile(`\d+(\.\d+)?`)

// Regexp for detecting batch tag numbers.
// Example: S02E01~13.
var batchReplaceExpr = regexp.MustCompile(`(\d+)~(\d+)`)

func strToFloat64(cur string) float64 {
	return utils.Must(strconv.ParseFloat(cur, 64))
}

// tagMergeBatchEpisodes will receive a tag represented by S0E1~12.
// it will transform it into S0E12 so the episode detection will only download episodes 13 and forward.
func tagMergeBatchEpisodes(tag string) string {
	matches := batchReplaceExpr.FindAllStringSubmatch(tag, -1)
	if len(matches) == 0 {
		return tag
	}
	values := matches[0][1:]
	return batchReplaceExpr.ReplaceAllString(
		tag,
		fmt.Sprint(utils.Max(utils.Map(values, strToFloat64)...)),
	)
}

// tagCompare receives 2 series tags, Example: S02E01 and S02E02.
// it will return the comparison of Tag1, Tag2.
// -1 = Tag1 < Tag2.
// 0 = Tag1 == Tag2.
// 1 = Tag1 > Tag2.
func tagCompare(a, b parser.SeasonEpisodeTag) int {
	if a.LastSeason() < b.LastSeason() {
		return -1
	}

	if a.LastSeason() > b.LastSeason() {
		return 1
	}

	if a.IsMultiEpisode() && !b.IsMultiEpisode() {
		return 1
	}

	if b.IsMultiEpisode() && !a.IsMultiEpisode() {
		return -1
	}

	if a.LastEpisode() < b.LastEpisode() {
		return -1
	}

	if a.LastEpisode() > b.LastEpisode() {
		return 1
	}

	return 0
}

// tagGetLatest is a pure function implementation for fetching the latest tag from a list of torrent entries.
func tagGetLatest(torrents []torrentclient.Torrent) parser.SeasonEpisodeTag {
	if len(torrents) == 0 {
		return parser.SeasonEpisodeTag{}
	}

	var latestTag parser.SeasonEpisodeTag

	for _, torrent := range torrents {
		tags := torrent.Tags
		seasonEpisodeTag := tags[len(tags)-1]
		meta := parser.Parse(seasonEpisodeTag)
		tag := meta.SeasonEpisodeTag

		if latestTag.IsZero() || tagCompare(tag, latestTag) > 0 {
			latestTag = tag
		}
	}

	return latestTag
}
