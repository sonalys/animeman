package discovery

import (
	"strings"
	"time"

	"github.com/sonalys/animeman/integrations/nyaa"
	"github.com/sonalys/animeman/internal/parser"
	"github.com/sonalys/animeman/internal/utils"
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
func filterTitleMatch(strippedTitles []string) func(e nyaa.Entry) bool {
	return func(nyaaEntry nyaa.Entry) bool {
		gotStrippedTitle := parser.StripTitle(nyaaEntry.Title)

		for _, title := range strippedTitles {
			if strings.EqualFold(gotStrippedTitle, title) {
				return true
			}
		}

		return false
	}
}
