package metadata

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

type (
	Tags []Tag

	Tag struct {
		Number   int            `json:"number,omitzero"`
		Episodes []EpisodeRange `json:"episodes,omitzero"`
	}

	EpisodeRange struct {
		Start float32 `json:"start,omitzero"`
		End   float32 `json:"end,omitzero"`
	}
)

func (e EpisodeRange) Compare(other EpisodeRange) int {
	maxA := max(e.Start, e.End)
	maxB := max(other.Start, other.End)

	if maxA < maxB {
		return -1
	}
	if maxA > maxB {
		return 1
	}
	return 0
}

func compareEpisodes(a, b []EpisodeRange) int {
	aa := append([]EpisodeRange(nil), a...)
	bb := append([]EpisodeRange(nil), b...)

	slices.SortFunc(aa, func(x, y EpisodeRange) int { return y.Compare(x) })
	slices.SortFunc(bb, func(x, y EpisodeRange) int { return y.Compare(x) })

	n := min(len(bb), len(aa))
	for i := range n {
		if c := aa[i].Compare(bb[i]); c != 0 {
			return c
		}
	}

	if len(aa) < len(bb) {
		return -1
	}
	if len(aa) > len(bb) {
		return 1
	}
	return 0
}

func containsEpisode(have []EpisodeRange, wanted EpisodeRange) bool {
	ws, we := episodeBounds(wanted)
	for _, h := range have {
		hs, he := episodeBounds(h)
		if hs <= ws && he >= we {
			return true
		}
	}
	return false
}

func episodeBounds(e EpisodeRange) (float32, float32) {
	end := e.Start
	if e.End != 0 {
		end = e.End
	}
	return e.Start, end
}

func (s Tags) String() string {
	parts := make([]string, 0, len(s))

	for _, season := range s {
		parts = append(parts, season.String())
	}

	return strings.Join(parts, " ")
}

func (t Tag) String() string {
	var parts []string

	for _, seasonTag := range t.Episodes {
		if seasonTag.End != 0 {
			parts = append(parts,
				fmt.Sprintf("S%dE%g-%g", t.Number, seasonTag.Start, seasonTag.End),
			)
		} else {
			parts = append(parts,
				fmt.Sprintf("S%dE%g", t.Number, seasonTag.Start),
			)
		}
	}

	// A season without explicit episodes is a season pack.
	if len(t.Episodes) == 0 {
		parts = append(parts, fmt.Sprintf("S%d", t.Number))
	}

	return strings.Join(parts, " ")
}

func (m Tags) Compare(other any) int {
	return compareTags(normalizeTags(m), normalizeTags(other))
}

func (t Tag) Compare(other any) int {
	return compareTags(normalizeTags(t), normalizeTags(other))
}

func (s Tags) IsZero() bool {
	return len(s) == 0
}

func (t Tag) IsZero() bool {
	return t.Number == 0
}

func (s Tags) LastEpisode() float32 {
	if len(s) == 0 {
		return -1
	}

	for _, v := range slices.Backward(s) {
		if len(v.Episodes) == 0 {
			continue
		}

		epRange := v.Episodes[len(v.Episodes)-1]
		if epRange.End != 0 {
			return epRange.End
		}
		return epRange.Start
	}

	return -1
}

func (s Tags) LastSeason() int {
	for _, tag := range slices.Backward(s) {
		if tag.Number != 0 {
			return tag.Number
		}
	}

	return 0
}

// Contains reports whether m describes a release that contains all of the
// semantic content requested by other. A season pack contains every episode
// in that season; an episode release does not contain the season pack itself.
func (m Tags) Contains(other Tags) bool {
	for _, wanted := range other {
		have := findSeason(m, wanted.Number)
		if have == nil {
			return false
		}
		if len(wanted.Episodes) == 0 {
			if len(have.Episodes) != 0 {
				return false
			}
			continue
		}
		if len(have.Episodes) == 0 {
			continue // season pack contains every episode in the season
		}
		for _, ep := range wanted.Episodes {
			if !containsEpisode(have.Episodes, ep) {
				return false
			}
		}
	}

	return true
}

func (t Tag) compare(other Tag) int {
	if t.Number < other.Number {
		return -1
	}
	if t.Number > other.Number {
		return 1
	}

	// A season with no explicit episodes is a season pack. A pack represents
	// the complete season and therefore sorts after episode-specific releases.
	sp, op := len(t.Episodes) == 0, len(other.Episodes) == 0
	if sp != op {
		if sp {
			return 1
		}
		return -1
	}
	return compareEpisodes(t.Episodes, other.Episodes)
}

func (tags *Tags) AppendSeason(number int) *Tag {
	if number <= 0 {
		number = 1
	}

	for i := range *tags {
		if (*tags)[i].Number == number {
			return &(*tags)[i]
		}
	}

	*tags = append(*tags, Tag{Number: number})
	*tags = sortSeasons(*tags)

	return &(*tags)[len(*tags)-1]
}

func sortSeasons(tags Tags) Tags {
	slices.SortFunc(tags, func(a, b Tag) int {
		if a.Number < b.Number {
			return -1
		}
		if a.Number > b.Number {
			return 1
		}
		return 0
	})

	out := tags[:0]
	for _, season := range tags {
		if len(out) > 0 && out[len(out)-1].Number == season.Number {
			out[len(out)-1].Episodes = append(out[len(out)-1].Episodes, season.Episodes...)
			continue
		}
		out = append(out, season)
	}
	tags = out

	return tags
}

func findSeason(seasons []Tag, number int) *Tag {
	for i := range seasons {
		if seasons[i].Number == number {
			return &seasons[i]
		}
	}
	return nil
}

func compareTags(a, b []Tag) int {
	switch {
	case len(a) == 0 && len(b) == 0:
		return 0
	case len(a) == 0:
		return -1
	case len(b) == 0:
		return 1
	}

	a = highestTags(a)
	b = highestTags(b)

	return a[0].compare(b[0])
}

func highestTags(tags []Tag) []Tag {
	if len(tags) == 0 {
		return nil
	}

	highest := tags[0]

	for _, tag := range tags[1:] {
		if tag.compare(highest) > 0 {
			highest = tag
		}
	}

	return []Tag{highest}
}

func appendEpisode(tags *Tags, season, fallbackSeason int, start, end string) {
	if season <= 0 {
		season = fallbackSeason
	}
	if season <= 0 {
		season = 1
	}

	s := tags.AppendSeason(season)

	startValue, _ := strconv.ParseFloat(start, 32)
	endValue := float32(0)
	if end != "" {
		value, _ := strconv.ParseFloat(end, 32)
		endValue = float32(value)
	}

	epRange := EpisodeRange{
		Start: float32(startValue),
		End:   endValue,
	}

	if !containsEpisode(s.Episodes, epRange) {
		s.Episodes = append(s.Episodes, EpisodeRange{
			Start: float32(startValue),
			End:   endValue,
		})
	}
}

func normalizeTags(tags any) []Tag {
	var result []Tag

	switch v := tags.(type) {
	case Tag:
		result = append(result, v)
	case Tags:
		result = append(result, v...)
	}

	slices.SortFunc(result, func(a, b Tag) int {
		return a.compare(b)
	})

	result = deduplicateTags(result)

	return result
}

func deduplicateTags(tags []Tag) []Tag {
	if len(tags) < 2 {
		return tags
	}

	result := make([]Tag, 0, len(tags))

	for _, tag := range tags {
		found := false

		for _, existing := range result {
			if tag.compare(existing) == 0 {
				found = true
				break
			}
		}

		if !found {
			result = append(result, tag)
		}
	}

	return result
}
