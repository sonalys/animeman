package discovery

import (
	"sync"
	"time"

	"github.com/sonalys/animeman/pkg/v1/animelist"
)

// ShowScanState tracks the scan history of a show to determine optimal polling intervals.
type ShowScanState struct {
	LastScanTime time.Time
	// Indicates if new episodes were found in the last scan, used to adjust intervals for completed shows
	FoundNewEpisodes bool
}

// IntervalTracker manages scan state for all shows and calculates adaptive intervals.
type IntervalTracker struct {
	mu            sync.RWMutex
	pollFrequency time.Duration
	state         map[string]ShowScanState // Key is the concatenated show titles
}

// NewIntervalTracker creates a new interval tracker with a configured poll frequency.
func NewIntervalTracker(pollFrequency time.Duration) *IntervalTracker {
	return &IntervalTracker{
		pollFrequency: pollFrequency,
		state:         make(map[string]ShowScanState),
	}
}

// getShowKey creates a unique key for a show based on its titles.
func getShowKey(titles []string) string {
	if len(titles) == 0 {
		return ""
	}
	// Use the first title as the key (titles are sorted in Entry creation)
	return titles[0]
}

// GetState retrieves the current scan state for a show.
func (it *IntervalTracker) GetState(entry animelist.Entry) ShowScanState {
	it.mu.RLock()
	defer it.mu.RUnlock()

	key := getShowKey(entry.Titles)
	state, exists := it.state[key]
	if !exists {
		return ShowScanState{}
	}
	return state
}

// UpdateState updates the scan state for a show after a discovery scan.
func (it *IntervalTracker) UpdateState(entry animelist.Entry, foundNewEpisodes bool) {
	it.mu.Lock()
	defer it.mu.Unlock()

	key := getShowKey(entry.Titles)
	state := it.state[key]
	state.LastScanTime = time.Now()
	state.FoundNewEpisodes = foundNewEpisodes

	it.state[key] = state
}

// ShouldScanNow determines if a show should be scanned based on its last scan time and interval.
// It should be scanned if next scan time is within the next poll frequency window, allowing for some flexibility in scheduling.
func (it *IntervalTracker) ShouldScanNow(entry animelist.Entry) bool {
	nextScanTime := it.GetNextScanTime(entry)
	return time.Until(nextScanTime) <= it.pollFrequency
}

// GetNextScanTime calculates the next optimal scan time for a show.
func (it *IntervalTracker) GetNextScanTime(entry animelist.Entry) time.Time {
	state := it.GetState(entry)
	if state.LastScanTime.IsZero() {
		// If we've never scanned this show, we can scan immediately
		return time.Now()
	}
	nextInterval := it.calculateNextInterval(entry)
	return state.LastScanTime.Add(nextInterval)
}

// calculateNextInterval determines the optimal polling interval based on show airing schedule.
// Dynamic intervals relative to the configured poll frequency:
// - For airing shows with episode schedule: interval based on next episode air date
// - For shows not yet aired: Interval increases as air date approaches (more frequent as it approaches)
// - For currently airing shows: Very short interval (1x poll frequency)
// - For completed shows with recent discoveries: Short interval (3x poll frequency)
// - For completed shows with no new episodes: Long interval (100x poll frequency)
func (it *IntervalTracker) calculateNextInterval(entry animelist.Entry) time.Duration {
	now := time.Now()
	state := it.GetState(entry)

	switch entry.AiringStatus {
	case animelist.AiringStatusAiring:
		for _, episode := range entry.EpisodeSchedule {
			if episode.AirDate.After(now) {
				timeUntilEpisode := episode.AirDate.Sub(now)

				if timeUntilEpisode > 7*24*time.Hour {
					return it.pollFrequency * 20
				}
				if timeUntilEpisode > 24*time.Hour {
					return it.pollFrequency * 5
				}
				if timeUntilEpisode > 12*time.Hour {
					return it.pollFrequency * 2
				}
				return it.pollFrequency
			}
		}
		return it.pollFrequency
	case animelist.AiringStatusAired:
		if !state.FoundNewEpisodes && !state.LastScanTime.IsZero() {
			return it.pollFrequency * 100
		}

		if state.FoundNewEpisodes && !state.LastScanTime.IsZero() {
			return it.pollFrequency * 3
		}

		return it.pollFrequency * 10
	default:
		timeUntilAir := entry.StartDate.Sub(now)

		if timeUntilAir > 30*24*time.Hour {
			return it.pollFrequency * 50
		}
		if timeUntilAir > 7*24*time.Hour {
			return it.pollFrequency * 20
		}
		if timeUntilAir > 12*time.Hour {
			return it.pollFrequency * 5
		}
		return it.pollFrequency
	}
}
