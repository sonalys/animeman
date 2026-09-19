package parser

import (
	"testing"

	"github.com/sonalys/animeman/internal/ports/animelist"
	"github.com/sonalys/animeman/internal/ports/torrentsource"
)

func Test_qualityScore(t *testing.T) {
	qualities := []string{"1080 HEVC", "1080", "720"}
	tests := []struct {
		name  string
		title string
		want  int
	}{
		{
			name:  "matches all keywords of first quality",
			title: "[EMBER] Show [1080p HEVC].mkv",
			want:  0,
		},
		{
			name:  "case insensitive",
			title: "Show S01E01 1080p hevc",
			want:  0,
		},
		{
			name:  "partial match of first quality falls to second",
			title: "Show S01E01 1080p x264",
			want:  1,
		},
		{
			name:  "matches third quality",
			title: "Show S01E01 720p",
			want:  2,
		},
		{
			name:  "no match",
			title: "Show S01E01",
			want:  len(qualities),
		},
		{
			// "1080 HEVC" has more words than "1080", but "1080" is higher priority,
			// so it must win over a title matching only "1080 HEVC" keywords twice.
			name:  "multi-word quality cannot outweigh higher quality",
			title: "Show 1080 HEVC HEVC",
			want:  0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := qualityScore(tt.title, qualities); got != tt.want {
				t.Errorf("qualityScore() = %d, want %d", got, tt.want)
			}
		})
	}
}

func Test_Prioritize_qualityPriority(t *testing.T) {
	qualities := []string{"1080 HEVC", "1080", "720"}

	torrent := func(title string, seeders int) torrentsource.Torrent {
		return torrentsource.Torrent{Title: title, Seeders: seeders}
	}

	t.Run("full quality match beats partial match and more seeders", func(t *testing.T) {
		results := []torrentsource.Torrent{
			torrent("Show S01E01 1080p x264", 100),
			torrent("Show S01E01 1080p HEVC", 1),
			torrent("Show S01E01 720p", 1000),
		}
		got := Prioritize(
			animelist.Entry{},
			results,
			torrentsource.SearchOptions{Qualities: qualities},
		)
		if got[0].Title != "Show S01E01 1080p HEVC" {
			t.Errorf("expected 1080p HEVC first, got %q", got[0].Title)
		}
	})

	t.Run("higher quality beats lower quality regardless of word count", func(t *testing.T) {
		results := []torrentsource.Torrent{
			torrent("Show S01E01 1080p HEVC", 100),
			torrent("Show S01E01 1080p", 100),
		}
		got := Prioritize(
			animelist.Entry{},
			results,
			torrentsource.SearchOptions{Qualities: qualities},
		)
		if got[0].Title != "Show S01E01 1080p HEVC" {
			t.Errorf("expected 1080p HEVC first, got %q", got[0].Title)
		}
	})

	t.Run("no qualities configured falls back to resolution", func(t *testing.T) {
		results := []torrentsource.Torrent{
			torrent("Show S01E01 720p", 1),
			torrent("Show S01E01 1080p HEVC", 1),
		}
		got := Prioritize(animelist.Entry{}, results, torrentsource.SearchOptions{})
		if got[0].Title != "Show S01E01 1080p HEVC" {
			t.Errorf("expected 1080p HEVC first by resolution, got %q", got[0].Title)
		}
	})
}
