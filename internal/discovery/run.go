package discovery

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/sonalys/animeman/integrations/nyaa"
	"github.com/sonalys/animeman/internal/parser"
	"github.com/sonalys/animeman/internal/utils"
	"github.com/sonalys/animeman/pkg/v1/animelist"
	"github.com/sonalys/animeman/pkg/v1/torrentclient"
)

// RunDiscovery controls the discovery routine,
// fetching entries from your anime list and looking for updates in Nyaa.si
// After finding updates, it will verify episode collision and dispatch it to your torrent client.
func (c *Controller) RunDiscovery(ctx context.Context) error {
	if err := c.TorrentRegenerateTags(ctx); err != nil {
		return fmt.Errorf("updating qBittorrent entries: %w", err)
	}
	log.Debug().Msgf("discovery started")
	entries, err := c.dep.AnimeListClient.GetCurrentlyWatching(ctx)
	if err != nil {
		return fmt.Errorf("fetching anime list: %w", err)
	}
	for _, entry := range entries {
		logger := log.With().Any("titles", entry.Titles).Logger()
		ctx := logger.WithContext(ctx)
		if err := c.DigestAnimeListEntry(ctx, entry); errors.Is(err, torrentclient.ErrUnauthorized) || errors.Is(err, context.Canceled) {
			return fmt.Errorf("failed to digest entry: %w", err)
		}
	}
	return nil
}

// episodeFilterNew will only return ParsedNyaa entries that are more recent than the given latestTag.
// excludeBatch is used when a show is airing or you have already downloaded some episodes of the season.
// excludeBatch avoids downloading a batch for episodes which you already have.
func episodeFilterNew(list []parser.ParsedNyaa, latestTag string, excludeBatch bool) []parser.ParsedNyaa {
	out := make([]parser.ParsedNyaa, 0, len(list))
	for _, nyaaEntry := range list {
		if excludeBatch && nyaaEntry.Meta.IsMultiEpisode {
			continue
		}
		// Make sure we only add episodes ahead of the current ones in the qBittorrent.
		// <= 0 is used to ensure we don't download the same episode multiple times, or older episodes.
		if tagCompare(nyaaEntry.SeasonEpisodeTag, latestTag) <= 0 {
			continue
		}
		// Some providers count episodes from season 1, some from season 2, example:
		// s02e19 when it should be s02e08. so we add continuity to download only next episode.
		if latestTag != "" && !nyaaEntry.Meta.TagIsNextEpisode(latestTag) {
			continue
		}
		latestTag = nyaaEntry.SeasonEpisodeTag
		out = append(out, nyaaEntry)
	}
	return out
}

var notWordDigitOrSpace = regexp.MustCompile("[^a-zA-Z 0-9]")

// calculateTitleSimilarityScore returns a value between 0 and 1 for how similar the titles are.
func calculateTitleSimilarityScore(originalTitle, title string) float64 {
	originalTitle = strings.ToLower(originalTitle)
	title = strings.ToLower(title)
	originalTitle = notWordDigitOrSpace.ReplaceAllString(originalTitle, "")
	title = notWordDigitOrSpace.ReplaceAllString(title, "")

	originalTitleWords := strings.Split(originalTitle, " ")
	titleWords := strings.Split(title, " ")
	wordCount := len(titleWords)

	var match int
outer:
	for _, curWord := range titleWords {
		for i, target := range originalTitleWords {
			if curWord == target {
				match++
				originalTitleWords = append(originalTitleWords[:i], originalTitleWords[i+1:]...)
				continue outer
			}
		}
	}
	return float64(match) / float64(wordCount)
}

// parseAndSort will digest the raw data from Nyaa into a parsed metadata struct `ParsedNyaa`.
// it will also sort the response by season and episode.
// it's important it returns a crescent season/episode list, so you don't download a recent episode and
// don't download the oldest ones in case you don't have all episodes since your latestTag.
func parseAndSort(animeListEntry animelist.Entry, entries []nyaa.Entry) []parser.ParsedNyaa {
	resp := utils.Map(entries, func(entry nyaa.Entry) parser.ParsedNyaa { return parser.NewParsedNyaa(entry) })
	sort.Slice(resp, func(i, j int) bool {
		cmp := tagCompare(resp[i].SeasonEpisodeTag, resp[j].SeasonEpisodeTag)
		if cmp != 0 {
			return cmp < 0
		}
		// For same tag, we compare vertical resolution, prioritizing better quality.
		cmp = resp[j].Meta.VerticalResolution - resp[i].Meta.VerticalResolution
		if cmp != 0 {
			return cmp < 0
		}
		var scoreI, scoreJ float64
		// Then we prioritize by title proximity score.
		for _, title := range animeListEntry.Titles {
			curScoreI := calculateTitleSimilarityScore(title, resp[i].Meta.Title)
			curScoreJ := calculateTitleSimilarityScore(title, resp[j].Meta.Title)
			if curScoreI > scoreI {
				scoreI = curScoreI
			}
			if curScoreJ > scoreJ {
				scoreJ = curScoreJ
			}
		}
		cmp = int((scoreJ - scoreI) * 100)
		if cmp != 0 {
			return cmp < 0
		}
		// Then prioritize number of seeds
		cmp = resp[j].Entry.Seeders - resp[i].Entry.Seeders
		if cmp != 0 {
			return cmp < 0
		}
		return cmp < 0
	})
	return resp
}

// getDownloadableEntries is responsible for filtering and ordering the raw Nyaa feed into valid downloadable torrents.
func getDownloadableEntries(
	animeListEntry animelist.Entry,
	entries []nyaa.Entry,
	latestTag string,
	animeStatus animelist.AiringStatus,
) []parser.ParsedNyaa {
	// If we don't have any episodes, and show is released, try to find a batch for all episodes.
	useBatch := latestTag == "" && animeStatus == animelist.AiringStatusAired
	parsedEntries := parseAndSort(animeListEntry, entries)
	if useBatch {
		log.Debug().Msg("anime is already aired, no downloaded entries. activating batch search")
		return utils.Filter(parsedEntries, filterBatchEntries)
	}
	return episodeFilterNew(parsedEntries, latestTag, !useBatch)
}

func (c *Controller) NyaaSearch(ctx context.Context, entry animelist.Entry) ([]nyaa.Entry, error) {
	logger := zerolog.Ctx(ctx)
	// Build search query for Nyaa.
	// For title we filter for english and original titles.
	strippedTitles := utils.Map(entry.Titles, func(title string) string { return parser.TitleStrip(title) })
	titleQuery := nyaa.QueryOr(strippedTitles)
	sourceQuery := nyaa.QueryOr(c.dep.Config.Sources)
	qualityQuery := nyaa.QueryOr(c.dep.Config.Qualitites)
	entries, err := c.dep.NYAA.List(ctx, titleQuery, sourceQuery, qualityQuery)
	logger.Debug().Int("count", len(entries)).Msg("found nyaa results for entry")
	if err != nil {
		return nil, fmt.Errorf("getting nyaa list: %w", err)
	}
	// Filters only entries after the anime started airing.
	viableResults := utils.Filter(entries,
		filterPublishedAfterDate(entry.StartDate),
		filterTitleMatch(entry),
	)
	logger.Debug().Int("count", len(entries)).Msg("viable nyaa entries")
	return viableResults, nil
}

// DigestAnimeListEntry receives an anime list entry and fetches the anime feed, looking for new content.
func (c *Controller) DigestAnimeListEntry(ctx context.Context, entry animelist.Entry) (err error) {
	logger := zerolog.Ctx(ctx)

	nyaaEntries, err := c.NyaaSearch(ctx, entry)
	// There should always be torrents for entries, if there aren't we can just exit the routine.
	if len(nyaaEntries) == 0 {
		logger.Debug().Any("title", entry.Titles).Msg("no nyaa entries found")
		return
	}
	latestTag, err := c.TorrentGetLatestEpisodes(ctx, entry)
	if err != nil {
		return fmt.Errorf("getting latest tag: %w", err)
	}
	*logger = logger.With().Str("latestTag", latestTag).Logger()

	logger.Debug().Msg("looking for torrent candidates")
	candidates := getDownloadableEntries(entry, nyaaEntries, latestTag, entry.AiringStatus)
	if len(candidates) == 0 {
		logger.Debug().Msg("no torrent candidates found")
		return
	}
	for _, nyaaEntry := range candidates {
		if err := c.TorrentDigestNyaa(ctx, entry, nyaaEntry); err != nil {
			logger.Error().Msgf("failed to digest nyaa entry: %s", err)
			continue
		}
	}
	logger.Debug().Msg("discovery finished")
	return
}
