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

func StripTitleSubtitle(title string) string {
	title = strings.Split(title, ": ")[0]
	title = strings.Split(title, ", ")[0]
	title = strings.Split(title, "- ")[0]
	title = strings.Split(title, ". ")[0]
	return title
}

type titleCleanOptions struct {
	removeDots bool
}

type TitleStripOptions interface {
	applyTitleStripOptions(*titleCleanOptions)
}

type optRemoveDots struct{}

func (o optRemoveDots) applyTitleStripOptions(opts *titleCleanOptions) {
	opts.removeDots = true
}

func RemoveDots() TitleStripOptions {
	return optRemoveDots{}
}

// StripTitle cleans title from sub-titles, tags and season / episode information.
// Example: [Source] Show: another story - S03E02 [1080p].mkv -> Show.
func StripTitle(title string, opts ...TitleStripOptions) string {
	options := titleCleanOptions{}
	for _, opt := range opts {
		opt.applyTitleStripOptions(&options)
	}

	if index := seasonIndexMatch(title); index != -1 {
		title = title[:index]
	}
	if index := episodeIndexMatch(title); index != -1 {
		title = title[:index]
	}
	title = regexp.MustCompile(`\s{2,}`).ReplaceAllString(title, " ")
	title = StripTitleSubtitle(title)
	if options.removeDots {
		title = strings.ReplaceAll(title, ".", " ")
	}
	title = strings.ReplaceAll(title, "\"", "")
	title = removeTags(title)
	return strings.TrimSpace(title)
}

func removeTags(title string) string {
	for _, expr := range titleCleanupExpr {
		title = expr.ReplaceAllString(title, "")
	}
	return title
}

// Parse will parse a title into a Metadata, extracting stripped title, tags, season and episode information.
func Parse(title string, opts ...TitleStripOptions) Metadata {
	resp := Metadata{
		Title:              StripTitle(title, opts...),
		VerticalResolution: parseVerticalResolution(title),
	}
	if tags := tagsExpr.FindAllStringSubmatch(title, -1); len(tags) > 0 {
		resp.Source = tags[0][1]
		resp.Tags = make([]string, 0, len(tags[1:]))
		for _, matches := range tags[1:] {
			resp.Tags = append(resp.Tags, matches[1])
		}
	}
	title = removeTags(title)
	resp.Episode, resp.IsMultiEpisode = ParseEpisode(title)
	resp.Season = ParseSeason(title)
	return resp
}

// TagBuildTitleSeasonEpisode builds a tag for filtering in your torrent client. Example: Show S03E02.
func (t Metadata) TagBuildTitleSeasonEpisode() string {
	return fmt.Sprintf("%s %s", t.buildTitle(), t.TagBuildSeasonEpisode())
}

// TagBuildTitleSeasonEpisode builds a tag for filtering in your torrent client. Example: Show S03E02.
func (t Metadata) TagBuildSeasonEpisode() string {
	resp := fmt.Sprintf("S%s", t.Season)
	if t.Episode != "" {
		resp = resp + fmt.Sprintf("E%s", t.Episode)
	}
	return resp
}

func filterAlphanumeric(s string) string {
	var result strings.Builder
	result.Grow(len(s))
	for i := 0; i < len(s); i++ {
		b := s[i]
		if ('a' <= b && b <= 'z') || ('A' <= b && b <= 'Z') || ('0' <= b && b <= '9') || b == ' ' {
			result.WriteByte(b)
		}
	}
	return result.String()
}

func (t Metadata) buildTitle() string {
	return strings.ToLower(filterAlphanumeric(t.Title))
}

// TagBuildBatch is used for when you download a torrent with multiple episodes.
func (t Metadata) TagBuildBatch() string {
	return fmt.Sprintf("%s S%s batch", t.buildTitle(), t.Season)
}

// TagBuildSeries builds a !Serie Name tag for you to be able to search all it's episodes with a tag.
func (t Metadata) TagBuildSeries() string {
	return "!" + t.buildTitle()
}

// TagsBuildTorrent builds all tags Animeman needs from your torrent client.
func (t Metadata) TagsBuildTorrent() []string {
	tags := []string{t.TagBuildSeries(), t.TagBuildTitleSeasonEpisode()}
	if t.IsMultiEpisode {
		tags = append(tags, t.TagBuildBatch())
	}
	return tags
}
