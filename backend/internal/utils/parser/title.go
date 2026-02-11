package parser

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/sonalys/animeman/internal/utils/tags"
)

// Regex for removing all annotations from a title, Examples: (Recoded), [1080p], .mkv.
var titleCleanupExpr = []*regexp.Regexp{
	// [anything inside brackets] or (parenthesis).
	regexp.MustCompile(`(\[.*?\])|(\(.*?\))`),
}

type titleCleanOptions struct {
	removeDots bool
}

type TitleStripOptions interface {
	applyTitleStripOptions(*titleCleanOptions)
}

// StripTitle cleans title from sub-titles, tags and season / episode information.
// Example: [Source] Show: another story - S03E02 [1080p].mkv -> Show.
func StripTitle(title string, opts ...TitleStripOptions) string {
	options := titleCleanOptions{}

	for _, opt := range opts {
		opt.applyTitleStripOptions(&options)
	}

	title = removeDotSpacing(title)

	if index := seasonIndexMatch(title); index != -1 {
		title = title[:index]
	}

	if index := episodeIndexMatch(title); index != -1 {
		title = title[:index]
	}

	title = removeQuotation(title)
	title = removeTags(title)
	title = removeTrailingNumbers(title)
	title = removeDashes(title)
	title = flattenMultispace(title)
	title = strings.TrimSpace(title)

	return title
}

func removeDashes(title string) string {
	return strings.ReplaceAll(title, "-", "")
}

func removeTrailingNumbers(title string) string {
	return strings.TrimRightFunc(title, func(r rune) bool {
		return r >= '0' && r <= '9'
	})
}

func removeDotSpacing(title string) string {
	dotReplaceRegexp := regexp.MustCompile(`([^ ])\.([^ ])`)
	title = dotReplaceRegexp.ReplaceAllString(title, "$1 $2")
	return title
}

func removeTags(title string) string {
	for _, expr := range titleCleanupExpr {
		title = expr.ReplaceAllString(title, "")
	}

	return title
}

func flattenMultispace(title string) string {
	return regexp.MustCompile(`\s{2,}`).ReplaceAllString(title, " ")
}

func removeQuotation(title string) string {
	return strings.ReplaceAll(title, "\"", "")
}

// Parse will parse a title into a Metadata, extracting stripped title, tags, season and episode information.
func Parse(title string, opts ...TitleStripOptions) Metadata {
	resp := Metadata{
		Title:              StripTitle(title, opts...),
		VerticalResolution: parseVerticalResolution(title),
		Tag:                tags.Tag{},
	}
	if tags := tagsExpr.FindAllStringSubmatch(title, -1); len(tags) > 0 {
		resp.Source = tags[0][1]
		resp.Labels = make([]string, 0, len(tags[1:]))
		for _, matches := range tags[1:] {
			resp.Labels = append(resp.Labels, matches[1])
		}
	}
	title = removeTags(title)

	resp.Tag.Episodes = ParseEpisode(title)
	resp.Tag.Seasons = []int{ParseSeason(title)}
	return resp
}

// TagBuildTitleSeasonEpisode builds a tag for filtering in your torrent client. Example: Show S03E02.
func (t Metadata) TagBuildTitleSeasonEpisode() string {
	return fmt.Sprintf("%s %s", t.buildTitle(), t.Tag.String())
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

// BuildSeriesTag builds a !Serie Name tag for you to be able to search all it's episodes with a tag.
func (t Metadata) BuildSeriesTag() string {
	return BuildTitleTag(t.Title)
}

// BuildTorrentTags builds all tags Animeman needs from your torrent client.
func (t Metadata) BuildTorrentTags() []string {
	tags := []string{t.BuildSeriesTag(), t.Tag.String()}
	return tags
}

// BuildTitleTag builds the torrent series tag. Example: !serie name.
func BuildTitleTag(title string) string {
	return "!" + strings.ToLower(filterAlphanumeric(title))
}
