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
	logger := getLogger(ctx)
	torrents := make([]torrentclient.Torrent, 0, 100)

	for _, title := range entry.Titles {
		req := &torrentclient.ListTorrentConfig{
			Tag: utils.Pointer(parser.BuildTitleTag(title)),
		}
		resp, err := c.dep.TorrentClient.List(ctx, req)

		if len(resp) == 0 {
			continue
		}

		logger.
			Trace().
			Str("tag", *req.Tag).
			Msg("identified entry tag on torrent client")

		if err != nil {
			return parser.SeasonEpisodeTag{}, fmt.Errorf("listing torrents: %w", err)
		}

		torrents = append(torrents, resp...)
	}

	latestTag := getLatestTag(torrents)
	if !latestTag.IsZero() {
		logger.
			Debug().
			Str("latestTag", latestTag.BuildTag()).
			Msg("identified latest tag on torrent client")
	}

	return latestTag, nil
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

// selectIdealTitle avoids kanji titles for example, preferring english ones.
func selectIdealTitle(titles []string) string {
	if len(titles) == 0 {
		return ""
	}

	for _, title := range titles {
		if strings.ContainsFunc(strings.ToLower(title), func(r rune) bool {
			return r >= 'a' && r <= 'z'
		}) {
			return title
		}
	}

	return titles[0]
}

// AddTorrentEntry receives an anime list entry and a downloadable torrent.
// It will configure all necessary metadata and send it to your torrent client.
func (c *Controller) AddTorrentEntry(ctx context.Context, animeListEntry animelist.Entry, parsedNyaa parser.ParsedNyaa) error {
	logger := getLogger(ctx)

	selectedTitle := selectIdealTitle(animeListEntry.Titles)

	meta := parsedNyaa.Meta.Clone()
	// Use nyaa metadata, but with anime list title.
	// This behavior avoids different sources creating different tags and downloading the same episode twice.
	meta.Title = selectedTitle
	tags := meta.BuildTorrentTags()

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
		tags := meta.BuildTorrentTags()

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
