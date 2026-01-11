package discovery

import (
	"regexp"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/sonalys/animeman/internal/integrations/nyaa"
	"github.com/sonalys/animeman/internal/parser"
	"github.com/sonalys/animeman/internal/utils"
	"github.com/sonalys/animeman/pkg/v1/animelist"
)

func filterBatchEntries(e parser.ParsedNyaa) bool { return e.ExtractedMetadata.Tag.IsMultiEpisode() }

// filterMetadata ensures that only coherent and expected nyaa entries are considered for donwload.
// This function avoids download unrelated torrents.
func filterMetadata(entry animelist.Entry) func(e nyaa.Entry) bool {
	strippedTitles := utils.Map(entry.Titles, func(title string) string { return parser.StripTitle(title) })

	return func(nyaaEntry nyaa.Entry) bool {
		publishedDate := utils.Must(time.Parse(time.RFC1123Z, nyaaEntry.PubDate))

		// Compares publishing date with anime start date, 2 days offset to prevent wrong timezone and hour precision.
		if publishedDate.Before(entry.StartDate.AddDate(0, 0, -2)) {
			log.
				Trace().
				Time("publishedDate", publishedDate).
				Time("startDate", entry.StartDate).
				Msg("discarding torrent candidate due to mismatch in publishedDate and startDate")

			return false
		}

		meta := parser.Parse(nyaaEntry.Title)

		// Check if nyaa entry episode is greater than the animelist episode count.
		if entry.NumEpisodes > 0 && meta.Tag.LastEpisode() > float64(entry.NumEpisodes) {
			log.
				Trace().
				Float64("lastEpisode", meta.Tag.LastEpisode()).
				Int("episodeCount", entry.NumEpisodes).
				Msg("discarding torrent candidate due to numEpisodes and metadata mismatch")

			return false
		}

		titles := append(entry.Titles, strippedTitles...)
		titles = utils.Deduplicate(titles)

		for _, originalTitle := range titles {
			if utils.MatchPrefixFlexible(meta.Title, originalTitle, ",.:`'\";-") {
				log.
					Trace().
					Str("nyaaTitle", meta.Title).
					Str("matchingTitlePrefix", originalTitle).
					Msg("torrent candidate matched original title prefix")

				return true
			}
		}

		log.
			Trace().
			Str("nyaaTitle", meta.Title).
			Strs("expectedTitlePrefixes", strippedTitles).
			Msg("discarding torrent candidate due to mismatching titles")

		return false
	}
}

func matchTitle(gotTitle, originalTitle string) bool {
	re := regexp.MustCompile(`[*\-:;.\\/ ]`)

	gotTitle = re.ReplaceAllString(gotTitle, "")
	originalTitle = re.ReplaceAllString(originalTitle, "")

	return utils.HasPrefixFold(gotTitle, originalTitle)
}
