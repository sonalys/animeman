package discovery

import (
	"sync"
	"time"

	"github.com/sonalys/animeman/pkg/v1/animelist"
)

// ShowScanState tracks the scan history of a show to determine optimal polling intervals.
type ShowScanState struct {
	NextScanTime time.Time
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

// getState retrieves the current scan state for a show.
func (it *IntervalTracker) getState(entry animelist.Entry) ShowScanState {
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
func (it *IntervalTracker) UpdateState(entry animelist.Entry, foundNewEpisodes bool) time.Time {
	key := getShowKey(entry.Titles)
	nextInterval := it.calculateNextInterval(entry)
	nextScanTime := time.Now().Add(nextInterval)

	it.mu.Lock()
	defer it.mu.Unlock()
	it.state[key] = ShowScanState{
		NextScanTime:     nextScanTime,
		FoundNewEpisodes: foundNewEpisodes,
	}

	return nextScanTime
}

// ShouldScanNow determines if a show should be scanned based on its last scan time and interval.
// It should be scanned if next scan time is within the next poll frequency window, allowing for some flexibility in scheduling.
func (it *IntervalTracker) ShouldScanNow(entry animelist.Entry) bool {
	nextScanTime := it.GetNextScanTime(entry)

	return time.Until(nextScanTime) <= it.pollFrequency
}

// GetNextScanTime calculates the next optimal scan time for a show.
func (it *IntervalTracker) GetNextScanTime(entry animelist.Entry) time.Time {
	state := it.getState(entry)

	if state.NextScanTime.IsZero() {
		// If we've never scanned this show, we can scan immediately
		return time.Now()
	}

	return state.NextScanTime
}

// calculateNextInterval determines the optimal polling interval based on show airing schedule.
func (it *IntervalTracker) calculateNextInterval(entry animelist.Entry) time.Duration {
	now := time.Now()
	state := it.getState(entry)

	if entry.AiringStatus == animelist.AiringStatusAired {
		if state.FoundNewEpisodes {
			return it.pollFrequency
		}

		// If the show has finished airing, and we didn't find new episodes in the last scan, we can check much less frequently.
		return 24 * time.Hour
	}

	nextEpisodeAirDate := entry.StartDate

	for _, episode := range entry.EpisodeSchedule {
		// Consider delay between episode release and when it becomes available for streaming.
		if !episode.AirDate.After(now.Add(-24 * time.Hour)) {
			continue
		}

		nextEpisodeAirDate = episode.AirDate
		break
	}

	timeUntilEpisode := time.Until(nextEpisodeAirDate)

	if timeUntilEpisode < 3*time.Hour {
		return it.pollFrequency
	}

	if timeUntilEpisode < 24*time.Hour {
		return time.Hour
	}

	return 6 * time.Hour
}
