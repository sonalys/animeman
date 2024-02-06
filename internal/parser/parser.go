package parser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// StripTitle returns only the main title, trimming everything after ':'.
func StripTitle(title string) string {
	title, _, found := strings.Cut(title, ":")
	if found {
		return title
	}
	return title
}

type ParsedTitle struct {
	Source         string
	Title          string
	Episode        string
	Season         string
	Tags           []string
	IsMultiEpisode bool
}

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
			episode, _ := strconv.ParseInt(matches[0][1], 10, 64)
			return fmt.Sprint(episode), false
		}
		episode1, _ := strconv.ParseInt(matches[0][1], 10, 64)
		episode2, _ := strconv.ParseInt(matches[1][1], 10, 64)
		return fmt.Sprintf("%d~%d", episode1, episode2), true
	}
	// Some scenarios are like Frieren Season 1
	return "0", true
}

var seasonExpr = []*regexp.Regexp{
	// 2nd season.
	regexp.MustCompile(`(\d+)(?:nd)|(?:rd)|(?:th)(?i:\sseason)`),
	// 2x15.
	regexp.MustCompile(`(\d+)(?:x\d+)`),
	// S02E15.
	regexp.MustCompile(`(?i:s)(\d+)`),
	// Season 1.
	regexp.MustCompile(`(?i:season\s)(\d+)`),
}

func matchSeason(title string) string {
	for _, expr := range seasonExpr {
		matches := expr.FindAllStringSubmatch(title, -1)
		if len(matches) == 0 || len(matches[0]) < 2 {
			continue
		}
		season, _ := strconv.ParseInt(matches[0][1], 10, 64)
		return fmt.Sprint(season)
	}
	return "0"
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

var titleCleanupExpr = []*regexp.Regexp{
	regexp.MustCompile(`(\[.*?\]|\(.*?\))`),
	regexp.MustCompile(`\..*$`),
}

func cleanWithRegex(expr *regexp.Regexp, value string) string {
	return expr.ReplaceAllString(value, "")
}

func cleanupTitle(title string) string {
	for _, expr := range titleCleanupExpr {
		title = cleanWithRegex(expr, title)
	}
	if index := matchSeasonIndex(title); index != -1 {
		title = title[:index]
	}
	if index := matchEpisodeIndex(title); index != -1 {
		title = title[:index]
	}
	title = strings.ReplaceAll(title, "  ", " ")
	return strings.TrimSpace(title)
}

func ParseTitle(title string) ParsedTitle {
	resp := ParsedTitle{
		Title: cleanupTitle(title),
	}
	if tags := tagsExpr.FindAllStringSubmatch(title, -1); len(tags) > 0 {
		resp.Source = tags[0][1]
		resp.Tags = make([]string, 0, len(tags[1:]))
		for _, matches := range tags[1:] {
			resp.Tags = append(resp.Tags, matches[1])
		}
	}
	resp.Episode, resp.IsMultiEpisode = matchEpisode(title)
	resp.Season = matchSeason(title)
	return resp
}
