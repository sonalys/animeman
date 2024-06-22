package parser

import (
	"fmt"
	"regexp"
	"strings"
)

// Regex for removing all annotations from a title, Examples: (Recoded), [1080p], .mkv.
var titleCleanupExpr = []*regexp.Regexp{
	// [anything inside brackets] or (parenthesis).
	regexp.MustCompile(`(\[.*?\])|(\(.*?\))`),
}

func TitleStripSubtitle(title string) string {
	title = strings.Split(title, ": ")[0]
	title = strings.Split(title, ", ")[0]
	title = strings.Split(title, "- ")[0]
	return title
}

// TitleStrip cleans title from sub-titles, tags and season / episode information.
// Example: [Source] Show: another story - S03E02 [1080p].mkv -> Show.
func TitleStrip(title string, cleanSubtitle bool) string {
	if index := seasonIndexMatch(title); index != -1 {
		title = title[:index]
	}
	if index := episodeIndexMatch(title); index != -1 {
		title = title[:index]
	}
	title = regexp.MustCompile(`\s{2,}`).ReplaceAllString(title, " ")
	if cleanSubtitle {
		title = TitleStripSubtitle(title)
	}
	title = strings.ReplaceAll(title, ".", " ")
	title = removeTags(title)
	return strings.TrimSpace(title)
}

func removeTags(title string) string {
	for _, expr := range titleCleanupExpr {
		title = expr.ReplaceAllString(title, "")
	}
	return title
}

// TitleParse will parse a title into a Metadata, extracting stripped title, tags, season and episode information.
func TitleParse(title string) Metadata {
	resp := Metadata{
		Title:              TitleStrip(title, false),
		VerticalResolution: qualityMatch(title),
	}
	if tags := tagsExpr.FindAllStringSubmatch(title, -1); len(tags) > 0 {
		resp.Source = tags[0][1]
		resp.Tags = make([]string, 0, len(tags[1:]))
		for _, matches := range tags[1:] {
			resp.Tags = append(resp.Tags, matches[1])
		}
	}
	title = removeTags(title)
	resp.Episode, resp.IsMultiEpisode = EpisodeParse(title)
	resp.Season = SeasonParse(title)
	return resp
}

// TagBuildTitleSeasonEpisode builds a tag for filtering in your torrent client. Example: Show S03E02.
func (t Metadata) TagBuildTitleSeasonEpisode() string {
	return fmt.Sprintf("%s %s", strings.ToLower(t.Title), t.TagBuildSeasonEpisode())
}

// TagBuildTitleSeasonEpisode builds a tag for filtering in your torrent client. Example: Show S03E02.
func (t Metadata) TagBuildSeasonEpisode() string {
	resp := fmt.Sprintf("S%s", t.Season)
	if t.Episode != "" {
		resp = resp + fmt.Sprintf("E%s", t.Episode)
	}
	return resp
}

// TagBuildBatch is used for when you download a torrent with multiple episodes.
func (t Metadata) TagBuildBatch() string {
	return fmt.Sprintf("%s S%s batch", strings.ToLower(t.Title), t.Season)
}

// TagBuildSeries builds a !Serie Name tag for you to be able to search all it's episodes with a tag.
func (t Metadata) TagBuildSeries() string {
	return "!" + strings.ToLower(t.Title)
}

// TagsBuildTorrent builds all tags Animeman needs from your torrent client.
func (t Metadata) TagsBuildTorrent() []string {
	tags := []string{t.TagBuildSeries(), t.TagBuildTitleSeasonEpisode()}
	if t.IsMultiEpisode {
		tags = append(tags, t.TagBuildBatch())
	}
	return tags
}
