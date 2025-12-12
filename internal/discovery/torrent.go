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

// findLatestTag will receive an anime list entry and return all torrents listed from the anime.
func (c *Controller) findLatestTag(ctx context.Context, entry animelist.Entry) (parser.SeasonEpisodeTag, error) {
	torrents := make([]torrentclient.Torrent, 0, 100)

	for i := range entry.Titles {
		// we should consider both metadata and titleEng, because your anime list has different titles available,
		// some torrent sources will use one, some will use the other, so to avoid duplication we check for both.
		metadata := parser.Parse(entry.Titles[i])

		resp, err := c.dep.TorrentClient.List(ctx, &torrentclient.ListTorrentConfig{
			Tag: utils.Pointer(metadata.TagBuildSeries()),
		})

		if err != nil {
			return parser.SeasonEpisodeTag{}, fmt.Errorf("listing torrents: %w", err)
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

func (c *Controller) buildTorrentName(title string, parsedNyaa parser.ParsedNyaa) string {
	var b strings.Builder

	if parsedNyaa.Meta.Source != "" {
		b.WriteString("[")
		b.WriteString(parsedNyaa.Meta.Source)
		b.WriteString("] ")
	}

	b.WriteString(title)
	b.WriteString(" ")
	b.WriteString(parsedNyaa.Meta.SeasonEpisodeTag.BuildTag())

	if parsedNyaa.Meta.VerticalResolution > 0 {
		b.WriteString(" ")
		fmt.Fprintf(&b, "[%dp]", parsedNyaa.Meta.VerticalResolution)
	}

	return b.String()
}

// AddTorrentEntry receives an anime list entry and a downloadable torrent.
// It will configure all necessary metadata and send it to your torrent client.
func (c *Controller) AddTorrentEntry(ctx context.Context, animeListEntry animelist.Entry, parsedNyaa parser.ParsedNyaa) error {
	logger := getLogger(ctx)

	selectedTitle := animeListEntry.Titles[0]

	meta := parsedNyaa.Meta.Clone()
	meta.Title = parser.StripTitle(meta.Title)
	tags := meta.TagsBuildTorrent()

	req := &torrentclient.AddTorrentConfig{
		Tags:     tags,
		URLs:     []string{parsedNyaa.Entry.Link},
		Category: c.dep.Config.Category,
		SavePath: c.TorrentGetDownloadPath(selectedTitle),
	}

	if c.dep.Config.RenameTorrent {
		req.Name = utils.Pointer(c.buildTorrentName(selectedTitle, parsedNyaa))
	}

	if err := c.dep.TorrentClient.AddTorrent(ctx, req); err != nil {
		return fmt.Errorf("adding torrents: %w", err)
	}

	logger.
		Info().
		Str("title", parsedNyaa.Entry.Title).
		Str("entry", selectedTitle).
		Str("path", req.SavePath).
		Int("detectedQuality", meta.VerticalResolution).
		Msg("torrent added")

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
		return fmt.Errorf("listing torrents: %w", err)
	}

	for _, torrent := range torrents {
		meta := parser.Parse(torrent.Name)
		meta.Title = parser.StripTitleSubtitle(meta.Title)
		tags := meta.TagsBuildTorrent()

		log.
			Info().
			Any("metadata", meta).
			Strs("tags", tags).
			Msgf("updating torrent tags")

		if err := c.dep.TorrentClient.AddTorrentTags(ctx, []string{torrent.Hash}, tags); err != nil {
			return fmt.Errorf("updating tags: %w", err)
		}
	}

	return nil
}
