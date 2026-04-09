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

// CalculateNextInterval determines the optimal polling interval based on show airing schedule.
// Dynamic intervals relative to the configured poll frequency:
// - For shows not yet aired: Interval increases as air date approaches (more frequent as it approaches)
// - For currently airing shows: Very short interval (1x poll frequency)
// - For completed shows with recent discoveries: Short interval (3x poll frequency)
// - For completed shows with no new episodes: Long interval (100x poll frequency)
func (it *IntervalTracker) CalculateNextInterval(entry animelist.Entry) time.Duration {
	now := time.Now()
	state := it.GetState(entry)

	// For shows that haven't aired yet
	if entry.StartDate.After(now) {
		timeUntilAir := entry.StartDate.Sub(now)

		// Scale interval based on proximity to air date, relative to poll frequency
		// Very far away (> 30 days): scan every 50x poll frequency
		if timeUntilAir > 30*24*time.Hour {
			return it.pollFrequency * 50
		}
		// Far away (7-30 days): scan every 20x poll frequency
		if timeUntilAir > 7*24*time.Hour {
			return it.pollFrequency * 20
		}
		// Close (< 7 days): scan every 5x poll frequency
		if timeUntilAir > 12*time.Hour {
			return it.pollFrequency * 5
		}
		// Very close (< 12 hours): scan every 1x poll frequency (frequent)
		return it.pollFrequency
	}

	// For currently airing shows - scan very frequently
	if entry.AiringStatus == animelist.AiringStatusAiring {
		return it.pollFrequency
	}

	// For completed shows
	if entry.AiringStatus == animelist.AiringStatusAired {
		// If we haven't found any new episodes in the last scan, use long interval
		if !state.FoundNewEpisodes && !state.LastScanTime.IsZero() {
			// Scan every 100x poll frequency for completed shows with no new content
			return it.pollFrequency * 100
		}

		// If we recently found episodes, scan more frequently
		if state.FoundNewEpisodes && !state.LastScanTime.IsZero() {
			// Scan every 3x poll frequency to catch remaining episodes
			return it.pollFrequency * 3
		}

		// First scan for completed show - moderate interval
		return it.pollFrequency * 10
	}

	// Unknown status - default to base poll frequency
	return it.pollFrequency
}

// ShouldScanNow determines if a show should be scanned based on its last scan time and interval.
func (it *IntervalTracker) ShouldScanNow(entry animelist.Entry) bool {
	return it.GetNextScanTime(entry).Before(time.Now())
}

// GetNextScanTime calculates the next optimal scan time for a show.
func (it *IntervalTracker) GetNextScanTime(entry animelist.Entry) time.Time {
	state := it.GetState(entry)
	if state.LastScanTime.IsZero() {
		// If we've never scanned this show, we can scan immediately
		return time.Now()
	}
	nextInterval := it.CalculateNextInterval(entry)
	return state.LastScanTime.Add(nextInterval)
}
