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
	tests := []struct {
		name             string
		pollFrequency    time.Duration
		entry            animelist.Entry
		setupState       func(*IntervalTracker, animelist.Entry)
		expectedInterval time.Duration
		description      string
	}{
		// FoundNewEpisodes scenarios
		{
			name:          "found new episodes returns poll frequency",
			pollFrequency: 5 * time.Minute,
			entry: animelist.Entry{
				Titles: []string{"New Episodes Found"},
			},
			setupState: func(it *IntervalTracker, entry animelist.Entry) {
				key := getShowKey(entry.Titles)
				it.mu.Lock()
				it.state[key] = ShowScanState{
					NextScanTime:     time.Now().Add(1 * time.Hour),
					FoundNewEpisodes: true,
				}
				it.mu.Unlock()
			},
			expectedInterval: 5 * time.Minute,
			description:      "should return poll frequency when FoundNewEpisodes is true",
		},

		// No episode schedule - Airing status
		{
			name:          "no episodes, airing status",
			pollFrequency: 5 * time.Minute,
			entry: animelist.Entry{
				Titles:          []string{"Currently Airing"},
				AiringStatus:    animelist.AiringStatusAiring,
				EpisodeSchedule: []animelist.EpisodeSchedule{},
			},
			expectedInterval: 5 * time.Minute,
			description:      "should return poll frequency for airing show with no episode schedule",
		},

		// No episode schedule - Aired status with end date scenarios
		{
			name:          "no episodes, aired status, no end date",
			pollFrequency: 5 * time.Minute,
			entry: animelist.Entry{
				Titles:          []string{"Completed Show"},
				AiringStatus:    animelist.AiringStatusAired,
				StartDate:       time.Now().Add(-100 * 24 * time.Hour),
				EndDate:         time.Time{}, // zero value
				EpisodeSchedule: []animelist.EpisodeSchedule{},
			},
			expectedInterval: 5 * time.Minute,
			description:      "should check recently ended shows frequently when no end date provided",
		},
		{
			name:          "no episodes, aired status, end date more than 7 days ago",
			pollFrequency: 5 * time.Minute,
			entry: animelist.Entry{
				Titles:          []string{"Old Show"},
				AiringStatus:    animelist.AiringStatusAired,
				StartDate:       time.Now().AddDate(0, 0, -200),
				EndDate:         time.Now().AddDate(0, 0, -20),
				EpisodeSchedule: []animelist.EpisodeSchedule{},
			},
			expectedInterval: 7 * 24 * time.Hour,
			description:      "should return 7 days for old completed shows",
		},
		{
			name:          "no episodes, aired status, end date within 7 days",
			pollFrequency: 5 * time.Minute,
			entry: animelist.Entry{
				Titles:          []string{"Recently Completed"},
				AiringStatus:    animelist.AiringStatusAired,
				StartDate:       time.Now().AddDate(0, 0, -30),
				EndDate:         time.Now().AddDate(0, 0, -2),
				EpisodeSchedule: []animelist.EpisodeSchedule{},
			},
			expectedInterval: 5 * time.Minute,
			description:      "should poll frequently for recently completed shows",
		},
		{
			name:          "no episodes, aired status, end date exactly 7 days ago (boundary)",
			pollFrequency: 5 * time.Minute,
			entry: animelist.Entry{
				Titles:          []string{"Boundary Show"},
				AiringStatus:    animelist.AiringStatusAired,
				StartDate:       time.Now().AddDate(0, 0, -40),
				EndDate:         time.Now().AddDate(0, 0, -6).Add(-20 * time.Hour),
				EpisodeSchedule: []animelist.EpisodeSchedule{},
			},
			expectedInterval: 5 * time.Minute,
			description:      "should use poll frequency when within 7 days of now",
		},

		// Episode schedule scenarios
		{
			name:          "start date within 3 hours",
			pollFrequency: 5 * time.Minute,
			entry: animelist.Entry{
				Titles:          []string{"Test Anime"},
				StartDate:       time.Now().Add(1 * time.Hour),
				EpisodeSchedule: []animelist.EpisodeSchedule{},
			},
			expectedInterval: 5 * time.Minute,
			description:      "should return poll frequency when show starts within 3 hours",
		},
		{
			name:          "start date within 24 hours but after 3 hours",
			pollFrequency: 5 * time.Minute,
			entry: animelist.Entry{
				Titles:          []string{"Test Anime"},
				StartDate:       time.Now().Add(12 * time.Hour),
				EpisodeSchedule: []animelist.EpisodeSchedule{},
			},
			expectedInterval: 1 * time.Hour,
			description:      "should return 1 hour when show starts between 3 and 24 hours",
		},
		{
			name:          "start date far in future",
			pollFrequency: 5 * time.Minute,
			entry: animelist.Entry{
				Titles:          []string{"Test Anime"},
				StartDate:       time.Now().AddDate(0, 0, 7),
				EpisodeSchedule: []animelist.EpisodeSchedule{},
			},
			expectedInterval: 24 * time.Hour,
			description:      "should return 24 hours when show starts more than 24 hours away",
		},
		{
			name:          "episode schedule within 3 hours",
			pollFrequency: 5 * time.Minute,
			entry: animelist.Entry{
				Titles:    []string{"Test Anime"},
				StartDate: time.Now().AddDate(0, 0, 1),
				EpisodeSchedule: []animelist.EpisodeSchedule{
					{
						AirDate: time.Now().Add(2 * time.Hour),
					},
				},
			},
			expectedInterval: 5 * time.Minute,
			description:      "should use episode schedule when closer than start date",
		},
		{
			name:          "episode schedule within 24 hours",
			pollFrequency: 5 * time.Minute,
			entry: animelist.Entry{
				Titles:    []string{"Test Anime"},
				StartDate: time.Now().AddDate(0, 0, 1),
				EpisodeSchedule: []animelist.EpisodeSchedule{
					{
						AirDate: time.Now().Add(6 * time.Hour),
					},
				},
			},
			expectedInterval: 1 * time.Hour,
			description:      "should return 1 hour when episode airs within 24 hours",
		},
		{
			name:          "multiple episodes, closest one determines interval",
			pollFrequency: 5 * time.Minute,
			entry: animelist.Entry{
				Titles:    []string{"Test Anime"},
				StartDate: time.Now().AddDate(0, 0, 7),
				EpisodeSchedule: []animelist.EpisodeSchedule{
					{
						AirDate: time.Now().AddDate(0, 0, 5),
					},
					{
						AirDate: time.Now().Add(2 * time.Hour),
					},
					{
						AirDate: time.Now().AddDate(0, 0, 10),
					},
				},
			},
			expectedInterval: 5 * time.Minute,
			description:      "should find the closest episode air date",
		},
		{
			name:          "past start date with future episodes",
			pollFrequency: 5 * time.Minute,
			entry: animelist.Entry{
				Titles:    []string{"Test Anime"},
				StartDate: time.Now().Add(-30 * 24 * time.Hour),
				EpisodeSchedule: []animelist.EpisodeSchedule{
					{
						AirDate: time.Now().Add(2 * time.Hour),
					},
				},
			},
			expectedInterval: 5 * time.Minute,
			description:      "should handle past start dates and use episodes correctly",
		},
		{
			name:          "all episodes in the past",
			pollFrequency: 5 * time.Minute,
			entry: animelist.Entry{
				Titles:    []string{"Test Anime"},
				StartDate: time.Now().Add(-30 * 24 * time.Hour),
				EpisodeSchedule: []animelist.EpisodeSchedule{
					{
						AirDate: time.Now().Add(-5 * 24 * time.Hour),
					},
					{
						AirDate: time.Now().Add(-2 * 24 * time.Hour),
					},
				},
			},
			expectedInterval: 24 * time.Hour,
			description:      "should use absolute values for past dates",
		},
		{
			name:          "edge case: just under 3 hours",
			pollFrequency: 5 * time.Minute,
			entry: animelist.Entry{
				Titles:          []string{"Test Anime"},
				StartDate:       time.Now().Add(3*time.Hour - 1*time.Second),
				EpisodeSchedule: []animelist.EpisodeSchedule{},
			},
			expectedInterval: 5 * time.Minute,
			description:      "should return poll frequency when just under 3 hours away",
		},
		{
			name:          "edge case: just over 3 hours",
			pollFrequency: 5 * time.Minute,
			entry: animelist.Entry{
				Titles:          []string{"Test Anime"},
				StartDate:       time.Now().Add(3*time.Hour + 1*time.Minute),
				EpisodeSchedule: []animelist.EpisodeSchedule{},
			},
			expectedInterval: 1 * time.Hour,
			description:      "should return 1 hour when just over 3 hours away",
		},
		{
			name:          "edge case: just under 24 hours",
			pollFrequency: 5 * time.Minute,
			entry: animelist.Entry{
				Titles:          []string{"Test Anime"},
				StartDate:       time.Now().Add(24*time.Hour - 1*time.Second),
				EpisodeSchedule: []animelist.EpisodeSchedule{},
			},
			expectedInterval: 1 * time.Hour,
			description:      "should return 1 hour when just under 24 hours away",
		},
		{
			name:          "edge case: just over 24 hours",
			pollFrequency: 5 * time.Minute,
			entry: animelist.Entry{
				Titles:          []string{"Test Anime"},
				StartDate:       time.Now().Add(24*time.Hour + 1*time.Minute),
				EpisodeSchedule: []animelist.EpisodeSchedule{},
			},
			expectedInterval: 24 * time.Hour,
			description:      "should return 24 hours when just over 24 hours away",
		},
		{
			name:          "different poll frequency",
			pollFrequency: 10 * time.Minute,
			entry: animelist.Entry{
				Titles:          []string{"Test Anime"},
				StartDate:       time.Now().Add(1 * time.Hour),
				EpisodeSchedule: []animelist.EpisodeSchedule{},
			},
			expectedInterval: 10 * time.Minute,
			description:      "should respect the configured poll frequency",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := NewIntervalTracker(tt.pollFrequency)
			if tt.setupState != nil {
				tt.setupState(tracker, tt.entry)
			}
			interval := tracker.calculateNextInterval(tt.entry)

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
