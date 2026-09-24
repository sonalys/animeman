package discovery

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/expr-lang/expr"
	"github.com/rs/zerolog/log"
	"github.com/sonalys/animeman/internal/pkg/metadata"
	"github.com/sonalys/animeman/internal/pkg/sliceutils"
	"github.com/sonalys/animeman/internal/pkg/stringutils"
	"github.com/sonalys/animeman/internal/ports/animelist"
	"github.com/sonalys/animeman/internal/ports/torrentclient"
	"github.com/sonalys/animeman/internal/ports/torrentsource"
)

// getLatestDownloadedTag returns the latest downloaded tag for a given anime in the torrent client.
func (c *Controller) getLatestDownloadedTag(
	ctx context.Context,
	entry animelist.Entry,
) (metadata.Tag, error) {
	logger := getLogger(ctx)
	torrents := make([]torrentclient.Torrent, 0, 100)

	cleanedTitles := sliceutils.Map(entry.Titles, func(title string) string {
		metadata := metadata.Parse(title, 1, nil)
		return metadata.PrimaryTitle
	})

	for _, title := range cleanedTitles {
		req := &torrentclient.ListTorrentConfig{
			Tag: new(buildTitleTag(title)),
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
			return metadata.Tag{}, fmt.Errorf("listing torrents: %w", err)
		}

		torrents = append(torrents, resp...)
	}

	latestTag := getLatestTag(torrents)
	if !latestTag.IsZero() {
		logger.
			Debug().
			Str("latestTag", latestTag.String()).
			Int("torrents", len(torrents)).
			Msg("identified latest tag on torrent client")
	} else {
		logger.
			Debug().
			Strs("titles", cleanedTitles).
			Int("torrents", len(torrents)).
			Msg("no latest tag found on torrent client")
	}

	return latestTag, nil
}

// buildTorrentDownloadPath returns a torrent path, creating a show folder if configured.
func (c *Controller) buildTorrentDownloadPath(title string) (path string) {
	if c.dep.Config.CreateShowFolder {
		return fmt.Sprintf("%s/%s", c.dep.Config.DownloadPath, title)
	}
	return c.dep.Config.DownloadPath
}

// buildTorrentName returns a torrent name based on the configured rename format and the parsed nyaa metadata.
// renameFormat is an expr-lang script for building the torrent name.
func (c *Controller) buildTorrentName(title string, torrent torrentsource.Torrent) string {
	var b strings.Builder

	env := map[string]any{
		"format": func(format string, input any) string {
			// If input is zero value or empty, returns empty string.
			valueOf := reflect.ValueOf(input)

			switch valueOf.Kind() {
			case reflect.Slice, reflect.Array, reflect.Map, reflect.String:
				if valueOf.Len() == 0 {
					return ""
				}
			case reflect.Pointer, reflect.Interface:
				if valueOf.IsNil() {
					return ""
				}
			default:
				if valueOf.IsZero() {
					return ""
				}
			}

			return fmt.Sprintf(format, input)
		},
		"title":              title,
		"releaseGroup":       torrent.Metadata.Group,
		"labels":             torrent.Metadata.Labels,
		"tag":                torrent.Metadata.Tags,
		"verticalResolution": torrent.Metadata.Resolutions.Highest(),
	}

	outputName, err := expr.Run(c.dep.Config.RenameFormat, env)
	if err != nil || outputName == "" {
		return torrent.Title
	}

	fmt.Fprintf(&b, "%v", outputName)

	return b.String()
}

func filterAlphanumeric(s string) string {
	var result strings.Builder
	result.Grow(len(s))
	for i := 0; i < len(s); i++ {
		b := s[i]
		if ('a' <= b && b <= 'z') || ('A' <= b && b <= 'Z') || ('0' <= b && b <= '9') || b == ' ' {
			result.WriteByte(b)
		}
	}
	return result.String()
}

func buildTitleTag(title string) string {
	return "!" + strings.ToLower(filterAlphanumeric(title))
}

func buildTorrentTags(title string, metadata metadata.Metadata) []string {
	return []string{
		buildTitleTag(title),
		metadata.Tags.String(),
	}
}

// AddTorrentEntry receives an anime list entry and a downloadable torrent.
// It will configure all necessary metadata and send it to your torrent client.
// It returns the torrent's info hash, empty when the source didn't provide one.
func (c *Controller) AddTorrentEntry(
	ctx context.Context,
	animeListEntry animelist.Entry,
	torrent torrentsource.Torrent,
) (string, error) {
	selectedTitle := animeListEntry.GetBestTitle()

	req := &torrentclient.AddTorrentConfig{
		Tags:     buildTorrentTags(selectedTitle, torrent.Metadata),
		URLs:     []string{torrent.Link},
		Category: c.dep.Config.Category,
		SavePath: c.buildTorrentDownloadPath(selectedTitle),
	}

	if c.dep.Config.RenameTorrent {
		req.Name = new(c.buildTorrentName(selectedTitle, torrent))
	}

	if err := c.dep.TorrentClient.AddTorrent(ctx, req); err != nil {
		return "", fmt.Errorf("adding torrents: %w", err)
	}

	log.
		Ctx(ctx).
		Debug().
		Str("name", torrent.Title).
		Msg("added torrent")

	return torrent.Hash, nil
}

// TorrentRegenerateTags will scan all torrents from the configured category and update their tags.
// This function exists for when you already have a collection of Anime categorized torrents.
// This function will tag all entries from the configured category for smart episode detection and filtering.
// entries are the anime list entries, used to normalize torrent names which were added
// under an alternative title back to the expected title tag.
func (c *Controller) TorrentRegenerateTags(ctx context.Context, entries []animelist.Entry) error {
	torrents, err := c.dep.TorrentClient.List(ctx, &torrentclient.ListTorrentConfig{
		Category: &c.dep.Config.Category,
		Tag:      new(""),
	})
	if err != nil {
		return fmt.Errorf("listing torrents: %w", err)
	}

	for _, torrent := range torrents {
		metadata := metadata.Parse(torrent.Name, 1, nil)

		// The torrent name might be based on an alternative title (e.g. "Show Name: Second Season").
		// Normalize it back to the anime list title so tag-based latest episode detection works.
		metadata.PrimaryTitle = normalizeTitle(metadata.PrimaryTitle, entries)

		if err := c.dep.TorrentClient.AddTorrentTags(
			ctx,
			[]string{torrent.Hash},
			buildTorrentTags(metadata.PrimaryTitle, metadata),
		); err != nil {
			return fmt.Errorf("updating tags: %w", err)
		}

		log.
			Ctx(ctx).
			Info().
			Str("torrentName", torrent.Name).
			Strs("tags", metadata.Labels).
			Msgf("generated torrent tags")
	}

	return nil
}

// normalizeTitle matches a torrent title against the anime list entries,
// returning the expected title from the closest matching entry.
// It handles alternative titles like "Show Name: Second Season" which would
// otherwise produce a series tag that never matches the anime list titles.
// If no entry matches with enough confidence, the original title is returned.
func normalizeTitle(torrentTitle string, entries []animelist.Entry) string {
	if len(entries) == 0 {
		return torrentTitle
	}

	const minSimilarity = 0.7

	for _, entry := range entries {
		bestScore := 0.0
		for _, title := range entry.Titles {
			score := stringutils.CalculateTextSimilarity(
				title,
				torrentTitle,
				torrentsource.IgnoreCharset,
			)
			bestScore = max(bestScore, score)
		}

		if bestScore >= minSimilarity {
			return entry.GetBestTitle()
		}
	}

	return torrentTitle
}
