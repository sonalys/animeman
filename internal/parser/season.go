package parser

import (
	"regexp"
	"strconv"
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
	// WIP: not working yet, identifies wrongly some cases.
	// Title 3.
	// regexp.MustCompile(`(?:[^-]\s+)(\d+)`),
}

// ParseSeason detects season on titles.
func ParseSeason(title string) int {
	for _, expr := range seasonExpr {
		matches := expr.FindAllStringSubmatch(title, -1)
		if len(matches) == 0 || len(matches[0]) < 2 {
			continue
		}

		seasonNumber, err := strconv.ParseInt(matches[0][1], 10, 64)
		if err != nil {
			continue
		}

		return int(seasonNumber)
	}

	return 0
}

// seasonIndexMatch is used for removing season from titles.
func seasonIndexMatch(title string) int {
	for i, expr := range seasonExpr {
		matches := expr.FindAllStringSubmatchIndex(title, -1)
		if len(matches) == 0 || len(matches[0]) < 2 {
			continue
		}

		startingIndex := matches[0][0]
		if i == 5 {
			startingIndex = matches[0][2] - 1
		}

		return startingIndex
	}
	return -1
}
