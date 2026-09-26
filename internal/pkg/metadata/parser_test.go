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

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComplexSeasonAndEpisodeNotation(t *testing.T) {
	tests := []struct {
		name  string
		input string
		eps   Tags
	}{
		{
			"Season - Episode",
			"[Erai-raws] Hell Mode: Yarikomizuki no Gamer wa Hai Settei no Isekai de Musou suru S2 - 13 [1080p HIDIVE WEB-DL AVC AAC][E1C62DE2] {Tags:L0;V9;C1;A=ja;S=en;}",
			[]Tag{
				{Number: 2, Episodes: []EpisodeRange{{Start: 13}}},
			},
		},
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
				{Number: 1, Episodes: []EpisodeRange{{Start: 1, End: 12}}},
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
					Episodes: []EpisodeRange{{Start: 1}, {Start: 2}},
				},
			},
		},
		{
			"two episodes with title",
			"[G] Show S2E1E2 The Title",
			[]Tag{
				{
					Number:   2,
					Episodes: []EpisodeRange{{Start: 1}, {Start: 2}},
				},
			},
		},
		{
			"half episode",
			"[G] Show S1E6.5",
			[]Tag{
				{Number: 1, Episodes: []EpisodeRange{{Start: 6.5}}},
			},
		},
		{
			"half episode with title",
			"[G] Show S1E6.5 Halfway There",
			[]Tag{
				{Number: 1, Episodes: []EpisodeRange{{Start: 6.5}}},
			},
		},
		{
			"roman season colon episode",
			"[G] Show Season IV: 3",
			[]Tag{
				{Number: 4, Episodes: []EpisodeRange{{Start: 3}}},
			},
		},
		{
			"roman season dash episode",
			"[G] Show Season IV - 3",
			[]Tag{
				{Number: 4, Episodes: []EpisodeRange{{Start: 3}}},
			},
		},
		{
			"bare episode range",
			"[G] Show E1-12",
			[]Tag{
				{Number: 1, Episodes: []EpisodeRange{{Start: 1, End: 12}}},
			},
		},
		{
			"multiple bare episodes",
			"[G] Show E1 E2 E3",
			[]Tag{
				{Number: 1, Episodes: []EpisodeRange{
					{Start: 1},
					{Start: 2},
					{Start: 3},
				}},
			},
		},
		{
			"Season x Episode",
			"[G] Show 2x12",
			[]Tag{
				{Number: 2, Episodes: []EpisodeRange{{Start: 12}}},
			},
		},
		{
			"Season x Episode",
			"[G] Show 2x1~12",
			[]Tag{
				{Number: 2, Episodes: []EpisodeRange{{Start: 1, End: 12}}},
			},
		},
		{
			"Season x Episode",
			"[G] Show 2x1-12",
			[]Tag{
				{Number: 2, Episodes: []EpisodeRange{{Start: 1, End: 12}}},
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
	Group             string            `json:"group,omitzero"`
	PrimaryTitle      string            `json:"primaryTitle,omitzero"`
	AlternateTitles   []string          `json:"alternateTitles,omitempty"`
	EpisodeTitle      string            `json:"episodeTitle,omitzero"`
	Year              int               `json:"year,omitzero"`
	Tags              []Tag             `json:"seasons,omitempty"`
	IsBatch           bool              `json:"isBatch,omitzero"`
	IsComplete        bool              `json:"isComplete,omitzero"`
	IsRemastered      bool              `json:"isRemastered,omitzero"`
	IsRepack          bool              `json:"isRepack,omitzero"`
	Resolutions       []string          `json:"resolutions,omitempty"`
	Dimensions        []string          `json:"dimensions,omitempty"`
	BitDepths         []string          `json:"bitDepths,omitempty"`
	Sources           []string          `json:"sources,omitempty"`
	Codecs            []string          `json:"codecs,omitempty"`
	AudioCodecs       []string          `json:"audioCodecs,omitempty"`
	AudioFlags        []string          `json:"audioFlags,omitempty"`
	SubtitleFlags     []string          `json:"subtitleFlags,omitempty"`
	SubtitleLanguages []string          `json:"subtitleLanguages,omitempty"`
	AudioLanguages    []string          `json:"audioLanguages,omitempty"`
	Encoders          []string          `json:"encoders,omitempty"`
	ReleaseFlags      []string          `json:"releaseFlags,omitempty"`
	Labels            []string          `json:"tags,omitempty"`
	Checksum          string            `json:"checksum,omitzero"`
	TagFields         map[string]string `json:"tagFields,omitempty"`
	Unknown           []string          `json:"unknown,omitempty"`
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
		Tags:              r.Tags,
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
		Labels:            r.Labels,
		Checksum:          r.Checksum,
		TagFields:         r.TagFields,
		Unknown:           r.Unknown,
	}
}

func TestCorpusTitlesAgainstJSON(t *testing.T) {
	files := []string{"testdata/nyaasi.sample.json", "testdata/nekobt.sample.json"}

	wd, _ := os.Getwd()

	for _, fn := range files {
		buffer, err := os.ReadFile(fn)
		if err != nil {
			t.Fatal(err)
		}

		var corpus corpusFile
		if err := json.Unmarshal(buffer, &corpus); err != nil {
			t.Fatalf("%s: %v", fn, err)
		}

		for i, tc := range corpus.Cases {
			t.Run(filepath.Base(fn)+"/"+strconv.Itoa(i), func(t *testing.T) {
				indexOf := bytes.Index(buffer, []byte(tc.Input))
				if indexOf > -1 {
					line := bytes.Count(buffer[:indexOf], []byte("\n"))
					t.Logf("%s:%d", path.Join(wd, fn), line)
				}

				r := Parse(tc.Input, 1, nil)
				got := projectResult(r)

				assert.Empty(t, cmp.Diff(tc.Expected, got, cmpopts.EquateEmpty()))

				primaryTitle := ParsePrimaryTitle(tc.Input)
				assert.Equal(t, r.PrimaryTitle, primaryTitle)

				// corpus.Cases[i].Expected = got
			})
		}

		// Utility to override expected results.
		// file, err := os.Create(fn)
		// require.NoError(t, err)
		// defer file.Close()

		// encoder := json.NewEncoder(file)
		// encoder.SetIndent("", "\t")

		// err = encoder.Encode(corpus)
		// require.NoError(t, err)
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

func Test_TitleParser(t *testing.T) {
	input := "Ascendance.of.a.Bookworm.S04E23.Charlottes.Baptism.1080p.CR.WEB-DL.JPN.AAC2.0.H.264.MSubs-ToonsHub.mkv"
	got := ParsePrimaryTitle(input)

	want := "Ascendance of a Bookworm"
	require.Equal(t, want, got)
}
