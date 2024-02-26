package discovery

import (
	"strconv"
	"time"

	"github.com/sonalys/animeman/integrations/nyaa"
	"github.com/sonalys/animeman/internal/parser"
	"github.com/sonalys/animeman/internal/utils"
	"golang.org/x/exp/constraints"
)

func strSliceToFloat(from []string) []float64 {
	out := make([]float64, 0, len(from))
	for _, cur := range from {
		out = append(out, utils.Must(strconv.ParseFloat(cur, 64)))
	}
	return out
}

func min[T constraints.Ordered](values ...T) (min T) {
	if len(values) == 0 {
		return
	}
	min = values[0]
	for i := range values {
		if values[i] < min {
			min = values[i]
		}
	}
	return min
}

func max[T constraints.Ordered](values ...T) (max T) {
	if len(values) == 0 {
		return
	}
	max = values[0]
	for i := range values {
		if values[i] > max {
			max = values[i]
		}
	}
	return max
}

func filterBatchEntries(e parser.ParsedNyaa) bool { return e.Meta.IsMultiEpisode }

// filterPublishedAfterDate creates a filter for nyaa.Entry only after a date t.
func filterPublishedAfterDate(t time.Time) func(e nyaa.Entry) bool {
	return func(e nyaa.Entry) bool {
		publishedDate := utils.Must(time.Parse(time.RFC1123Z, e.PubDate))
		return publishedDate.After(t)
	}
}
