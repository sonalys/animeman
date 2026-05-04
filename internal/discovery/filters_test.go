package discovery

import (
	"testing"
	"time"

	"github.com/sonalys/animeman/internal/integrations/nyaa"
	"github.com/sonalys/animeman/internal/utils"
	"github.com/sonalys/animeman/pkg/v1/animelist"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_matchTitle(t *testing.T) {
	testCases := []struct {
		name string
		a, b string
		want bool
	}{
		{
			name: "do not match different titles",
			a:    "anime 1",
			b:    "anime 2",
			want: false,
		},
		{
			name: "match prefix",
			a:    "anime 1: subtitle",
			b:    "anime 1",
			want: true,
		},
		{
			name: "ignore special characters",
			a:    "anime 1/2",
			b:    "anime 1-2",
			want: true,
		},
		{
			name: "ignore space",
			a:    "anime1/2",
			b:    "anime 1-2",
			want: true,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			match := utils.MatchPrefixFlexible(tC.a, tC.b, ignoreCharset)
			require.Equal(t, tC.want, match)
		})
	}
}

func Test_filterMetadata(t *testing.T) {
	// Helper to create RFC1123Z time strings
	fmtTime := func(t time.Time) string { return t.Format(time.RFC1123Z) }
	now := time.Now().Truncate(time.Second)

	tests := []struct {
		name       string
		entry      animelist.Entry
		nyaaItem   nyaa.Item
		wantResult bool
	}{
		{
			name: "Valid entry - title match and date after start",
			entry: animelist.Entry{
				Titles:      []string{"Frieren"},
				StartDate:   now.AddDate(0, 0, -1),
				NumEpisodes: 28,
			},
			nyaaItem: nyaa.Item{
				Title:   "[Subs] Frieren - 01.mkv",
				PubDate: fmtTime(now),
			},
			wantResult: true,
		},
		{
			name: "Invalid - published before start date (beyond 2-day offset)",
			entry: animelist.Entry{
				Titles:    []string{"Frieren"},
				StartDate: now,
			},
			nyaaItem: nyaa.Item{
				Title:   "[Subs] Frieren - 01.mkv",
				PubDate: fmtTime(now.AddDate(0, 0, -5)),
			},
			wantResult: false,
		},
		{
			name: "Invalid - episode number exceeds series total",
			entry: animelist.Entry{
				Titles:      []string{"Frieren"},
				StartDate:   now.AddDate(0, 0, -1),
				NumEpisodes: 12,
			},
			nyaaItem: nyaa.Item{
				Title:   "[Subs] Frieren - 13.mkv",
				PubDate: fmtTime(now),
			},
			wantResult: false,
		},
		{
			name: "Valid - Flexible title matching with charset characters",
			entry: animelist.Entry{
				Titles: []string{"Frieren: Beyond Journey's End"},
			},
			nyaaItem: nyaa.Item{
				Title:   "Frieren - Beyond Journey's End - 01",
				PubDate: fmtTime(now),
			},
			wantResult: true,
		},
		{
			name: "Invalid - Completely different title",
			entry: animelist.Entry{
				Titles: []string{"One Piece"},
			},
			nyaaItem: nyaa.Item{
				Title:   "Naruto - 01",
				PubDate: fmtTime(now),
			},
			wantResult: false,
		},
		{
			name: "Valid - Season to the original title",
			entry: animelist.Entry{
				Titles: []string{"One Piece 2nd season"},
			},
			nyaaItem: nyaa.Item{
				Title:   "One Piece",
				PubDate: fmtTime(now),
			},
			wantResult: true,
		},
		{
			name: "Valid - Trailing number should still match",
			entry: animelist.Entry{
				Titles: []string{"One Piece 2nd season"},
			},
			nyaaItem: nyaa.Item{
				Title:   "One Piece 2nd season",
				PubDate: fmtTime(now),
			},
			wantResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := filterMetadata(tt.entry, &FilterData{DiscardedMap: make(map[DiscardReason]uint)})
			got := filter(tt.nyaaItem)
			assert.Equal(t, tt.wantResult, got)
		})
	}
}
