package parser

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/sonalys/animeman/integrations/qbittorrent"
)

type ParsedTitle struct {
	Source         string
	Title          string
	Episode        string
	Season         string
	Tags           []string
	IsMultiEpisode bool
}

var titleCleanupExpr = []*regexp.Regexp{
	regexp.MustCompile(`(\[.*?\]|\(.*?\))`),
	regexp.MustCompile(`\..*$`),
}

func cleanWithRegex(expr *regexp.Regexp, value string) string {
	return expr.ReplaceAllString(value, "")
}

func StripTitle(title string) string {
	for _, expr := range titleCleanupExpr {
		title = cleanWithRegex(expr, title)
	}
	if index := matchSeasonIndex(title); index != -1 {
		title = title[:index]
	}
	if index := matchEpisodeIndex(title); index != -1 {
		title = title[:index]
	}
	title = strings.Split(title, ": ")[0]
	title = strings.Split(title, ", ")[0]
	title = strings.Split(title, "- ")[0]
	title = strings.ReplaceAll(title, "  ", " ")
	return strings.TrimSpace(title)
}

func ParseTitle(title string) ParsedTitle {
	resp := ParsedTitle{
		Title: StripTitle(title),
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

func (t ParsedTitle) BuildSeasonEpisodeTag() string {
	resp := fmt.Sprintf("%s S%s", t.Title, t.Season)
	if !t.IsMultiEpisode {
		resp = resp + fmt.Sprintf("E%s", t.Episode)
	}
	return resp
}

func (t ParsedTitle) BuildBatchTag() string {
	return fmt.Sprintf("%s S%s batch", t.Title, t.Season)
}

func (t ParsedTitle) BuildSeriesTag() string {
	return "!" + t.Title
}

func (t ParsedTitle) BuildTorrentTags() qbittorrent.Tags {
	tags := qbittorrent.Tags{t.BuildSeriesTag(), t.BuildSeasonEpisodeTag()}
	if t.IsMultiEpisode {
		tags = append(tags, t.BuildBatchTag())
	}
	return tags
}
