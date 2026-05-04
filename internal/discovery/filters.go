package discovery

import (
	"time"

	"github.com/sonalys/animeman/internal/integrations/nyaa"
	"github.com/sonalys/animeman/internal/parser"
	"github.com/sonalys/animeman/internal/utils"
	"github.com/sonalys/animeman/pkg/v1/animelist"
)

const ignoreCharset = " \t!,.:`'\"/\\;-[](){}*【】"

// filterMetadata ensures that only coherent and expected nyaa entries are considered for donwload.
// This function avoids download unrelated torrents.
func filterMetadata(
	entry animelist.Entry,
	filterData *FilterData,
) func(e nyaa.Item) bool {
	return func(nyaaEntry nyaa.Item) bool {
		publishedDate := utils.Must(time.Parse(time.RFC1123Z, nyaaEntry.PubDate))

		// Compares publishing date with anime start date, 2 days offset to prevent wrong timezone and hour precision.
		if publishedDate.Before(entry.StartDate.AddDate(0, 0, -2)) {
			filterData.DiscardReason[DiscardReasonPublishedDateMismatch]++

			return false
		}

		meta := parser.Parse(nyaaEntry.Title, 1)

		// Check if nyaa entry episode is greater than the animelist episode count.
		if entry.NumEpisodes > 0 && meta.Tag.LastEpisode() > float64(entry.NumEpisodes) {
			filterData.DiscardReason[DiscardReasonEpisodeCountMismatch]++

			return false
		}

		nyaaTitleWithoutTags := parser.StripTags(nyaaEntry.Title)

		for _, originalTitle := range entry.Titles {
			// Remove season information from the original title, as it is not always present in the nyaa entry.
			originalTitleWithoutSeason := parser.StripSeason(originalTitle)
			originalTitleWithoutSubtitle := parser.StripSubtitle(originalTitleWithoutSeason)

			if utils.MatchPrefixFlexible(nyaaTitleWithoutTags, originalTitleWithoutSubtitle, ignoreCharset) {
				return true
			}
		}

		filterData.DiscardReason[DiscardReasonTitleMismatch]++

		return false
	}
}
