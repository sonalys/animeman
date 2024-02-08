package discovery

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	"github.com/rs/zerolog/log"
	"github.com/sonalys/animeman/internal/parser"
	"github.com/sonalys/animeman/internal/utils"
	"github.com/sonalys/animeman/pkg/v1/animelist"
	"github.com/sonalys/animeman/pkg/v1/torrentclient"
	"golang.org/x/exp/constraints"
)

var numberExpr = regexp.MustCompile(`\d+`)
var batchReplaceExpr = regexp.MustCompile(`(\d+)~(\d+)`)

func strSliceToInt(from []string) []int64 {
	out := make([]int64, 0, len(from))
	for _, cur := range from {
		out = append(out, utils.Must(strconv.ParseInt(cur, 10, 64)))
	}
	return out
}

func min[T constraints.Ordered](values ...T) (min T) {
	if len(values) == 0 {
		return
	}
	min = values[0]
	for i := range values {
		if values[i] < min {
			min = values[i]
		}
	}
	return min
}

func max[T constraints.Ordered](values ...T) (max T) {
	if len(values) == 0 {
		return
	}
	max = values[0]
	for i := range values {
		if values[i] > max {
			max = values[i]
		}
	}
	return max
}

func mergeBatchEpisodes(tag string) string {
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

func compareTags(a, b string) int {
	if a == "" && b != "" {
		return -1
	}
	if a != "" && b == "" {
		return 1
	}
	a = mergeBatchEpisodes(a)
	b = mergeBatchEpisodes(b)
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

func getLatestTag(torrents ...torrentclient.Torrent) string {
	var latestTag string
	for _, torrent := range torrents {
		tags := torrent.Tags
		seasonEpisodeTag := tags[len(tags)-1]
		if compareTags(seasonEpisodeTag, latestTag) > 0 {
			latestTag = seasonEpisodeTag
		}
	}
	return latestTag
}

func (c *Controller) GetLatestTag(ctx context.Context, entry animelist.Entry) (string, error) {
	// check if torrent already exists, if so we skip it.
	title := parser.ParseTitle(entry.Title)
	titleEng := parser.ParseTitle(entry.TitleEng)
	torrents1, err := c.dep.TorrentClient.List(ctx, torrentclient.Tag(title.BuildSeriesTag()))
	if err != nil {
		return "", fmt.Errorf("listing torrents: %w", err)
	}
	torrents2, err := c.dep.TorrentClient.List(ctx, torrentclient.Tag(titleEng.BuildSeriesTag()))
	if err != nil {
		return "", fmt.Errorf("listing torrents: %w", err)
	}
	return getLatestTag(append(torrents1, torrents2...)...), nil
}

func (c *Controller) DigestNyaaTorrent(ctx context.Context, entry animelist.Entry, nyaaEntry TaggedNyaa) error {
	if nyaaEntry.meta.IsMultiEpisode && entry.AiringStatus == animelist.AiringStatusAiring {
		log.Debug().Msgf("torrent dropped: multi-episode for currently airing")
		return nil
	}
	var savePath torrentclient.SavePath
	if c.dep.Config.CreateShowFolder {
		savePath = torrentclient.SavePath(fmt.Sprintf("%s/%s", c.dep.Config.DownloadPath, entry.GetTitle()))
	} else {
		savePath = torrentclient.SavePath(c.dep.Config.DownloadPath)
	}
	tags := nyaaEntry.meta.BuildTorrentTags()
	err := c.dep.TorrentClient.AddTorrent(ctx,
		tags,
		savePath,
		torrentclient.TorrentURL{nyaaEntry.entry.Link},
		torrentclient.Category(c.dep.Config.Category),
	)
	if err != nil {
		return fmt.Errorf("adding torrents: %w", err)
	}
	log.Info().
		Str("savePath", string(savePath)).
		Strs("tag", tags).
		Msgf("torrent added")
	return nil
}
