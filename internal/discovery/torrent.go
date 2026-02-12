package discovery

import (
	"context"
	"fmt"
	"strings"
	"unicode"

	"github.com/rs/zerolog/log"
	"github.com/sonalys/animeman/internal/parser"
	"github.com/sonalys/animeman/internal/tags"
	"github.com/sonalys/animeman/internal/utils"
	"github.com/sonalys/animeman/pkg/v1/animelist"
	"github.com/sonalys/animeman/pkg/v1/torrentclient"
)

// findLatestTag will receive an anime list entry and return all torrents listed from the anime.
func (c *Controller) findLatestTag(ctx context.Context, entry animelist.Entry) (tags.Tag, error) {
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
			return tags.Tag{}, fmt.Errorf("listing torrents: %w", err)
		}

		torrents = append(torrents, resp...)
	}

	latestTag := getLatestTag(torrents)
	if !latestTag.IsZero() {
		logger.
			Debug().
			Str("latestTag", latestTag.String()).
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

	if parsedNyaa.ExtractedMetadata.Source != "" {
		b.WriteString("[")
		b.WriteString(parsedNyaa.ExtractedMetadata.Source)
		b.WriteString("] ")
	}

	b.WriteString(title)

	tag := parsedNyaa.ExtractedMetadata.Tag

	// Avoid printing S1 on titles, since lots of shows and movies dont require this notation.
	if tag.LastEpisode() > 0 {
		b.WriteString(" ")
		b.WriteString(tag.String())
	}

	if parsedNyaa.ExtractedMetadata.VerticalResolution > 0 {
		b.WriteString(" ")
		fmt.Fprintf(&b, "[%dp]", parsedNyaa.ExtractedMetadata.VerticalResolution)
	}

	return b.String()
}

// selectIdealTitle avoids kanji titles for example, preferring english ones.
func selectIdealTitle(titles []string) string {
	if len(titles) == 0 {
		return ""
	}

	for _, t := range titles {
		if isASCII(t) {
			return t
		}
	}

	// Fallback to first element if no ASCII title is found
	return titles[0]
}

func isASCII(s string) bool {
	for _, c := range s {
		if c > unicode.MaxASCII {
			return false
		}
	}
	return true
}

// AddTorrentEntry receives an anime list entry and a downloadable torrent.
// It will configure all necessary metadata and send it to your torrent client.
func (c *Controller) AddTorrentEntry(ctx context.Context, animeListEntry animelist.Entry, parsedNyaa parser.ParsedNyaa) error {
	selectedTitle := selectIdealTitle(animeListEntry.Titles)

	meta := parsedNyaa.ExtractedMetadata.Clone()
	// Use nyaa metadata, but with anime list title.
	// This behavior avoids different sources creating different tags and downloading the same episode twice.
	meta.Title = selectedTitle
	tags := meta.BuildTorrentTags()

	req := &torrentclient.AddTorrentConfig{
		Tags:     tags,
		URLs:     []string{parsedNyaa.NyaaTorrent.Link},
		Category: c.dep.Config.Category,
		SavePath: c.TorrentGetDownloadPath(selectedTitle),
	}

	if c.dep.Config.RenameTorrent {
		req.Name = utils.Pointer(c.buildTorrentName(selectedTitle, parsedNyaa))
	}

	if err := c.dep.TorrentClient.AddTorrent(ctx, req); err != nil {
		return fmt.Errorf("adding torrents: %w", err)
	}

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
		meta := parser.Parse(torrent.Name, 1)
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
