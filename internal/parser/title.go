package parser

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/sonalys/animeman/internal/tags"
	"github.com/sonalys/animeman/internal/utils"
)

// Regex for removing all annotations from a title, Examples: (Recoded), [1080p], .mkv.
var titleCleanupExpr = []*regexp.Regexp{
	// [anything inside brackets] or (parenthesis).
	regexp.MustCompile(`(\[.*?\])|(\(.*?\))`),
}

func StripSeason(title string) string {
	if index := seasonIndexMatch(title); index != -1 {
		title = title[:index]
	}

	return title
}

// StripTitle cleans title from sub-titles, tags and season / episode information.
// Example: [Source] Show: another story - S03E02 [1080p].mkv -> Show.
func StripTitle(title string) string {
	title = removeDotSpacing(title)

	if index := seasonIndexMatch(title); index != -1 {
		title = title[:index]
	}

	if index := episodeIndexMatch(title); index != -1 {
		title = title[:index]
	}

	title = StripTags(title)
	title = strings.TrimSpace(title)

	return title
}

func StripSubtitle(title string) string {
	indexOf := strings.LastIndex(title, ": ")
	if indexOf == -1 {
		return title
	}

	return title[:indexOf]
}

func removeDotSpacing(title string) string {
	dotReplaceRegexp := regexp.MustCompile(`([^ ])\.([^ ])`)
	title = dotReplaceRegexp.ReplaceAllString(title, "$1 $2")
	return title
}

func StripTags(title string) string {
	for _, expr := range titleCleanupExpr {
		title = expr.ReplaceAllString(title, "")
	}

	return strings.TrimSpace(title)
}

// Parse will parse a title into a Metadata, extracting stripped title, tags, season and episode information.
func Parse(title string, fallbackSeason int, sources []string) Metadata {
	normalizedTitle := strings.ToLower(title)
	normalizedSources := utils.Transform(sources, func(source string) string {
		return strings.ToLower(source)
	})

	resp := Metadata{
		Title:              StripTitle(title),
		VerticalResolution: parseVerticalResolution(title),
		Tag:                tags.Tag{},
		// Source is extracted from the title if it matches any of the provided sources.
		// Better than guessing the source as the first tag match.
		ReleaseGroup: func() string {
			sourceIndex := slices.IndexFunc(
				normalizedSources,
				func(normalizedSource string) bool { return strings.Contains(normalizedTitle, normalizedSource) },
			)
			if sourceIndex != -1 {
				return sources[sourceIndex]
			}
			return ""
		}(),
	}

	if tags := tagsExpr.FindAllStringSubmatch(title, -1); len(tags) > 0 {
		// If the title starts with a tag, we assume it's the source and set it as such.
		// Example: [Source] Show - S03E02 [1080p].mkv
		if title[0] == '[' {
			resp.ReleaseGroup = tags[0][1]
			tags = tags[1:]
		}

		resp.Labels = make([]string, 0, len(tags))
		for _, matches := range tags {
			resp.Labels = append(resp.Labels, strings.Split(matches[1], " ")...)
		}
	}

	title = StripTags(title)

	resp.Tag.Episodes = ParseEpisode(title)

	if detectedSeason := ParseSeason(title); detectedSeason > 0 {
		resp.Tag.Seasons = []int{detectedSeason}
	} else {
		resp.Tag.Seasons = []int{fallbackSeason}
	}

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
