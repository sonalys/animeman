package parser

import "strconv"

// Metadata is a digested metadata struct parsed from titles.
type Metadata struct {
	Source             string
	Title              string
	Episode            string
	Season             string
	Tags               []string
	VerticalResolution int
	// Is true when the title contains no episode information or multiple episodes.
	// Examples: Show S01, Show S01E01~13.
	IsMultiEpisode bool
}

// Returns false if same season and episode difference is bigger than 1.
// otherwise returns true.
func (m *Metadata) TagIsNextEpisode(latest string) bool {
	latestSeason := ParseSeason(latest)
	latestEpisode, isMulti := ParseEpisode(latest)
	// Avoids panic converting 6.5 for example to int.
	if isMulti {
		return true
	}
	if m.Season == latestSeason {
		epCur, _ := strconv.ParseFloat(m.Episode, 64)
		epLatest, _ := strconv.ParseFloat(latestEpisode, 64)
		return epCur <= epLatest+1
	}
	return true
}

func (m Metadata) Clone() Metadata {
	return Metadata{
		Source:             m.Source,
		Title:              m.Title,
		Episode:            m.Episode,
		Season:             m.Season,
		Tags:               append([]string{}, m.Tags...),
		VerticalResolution: m.VerticalResolution,
		IsMultiEpisode:     m.IsMultiEpisode,
	}
}
