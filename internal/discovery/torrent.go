package discovery

import (
	"context"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/sonalys/animeman/internal/parser"
	"github.com/sonalys/animeman/internal/utils"
	"github.com/sonalys/animeman/pkg/v1/animelist"
	"github.com/sonalys/animeman/pkg/v1/torrentclient"
)

// TorrentGetLatestEpisodes will receive an anime list entry and return all torrents listed from the anime.
func (c *Controller) TorrentGetLatestEpisodes(ctx context.Context, entry animelist.Entry) (string, error) {
	var torrents []torrentclient.Torrent
	for i := range entry.Titles {
		// we should consider both title and titleEng, because your anime list has different titles available,
		// some torrent sources will use one, some will use the other, so to avoid duplication we check for both.
		resp, err := c.dep.TorrentClient.List(ctx, &torrentclient.ListTorrentConfig{
			Tag: utils.Pointer(parser.TitleParse(entry.Titles[i]).TagBuildSeries()),
		})
		if err != nil {
			return "", fmt.Errorf("listing torrents: %w", err)
		}
		torrents = append(torrents, resp...)
	}
	return tagGetLatest(torrents), nil
}

// TorrentGetDownloadPath returns a torrent path, creating a show folder if configured.
func (c *Controller) TorrentGetDownloadPath(title string) (path string) {
	if c.dep.Config.CreateShowFolder {
		return fmt.Sprintf("%s/%s", c.dep.Config.DownloadPath, title)
	}
	return c.dep.Config.DownloadPath
}

func (c *Controller) buildTorrentName(entry animelist.Entry, parsedNyaa parser.ParsedNyaa) string {
	var b strings.Builder
	if parsedNyaa.Meta.Source != "" {
		b.WriteString("[")
		b.WriteString(parsedNyaa.Meta.Source)
		b.WriteString("] ")
	}
	b.WriteString(entry.Titles[0])
	b.WriteString(" ")
	b.WriteString(parsedNyaa.SeasonEpisodeTag)
	if parsedNyaa.Meta.VerticalResolution > 0 {
		b.WriteString(" ")
		b.WriteString(fmt.Sprintf("[%d]", parsedNyaa.Meta.VerticalResolution))
	}
	return b.String()
}

// TorrentDigestNyaa receives an anime list entry and a downloadable torrent.
// It will configure all necessary metadata and send it to your torrent client.
func (c *Controller) TorrentDigestNyaa(ctx context.Context, entry animelist.Entry, parsedNyaa parser.ParsedNyaa) error {
	savePath := c.TorrentGetDownloadPath(entry.Titles[0])
	tags := parsedNyaa.Meta.TagsBuildTorrent()
	req := &torrentclient.AddTorrentConfig{
		Tags:     tags,
		URLs:     []string{parsedNyaa.Entry.Link},
		Category: c.dep.Config.Category,
		SavePath: savePath,
	}
	if c.dep.Config.RenameTorrent {
		req.Name = utils.Pointer(c.buildTorrentName(entry, parsedNyaa))
	}
	if err := c.dep.TorrentClient.AddTorrent(ctx, req); err != nil {
		return fmt.Errorf("adding torrents: %w", err)
	}
	log.Info().Str("savePath", string(savePath)).Strs("tag", tags).Msgf("torrent added")
	return nil
}

// TorrentRegenerateTags will scan all torrents from the configured category and update their tags.
// This function exists for when you already have a collection of Anime categorized torrents.
// This function will tag all entries from the configured category for smart episode detection and filtering.
func (c *Controller) TorrentRegenerateTags(ctx context.Context) error {
	torrents, err := c.dep.TorrentClient.List(ctx, &torrentclient.ListTorrentConfig{
		Category: &c.dep.Config.Category,
		Tag:      utils.Pointer(""),
	})
	if err != nil {
		return fmt.Errorf("listing: %w", err)
	}
	for _, torrent := range torrents {
		parsedTitle := parser.TitleParse(torrent.Name)
		tags := parsedTitle.TagsBuildTorrent()
		log.Info().Any("metadata", parsedTitle).Strs("tags", tags).Msgf("updating torrent tags")
		if err := c.dep.TorrentClient.AddTorrentTags(ctx, []string{torrent.Hash}, tags); err != nil {
			return fmt.Errorf("updating tags: %w", err)
		}
	}
	return nil
}
