package discovery

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

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

	log.
		Debug().
		Msgf("discovery started")

	ctx = log.Logger.WithContext(ctx)

	if err := c.TorrentRegenerateTags(ctx); err != nil {
		return fmt.Errorf("updating qBittorrent entries: %w", err)
	}

	entries, err := c.dep.AnimeListClient.GetCurrentlyWatching(ctx)
	if err != nil {
		return fmt.Errorf("fetching anime list: %w", err)
	}

	for _, entry := range entries {
		if !entry.StartDate.IsZero() && entry.StartDate.After(time.Now()) {
			log.
				Debug().
				Strs("title", entry.Titles).
				Msgf("skipping entry as it hasn't aired yet")
			continue
		}

		log.
			Debug().
			Strs("title", entry.Titles).
			Msgf("processing entry")

		if err := c.DiscoverEntry(ctx, entry); errors.Is(err, torrentclient.ErrUnauthorized) || errors.Is(err, context.Canceled) {
			return fmt.Errorf("failed to digest entry: %w", err)
		}
	}

	log.
		Debug().
		Int("entries", len(entries)).
		Dur("duration", time.Since(t1)).
		Msg("discovery finished")
	return nil
}

// filterNewEpisodes will only return ParsedNyaa entries that are more recent than the given latestTag.
// excludeBatch is used when a show is airing or you have already downloaded some episodes of the season.
// excludeBatch avoids downloading a batch for episodes which you already have.
func filterNewEpisodes(list []parser.ParsedNyaa, latestTag parser.SeasonEpisodeTag) []parser.ParsedNyaa {
	out := make([]parser.ParsedNyaa, 0, len(list))

	currentTag := latestTag

	for _, nyaaEntry := range list {
		// Avoid re-downloading episodes we already have, on batches.
		if !latestTag.IsZero() && nyaaEntry.Meta.SeasonEpisodeTag.IsMultiEpisode() {
			continue
		}

		// Make sure we only add episodes ahead of the current ones in the qBittorrent.
		// <= 0 is used to ensure we don't download the same episode multiple times, or older episodes.
		if tagCompare(nyaaEntry.Meta.SeasonEpisodeTag, currentTag) <= 0 {
			continue
		}

		// Some providers count episodes from season 1, some from season 2, example:
		// s02e19 when it should be s02e08. so we add continuity to download only next episode.
		if currentTag.IsZero() && !currentTag.Before(nyaaEntry.Meta.SeasonEpisodeTag) {
			continue
		}

		currentTag = nyaaEntry.Meta.SeasonEpisodeTag
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
		cmp := tagCompare(first.Meta.SeasonEpisodeTag, second.Meta.SeasonEpisodeTag)
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
	latestTag parser.SeasonEpisodeTag,
	animeStatus animelist.AiringStatus,
) []parser.ParsedNyaa {
	parsedEntries := parseAndSortResults(animeListEntry, entries)
	newEpisodes := filterNewEpisodes(parsedEntries, latestTag)

	batchOnly := latestTag.IsZero() && animeStatus == animelist.AiringStatusAired

	if batchOnly {
		newEpisodes = utils.Filter(newEpisodes, filterBatchEntries)
	}

	log.
		Trace().
		Int("newEpisodes", len(newEpisodes)).
		Bool("batchOnly", batchOnly).
		Msg("filtered viable torrent candidates for new episodes")

	return newEpisodes
}

func (c *Controller) NyaaSearch(ctx context.Context, entry animelist.Entry) ([]nyaa.Entry, error) {
	logger := getLogger(ctx)

	// Build search query for Nyaa.
	// For title we filter for english and original titles.
	strippedTitles := utils.Map(entry.Titles, func(title string) string { return strings.ToLower(parser.StripTitle(title)) })
	titleQuery := nyaa.QueryOr(utils.Deduplicate(strippedTitles))

	sourceQuery := nyaa.QueryOr(c.dep.Config.Sources)
	qualityQuery := nyaa.QueryOr(c.dep.Config.Qualitites)

	entries, err := c.dep.NYAA.List(ctx, titleQuery, sourceQuery, qualityQuery)
	if err != nil {
		return nil, fmt.Errorf("getting nyaa list: %w", err)
	}

	logger.
		Debug().
		Int("count", len(entries)).
		Msg("searched nyaa for torrent candidates")

	if len(entries) == 0 {
		return nil, nil
	}

	entries = utils.Filter(entries,
		filterMetadata(entry),
	)

	if len(entries) == 0 {
		logger.
			Debug().
			Strs("titles", strippedTitles).
			Msg("no nyaa result matching title filter")
	}

	return entries, nil
}

// DiscoverEntry receives an anime list entry and fetches the anime feed, looking for new content.
func (c *Controller) DiscoverEntry(ctx context.Context, entry animelist.Entry) error {
	logger := getLogger(ctx)

	torrentResults, err := c.NyaaSearch(ctx, entry)
	if err != nil {
		return fmt.Errorf("searching torrent for anime: %w", err)
	}

	if len(torrentResults) == 0 {
		return nil
	}

	latestTag, err := c.findLatestTag(ctx, entry)
	if err != nil {
		return fmt.Errorf("finding latest anime season episode tag: %w", err)
	}

	episodesTorrents := filterEpisodes(entry, torrentResults, latestTag, entry.AiringStatus)

	if len(episodesTorrents) == 0 {
		logger.
			Debug().
			Int("searchResults", len(torrentResults)).
			Str("latestTag", latestTag.BuildTag()).
			Msg("no new episodes identified")

		return nil
	}

	for _, episodeTorrent := range episodesTorrents {
		if err := c.AddTorrentEntry(ctx, entry, episodeTorrent); err != nil {
			return fmt.Errorf("adding torrent to client: %w", err)
		}
	}

	return nil
}
