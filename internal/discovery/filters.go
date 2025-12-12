package discovery

import (
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/sonalys/animeman/integrations/nyaa"
	"github.com/sonalys/animeman/internal/parser"
	"github.com/sonalys/animeman/internal/utils"
	"github.com/sonalys/animeman/pkg/v1/animelist"
)

func filterBatchEntries(e parser.ParsedNyaa) bool { return e.Meta.SeasonEpisodeTag.IsMultiEpisode() }

// filterMetadata ensures that only coherent and expected nyaa entries are considered for donwload.
// This function avoids download unrelated torrents.
func filterMetadata(entry animelist.Entry) func(e nyaa.Entry) bool {
	originalTitles := utils.Map(entry.Titles, func(title string) string { return strings.ToLower(title) })
	strippedTitles := utils.Map(originalTitles, func(title string) string { return parser.StripTitle(title) })

	return func(nyaaEntry nyaa.Entry) bool {
		publishedDate := utils.Must(time.Parse(time.RFC1123Z, nyaaEntry.PubDate))

		if publishedDate.Before(entry.StartDate) {
			log.
				Trace().
				Time("publishedDate", publishedDate).
				Time("startDate", entry.StartDate).
				Msg("discarding torrent candidate due to mismatch in publishedDate and startDate")

			return false
		}

		meta := parser.Parse(nyaaEntry.Title)

		// Check if nyaa entry episode is greater than the animelist episode count.
		if meta.SeasonEpisodeTag.LastEpisode() > float64(entry.NumEpisodes) {
			log.
				Trace().
				Float64("lastEpisode", meta.SeasonEpisodeTag.LastEpisode()).
				Int("episodeCount", entry.NumEpisodes).
				Msg("discarding torrent candidate due to numEpisodes and metadata mismatch")

			return false
		}

		// Try to match stripped title afterwards.
		for _, strippedTitle := range strippedTitles {
			if strings.EqualFold(meta.Title, strippedTitle) {
				return true
			}
		}

		log.
			Trace().
			Str("nyaaTitle", meta.Title).
			Strs("expectedTitles", strippedTitles).
			Msg("discarding torrent candidate due to mismatching titles")

		return false
	}
}
