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

// Guarantees that the main title matches for the anime list entry and the nyaa entry.
func filterTitleMatch(alEntry animelist.Entry) func(e nyaa.Entry) bool {
	parsedTitles := utils.Map(alEntry.Titles, func(title string) string {
		return parser.Parse(title).Title
	})
	return func(e nyaa.Entry) bool {
		meta := parser.Parse(e.Title)

		for _, parsedTitle := range parsedTitles {
			if strings.EqualFold(meta.Title, parsedTitle) {
				return true
			}
		}

		return false
	}
}
