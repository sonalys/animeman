package parser

import (
	"regexp"
	"strconv"
)

var qualityExpr = []*regexp.Regexp{
	// 1080p, 720p.
	regexp.MustCompile(`(\d+)p`),
	// 1920x1080
	regexp.MustCompile(`\d{3,4}x(\d{3,4})`),
}

// parseVerticalResolution detects quality from title.
func parseVerticalResolution(title string) int {
	for _, expr := range qualityExpr {
		matches := expr.FindAllStringSubmatch(title, -1)
		if len(matches) == 0 || len(matches[0]) < 2 {
			continue
		}

		value, _ := strconv.ParseInt(matches[0][1], 10, 64)

		return int(value)
	}
	return -1
}
