package discovery

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

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
	t1 := time.Now()
	log.Debug().Msgf("discovery started")

	if err := c.TorrentRegenerateTags(ctx); err != nil {
		return fmt.Errorf("updating qBittorrent entries: %w", err)
	}

	entries, err := c.dep.AnimeListClient.GetCurrentlyWatching(ctx)
	if err != nil {
		return fmt.Errorf("fetching anime list: %w", err)
	}

	for _, entry := range entries {
		logger := log.With().Any("titles", entry.Titles).Logger()
		ctx := logger.WithContext(ctx)
		if err := c.DiscoverEntry(ctx, entry); errors.Is(err, torrentclient.ErrUnauthorized) || errors.Is(err, context.Canceled) {
			return fmt.Errorf("failed to digest entry: %w", err)
		}
	}

	log.Debug().Int("entries", len(entries)).Dur("duration", time.Since(t1)).Msg("discovery finished")
	return nil
}

// filterNewEpisodes will only return ParsedNyaa entries that are more recent than the given latestTag.
// excludeBatch is used when a show is airing or you have already downloaded some episodes of the season.
// excludeBatch avoids downloading a batch for episodes which you already have.
func filterNewEpisodes(list []parser.ParsedNyaa, latestTag string, excludeBatch bool) []parser.ParsedNyaa {
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

// parseAndSortResults will digest the raw data from Nyaa into a parsed metadata struct `ParsedNyaa`.
// it will also sort the response by season and episode.
// it's important it returns a crescent season/episode list, so you don't download a recent episode and
// don't download the oldest ones in case you don't have all episodes since your latestTag.
func parseAndSortResults(animeListEntry animelist.Entry, entries []nyaa.Entry) []parser.ParsedNyaa {
	parsedEntries := utils.Map(entries, func(entry nyaa.Entry) parser.ParsedNyaa { return parser.NewParsedNyaa(entry) })

	smallerFunc := func(i, j int) bool {
		first := parsedEntries[i]
		second := parsedEntries[j]

		// Sort first by season/episode tag.
		cmp := tagCompare(first.SeasonEpisodeTag, second.SeasonEpisodeTag)
		if cmp != 0 {
			return cmp < 0
		}
		// Then resolution.
		cmp = second.Meta.VerticalResolution - first.Meta.VerticalResolution
		if cmp != 0 {
			return cmp < 0
		}
		// Then title similarity.
		titleSimilarityI := utils.Max(utils.Map(animeListEntry.Titles, func(curTitle string) float64 {
			return calculateTitleSimilarityScore(curTitle, first.Meta.Title)
		})...)
		titleSimilarityJ := utils.Max(utils.Map(animeListEntry.Titles, func(curTitle string) float64 {
			return calculateTitleSimilarityScore(curTitle, second.Meta.Title)
		})...)

		if titleSimilarityI != titleSimilarityJ {
			return titleSimilarityI > titleSimilarityJ
		}

		// Then prioritize number of seeds
		return first.Entry.Seeders > second.Entry.Seeders
	}

	sort.Slice(parsedEntries, smallerFunc)

	return parsedEntries
}

// filterEpisodes is responsible for filtering and ordering the raw Nyaa feed into valid downloadable torrents.
func filterEpisodes(
	animeListEntry animelist.Entry,
	entries []nyaa.Entry,
	latestTag string,
	animeStatus animelist.AiringStatus,
) []parser.ParsedNyaa {
	// If we don't have any episodes, and show is released, try to find a batch for all episodes.
	useBatch := latestTag == "" && animeStatus == animelist.AiringStatusAired
	parsedEntries := parseAndSortResults(animeListEntry, entries)
	if useBatch {
		log.Debug().Msg("anime is already aired, no downloaded entries. activating batch search")
		return utils.Filter(parsedEntries, filterBatchEntries)
	}
	return filterNewEpisodes(parsedEntries, latestTag, !useBatch)
}

func (c *Controller) NyaaSearch(ctx context.Context, entry animelist.Entry) ([]nyaa.Entry, error) {
	logger := zerolog.Ctx(ctx)
	// Build search query for Nyaa.
	// For title we filter for english and original titles.
	strippedTitles := utils.Map(entry.Titles, func(title string) string { return parser.StripTitle(title) })
	titleQuery := nyaa.QueryOr(strippedTitles)
	sourceQuery := nyaa.QueryOr(c.dep.Config.Sources)
	qualityQuery := nyaa.QueryOr(c.dep.Config.Qualitites)

	entries, err := c.dep.NYAA.List(ctx, titleQuery, sourceQuery, qualityQuery)
	if err != nil {
		return nil, fmt.Errorf("getting nyaa list: %w", err)
	}
	logger.Debug().Int("count", len(entries)).Msg("found nyaa results for entry")
	// Filters only entries after the anime started airing.
	viableResults := utils.Filter(entries,
		filterPublishedAfterDate(entry.StartDate),
		filterTitleMatch(entry),
	)
	if len(viableResults) > 0 {
		logger.Debug().Int("count", len(viableResults)).Msg("found viable nyaa entries")
	}
	return viableResults, nil
}

// DiscoverEntry receives an anime list entry and fetches the anime feed, looking for new content.
func (c *Controller) DiscoverEntry(ctx context.Context, anime animelist.Entry) (err error) {
	logger := zerolog.Ctx(ctx)

	torrentResults, err := c.NyaaSearch(ctx, anime)
	if err != nil {
		return fmt.Errorf("searching torrent for anime: %w", err)
	}

	if len(torrentResults) == 0 {
		logger.Debug().Any("title", anime.Titles).Msg("no nyaa entries found")
		return
	}

	latestTag, err := c.findLatestTag(ctx, anime)
	if err != nil {
		return fmt.Errorf("finding latest anime season episode tag: %w", err)
	}
	*logger = logger.With().Str("latestTag", latestTag).Logger()

	logger.Debug().Msg("looking for torrent candidates")
	episodesTorrents := filterEpisodes(anime, torrentResults, latestTag, anime.AiringStatus)

	if len(episodesTorrents) == 0 {
		logger.Debug().Msg("no new episodes found")
		return
	}

	for _, episodeTorrent := range episodesTorrents {
		if err := c.AddTorrentEntry(ctx, anime, episodeTorrent); err != nil {
			logger.Error().Msgf("failed to digest nyaa entry: %s", err)
			continue
		}
	}

	return
}
