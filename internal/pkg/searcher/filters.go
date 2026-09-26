package searcher

import (
	"slices"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/sonalys/animeman/internal/pkg/metadata"
	"github.com/sonalys/animeman/internal/pkg/sliceutils"
	"github.com/sonalys/animeman/internal/pkg/stringutils"
	"github.com/sonalys/animeman/internal/ports/animelist"
	"github.com/sonalys/animeman/internal/ports/torrentsource"
)

type filterStats map[string]uint

func newFilterStats() filterStats {
	return make(filterStats)
}

func (f *filterStats) Inc(name string) {
	(*f)[name]++
}

// matchSources keeps only torrents whose title contains one of the
// configured sources (release groups), mirroring how the parser extracts
// the release group.
func matchSources(
	sources []string,
) func(ignoreCounter func(string)) func(torrentsource.Torrent) bool {
	normalizedSources := sliceutils.Map(sources, strings.ToLower)

	return func(ignoreCounter func(string)) func(torrentsource.Torrent) bool {
		return func(torrent torrentsource.Torrent) bool {
			if len(sources) == 0 {
				return true
			}

			if torrent.Metadata.Group == "" {
				ignoreCounter("missingReleaseGroup")
				return false
			}

			if slices.Contains(normalizedSources, strings.ToLower(torrent.Metadata.Group)) {
				return true
			}

			ignoreCounter("releaseGroupWhitelist")

			log.
				Trace().
				Strs("sourceWhitelist", sources).
				Str("source", torrent.Metadata.Group).
				Msg("torrent source is not whitelisted")

			return false
		}
	}
}

func minSeeders(minSeeders int) func(ignoreCounter func(string)) func(torrentsource.Torrent) bool {
	return func(ignoreCounter func(string)) func(torrentsource.Torrent) bool {
		return func(torrent torrentsource.Torrent) bool {
			if torrent.Seeders < minSeeders {
				ignoreCounter("minSeeders")

				log.
					Trace().
					Int("seeders", torrent.Seeders).
					Int("minSeeders", minSeeders).
					Msg("torrent doesn't have enough seeders")

				return false
			}

			return true
		}
	}
}

func matchStartDate(
	entry animelist.Entry,
) func(ignoreCounter func(string)) func(torrentsource.Torrent) bool {
	return func(ignoreCounter func(string)) func(torrentsource.Torrent) bool {
		return func(torrent torrentsource.Torrent) bool {
			// Compares publishing date with anime start date, 2 days offset to prevent wrong timezone and hour precision.
			if !entry.StartDate.IsZero() &&
				torrent.PublishedAt.Before(entry.StartDate.AddDate(0, 0, -2)) {
				ignoreCounter("startDateMismatch")

				log.
					Trace().
					Time("publishedAt", torrent.PublishedAt).
					Time("startDate", entry.StartDate).
					Msg("torrent was published before show airing date")

				return false
			}

			return true
		}
	}
}

func matchEpisodeCount(
	entry animelist.Entry,
	sources []string,
) func(ignoreCounter func(string)) func(torrentsource.Torrent) bool {
	return func(ignoreCounter func(string)) func(torrentsource.Torrent) bool {
		return func(torrent torrentsource.Torrent) bool {
			// If ep number is greater than season ep count, should be removed.
			// This can happen when certain sources mark S2 but use absolute ep number, so they start like S2E13 instead of S2E01.
			// If there's only a single source, then this won't be a problem.
			if len(sources) > 1 && entry.NumEpisodes != 0 {
				if torrent.Metadata.Tags.LastEpisode() > float32(entry.NumEpisodes) {
					ignoreCounter("epMismatch")

					log.
						Trace().
						Float32("episode", torrent.Metadata.Tags.LastEpisode()).
						Int("totalEpisodes", entry.NumEpisodes).
						Msg("torrent detected episode tag is above the season total episode count")

					return false
				}
			}

			return true
		}
	}
}

func newerEpisode(
	latestTag metadata.Tag,
) func(ignoreCounter func(string)) func(torrentsource.Torrent) bool {
	return func(ignoreCounter func(string)) func(torrentsource.Torrent) bool {
		return func(torrent torrentsource.Torrent) bool {
			if latestTag.IsZero() {
				return true
			}

			if !torrent.Metadata.Tags.IsZero() &&
				torrent.Metadata.Tags.Compare(latestTag) > 0 {
				return true
			}

			ignoreCounter("oldEpisode")

			log.
				Trace().
				Stringer("latestTag", latestTag).
				Stringer("tag", torrent.Metadata.Tags).
				Msg("torrent is older than latestTag")

			return false
		}
	}
}

func matchTitlePrefix(
	titles []string,
) func(ignoreCounter func(string)) func(torrentsource.Torrent) bool {
	cleanedTitles := sliceutils.Map(titles, func(title string) string {
		return metadata.ParsePrimaryTitle(title)
	})

	return func(ignoreCounter func(string)) func(torrentsource.Torrent) bool {
		return func(torrent torrentsource.Torrent) bool {
			for _, title := range cleanedTitles {
				if stringutils.MatchPrefixFlexible(
					torrent.Metadata.PrimaryTitle,
					title,
					torrentsource.IgnoreCharset,
				) {
					return true
				}
			}

			log.
				Trace().
				Strs("titles", cleanedTitles).
				Str("title", torrent.Metadata.PrimaryTitle).
				Msg("torrent primary title did not match any of the entry titles")

			ignoreCounter("titlePrefix")

			return false
		}
	}
}
