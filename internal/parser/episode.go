package parser

import (
	"fmt"
	"regexp"
)

// Anything that is inside [].
var tagsExpr = regexp.MustCompile(`\[([^\[\]]*)\]`)

// Examples: 6, 6.5, 1~12, 1 ~ 12, 1-12, 1 - 12.
const episodeGroup = `(\d+(?:\.\d+)?|(?:\s?~|-\s?-\s?\d+))`

var episodeExpr = []*regexp.Regexp{
	// 0x15.
	regexp.MustCompile(`\d+x` + episodeGroup),
	// - 15.
	regexp.MustCompile(`(?i:[^season])\s` + episodeGroup + `(?:\W|$)`),
	// S02E15.
	regexp.MustCompile(`(?i:e)` + episodeGroup),
}

func matchEpisode(title string) (string, bool) {
	for _, expr := range episodeExpr {
		matches := expr.FindAllStringSubmatch(title, -1)
		if len(matches) == 0 || len(matches[0]) < 2 {
			continue
		}
		if len(matches) == 1 {
			episode := parseInt(matches[0][1])
			return fmt.Sprint(episode), false
		}
		episode1 := parseInt(matches[0][1])
		episode2 := parseInt(matches[1][1])
		// Stringify episode number to avoid left digits, example: 02.
		// Reason: we want an exact match for tags, so E02 and E2 wouldn't match.
		return fmt.Sprintf("%d~%d", episode1, episode2), true
	}
	// Some scenarios are like Frieren Season 1
	return "", true
}

func matchEpisodeIndex(title string) int {
	for _, expr := range episodeExpr {
		matches := expr.FindAllStringSubmatchIndex(title, -1)
		if len(matches) == 0 || len(matches[0]) < 2 {
			continue
		}
		return matches[0][0]
	}
	return -1
}
