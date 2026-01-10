package tags

import (
	"fmt"
	"strings"
)

type Tag struct {
	// Seasons represents which seasons are specified in a tag.
	// Tags can represent multiple seasons, like in S01-12.
	Seasons []int

	// Episodes represents which episodes or episode ranges are specified in a tag.
	// float64 because of half episodes like 6.5.
	Episodes []float64
}

var Zero Tag

func SeasonEpisode(season int, episode float64) Tag {
	return Tag{
		Seasons:  []int{season},
		Episodes: []float64{episode},
	}
}

func (t Tag) FirstSeason() int {
	if t.IsZero() {
		return 0
	}

	if len(t.Seasons) == 0 {
		return 1
	}

	firstSeason := t.Seasons[0]

	for i := 1; i < len(t.Seasons); i++ {
		if t.Seasons[i] < firstSeason {
			firstSeason = t.Seasons[i]
		}
	}

	return firstSeason
}

func (t Tag) LastSeason() int {
	if t.IsZero() {
		return 0
	}

	if len(t.Seasons) == 0 {
		return 1
	}

	lastSeason := t.Seasons[0]

	for i := 1; i < len(t.Seasons); i++ {
		if t.Seasons[i] > lastSeason {
			lastSeason = t.Seasons[i]
		}
	}

	return lastSeason
}

func (t Tag) FirstEpisode() float64 {
	if t.IsZero() {
		return 0
	}

	if len(t.Episodes) == 0 {
		return 1
	}

	firstEpisode := t.Episodes[0]

	for i := 1; i < len(t.Episodes); i++ {
		if t.Episodes[i] < firstEpisode {
			firstEpisode = t.Episodes[i]
		}
	}

	return firstEpisode
}

func (t Tag) LastEpisode() float64 {
	if t.IsZero() {
		return 0
	}

	if len(t.Episodes) == 0 {
		return 1
	}

	lastEpisode := t.Episodes[0]

	for i := 1; i < len(t.Episodes); i++ {
		if t.Episodes[i] > lastEpisode {
			lastEpisode = t.Episodes[i]
		}
	}

	return lastEpisode
}

// Returns false if same season and episode difference is bigger than 1.
// otherwise returns true.
func (t *Tag) Before(other Tag) bool {
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

// String builds a tag for filtering in your torrent client. Example: Show S03E02.
func (t Tag) String() string {
	var b strings.Builder

	switch {
	case len(t.Seasons) == 1:
		fmt.Fprintf(&b, "S%d", t.Seasons[0])
	case len(t.Seasons) == 2:
		fmt.Fprintf(&b, "S%d-%d", t.Seasons[0], t.Seasons[1])
	}

	switch {
	case len(t.Episodes) == 1:
		fmt.Fprintf(&b, "E%v", t.Episodes[0])
	case len(t.Episodes) == 2:
		fmt.Fprintf(&b, "E%v-%v", t.Episodes[0], t.Episodes[1])

	}

	return b.String()
}

func (t Tag) IsZero() bool {
	return len(t.Seasons) == 0 && len(t.Episodes) == 0
}

func (t Tag) IsMultiEpisode() bool {
	return len(t.Episodes) == 0 || len(t.Episodes) > 1
}
