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
	originalTitles := utils.Map(entry.Titles, func(title string) string { return strings.ToLower(title) })
	strippedTitles := utils.Map(originalTitles, func(title string) string { return parser.StripTitle(title) })

	return func(nyaaEntry nyaa.Entry) bool {
		gotTitle := strings.ToLower(nyaaEntry.Title)
		gotStrippedTitle := parser.StripTitle(gotTitle)

		// Try to match stripped title afterwards.
		for _, strippedTitle := range strippedTitles {
			if strings.EqualFold(gotStrippedTitle, strippedTitle) {
				return true
			}
		}

		return false
	}
}
