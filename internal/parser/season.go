package parser

import (
	"regexp"
	"strings"
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
	// 3 - 04
	regexp.MustCompile(`(\d+)\s*-\s*(?:\d+)`),
}

// ParseSeason detects season on titles.
func ParseSeason(title string) string {
	for _, expr := range seasonExpr {
		matches := expr.FindAllStringSubmatch(title, -1)
		if len(matches) == 0 || len(matches[0]) < 2 {
			continue
		}
		return strings.TrimLeft(matches[0][1], "0")
	}
	return "1"
}

// seasonIndexMatch is used for removing season from titles.
func seasonIndexMatch(title string) int {
	for _, expr := range seasonExpr {
		matches := expr.FindAllStringSubmatchIndex(title, -1)
		if len(matches) == 0 || len(matches[0]) < 2 {
			continue
		}
		return matches[0][0]
	}
	return -1
}
