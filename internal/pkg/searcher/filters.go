package searcher

import (
	"slices"
	"strings"

	"github.com/sonalys/animeman/internal/pkg/sliceutils"
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

			if torrent.Metadata.ReleaseGroup == "" {
				ignoreCounter("missingReleaseGroup")
				return false
			}

			if slices.Contains(normalizedSources, strings.ToLower(torrent.Metadata.ReleaseGroup)) {
				return true
			}

			ignoreCounter("releaseGroupWhitelist")

			return false
		}
	}
}

func minSeeders(minSeeders int) func(ignoreCounter func(string)) func(torrentsource.Torrent) bool {
	return func(ignoreCounter func(string)) func(torrentsource.Torrent) bool {
		return func(torrent torrentsource.Torrent) bool {
			if torrent.Seeders < minSeeders {
				ignoreCounter("minSeeders")

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
				if torrent.Metadata.Tag.FirstEpisode() > float64(entry.NumEpisodes) {
					ignoreCounter("epMismatch")
					return false
				}
			}

			return true
		}
	}
}
