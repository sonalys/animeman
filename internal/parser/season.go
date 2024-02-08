package parser

import (
	"fmt"
	"regexp"
)

var seasonExpr = []*regexp.Regexp{
	// 2nd season.
	regexp.MustCompile(`(\d+)(?:(?:nd)|(?:rd)|(?:th))(?i:\sseason)`),
	// 2x15.
	regexp.MustCompile(`(\d+)(?:x\d+)`),
	// S02E15.
	regexp.MustCompile(`(?i:s)(\d+)`),
	// Season 1.
	regexp.MustCompile(`(?i:season\s)(\d+)`),
}

// matchSeason detects season on titles.
func matchSeason(title string) string {
	for _, expr := range seasonExpr {
		matches := expr.FindAllStringSubmatch(title, -1)
		if len(matches) == 0 || len(matches[0]) < 2 {
			continue
		}
		// Stringify season number to avoid left digits, example: 02.
		// Reason: we want an exact match for tags, so S02 and S2 wouldn't match.
		season := parseInt(matches[0][1])
		return fmt.Sprint(season)
	}
	return "1"
}

// matchSeasonIndex is used for removing season from titles.
func matchSeasonIndex(title string) int {
	for _, expr := range seasonExpr {
		matches := expr.FindAllStringSubmatchIndex(title, -1)
		if len(matches) == 0 || len(matches[0]) < 2 {
			continue
		}
		return matches[0][0]
	}
	return -1
}
