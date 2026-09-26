package metadata

import (
	"strconv"
	"strings"
)

type Resolutions []string

func (r Resolutions) Highest() string {
	var highest string
	var highestHeight int

	for _, resolution := range r {
		height, err := strconv.Atoi(strings.TrimSuffix(strings.ToLower(resolution), "p"))
		if err != nil {
			continue
		}

		if height > highestHeight {
			highestHeight = height
			highest = resolution
		}
	}

	return highest
}

func (r Resolutions) Compare(other Resolutions) int {
	a := r.Highest()
	b := other.Highest()

	if a == b {
		return 0
	}

	av := resolutionHeight(a)
	bv := resolutionHeight(b)

	switch {
	case av < bv:
		return -1
	default:
		return 1
	}
}

func resolutionHeight(resolution string) int {
	resolution = strings.TrimSpace(strings.ToLower(resolution))
	resolution = strings.TrimSuffix(resolution, "p")

	height, err := strconv.Atoi(resolution)
	if err != nil {
		return 0
	}

	return height
}
