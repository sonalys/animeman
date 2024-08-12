package parser

import (
	"fmt"
	"regexp"
	"strings"
)

// Anything that is inside [].
var tagsExpr = regexp.MustCompile(`\[([^\[\]]*)\]`)

const episodeRegexExpr = `(\d+(?:\.\d(?:\D|$))?)`

// Examples: 6, 6.5, 1~12, 1 ~ 12, 1-12, 1 - 12.
const episodeGroup = `(?:` + episodeRegexExpr + `(?:\s*[~\-]\s*` + episodeRegexExpr + `)?)`

var episodeExpr = []*regexp.Regexp{
	// Title - 05.
	regexp.MustCompile(` - ` + episodeGroup),
	// E15 or S02E15.
	regexp.MustCompile(`(?i)e` + episodeGroup),
	// 0x15.
	regexp.MustCompile(`(?i)\W\d+x` + episodeGroup),
}

// EpisodeParse detects episodes on titles.
func EpisodeParse(title string) (string, bool) {
	for _, expr := range episodeExpr {
		matches := expr.FindAllStringSubmatch(title, -1)
		if len(matches) == 0 {
			continue
		}
		group := matches[0]
		if group[2] == "" {
			return strings.TrimLeft(group[1], "0"), false
		}
		episode1 := strings.TrimLeft(group[1], "0")
		episode2 := strings.TrimLeft(group[2], "0")
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
