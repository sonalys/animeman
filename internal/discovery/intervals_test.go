package discovery

import (
	"testing"
	"time"

	"github.com/sonalys/animeman/pkg/v1/animelist"
	"github.com/stretchr/testify/assert"
)

func TestNewIntervalTracker(t *testing.T) {
	pollFrequency := 5 * time.Minute
	tracker := NewIntervalTracker(pollFrequency)

	assert.NotNil(t, tracker)
	assert.Equal(t, pollFrequency, tracker.pollFrequency)
	assert.NotNil(t, tracker.state)
	assert.Empty(t, tracker.state)
}

func TestGetShowKey(t *testing.T) {
	tests := []struct {
		name     string
		titles   []string
		expected string
	}{
		{
			name:     "single title",
			titles:   []string{"Attack on Titan"},
			expected: "Attack on Titan",
		},
		{
			name:     "multiple titles",
			titles:   []string{"Attack on Titan", "進撃の巨人"},
			expected: "Attack on Titan",
		},
		{
			name:     "empty titles",
			titles:   []string{},
			expected: "",
		},
		{
			name:     "single empty string",
			titles:   []string{""},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getShowKey(tt.titles)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUpdateState(t *testing.T) {
	tracker := NewIntervalTracker(5 * time.Minute)
	entry := animelist.Entry{
		Titles:       []string{"Test Anime"},
		AiringStatus: animelist.AiringStatusAiring,
		StartDate:    time.Now().AddDate(0, 0, 1),
	}

	before := time.Now()
	nextScanTime := tracker.UpdateState(entry, true)
	after := time.Now()

	// Verify the returned next scan time is in the future
	assert.True(t, nextScanTime.After(before))
	assert.True(t, nextScanTime.After(after))

	// Verify state was stored
	state := tracker.getState(entry)
	assert.True(t, state.FoundNewEpisodes)
	assert.Equal(t, nextScanTime, state.NextScanTime)
}

func TestUpdateState_FoundNewEpisodes(t *testing.T) {
	tests := []struct {
		name             string
		foundNewEpisodes bool
		airingStatus     animelist.AiringStatus
		expectedFoundNew bool
	}{
		{
			name:             "found new episodes",
			foundNewEpisodes: true,
			airingStatus:     animelist.AiringStatusAiring,
			expectedFoundNew: true,
		},
		{
			name:             "no new episodes",
			foundNewEpisodes: false,
			airingStatus:     animelist.AiringStatusAired,
			expectedFoundNew: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := NewIntervalTracker(5 * time.Minute)
			entry := animelist.Entry{
				Titles:       []string{tt.name},
				AiringStatus: tt.airingStatus,
				StartDate:    time.Now(),
			}

			tracker.UpdateState(entry, tt.foundNewEpisodes)
			state := tracker.getState(entry)

			assert.Equal(t, tt.expectedFoundNew, state.FoundNewEpisodes)
		})
	}
}

func TestGetNextScanTime(t *testing.T) {
	tests := []struct {
		name             string
		setupFunc        func(*IntervalTracker, animelist.Entry)
		entry            animelist.Entry
		shouldBeNow      bool
		shouldBeInFuture bool
	}{
		{
			name: "never scanned before",
			setupFunc: func(it *IntervalTracker, entry animelist.Entry) {
				// Don't update state, simulating first scan
			},
			entry: animelist.Entry{
				Titles: []string{"First Scan Anime"},
			},
			shouldBeNow: true,
		},
		{
			name: "has been scanned",
			setupFunc: func(it *IntervalTracker, entry animelist.Entry) {
				it.UpdateState(entry, false)
			},
			entry: animelist.Entry{
				Titles:       []string{"Already Scanned"},
				AiringStatus: animelist.AiringStatusAiring,
				StartDate:    time.Now(),
			},
			shouldBeInFuture: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := NewIntervalTracker(5 * time.Minute)
			tt.setupFunc(tracker, tt.entry)

			nextScanTime := tracker.GetNextScanTime(tt.entry)

			if tt.shouldBeNow {
				// Should be approximately now (within a second)
				assert.True(t, time.Since(nextScanTime) < 1*time.Second)
			}

			if tt.shouldBeInFuture {
				assert.True(t, nextScanTime.After(time.Now()))
			}
		})
	}
}

func TestShouldScanNow(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(*IntervalTracker, animelist.Entry)
		entry       animelist.Entry
		expected    bool
		description string
	}{
		{
			name: "first scan ever",
			setupFunc: func(it *IntervalTracker, entry animelist.Entry) {
				// Don't update state
			},
			entry: animelist.Entry{
				Titles: []string{"First Scan"},
			},
			expected:    true,
			description: "should scan if never scanned before",
		},
		{
			name: "just scanned",
			setupFunc: func(it *IntervalTracker, entry animelist.Entry) {
				it.UpdateState(entry, false)
			},
			entry: animelist.Entry{
				Titles:       []string{"Just Scanned"},
				AiringStatus: animelist.AiringStatusAiring,
				StartDate:    time.Now(),
			},
			expected:    true,
			description: "should scan within poll frequency window after update",
		},
		{
			name: "scan overdue",
			setupFunc: func(it *IntervalTracker, entry animelist.Entry) {
				tracker := it
				key := getShowKey(entry.Titles)
				tracker.mu.Lock()
				tracker.state[key] = ShowScanState{
					NextScanTime:     time.Now().Add(-1 * time.Minute),
					FoundNewEpisodes: false,
				}
				tracker.mu.Unlock()
			},
			entry: animelist.Entry{
				Titles: []string{"Overdue"},
			},
			expected:    true,
			description: "should scan if next scan time is in the past",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := NewIntervalTracker(5 * time.Minute)
			tt.setupFunc(tracker, tt.entry)

			result := tracker.ShouldScanNow(tt.entry)
			assert.Equal(t, tt.expected, result, tt.description)
		})
	}
}

func TestCalculateNextInterval(t *testing.T) {
	pollFrequency := 5 * time.Minute

	tests := []struct {
		name             string
		airingStatus     animelist.AiringStatus
		foundNewEpisodes bool
		episodeSchedule  []animelist.EpisodeSchedule
		startDate        time.Time
		expectedInterval time.Duration
		description      string
	}{
		{
			name:             "aired show with new episodes",
			airingStatus:     animelist.AiringStatusAired,
			foundNewEpisodes: true,
			expectedInterval: pollFrequency,
			description:      "aired show with new episodes should use base poll frequency",
		},
		{
			name:             "aired show without new episodes",
			airingStatus:     animelist.AiringStatusAired,
			foundNewEpisodes: false,
			expectedInterval: 24 * time.Hour,
			description:      "aired show without new episodes should use 24 hour interval",
		},
		{
			name:         "airing show with episode in <3 hours",
			airingStatus: animelist.AiringStatusAiring,
			startDate:    time.Now().AddDate(0, 0, -1),
			episodeSchedule: []animelist.EpisodeSchedule{
				{
					AirDate: time.Now().Add(1 * time.Hour),
				},
			},
			expectedInterval: pollFrequency,
			description:      "airing show with episode in <3 hours should use base poll frequency",
		},
		{
			name:         "airing show with episode in 3-24 hours",
			airingStatus: animelist.AiringStatusAiring,
			startDate:    time.Now().AddDate(0, 0, -1),
			episodeSchedule: []animelist.EpisodeSchedule{
				{
					AirDate: time.Now().Add(12 * time.Hour),
				},
			},
			expectedInterval: 1 * time.Hour,
			description:      "airing show with episode in 3-24 hours should use 1 hour interval",
		},
		{
			name:         "airing show with episode in >24 hours",
			airingStatus: animelist.AiringStatusAiring,
			startDate:    time.Now().AddDate(0, 0, -1),
			episodeSchedule: []animelist.EpisodeSchedule{
				{
					AirDate: time.Now().Add(48 * time.Hour),
				},
			},
			expectedInterval: 6 * time.Hour,
			description:      "airing show with episode in >24 hours should use 6 hour interval",
		},
		{
			name:         "airing show with no upcoming episodes",
			airingStatus: animelist.AiringStatusAiring,
			startDate:    time.Now().AddDate(0, 0, 3),
			episodeSchedule: []animelist.EpisodeSchedule{
				{
					AirDate: time.Now().Add(-24 * time.Hour),
				},
			},
			expectedInterval: 6 * time.Hour,
			description:      "airing show with no upcoming episodes should use 6 hour interval",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := NewIntervalTracker(pollFrequency)
			entry := animelist.Entry{
				Titles:          []string{tt.name},
				AiringStatus:    tt.airingStatus,
				StartDate:       tt.startDate,
				EpisodeSchedule: tt.episodeSchedule,
			}

			if tt.foundNewEpisodes && tt.airingStatus == animelist.AiringStatusAired {
				tracker.UpdateState(entry, true)
			}

			interval := tracker.calculateNextInterval(entry)
			assert.Equal(t, tt.expectedInterval, interval, tt.description)
		})
	}
}

func TestShouldScanNow_WithinPollFrequency(t *testing.T) {
	pollFrequency := 5 * time.Minute
	tracker := NewIntervalTracker(pollFrequency)

	entry := animelist.Entry{
		Titles:       []string{"Test"},
		AiringStatus: animelist.AiringStatusAiring,
		StartDate:    time.Now(),
	}

	// Manually set next scan time to be far in the future
	key := getShowKey(entry.Titles)
	tracker.mu.Lock()
	tracker.state[key] = ShowScanState{
		NextScanTime:     time.Now().Add(10 * time.Minute),
		FoundNewEpisodes: false,
	}
	tracker.mu.Unlock()

	// Should not scan since next scan time is outside poll frequency window
	assert.False(t, tracker.ShouldScanNow(entry))
}

func TestIntegration_FullScanCycle(t *testing.T) {
	pollFrequency := 5 * time.Minute
	tracker := NewIntervalTracker(pollFrequency)

	entry := animelist.Entry{
		Titles:       []string{"My Hero Academia"},
		AiringStatus: animelist.AiringStatusAiring,
		StartDate:    time.Now().AddDate(0, 0, -100),
		EpisodeSchedule: []animelist.EpisodeSchedule{
			{
				AirDate: time.Now().Add(2 * time.Hour),
			},
		},
	}

	// First scan - should be due now
	assert.True(t, tracker.ShouldScanNow(entry))
	assert.Equal(t, ShowScanState{}, tracker.getState(entry))

	// Update after finding new episodes
	nextScan := tracker.UpdateState(entry, true)
	assert.True(t, nextScan.After(time.Now()))

	// Should scan within poll frequency window
	assert.True(t, tracker.ShouldScanNow(entry))

	// Verify state was updated
	state := tracker.getState(entry)
	assert.True(t, state.FoundNewEpisodes)

	// Manually advance the next scan time
	key := getShowKey(entry.Titles)
	tracker.mu.Lock()
	tracker.state[key] = ShowScanState{
		NextScanTime:     time.Now().Add(-1 * time.Minute),
		FoundNewEpisodes: true,
	}
	tracker.mu.Unlock()

	// Should scan now after time advanced
	assert.True(t, tracker.ShouldScanNow(entry))
}
