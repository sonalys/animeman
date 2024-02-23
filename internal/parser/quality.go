package parser

import (
	"regexp"
)

var qualityExpr = []*regexp.Regexp{
	// 1080p, 720p.
	regexp.MustCompile(`(\d+)p`),
	// 1920x1080
	regexp.MustCompile(`\d{3,4}x(\d{3,4})`),
}

// qualityMatch detects quality from title.
func qualityMatch(title string) int {
	for _, expr := range qualityExpr {
		matches := expr.FindAllStringSubmatch(title, -1)
		if len(matches) == 0 || len(matches[0]) < 2 {
			continue
		}
		return parseInt(matches[0][1])
	}
	return -1
}
