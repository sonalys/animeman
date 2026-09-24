package metadata

import (
	"bytes"
	"encoding/json"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestComplexSeasonAndEpisodeNotation(t *testing.T) {
	tests := []struct {
		name  string
		input string
		eps   Tags
	}{
		{"multiple seasons", "[G] Show S1 S2 S3", []Tag{{Number: 1}, {Number: 2}, {Number: 3}}},
		{
			"multiple seasons with commas",
			"[G] Show S1, S2, S3",
			[]Tag{{Number: 1}, {Number: 2}, {Number: 3}},
		},
		{"season range", "[G] Show S1-3", []Tag{{Number: 1}, {Number: 2}, {Number: 3}}},
		{
			"season range with S prefix",
			"[G] Show S1-S3",
			[]Tag{{Number: 1}, {Number: 2}, {Number: 3}},
		},
		{
			"season range with E range",
			"[G] Show S1-3 E1-12",
			[]Tag{
				{Number: 1, Episodes: []EpisodeRange{{Start: 1, End: 12, Raw: "E1-12"}}},
				{Number: 2},
				{Number: 3},
			},
		},
		{
			"two episodes",
			"[G] Show S2E1E2",
			[]Tag{
				{
					Number:   2,
					Episodes: []EpisodeRange{{Start: 1, Raw: "E1"}, {Start: 2, Raw: "E2"}},
				},
			},
		},
		{
			"two episodes with title",
			"[G] Show S2E1E2 The Title",
			[]Tag{
				{
					Number:   2,
					Episodes: []EpisodeRange{{Start: 1, Raw: "E1"}, {Start: 2, Raw: "E2"}},
				},
			},
		},
		{
			"half episode",
			"[G] Show S1E6.5",
			[]Tag{
				{Number: 1, Episodes: []EpisodeRange{{Start: 6.5, Raw: "E6.5"}}},
			},
		},
		{
			"half episode with title",
			"[G] Show S1E6.5 Halfway There",
			[]Tag{
				{Number: 1, Episodes: []EpisodeRange{{Start: 6.5, Raw: "E6.5"}}},
			},
		},
		{
			"roman season colon episode",
			"[G] Show Season IV: 3",
			[]Tag{
				{Number: 4, Episodes: []EpisodeRange{{Start: 3, Raw: "Season IV: 3"}}},
			},
		},
		{
			"roman season dash episode",
			"[G] Show Season IV - 3",
			[]Tag{
				{Number: 4, Episodes: []EpisodeRange{{Start: 3, Raw: "Season IV - 3"}}},
			},
		},
		{
			"bare episode range",
			"[G] Show E1-12",
			[]Tag{
				{Number: 1, Episodes: []EpisodeRange{{Start: 1, End: 12, Raw: "E1-12"}}},
			},
		},
		{
			"multiple bare episodes",
			"[G] Show E1 E2 E3",
			[]Tag{
				{Number: 1, Episodes: []EpisodeRange{
					{Start: 1, Raw: "E1"},
					{Start: 2, Raw: "E2"},
					{Start: 3, Raw: "E3"},
				}},
			},
		},
		{
			"Season x Episode",
			"[G] Show 2x12",
			[]Tag{
				{Number: 2, Episodes: []EpisodeRange{{Start: 12, Raw: "12"}}},
			},
		},
		{
			"Season x Episode",
			"[G] Show 2x1~12",
			[]Tag{
				{Number: 2, Episodes: []EpisodeRange{{Start: 1, End: 12, Raw: "1~12"}}},
			},
		},
		{
			"Season x Episode",
			"[G] Show 2x1-12",
			[]Tag{
				{Number: 2, Episodes: []EpisodeRange{{Start: 1, End: 12, Raw: "1-12"}}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Parse(tt.input, 1, nil)
			if strings.HasPrefix(r.PrimaryTitle, "Show S") ||
				strings.Contains(r.PrimaryTitle, "E1") ||
				strings.Contains(r.PrimaryTitle, "E6") {
				t.Fatalf("primary title retained episode notation: %q", r.PrimaryTitle)
			}
			require.Equal(t, tt.eps, r.Tags)
		})
	}
}

func TestEpisodeTitleNotConfusedWithChainedOrHalfEpisodes(t *testing.T) {
	for _, tt := range []struct{ input, want string }{
		{"[G] Show S1E12 The Title", "The Title"},
		{"[G] Show S1E1E2 The Title", "The Title"},
		{"[G] Show S1E6.5 Halfway There", "Halfway There"},
	} {
		r := Parse(tt.input, 1, nil)
		if r.EpisodeTitle != tt.want {
			t.Fatalf("%q: episode title = %q, want %q", tt.input, r.EpisodeTitle, tt.want)
		}
	}
}

func TestTokenizerPreservesStructure(t *testing.T) {
	s := `[G] Title (JP; Alt) [1080p HEVC, AAC][ABC123] {Tags:A=ja;S=en,fr;}`
	toks := Tokenize(s)
	if len(toks) < 7 {
		t.Fatalf("too few tokens: %#v", toks)
	}
	if toks[0].Kind != TokenBracket || toks[0].Text != "G" {
		t.Fatalf("group token: %#v", toks[0])
	}
	if toks[1].Kind != TokenText || !strings.Contains(toks[1].Text, "Title") {
		t.Fatalf("text token: %#v", toks[1])
	}
	var kinds []TokenKind
	for _, x := range toks {
		kinds = append(kinds, x.Kind)
	}
	if kinds[len(kinds)-1] != TokenTagBlock {
		t.Fatalf("last token = %v", kinds[len(kinds)-1])
	}
}

type goldenResult struct {
	Group             string            `json:"group"`
	PrimaryTitle      string            `json:"primaryTitle"`
	AlternateTitles   []string          `json:"alternateTitles"`
	EpisodeTitle      string            `json:"episodeTitle"`
	Year              int               `json:"year"`
	Seasons           []Tag             `json:"seasons"`
	IsBatch           bool              `json:"isBatch"`
	IsComplete        bool              `json:"isComplete"`
	IsRemastered      bool              `json:"isRemastered"`
	IsRepack          bool              `json:"isRepack"`
	Resolutions       []string          `json:"resolutions"`
	Dimensions        []string          `json:"dimensions"`
	BitDepths         []string          `json:"bitDepths"`
	Sources           []string          `json:"sources"`
	Codecs            []string          `json:"codecs"`
	AudioCodecs       []string          `json:"audioCodecs"`
	AudioFlags        []string          `json:"audioFlags"`
	SubtitleFlags     []string          `json:"subtitleFlags"`
	SubtitleLanguages []string          `json:"subtitleLanguages"`
	AudioLanguages    []string          `json:"audioLanguages"`
	Encoders          []string          `json:"encoders"`
	ReleaseFlags      []string          `json:"releaseFlags"`
	Tags              []string          `json:"tags"`
	Checksum          string            `json:"checksum"`
	TagFields         map[string]string `json:"tagFields"`
	Unknown           []string          `json:"unknown"`
}

type corpusCase struct {
	Input    string       `json:"input"`
	Expected goldenResult `json:"expected"`
}

type corpusFile struct {
	Cases []corpusCase `json:"cases"`
}

func projectResult(r Metadata) goldenResult {
	return goldenResult{
		Group:             r.Group,
		PrimaryTitle:      r.PrimaryTitle,
		AlternateTitles:   r.AlternateTitles,
		EpisodeTitle:      r.EpisodeTitle,
		Year:              r.Year,
		Seasons:           r.Tags,
		IsBatch:           r.IsBatch,
		IsComplete:        r.IsComplete,
		IsRemastered:      r.IsRemastered,
		IsRepack:          r.IsRepack,
		Resolutions:       r.Resolutions,
		Dimensions:        r.Dimensions,
		BitDepths:         r.BitDepths,
		Sources:           r.Sources,
		Codecs:            r.Codecs,
		AudioCodecs:       r.AudioCodecs,
		AudioFlags:        r.AudioFlags,
		SubtitleFlags:     r.SubtitleFlags,
		SubtitleLanguages: r.SubtitleLanguages,
		AudioLanguages:    r.AudioLanguages,
		Encoders:          r.Encoders,
		ReleaseFlags:      r.ReleaseFlags,
		Tags:              r.Labels,
		Checksum:          r.Checksum,
		TagFields:         r.TagFields,
		Unknown:           r.Unknown,
	}
}

func TestCorpusTitlesAgainstJSON(t *testing.T) {
	files := []string{"testdata/nyaasi.sample.json", "testdata/nekobt.sample.json"}

	wd, _ := os.Getwd()

	for _, fn := range files {
		b, err := os.ReadFile(fn)
		if err != nil {
			t.Fatal(err)
		}
		var corpus corpusFile
		if err := json.Unmarshal(b, &corpus); err != nil {
			t.Fatalf("%s: %v", fn, err)
		}
		for i, tc := range corpus.Cases {
			t.Run(filepath.Base(fn)+"/"+strconv.Itoa(i), func(t *testing.T) {
				indexOf := bytes.Index(b, []byte(tc.Input))
				if indexOf > -1 {
					line := bytes.Count(b[:indexOf], []byte("\n"))
					t.Logf("%s:%d", path.Join(wd, fn), line)
				}

				r := Parse(tc.Input, 1, nil)
				require.Equal(t, tc.Input, r.Raw)
				got := projectResult(r)

				ss, _ := json.Marshal(got.Seasons)
				t.Logf("\"seasons\": %s", ss)

				require.Equal(t, tc.Expected, got)
			})
		}
	}
}

func eq[T comparable](t *testing.T, got, want T) {
	t.Helper()
	if got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}
func has(t *testing.T, xs []string, want string) {
	t.Helper()
	if slices.Contains(xs, want) {
		return
	}
	t.Fatalf("%q not found in %#v", want, xs)
}

func TestFilenameStyleTitles(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantTitle string
		season    int
		episode   string
		res       string
		codec     string
		audio     string
		group     string
	}{
		{
			"dot separated release",
			"Show.name.S02E19.subtitle.here.1080p.TAG.AAC2.0.H.264-VARYG.mkv",
			"Show name",
			2,
			"19",
			"1080p",
			"H.264",
			"AAC2.0",
			"VARYG",
		},
		{"tilde season range", "Show.S1~12", "Show", 0, "", "", "", "", ""},
		{"tilde episode range", "Show.S1E1~12", "Show", 1, "1", "", "", "", ""},
		{
			"decimal episode filename",
			"Show.S01E6.5.Half.Episode.1080p.WEB-DL.mkv",
			"Show",
			1,
			"6.5",
			"1080p",
			"",
			"",
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Parse(tt.input, 1, nil)
			eq(t, r.PrimaryTitle, tt.wantTitle)
			if tt.res != "" {
				has(t, r.Resolutions, tt.res)
			}
			if tt.codec != "" {
				has(t, r.Codecs, tt.codec)
			}
			if tt.audio != "" {
				has(t, r.AudioCodecs, tt.audio)
			}
			if tt.group != "" {
				eq(t, r.Group, tt.group)
			}
		})
	}
}

func TestMetadataCompareSeasonPackBeforeEpisodeRelease(t *testing.T) {
	tests := []struct {
		name   string
		first  Metadata
		second Metadata
		want   int
	}{
		{
			name: "season pack before episode",
			first: Metadata{
				PrimaryTitle: "Show",
				Tags: []Tag{
					{Number: 2},
				},
			},
			second: Metadata{
				PrimaryTitle: "Show",
				Tags: []Tag{
					{
						Number: 2,
						Episodes: []EpisodeRange{
							{Start: 1},
						},
					},
				},
			},
			want: 1,
		},
		{
			name: "season pack before episode range",
			first: Metadata{
				PrimaryTitle: "Show",
				Tags: []Tag{
					{Number: 2},
				},
			},
			second: Metadata{
				PrimaryTitle: "Show",
				Tags: []Tag{
					{
						Number: 2,
						Episodes: []EpisodeRange{
							{Start: 1, End: 12},
						},
					},
				},
			},
			want: 1,
		},
		{
			name: "episode is after season pack",
			first: Metadata{
				PrimaryTitle: "Show",
				Tags: []Tag{
					{
						Number: 2,
						Episodes: []EpisodeRange{
							{Start: 1},
						},
					},
				},
			},
			second: Metadata{
				PrimaryTitle: "Show",
				Tags: []Tag{
					{Number: 2},
				},
			},
			want: -1,
		},
		{
			name: "same season pack",
			first: Metadata{
				PrimaryTitle: "Show",
				Tags: []Tag{
					{Number: 2},
				},
			},
			second: Metadata{
				PrimaryTitle: "Show",
				Tags: []Tag{
					{Number: 2},
				},
			},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.first.Tags.Compare(tt.second.Tags)
			if got != tt.want {
				t.Fatalf("Compare() = %d, want %d", got, tt.want)
			}
		})
	}
}
