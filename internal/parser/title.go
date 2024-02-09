package parser

import (
	"fmt"
	"regexp"
	"strings"
)

// Metadata is a digested metadata struct parsed from titles.
type Metadata struct {
	Source  string
	Title   string
	Episode string
	Season  string
	Tags    []string
	// Is true when the title contains no episode information or multiple episodes.
	// Examples: Show S01, Show S01E01~13.
	IsMultiEpisode bool
}

// Regex for removing all annotations from a title, Examples: (Recoded), [1080p], .mkv.
var titleCleanupExpr = []*regexp.Regexp{
	regexp.MustCompile(`(\[.*?\]|\(.*?\))`),
	regexp.MustCompile(`\..*$`),
}

// TitleStrip cleans title from sub-titles, tags and season / episode information.
// Example: [Source] Show: another story - S03E02 [1080p].mkv -> Show.
func TitleStrip(title string) string {
	for _, expr := range titleCleanupExpr {
		title = expr.ReplaceAllString(title, "")
	}
	if index := matchSeasonIndex(title); index != -1 {
		title = title[:index]
	}
	if index := episodeIndexMatch(title); index != -1 {
		title = title[:index]
	}
	title = strings.Split(title, ": ")[0]
	title = strings.Split(title, ", ")[0]
	title = strings.Split(title, "- ")[0]
	title = strings.ReplaceAll(title, "  ", " ")
	return strings.TrimSpace(title)
}

// TitleParse will parse a title into a Metadata, extracting stripped title, tags, season and episode information.
func TitleParse(title string) Metadata {
	resp := Metadata{
		Title: TitleStrip(title),
	}
	if tags := tagsExpr.FindAllStringSubmatch(title, -1); len(tags) > 0 {
		resp.Source = tags[0][1]
		resp.Tags = make([]string, 0, len(tags[1:]))
		for _, matches := range tags[1:] {
			resp.Tags = append(resp.Tags, matches[1])
		}
	}
	resp.Episode, resp.IsMultiEpisode = episodeMatch(title)
	resp.Season = matchSeason(title)
	return resp
}

// TagBuildSeasonEpisode builds a tag for filtering in your torrent client. Example: Show S03E02.
func (t Metadata) TagBuildSeasonEpisode() string {
	resp := fmt.Sprintf("%s S%s", t.Title, t.Season)
	if t.Episode != "" {
		resp = resp + fmt.Sprintf("E%s", t.Episode)
	}
	return resp
}

// TagBuildBatch is used for when you download a torrent with multiple episodes.
func (t Metadata) TagBuildBatch() string {
	return fmt.Sprintf("%s S%s batch", t.Title, t.Season)
}

// TagBuildSeries builds a !Serie Name tag for you to be able to search all it's episodes with a tag.
func (t Metadata) TagBuildSeries() string {
	return "!" + t.Title
}

// TagsBuildTorrent builds all tags Animeman needs from your torrent client.
func (t Metadata) TagsBuildTorrent() []string {
	tags := []string{t.TagBuildSeries(), t.TagBuildSeasonEpisode()}
	if t.IsMultiEpisode {
		tags = append(tags, t.TagBuildBatch())
	}
	return tags
}
