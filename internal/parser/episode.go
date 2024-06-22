package parser

import (
	"fmt"
	"regexp"
	"strings"
)

// Anything that is inside [].
var tagsExpr = regexp.MustCompile(`\[([^\[\]]*)\]`)

// Examples: 6, 6.5, 1~12, 1 ~ 12, 1-12, 1 - 12.
const episodeGroup = `(\d+(?:\.\d+)?|(?:\s?~|-\s?-\s?\d+))`

var episodeExpr = []*regexp.Regexp{
	// Title - 05.
	regexp.MustCompile(` - ` + episodeGroup),
	// E15 or S02E15.
	regexp.MustCompile(`(?i:e)` + episodeGroup),
	// 0x15.
	regexp.MustCompile(`\d+(?i:x)` + episodeGroup),
}

// EpisodeParse detects episodes on titles.
func EpisodeParse(title string) (string, bool) {
	for _, expr := range episodeExpr {
		matches := expr.FindAllStringSubmatch(title, -1)
		if len(matches) == 0 || len(matches[0]) < 2 {
			continue
		}
		if len(matches) == 1 {
			return strings.TrimLeft(matches[0][1], "0"), false
		}
		episode1 := strings.TrimLeft(matches[0][1], "0")
		episode2 := strings.TrimLeft(matches[1][1], "0")
		// Stringify episode number to avoid left digits, example: 02.
		// Reason: we want an exact match for tags, so E02 and E2 wouldn't match.
		return fmt.Sprintf("%s~%s", episode1, episode2), true
	}
	// Some scenarios are like Frieren Season 1
	return "", true
}

// episodeIndexMatch is used for filtering episodes out of titles.
func episodeIndexMatch(title string) int {
	for _, expr := range episodeExpr {
		matches := expr.FindAllStringSubmatchIndex(title, -1)
		if len(matches) == 0 || len(matches[0]) < 2 {
			continue
		}
		return matches[0][0]
	}
	return -1
}
