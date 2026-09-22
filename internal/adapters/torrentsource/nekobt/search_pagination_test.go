package nekobt

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sonalys/animeman/internal/ports/animelist"
	"github.com/sonalys/animeman/internal/ports/torrentsource"
	"github.com/sonalys/animeman/internal/tags"
	"github.com/stretchr/testify/require"
)

func Test_shouldPaginate(t *testing.T) {
	tests := []struct {
		name      string
		items     []item
		latestTag tags.Tag
		want      bool
	}{
		{
			name:      "no latest tag, never paginate",
			items:     []item{{Title: "Show S1E1"}},
			latestTag: tags.Tag{},
			want:      false,
		},
		{
			name:      "empty page, stop",
			items:     nil,
			latestTag: tags.Tag{Seasons: []int{1}, Episodes: []float64{5}},
			want:      false,
		},
		{
			// A result at or below the latest downloaded tag ends the search.
			name:      "smallest older than latest, stop",
			items:     []item{{Title: "Show S1E1"}, {Title: "Show S1E3"}},
			latestTag: tags.Tag{Seasons: []int{1}, Episodes: []float64{5}},
			want:      false,
		},
		{
			// Smallest tag S1E5 equals latest: nothing older is needed.
			name:      "smallest equals latest, stop",
			items:     []item{{Title: "Show S1E5"}, {Title: "Show S1E7"}},
			latestTag: tags.Tag{Seasons: []int{1}, Episodes: []float64{5}},
			want:      false,
		},
		{
			// Smallest tag S1E7 is newer than latest S1E5: paginate.
			name:      "smallest newer than latest, paginate",
			items:     []item{{Title: "Show S1E7"}},
			latestTag: tags.Tag{Seasons: []int{1}, Episodes: []float64{5}},
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, shouldPaginate(tt.items, tt.latestTag))
		})
	}
}

// feed renders a torznab rss response with the given episode numbers.
func feed(eps ...int) string {
	items := ""
	for _, ep := range eps {
		items += fmt.Sprintf(`
			<item>
				<title>Show S1E%d</title>
				<link>https://example.com/%d</link>
				<pubDate>Mon, 01 Jan 2024 00:00:00 +0000</pubDate>
				<attr name="seeders" value="1"/>
				<attr name="infohash" value="hash%d"/>
			</item>`, ep, ep, ep)
	}
	return fmt.Sprintf(`<rss><channel><title>test</title>%s</channel></rss>`, items)
}

// resolveHandler serves the JSON API /media/resolve endpoint.
func resolveHandler(t *testing.T, requested *[]string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/media/resolve" {
			id := r.URL.Query().Get("id")
			*requested = append(*requested, id)
			w.Write([]byte(`{"id":"s1"}`))
			return
		}
		t.Errorf("unexpected request: %s", r.URL)
	}
}

func Test_Search_paginates(t *testing.T) {
	// Page 0 only has results newer than the latest torrent, so the next page
	// is requested before stopping at its empty feed.
	pages := map[int]string{
		0:        feed(3, 4),
		pageSize: feed(),
	}

	var requested []int
	var resolved []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/media/resolve" {
			resolved = append(resolved, r.URL.Query().Get("id"))
			w.Write([]byte(`{"id":"s1"}`))
			return
		}
		offset := r.URL.Query().Get("offset")
		limit := r.URL.Query().Get("limit")
		require.Equal(t, "100", limit)

		var page int
		fmt.Sscanf(offset, "%d", &page)
		requested = append(requested, page)
		w.Write([]byte(pages[page]))
	}))
	defer srv.Close()

	api := New(srv.Client(), Config{})
	// Point the adapter at the test server.
	previousURL := TORZNAB_URL
	previousJSON := JSON_URL
	t.Cleanup(func() { TORZNAB_URL = previousURL; JSON_URL = previousJSON })
	TORZNAB_URL = srv.URL
	JSON_URL = srv.URL

	torrents, err := api.Search(context.Background(), entryWithID(), torrentsource.SearchOptions{
		LatestTag: tags.Tag{Seasons: []int{1}, Episodes: []float64{2}},
	})
	require.NoError(t, err)
	require.Equal(t, []int{0, pageSize}, requested)
	// The anilist id was resolved exactly once, then cached.
	require.Equal(t, []string{"anilist-123"}, resolved)
	// Episodes 3 and 4 are newer than latest S1E2.
	require.Len(t, torrents, 2)
}

func Test_Search_stopsWhenPageHasNewer(t *testing.T) {
	pages := map[int]string{
		0: feed(1, 2),
	}

	var requested []int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/media/resolve" {
			w.Write([]byte(`{"id":"s1"}`))
			return
		}
		var page int
		fmt.Sscanf(r.URL.Query().Get("offset"), "%d", &page)
		requested = append(requested, page)
		w.Write([]byte(pages[page]))
	}))
	defer srv.Close()

	api := New(srv.Client(), Config{})
	previousURL := TORZNAB_URL
	previousJSON := JSON_URL
	t.Cleanup(func() { TORZNAB_URL = previousURL; JSON_URL = previousJSON })
	TORZNAB_URL = srv.URL
	JSON_URL = srv.URL

	_, err := api.Search(context.Background(), entryWithID(), torrentsource.SearchOptions{
		LatestTag: tags.Tag{Seasons: []int{1}, Episodes: []float64{2}},
	})
	require.NoError(t, err)
	// First page already contains newer episodes, no second request.
	require.Equal(t, []int{0}, requested)
}

func Test_Search_emptyPageStops(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/media/resolve" {
			w.Write([]byte(`{"id":"s1"}`))
			return
		}
		w.Write([]byte(feed()))
	}))
	defer srv.Close()

	api := New(srv.Client(), Config{})
	previousURL := TORZNAB_URL
	previousJSON := JSON_URL
	t.Cleanup(func() { TORZNAB_URL = previousURL; JSON_URL = previousJSON })
	TORZNAB_URL = srv.URL
	JSON_URL = srv.URL

	torrents, err := api.Search(context.Background(), entryWithID(), torrentsource.SearchOptions{
		LatestTag: tags.Tag{Seasons: []int{1}, Episodes: []float64{2}},
	})
	require.NoError(t, err)
	require.Empty(t, torrents)
}

// entryWithID returns an entry with an anilist id so resolveMediaID works.
func entryWithID() animelist.Entry {
	return animelist.Entry{
		AnilistID: 123,
		Titles:    []string{"Show"},
	}
}
