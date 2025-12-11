package parser

import (
	"fmt"
	"strings"
)

// Metadata is a digested metadata struct parsed from titles.
type Metadata struct {
	Source             string
	Title              string
	SeasonEpisodeTag   SeasonEpisodeTag
	Tags               []string
	VerticalResolution int
}

type SeasonEpisodeTag struct {
	Season  []int
	Episode []float64
}

func (t SeasonEpisodeTag) FirstSeason() int {
	if len(t.Season) == 0 {
		return 1
	}

	firstSeason := t.Season[0]

	for i := 1; i < len(t.Season); i++ {
		if t.Season[i] < firstSeason {
			firstSeason = t.Season[i]
		}
	}

	return firstSeason
}

func (t SeasonEpisodeTag) LastSeason() int {
	if len(t.Season) == 0 {
		return 1
	}

	lastSeason := t.Season[0]

	for i := 1; i < len(t.Season); i++ {
		if t.Season[i] > lastSeason {
			lastSeason = t.Season[i]
		}
	}

	return lastSeason
}

func (t SeasonEpisodeTag) FirstEpisode() float64 {
	if len(t.Episode) == 0 {
		return 1
	}

	firstEpisode := t.Episode[0]

	for i := 1; i < len(t.Episode); i++ {
		if t.Episode[i] < firstEpisode {
			firstEpisode = t.Episode[i]
		}
	}

	return firstEpisode
}

func (t SeasonEpisodeTag) LastEpisode() float64 {
	if len(t.Episode) == 0 {
		return 1
	}

	lastEpisode := t.Episode[0]

	for i := 1; i < len(t.Episode); i++ {
		if t.Episode[i] > lastEpisode {
			lastEpisode = t.Episode[i]
		}
	}

	return lastEpisode
}

// Returns false if same season and episode difference is bigger than 1.
// otherwise returns true.
func (t *SeasonEpisodeTag) Before(other SeasonEpisodeTag) bool {
	if t.LastSeason() > other.FirstSeason() {
		return false
	}

	if t.LastSeason() == other.FirstSeason() {
		if t.LastEpisode() >= other.FirstEpisode() {
			return false
		}
	}

	return true
}

func (m Metadata) Clone() Metadata {
	return Metadata{
		Source:             m.Source,
		Title:              m.Title,
		SeasonEpisodeTag:   m.SeasonEpisodeTag,
		Tags:               append([]string{}, m.Tags...),
		VerticalResolution: m.VerticalResolution,
	}
}

// TagBuildTitleSeasonEpisode builds a tag for filtering in your torrent client. Example: Show S03E02.
func (t SeasonEpisodeTag) BuildTag() string {
	var b strings.Builder

	switch {
	case len(t.Season) == 1:
		fmt.Fprintf(&b, "S%d", t.Season[0])
	case len(t.Season) == 2:
		fmt.Fprintf(&b, "S%d-%d", t.Season[0], t.Season[1])
	}

	switch {
	case len(t.Episode) == 1:
		fmt.Fprintf(&b, "E%v", t.Episode[0])
	case len(t.Episode) == 2:
		fmt.Fprintf(&b, "E%v-%v", t.Episode[0], t.Episode[1])

	}

	return b.String()
}

func (t SeasonEpisodeTag) IsZero() bool {
	return len(t.Season) == 0 && len(t.Episode) == 0
}

func (t SeasonEpisodeTag) IsMultiEpisode() bool {
	return len(t.Episode) == 0 || len(t.Episode) > 1
}
