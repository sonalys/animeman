package discovery

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	"github.com/rs/zerolog/log"
	"github.com/sonalys/animeman/integrations/myanimelist"
	"github.com/sonalys/animeman/integrations/nyaa"
	"github.com/sonalys/animeman/integrations/qbittorrent"
	"github.com/sonalys/animeman/internal/parser"
	"github.com/sonalys/animeman/internal/utils"
)

var numberExpr = regexp.MustCompile(`\d+`)

func strSliceToInt(from []string) []int64 {
	out := make([]int64, 0, len(from))
	for _, cur := range from {
		out = append(out, utils.Must(strconv.ParseInt(cur, 10, 64)))
	}
	return out
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func compareTags(a, b string) int {
	if a == "" && b != "" {
		return -1
	}
	if a != "" && b == "" {
		return 1
	}
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

func getLatestTag(torrents ...qbittorrent.Torrent) string {
	var latestTag string
	for _, torrent := range torrents {
		tags := torrent.GetTags()
		seasonEpisodeTag := tags[len(tags)-1]
		if compareTags(seasonEpisodeTag, latestTag) > 0 {
			latestTag = seasonEpisodeTag
		}
	}
	return latestTag
}

func (c *Controller) GetLatestTag(ctx context.Context, entry myanimelist.AnimeListEntry) (string, error) {
	// check if torrent already exists, if so we skip it.
	title := parser.ParseTitle(entry.Title)
	titleEng := parser.ParseTitle(entry.TitleEng)
	torrents1, err := c.dep.QB.List(ctx, qbittorrent.Tag(title.BuildSeriesTag()))
	if err != nil {
		return "", fmt.Errorf("listing torrents: %w", err)
	}
	torrents2, err := c.dep.QB.List(ctx, qbittorrent.Tag(titleEng.BuildSeriesTag()))
	if err != nil {
		return "", fmt.Errorf("listing torrents: %w", err)
	}
	return getLatestTag(append(torrents1, torrents2...)...), nil
}

func (c *Controller) DigestNyaaTorrent(ctx context.Context, entry myanimelist.AnimeListEntry, torrent nyaa.Entry) (bool, error) {
	meta := parser.ParseTitle(torrent.Title)
	if meta.IsMultiEpisode && entry.AiringStatus == myanimelist.AiringStatusAiring {
		log.Debug().Msgf("torrent dropped: multi-episode for currently airing")
		return false, nil
	}
	latestTag, err := c.GetLatestTag(ctx, entry)
	if err != nil {
		return false, fmt.Errorf("getting latest tag: %w", err)
	}
	// Check if qBittorrent client already has an episode after the current one.
	tagCompare := compareTags(meta.BuildSeasonEpisodeTag(), latestTag)
	if tagCompare <= 0 {
		return false, nil
	}
	var savePath qbittorrent.SavePath
	if c.dep.Config.CreateShowFolder {
		savePath = qbittorrent.SavePath(fmt.Sprintf("%s/%s", c.dep.Config.DownloadPath, entry.GetTitle()))
	} else {
		savePath = qbittorrent.SavePath(c.dep.Config.DownloadPath)
	}
	tags := meta.BuildTorrentTags()
	err = c.dep.QB.AddTorrent(ctx,
		tags,
		savePath,
		qbittorrent.TorrentURL{torrent.Link},
		qbittorrent.Category(c.dep.Config.Category),
	)
	if err != nil {
		return false, fmt.Errorf("adding torrents: %w", err)
	}
	log.Info().
		Str("savePath", string(savePath)).
		Strs("tag", tags).
		Msgf("torrent added")
	return true, nil
}
