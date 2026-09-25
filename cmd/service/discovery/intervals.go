package discovery

import (
	"sync"
	"time"

	"github.com/sonalys/animeman/internal/ports/animelist"
)

type (
	// showScanState tracks the scan history of a show to determine optimal polling intervals.
	showScanState struct {
		NextScanTime time.Time
		// Indicates if new episodes were found in the last scan, used to adjust intervals for completed shows
		FoundNewEpisodes bool
	}

	// intervalTracker manages scan state for all shows and calculates adaptive intervals.
	intervalTracker struct {
		mu            sync.RWMutex
		pollFrequency time.Duration
		state         map[string]showScanState // Key is the concatenated show titles
	}
)

// newIntervalTracker creates a new interval tracker with a configured poll frequency.
func newIntervalTracker(pollFrequency time.Duration) *intervalTracker {
	return &intervalTracker{
		pollFrequency: pollFrequency,
		state:         make(map[string]showScanState),
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
func (it *intervalTracker) getState(entry animelist.Entry) showScanState {
	it.mu.RLock()
	defer it.mu.RUnlock()

	key := getShowKey(entry.Titles)
	state, exists := it.state[key]
	if !exists {
		return showScanState{}
	}
	return state
}

// updateState updates the scan state for a show after a discovery scan.
func (it *intervalTracker) updateState(entry animelist.Entry, foundNewEpisodes bool) time.Time {
	key := getShowKey(entry.Titles)
	nextInterval := it.calculateNextInterval(entry)
	nextScanTime := time.Now().Add(nextInterval)

	it.mu.Lock()
	defer it.mu.Unlock()
	it.state[key] = showScanState{
		NextScanTime:     nextScanTime,
		FoundNewEpisodes: foundNewEpisodes,
	}

	return nextScanTime
}

// shouldScanNow determines if a show should be scanned based on its last scan time and interval.
// It should be scanned if next scan time is within the next poll frequency window, allowing for some flexibility in scheduling.
func (it *intervalTracker) shouldScanNow(entry animelist.Entry) bool {
	nextScanTime := it.getNextScanTime(entry)

	return time.Until(nextScanTime) <= it.pollFrequency
}

// getNextScanTime calculates the next optimal scan time for a show.
func (it *intervalTracker) getNextScanTime(entry animelist.Entry) time.Time {
	state := it.getState(entry)

	if state.NextScanTime.IsZero() {
		// If we've never scanned this show, we can scan immediately
		return time.Now()
	}

	return state.NextScanTime
}

// calculateNextInterval determines the optimal polling interval based on show airing schedule.
func (it *intervalTracker) calculateNextInterval(entry animelist.Entry) time.Duration {
	now := time.Now()
	hasEpisodeSchedule := len(entry.EpisodeSchedule) > 0
	state := it.getState(entry)

	if state.FoundNewEpisodes {
		// If we found new episodes in the last scan, we can check more frequently for updates
		return it.pollFrequency
	}

	if !hasEpisodeSchedule {
		switch entry.AiringStatus {
		case animelist.AiringStatusAiring:
			return it.pollFrequency
		case animelist.AiringStatusAired:
			endDate := entry.EndDate

			// If we don't have an end date, but the show has started, we can assume it might be ongoing and check more frequently.
			if endDate.IsZero() {
				endDate = entry.StartDate.AddDate(1, 0, 0)
			}

			switch {
			case endDate.IsZero():
				return 24 * time.Hour
			case endDate.Before(now.AddDate(0, 0, -7)):
				return 7 * 24 * time.Hour
			default:
				return it.pollFrequency
			}
		}
	}

	timeUntilRelease := entry.StartDate.Sub(now)
	timeUntilRelease = max(timeUntilRelease, -timeUntilRelease)

	// Find the closest episode air date.
	for _, episode := range entry.EpisodeSchedule {
		delta := episode.AirDate.Sub(now)
		delta = max(delta, -delta)

		if delta < timeUntilRelease {
			timeUntilRelease = delta
		}
	}

	if timeUntilRelease < 3*time.Hour {
		return it.pollFrequency
	}

	if timeUntilRelease < 24*time.Hour {
		return time.Hour
	}

	return 24 * time.Hour
}
