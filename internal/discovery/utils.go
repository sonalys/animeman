package discovery

import (
	"strings"
	"time"

	"github.com/sonalys/animeman/integrations/nyaa"
	"github.com/sonalys/animeman/internal/parser"
	"github.com/sonalys/animeman/internal/utils"
	"github.com/sonalys/animeman/pkg/v1/animelist"
)

func filterBatchEntries(e parser.ParsedNyaa) bool { return e.Meta.SeasonEpisodeTag.IsMultiEpisode() }

// filterPublishedAfterDate creates a filter for nyaa.Entry only after a date t.
func filterPublishedAfterDate(t time.Time) func(e nyaa.Entry) bool {
	return func(e nyaa.Entry) bool {
		publishedDate := utils.Must(time.Parse(time.RFC1123Z, e.PubDate))
		return publishedDate.After(t)
	}
}

// filterTitleMatch guarantees that the main title matches for the anime list entry and the nyaa entry.
func filterTitleMatch(entry animelist.Entry) func(e nyaa.Entry) bool {
	loweredTitles := utils.Map(entry.Titles, func(title string) string { return strings.ToLower(title) })
	strippedTitles := utils.Map(entry.Titles, func(title string) string { return parser.StripTitle(title) })

	return func(nyaaEntry nyaa.Entry) bool {
		nyaaLoweredTitle := strings.ToLower(nyaaEntry.Title)
		nyaaStrippedTitle := parser.StripTitle(nyaaLoweredTitle)

		// Try to match original title first.
		for _, title := range loweredTitles {
			if strings.Contains(nyaaLoweredTitle, title) {
				return true
			}
		}

		// Try to match stripped title afterwards.
		for _, title := range strippedTitles {
			if strings.EqualFold(nyaaStrippedTitle, title) {
				return true
			}
		}

		return false
	}
}
