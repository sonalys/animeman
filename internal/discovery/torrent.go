package discovery

import (
	"context"
	"fmt"
	"regexp"

	"github.com/rs/zerolog/log"
	"github.com/sonalys/animeman/internal/parser"
	"github.com/sonalys/animeman/pkg/v1/animelist"
	"github.com/sonalys/animeman/pkg/v1/torrentclient"
)

// Regexp for detecting numbers.
var numberExpr = regexp.MustCompile(`\d+`)

// Regexp for detecting batch tag numbers.
// Example: S02E01~13.
var batchReplaceExpr = regexp.MustCompile(`(\d+)~(\d+)`)

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
		fmt.Sprint(max(strSliceToInt(values)...)),
	)
}

// tagCompare receives 2 series tags, Example: S02E01 and S02E02.
// it will return the comparison of Tag1, Tag2.
// -1 = Tag1 < Tag2.
// 0 = Tag1 == Tag2.
// 1 = Tag1 > Tag2.
func tagCompare(a, b string) int {
	if a == "" && b != "" {
		return -1
	}
	if a != "" && b == "" {
		return 1
	}
	a = tagMergeBatchEpisodes(a)
	b = tagMergeBatchEpisodes(b)
	aNums := strSliceToInt(numberExpr.FindAllString(a, -1))
	bNums := strSliceToInt(numberExpr.FindAllString(b, -1))
	lenA, lenB := len(aNums), len(bNums)
	minSize := min(lenA, lenB)
	for i := 0; i < minSize; i++ {
		if aNums[i] > bNums[i] {
			return 1
		}
		if aNums[i] < bNums[i] {
			return -1
		}
	}
	// Case for S3 and S3E2, we want the smaller one, which is more inclusive.
	if lenA < lenB {
		return 1
	}
	if lenA > lenB {
		return -1
	}
	return 0
}

// tagGetLatest is a pure function implementation for fetching the latest tag from a list of torrent entries.
func tagGetLatest(torrents ...torrentclient.Torrent) string {
	var latestTag string
	for _, torrent := range torrents {
		tags := torrent.Tags
		seasonEpisodeTag := tags[len(tags)-1]
		if tagCompare(seasonEpisodeTag, latestTag) > 0 {
			latestTag = seasonEpisodeTag
		}
	}
	return latestTag
}

// TagGetLatest will receive an anime list entry and return all torrents listed from the anime.
func (c *Controller) TagGetLatest(ctx context.Context, entry animelist.Entry) (string, error) {
	var torrents []torrentclient.Torrent
	for i := range entry.Titles {
		// check if torrent already exists, if so we skip it.
		title := parser.TitleParse(entry.Titles[i])
		// we should consider both title and titleEng, because your anime list has different titles available,
		// some torrent sources will use one, some will use the other, so to avoid duplication we check for both.
		torrents1, err := c.dep.TorrentClient.List(ctx, torrentclient.Tag(title.TagBuildSeries()))
		if err != nil {
			return "", fmt.Errorf("listing torrents: %w", err)
		}
		torrents = append(torrents, torrents1...)
	}
	return tagGetLatest(torrents...), nil
}

// torrentGetPath returns a torrent path, creating a show folder if configured.
func (c *Controller) torrentGetPath(title string) (path torrentclient.SavePath) {
	if c.dep.Config.CreateShowFolder {
		return torrentclient.SavePath(fmt.Sprintf("%s/%s", c.dep.Config.DownloadPath, title))
	}
	return torrentclient.SavePath(c.dep.Config.DownloadPath)
}

// DigestNyaaTorrent receives an anime list entry and a downloadable torrent.
// It will configure all necessary metadata and send it to your torrent client.
func (c *Controller) DigestNyaaTorrent(ctx context.Context, entry animelist.Entry, nyaaEntry ParsedNyaa) error {
	tags := nyaaEntry.meta.TagsBuildTorrent()
	savePath := c.torrentGetPath(entry.Titles[0])
	torrentURL := torrentclient.TorrentURL{nyaaEntry.entry.Link}
	category := torrentclient.Category(c.dep.Config.Category)
	if err := c.dep.TorrentClient.AddTorrent(ctx, tags, savePath, torrentURL, category); err != nil {
		return fmt.Errorf("adding torrents: %w", err)
	}
	log.Info().Str("savePath", string(savePath)).Strs("tag", tags).Msgf("torrent added")
	return nil
}