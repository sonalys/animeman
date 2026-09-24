package metadata

import (
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

type EpisodeRange struct {
	Start float32
	End   float32
	Raw   string
}

type Tag struct {
	Number   int            `json:"Number"`
	Episodes []EpisodeRange `json:"Episodes"`
}

type Metadata struct {
	Raw               string
	Group             string
	PrimaryTitle      string
	AlternateTitles   []string
	EpisodeTitle      string
	Year              int
	Tags              Tags
	IsBatch           bool
	IsComplete        bool
	IsRemastered      bool
	IsRepack          bool
	Resolutions       Resolutions
	Dimensions        []string
	BitDepths         []string
	Sources           []string
	Codecs            []string
	AudioCodecs       []string
	AudioFlags        []string
	SubtitleFlags     []string
	SubtitleLanguages []string
	AudioLanguages    []string
	Encoders          []string
	ReleaseFlags      []string
	Labels            []string
	Checksum          string
	TagFields         map[string]string
	Unknown           []string
	Tokens            []Token
}

var (
	seasonEpRE = regexp.MustCompile(
		`(?i)\bS(\d{1,3})\s*E(\d+(?:\.\d+)?)(?:\s*[-~]\s*(?:E)?(\d+(?:\.\d+)?))?(?:E\d+(?:\.\d+)?)*\b`,
	)

	seasonRangeRE = regexp.MustCompile(
		`(?i)\bS(\d{1,3})\s*(?:-{1,2}|~{1})\s*S?(\d{1,3})\b`,
	)

	seasonOnlyRE = regexp.MustCompile(
		`(?i)\bS(\d{1,3})\b`,
	)

	episodeRE = regexp.MustCompile(
		`(?i)E(\d+(?:\.\d+)?)(?:\s*(?:-|~)\s*E?(\d+(?:\.\d+)?))?`,
	)

	// Numeric season/episode form, e.g. "2 - 12".
	numericSeasonEpisodeRE = regexp.MustCompile(`(?i)\b([1-9]\d?)\s*-\s*(\d{1,4})\b`)

	// Season x Episode form, e.g. "2x1", "2x1-12", "2x1~12".
	seasonXEpisodeRE = regexp.MustCompile(
		`(?i)\b([1-9]\d?)\s*[x×]\s*(\d{1,4})(?:\s*[-~]\s*(\d{1,4}))?\b`,
	)

	// Season 4 / Season IV
	seasonWordRE = regexp.MustCompile(
		`(?i)\bseason\s+(\d{1,3}|[IVXLCDM]+)\b`,
	)

	// 4th season
	seasonOrdinalRE = regexp.MustCompile(
		`(?i)\b(\d{1,3})(?:st|nd|rd|th)\s+season\b`,
	)

	// IV season
	seasonRomanWordRE = regexp.MustCompile(
		`(?i)\b([IVXLCDM]+)\s+season\b`,
	)

	// Season-tagged dangling episodes:
	//
	//   S4: 13
	//   Season 4: 13
	//   Season IV: 13
	//   4th season: 13
	//   III: 13
	seasonTaggedBareEpisodeRE = regexp.MustCompile(
		`(?i)\b(?:S(\d{1,3})|season\s+(\d{1,3}|[IVXLCDM]+)|(\d{1,3})(?:st|nd|rd|th)\s+season|([IVXLCDM]+))\s*[:\-]\s*(\d{1,4})(?:\.[A-Za-z0-9]+)?(?:\s|$)`,
	)

	// Dangling episode:
	//
	//   Show - 13
	//   Show - 13.mkv
	//
	// stripFilenameExtension() is called before parseSeasonEpisode(),
	// so this regex intentionally knows nothing about .mkv.
	bareTrailingEpisodeRE = regexp.MustCompile(
		`(?i)(?:^|\s)-\s*(\d{1,4})(?:\.[A-Za-z0-9]+)?(?:\s|$)`,
	)

	yearRE = regexp.MustCompile(`\b(?:19|20)\d{2}\b`)

	epRangeRE = regexp.MustCompile(
		`(?i)\b(\d{1,4})\s*(?:-|~)\s*(\d{1,4})\b`,
	)

	resRE = regexp.MustCompile(
		`(?i)\b(?:2160|1440|1080|720|540|480|360|240)p\b`,
	)

	dimRE = regexp.MustCompile(`\b\d{3,5}x\d{3,5}\b`)

	checksumRE = regexp.MustCompile(`(?i)\b[0-9a-f]{8}\b`)

	codecRE = regexp.MustCompile(
		`(?i)\b(?:AV1|HEVC|H\.265|H265|AVC|H\.264|H264|x265|x264)\b`,
	)

	audioCodecRE = regexp.MustCompile(
		`(?i)\b(?:AAC(?:2\.0)?|Opus|E-?AC-?3|DDP(?:2\.0)?|DD2\.0|DTS-HD(?:\s+MA)?|DTS|FLAC|AC3)\b`,
	)
)

func Parse(raw string, fallbackSeason int, sources []string) Metadata {
	r := Metadata{Raw: raw, TagFields: map[string]string{},
		Tokens: Tokenize(raw)}
	r.Group = extractGroup(r.Tokens)
	if r.Group == "" {
		r.Group = extractFilenameGroup(raw)
	}
	parts := splitTopLevel(raw, '|')
	if len(parts) > 1 {
		for _, p := range parts[1:] {
			x := strings.TrimSpace(p)
			if i := strings.Index(x, "("); i >= 0 {
				x = strings.TrimSpace(x[:i])
			}
			if x != "" {
				r.AlternateTitles = append(r.AlternateTitles, x)
			}
		}
	}
	r.Checksum = extractChecksum(r.Tokens)
	r.IsBatch = containsAnyCI(raw, "batch", "mini-batch")
	r.IsComplete = containsAnyCI(raw, "complete series", "complete", "01-", "01 ~")
	r.IsRemastered = containsAnyCI(raw, "remastered")
	r.IsRepack = containsAnyCI(raw, "repack")
	if containsAnyCI(raw, "weekly") {
		r.Labels = append(r.Labels, "weekly")
	}
	for _, t := range r.Tokens {
		if t.Kind == TokenTagBlock {
			parseTagBlock(t.Text, &r)
		}
	}
	parseTech(raw, &r)
	parseSeasonEpisode(raw, &r)
	parseYear(raw, &r)
	parseSemanticGroups(r.Tokens, &r)
	parseMainText(raw, &r)
	dedupe(&r.AlternateTitles)
	dedupe(&r.Labels)
	dedupe(&r.Resolutions)
	dedupe(&r.Dimensions)
	dedupe(&r.BitDepths)
	dedupe(&r.Sources)
	dedupe(&r.Codecs)
	dedupe(&r.AudioCodecs)
	dedupe(&r.AudioFlags)
	dedupe(&r.SubtitleFlags)
	dedupe(&r.SubtitleLanguages)
	dedupe(&r.AudioLanguages)
	dedupe(&r.Encoders)
	dedupe(&r.ReleaseFlags)
	return r
}

func extractGroup(ts []Token) string {
	for _, t := range ts {
		if t.Kind == TokenBracket {
			x := strings.TrimSpace(t.Text)
			if x != "" {
				return x
			}
		}
	}
	return ""
}
func extractFilenameGroup(raw string) string {
	s := stripFilenameExtension(strings.TrimSpace(raw))
	if m := regexp.MustCompile(`(?i)-([A-Za-z][A-Za-z0-9_-]*)$`).FindStringSubmatch(s); m != nil {
		g := m[1]
		if !containsAnyCI(g, "DL", "WEBRip", "WEB-DL", "BluRay", "BD", "AMZN", "CR") {
			return g
		}
	}
	return ""
}

func extractChecksum(ts []Token) string {
	for _, t := range ts {
		if t.Kind == TokenBracket {
			x := strings.TrimSpace(t.Text)
			if checksumRE.MatchString(x) && strings.TrimSpace(checksumRE.FindString(x)) == x {
				return strings.ToUpper(x)
			}
		}
	}
	return ""
}

func maskBracketContent(s string) string {
	b := []byte(s)
	depth := 0
	for i := range b {
		switch b[i] {
		case '[':
			depth++
			if depth > 0 {
				b[i] = ' '
			}
		case ']':
			if depth > 0 {
				b[i] = ' '
			}
			if depth > 0 {
				depth--
			}
		default:
			if depth > 0 {
				b[i] = ' '
			}
		}
	}
	return string(b)
}

func appendSeason(r *Metadata, number int) *Tag {
	if number <= 0 {
		number = 1
	}
	for i := range r.Tags {
		if r.Tags[i].Number == number {
			return &r.Tags[i]
		}
	}
	r.Tags = append(r.Tags, Tag{Number: number})
	return &r.Tags[len(r.Tags)-1]
}

func appendEpisode(r *Metadata, season int, start, end, raw string) {
	s := appendSeason(r, season)

	startValue, _ := strconv.ParseFloat(start, 32)
	endValue := float32(0)
	if end != "" {
		value, _ := strconv.ParseFloat(end, 32)
		endValue = float32(value)
	}

	s.Episodes = append(s.Episodes, EpisodeRange{
		Start: float32(startValue),
		End:   endValue,
		Raw:   raw,
	})
}

func hasEpisodes(r *Metadata) bool {
	for _, season := range r.Tags {
		if len(season.Episodes) > 0 {
			return true
		}
	}
	return false
}

func seasonBefore(pos int, seasonMatches [][]int, s string) int {
	bestPos := -1
	bestSeason := 0
	for _, m := range seasonMatches {
		if m[0] >= pos || m[2] < 0 || m[3] < 0 || m[0] < bestPos {
			continue
		}
		n, err := strconv.Atoi(s[m[2]:m[3]])
		if err == nil {
			bestPos = m[0]
			bestSeason = n
		}
	}
	return bestSeason
}

func sortSeasons(r *Metadata) {
	slices.SortFunc(r.Tags, func(a, b Tag) int {
		if a.Number < b.Number {
			return -1
		}
		if a.Number > b.Number {
			return 1
		}
		return 0
	})

	out := r.Tags[:0]
	for _, season := range r.Tags {
		if len(out) > 0 && out[len(out)-1].Number == season.Number {
			out[len(out)-1].Episodes = append(out[len(out)-1].Episodes, season.Episodes...)
			continue
		}
		out = append(out, season)
	}
	r.Tags = out
}

func dedupeSeasonEpisodes(r *Metadata) {
	for si := range r.Tags {
		seen := make(map[EpisodeRange]struct{}, len(r.Tags[si].Episodes))
		out := r.Tags[si].Episodes[:0]

		for _, ep := range r.Tags[si].Episodes {
			key := EpisodeRange{
				Start: ep.Start,
				End:   ep.End,
			}

			if _, ok := seen[key]; ok {
				continue
			}

			seen[key] = struct{}{}
			out = append(out, ep)
		}

		r.Tags[si].Episodes = out
	}
}

func parseSeasonEpisode(s string, r *Metadata) {
	// Ignore bracketed release/technical metadata while extracting title
	// notation. This prevents values such as "E3" in a checksum like
	// "4AE3A605" from becoming an episode.
	parseS := maskBracketContent(s)
	// ------------------------------------------------------------
	// 1. Explicit SxxExx notation
	// ------------------------------------------------------------

	// Seasons: collect explicit ranges first so S1-3 is not also
	// emitted as S1.
	seasonRanges := seasonRangeRE.FindAllStringSubmatchIndex(parseS, -1)

	for _, m := range seasonRanges {
		start, _ := strconv.Atoi(s[m[2]:m[3]])
		end, _ := strconv.Atoi(parseS[m[4]:m[5]])

		if start >= end {
			continue
		}

		for season := start; season <= end; season++ {
			appendSeason(r, season)
		}
	}

	for _, m := range seasonOnlyRE.FindAllStringSubmatchIndex(parseS, -1) {
		if m[1] < len(s) && s[m[1]] >= '0' && s[m[1]] <= '9' {
			continue
		}

		insideRange := false
		for _, rr := range seasonRanges {
			if m[0] >= rr[0] && m[0] < rr[1] {
				insideRange = true
				break
			}
		}

		if insideRange {
			continue
		}

		n, _ := strconv.Atoi(s[m[2]:m[3]])
		appendSeason(r, n)
	}

	// ------------------------------------------------------------
	// 2. Long-form / ordinal / Roman season tags
	// ------------------------------------------------------------

	// Keep the full match around so long-form season parsing does not also
	// add a second season when the same season is part of "Season IV: 3".
	seasonTaggedMatches := seasonTaggedBareEpisodeRE.FindAllStringSubmatchIndex(parseS, -1)

	// Season 4 / Season IV
	for _, m := range seasonWordRE.FindAllStringSubmatchIndex(parseS, -1) {
		rawSeason := s[m[2]:m[3]]
		if isRomanSeasonTag(rawSeason) && overlapsMatch(m[0], m[1], seasonTaggedMatches) {
			continue
		}

		if n, ok := parseSeasonNumber(rawSeason); ok {
			appendSeason(r, n)
		}
	}

	// 4th season / 21st season
	for _, m := range seasonOrdinalRE.FindAllStringSubmatchIndex(parseS, -1) {
		n, err := strconv.Atoi(s[m[2]:m[3]])
		if err == nil {
			appendSeason(r, n)
		}
	}

	// IV season
	for _, m := range seasonRomanWordRE.FindAllStringSubmatchIndex(parseS, -1) {
		if overlapsMatch(m[0], m[1], seasonTaggedMatches) {
			continue
		}
		rawSeason := s[m[2]:m[3]]

		if n, ok := romanToInt(rawSeason); ok {
			appendSeason(r, n)
		}
	}

	// ------------------------------------------------------------
	// 3. Explicit SxxExx notation
	// ------------------------------------------------------------

	// The episode regex also matches the E portion of SxxExx. Keep the
	// season from the combined form here.
	for _, m := range seasonEpRE.FindAllStringSubmatchIndex(parseS, -1) {
		n, err := strconv.Atoi(s[m[2]:m[3]])
		if err == nil {
			appendSeason(r, n)
		}
	}

	// ------------------------------------------------------------
	// 4. Explicit E notation
	// ------------------------------------------------------------

	// Supports:
	//   E1
	//   E1.5
	//   E1-12
	//   E1-E12
	//   E1E2
	for _, m := range episodeRE.FindAllStringSubmatchIndex(parseS, -1) {
		start, end := s[m[2]:m[3]], ""

		if m[4] >= 0 {
			end = s[m[4]:m[5]]
		}

		season := seasonBefore(m[0], seasonEpRE.FindAllStringSubmatchIndex(parseS, -1), parseS)
		appendEpisode(r, season, start, end, s[m[0]:m[1]])
	}

	// ------------------------------------------------------------
	// 4. Season-tagged dangling episodes
	// ------------------------------------------------------------

	// Examples:
	//
	//   S4: 13
	//   Season 4: 13
	//   Season IV: 13
	//   4th season: 13
	//   III: 13
	//
	// This is deliberately handled separately from ordinary bare
	// episodes because the season tag gives us strong context.
	for _, m := range seasonTaggedMatches {
		// S1-3 is a season range, not season 1 episode 3.
		isSeasonRange := false
		for _, rr := range seasonRanges {
			if m[0] >= rr[0] && m[0] < rr[1] {
				isSeasonRange = true
				break
			}
		}
		if isSeasonRange {
			continue
		}

		var seasonRaw string

		switch {
		case m[2] >= 0:
			// S4
			seasonRaw = s[m[2]:m[3]]

		case m[4] >= 0:
			// Season 4 / Season IV
			seasonRaw = s[m[4]:m[5]]

		case m[6] >= 0:
			// 4th season
			seasonRaw = s[m[6]:m[7]]

		case m[8] >= 0:
			// III
			seasonRaw = s[m[8]:m[9]]
		}

		season, ok := parseSeasonNumber(seasonRaw)
		if !ok {
			continue
		}
		appendSeason(r, season)

		epStart := m[10]
		epEnd := m[11]

		appendEpisode(r, season, s[epStart:epEnd], "", s[m[0]:m[1]])
	}

	// ------------------------------------------------------------
	// 5. Numeric season/episode notation
	//
	//   Show 2 - 12
	//
	// This is stronger than a generic numeric range because it occurs in
	// title text immediately before the release metadata.
	numericSeasonMatches := numericSeasonEpisodeRE.FindAllStringSubmatchIndex(parseS, -1)
	for _, m := range numericSeasonMatches {
		season, _ := strconv.Atoi(parseS[m[2]:m[3]])
		episode := parseS[m[4]:m[5]]
		appendSeason(r, season)
		appendEpisode(r, season, episode, "", s[m[0]:m[1]])
	}

	// ------------------------------------------------------------
	// 6. Season x Episode notation
	//
	//   2x1
	//   2x1-12
	//   2x1~12
	//
	// The episode Raw value intentionally contains only the episode part,
	// matching the representation used by the other episode notations.
	// ------------------------------------------------------------

	seasonXMatches := seasonXEpisodeRE.FindAllStringSubmatchIndex(parseS, -1)
	for _, m := range seasonXMatches {
		season, _ := strconv.Atoi(parseS[m[2]:m[3]])
		start := s[m[4]:m[5]]
		end := ""

		rawEnd := m[5]

		if m[6] >= 0 {
			end = s[m[6]:m[7]]
			rawEnd = m[7]
		}

		appendSeason(r, season)
		appendEpisode(r, season, start, end, s[m[4]:rawEnd])
	}

	// ------------------------------------------------------------
	// 7. Bare numeric episode ranges
	// ------------------------------------------------------------

	// Examples:
	//   0501 ~ 0600
	//   1-12
	//
	// Exclude years and ranges that are actually season ranges.
	for _, m := range epRangeRE.FindAllStringSubmatchIndex(parseS, -1) {
		if overlapsMatch(m[0], m[1], numericSeasonMatches) ||
			overlapsMatch(m[0], m[1], seasonXMatches) {
			continue
		}
		a, _ := strconv.Atoi(parseS[m[2]:m[3]])
		b, _ := strconv.Atoi(parseS[m[4]:m[5]])

		if a >= 1900 && a <= 2099 &&
			b >= 1900 && b <= 2099 {
			continue
		}

		seasonRange := false

		for _, rr := range seasonRanges {
			if m[0] >= rr[0] && m[0] < rr[1] {
				seasonRange = true
				break
			}
		}

		if seasonRange {
			continue
		}

		season := seasonBefore(m[0], seasonOnlyRE.FindAllStringSubmatchIndex(parseS, -1), parseS)
		appendEpisode(r, season, s[m[2]:m[3]], s[m[4]:m[5]], s[m[0]:m[1]])
	}

	// ------------------------------------------------------------
	// 6. Dangling episode after "-":
	//
	//   Show - 13
	//   Show - 13.mkv
	//
	// The extension is allowed here because Parse receives the raw
	// filename, before stripFilenameExtension() is called.
	// ------------------------------------------------------------

	for _, m := range bareTrailingEpisodeRE.FindAllStringSubmatchIndex(parseS, -1) {
		if overlapsMatch(m[0], m[1], numericSeasonMatches) {
			continue
		}
		// A season-tagged form such as "II - 12" has already been
		// consumed above. Do not also parse its trailing "- 12" as a
		// bare episode (which would incorrectly default to season 1).
		if overlapsMatch(m[0], m[1], seasonTaggedMatches) {
			continue
		}
		start := s[m[2]:m[3]]

		// Avoid interpreting a year as an episode.
		n, _ := strconv.Atoi(start)
		if n >= 1900 && n <= 2099 {
			continue
		}

		season := seasonBefore(m[0], seasonOnlyRE.FindAllStringSubmatchIndex(parseS, -1), parseS)
		appendEpisode(r, season, start, "", s[m[0]:m[1]])
	}

	sortSeasons(r)
	dedupeSeasonEpisodes(r)
}

func isRomanSeasonTag(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !strings.ContainsRune("ivxlcdmIVXLCDM", r) {
			return false
		}
	}
	return true
}

func overlapsMatch(start, end int, matches [][]int) bool {
	for _, m := range matches {
		if start < m[1] && end > m[0] {
			return true
		}
	}
	return false
}

func (e EpisodeRange) compare(other EpisodeRange) int {
	maxA := max(e.Start, e.End)
	maxB := max(other.Start, other.End)

	if maxA < maxB {
		return -1
	}

	if maxA > maxB {
		return 1
	}

	return 0
}

func (t Tag) compare(other Tag) int {
	if t.Number < other.Number {
		return -1
	}
	if t.Number > other.Number {
		return 1
	}

	// A season with no explicit episodes is a season pack. A pack represents
	// the complete season and therefore sorts after episode-specific releases.
	sp, op := len(t.Episodes) == 0, len(other.Episodes) == 0
	if sp != op {
		if sp {
			return 1
		}
		return -1
	}
	return compareEpisodes(t.Episodes, other.Episodes)
}

func compareEpisodes(a, b []EpisodeRange) int {
	aa := append([]EpisodeRange(nil), a...)
	bb := append([]EpisodeRange(nil), b...)

	slices.SortFunc(aa, func(x, y EpisodeRange) int { return y.compare(x) })
	slices.SortFunc(bb, func(x, y EpisodeRange) int { return y.compare(x) })

	n := min(len(bb), len(aa))

	for i := range n {
		if c := aa[i].compare(bb[i]); c != 0 {
			return c
		}
	}

	if len(aa) < len(bb) {
		return -1
	}

	if len(aa) > len(bb) {
		return 1
	}
	return 0
}

// Contains reports whether m describes a release that contains all of the
// semantic content requested by other. A season pack contains every episode
// in that season; an episode release does not contain the season pack itself.
func (m Tags) Contains(other Tags) bool {
	for _, wanted := range other {
		have := findSeason(m, wanted.Number)
		if have == nil {
			return false
		}
		if len(wanted.Episodes) == 0 {
			if len(have.Episodes) != 0 {
				return false
			}
			continue
		}
		if len(have.Episodes) == 0 {
			continue // season pack contains every episode in the season
		}
		for _, ep := range wanted.Episodes {
			if !containsEpisode(have.Episodes, ep) {
				return false
			}
		}
	}

	return true
}

func findSeason(seasons []Tag, number int) *Tag {
	for i := range seasons {
		if seasons[i].Number == number {
			return &seasons[i]
		}
	}
	return nil
}

func containsEpisode(have []EpisodeRange, wanted EpisodeRange) bool {
	ws, we := episodeBounds(wanted)
	for _, h := range have {
		hs, he := episodeBounds(h)
		if hs <= ws && he >= we {
			return true
		}
	}
	return false
}

func episodeBounds(e EpisodeRange) (float32, float32) {
	end := e.Start
	if e.End != 0 {
		end = e.End
	}
	return e.Start, end
}

func parseYear(s string, r *Metadata) {
	if m := yearRE.FindStringSubmatch(s); m != nil {
		r.Year, _ = strconv.Atoi(m[0])
	}
}

func parseTech(s string, r *Metadata) {
	r.Resolutions = resRE.FindAllString(s, -1)
	r.Dimensions = dimRE.FindAllString(s, -1)
	for _, x := range regexp.MustCompile(`(?i)\b(?:8|10|12)-?bit\b`).FindAllString(s, -1) {
		r.BitDepths = append(r.BitDepths, x)
	}
	for _, x := range codecRE.FindAllString(s, -1) {
		r.Codecs = append(r.Codecs, normalizeCodec(x))
	}
	r.AudioCodecs = audioCodecRE.FindAllString(s, -1)
	for _, x := range []string{"WEB-DL", "WEBRip", "CTHP", "WebRip", "BD", "BluRay", "AMZN", "CR", "HIDI", "HIDIVE", "IQIYI", "BILI", "LIV", "OV", "YTB", "VHS"} {
		if containsCI(s, x) {
			r.Sources = append(r.Sources, x)
		}
	}
	for _, x := range []string{"Dual Audio", "Dual-Audio", "DUAL", "MULTi", "Multi-Audio", "Multi-Subs", "MultiSub", "English Dub", "English-Sub", "Korean Audio", "D-SUB", "M-SUB"} {
		if containsCI(s, x) {
			if strings.Contains(strings.ToLower(x), "sub") {
				r.SubtitleFlags = append(r.SubtitleFlags, x)
			} else {
				r.AudioFlags = append(r.AudioFlags, x)
			}
		}
	}
	for _, x := range regexp.MustCompile(`(?i)\b(?:NVENC|Veryslow)\b`).FindAllString(s, -1) {
		r.Encoders = append(r.Encoders, x)
	}
	if containsCI(s, "END") {
		r.ReleaseFlags = append(r.ReleaseFlags, "END")
	}
}

func normalizeCodec(s string) string {
	u := strings.ToUpper(s)
	if u == "H265" {
		return "H.265"
	}
	if u == "H264" {
		return "H.264"
	}
	return s
}

func parseSemanticGroups(ts []Token, r *Metadata) {
	bracketCount := 0
	for _, t := range ts {
		if t.Kind == TokenBracket {
			bracketCount++
			if bracketCount == 1 || strings.EqualFold(strings.TrimSpace(t.Text), r.Checksum) {
				continue
			}
			classifyBracket(t.Text, r)
		}
		if t.Kind == TokenParen {
			classifyParen(t.Text, r)
		}
	}
}

func classifyBracket(x string, r *Metadata) {
	x = strings.TrimSpace(x)
	if x == "" {
		return
	}
	if checksumRE.MatchString(x) && strings.TrimSpace(checksumRE.FindString(x)) == x {
		return
	}
	if containsCI(x, "batch") {
		r.ReleaseFlags = append(r.ReleaseFlags, x)
		r.Labels = append(r.Labels, x)
		return
	}
	// technical bracket is already represented by dedicated fields
	if resRE.MatchString(x) || codecRE.MatchString(x) || audioCodecRE.MatchString(x) ||
		containsCI(x, "WEB") ||
		containsCI(x, "BD") ||
		containsCI(x, "Multi") ||
		containsCI(x, "Dub") {
		return
	}
	if containsCI(x, "Complete") || containsCI(x, "Uncensored") || containsCI(x, "weekly") {
		r.ReleaseFlags = append(r.ReleaseFlags, x)
		r.Labels = append(r.Labels, x)
		return
	}
	if !containsAnyCI(x, "Mini-Batch") {
		r.Unknown = append(r.Unknown, x)
	}
}

func classifyParen(x string, r *Metadata) {
	x = strings.TrimSpace(x)
	if x == "" {
		return
	}
	if resRE.MatchString(x) || codecRE.MatchString(x) || audioCodecRE.MatchString(x) ||
		containsAnyCI(x, "WEB", "BD") {
		return
	}
	if containsAnyCI(x, "Multi-Subs", "Multi-Audio", "Dual-Audio", "English-Sub", "Korean Audio") {
		parts := strings.Split(x, ",")
		var alias []string
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if containsAnyCI(
				p,
				"Multi-Subs",
				"Multi-Audio",
				"Dual-Audio",
				"English-Sub",
				"Korean Audio",
			) {
				break
			}
			if p != "" {
				alias = append(alias, p)
			}
		}
		if len(alias) > 0 {
			r.AlternateTitles = append(r.AlternateTitles, strings.Join(alias, ", "))
		}
		r.ReleaseFlags = append(r.ReleaseFlags, x)
		return
	}
	if strings.Contains(x, "|") || strings.Contains(x, ";") {
		for _, p := range splitAlias(x) {
			r.AlternateTitles = append(r.AlternateTitles, p)
		}
		return
	}
	if yearRE.MatchString(x) {
		return
	}
	if containsCI(x, "Multi-Subs") || containsCI(x, "Multi-Audio") || containsCI(x, "Dual-Audio") ||
		containsCI(x, "English-Sub") ||
		containsCI(x, "Korean Audio") {
		r.ReleaseFlags = append(r.ReleaseFlags, x)
		return
	}
	// Parentheses containing separators are usually aliases, except obvious release notes.
	if strings.Contains(x, "|") || strings.Contains(x, ";") {
		for _, p := range splitAlias(x) {
			r.AlternateTitles = append(r.AlternateTitles, p)
		}
		return
	}
	if strings.ContainsAny(x, ";,|") {
		for _, p := range splitAlias(x) {
			r.AlternateTitles = append(r.AlternateTitles, p)
		}
		return
	}
	if !containsAnyCI(x, "weekly", "batch", "dual", "multi-sub", "multi-audio") {
		r.AlternateTitles = append(r.AlternateTitles, x)
	}
}

func splitAlias(x string) []string {
	var out []string
	for _, p := range strings.FieldsFunc(x, func(r rune) bool { return r == '|' || r == ';' }) {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func parseMainText(raw string, r *Metadata) {
	s := raw
	if r.Group != "" {
		if i := strings.Index(s, "["); i == 0 {
			if j := matchingClose(s, 0, '[', ']'); j >= 0 {
				s = s[j+1:]
			}
		}
	}
	s = strings.TrimSpace(s)
	// Titles can arrive as torrent filenames. Remove only the terminal filename
	// extension; dots elsewhere are meaningful title separators.
	s = stripFilenameExtension(s)
	// Remove all explicit technical/metadata groups, retaining plain text.
	var b strings.Builder
	for i := 0; i < len(s); {
		switch s[i] {
		case '[', '(', '{':
			close := map[byte]byte{'[': ']', '(': ')', '{': '}'}[s[i]]
			j := matchingClose(s, i, rune(s[i]), rune(close))
			if j >= 0 {
				b.WriteByte(' ')
				i = j + 1
				continue
			}
		}
		b.WriteByte(s[i])
		i++
	}
	s = strings.TrimSpace(b.String())
	if i := strings.Index(s, "|"); i >= 0 {
		s = strings.TrimSpace(s[:i])
	}
	s = strings.ReplaceAll(s, "|", " ")
	// Strip tag block residue and common trailing release suffixes.
	s = regexp.MustCompile(`\s+\{Tags:.*$`).ReplaceAllString(s, "")
	// Filename-style release names use dots as separators. Preserve dots inside
	// decimal episode numbers and codec/channel notation, but turn structural
	// dots into spaces when they separate title/metadata components.
	s = normalizeFilenameSeparators(s)
	s = regexp.MustCompile(`(?i)\s+-[A-Za-z0-9]+\s*$`).ReplaceAllString(s, "")
	s = strings.Join(strings.Fields(s), " ")
	if i := strings.Index(
		s,
		" - ",
	); i >= 0 { // retain hyphen in ordinary names unless followed by an episode-like number
		right := strings.TrimSpace(s[i+3:])
		if regexp.MustCompile(`^\d{1,4}\b`).MatchString(right) {
			s = strings.TrimSpace(s[:i])
		}
	}
	// Remove the combined SxxExx form first. Removing only E notation would
	// leave the season marker (for example, S03E11 -> S03).
	s = seasonEpRE.ReplaceAllString(s, "")
	if hasEpisodes(r) {
		s = episodeRE.ReplaceAllString(s, "")
		s = epRangeRE.ReplaceAllString(s, "")
	}

	if len(r.Tags) > 0 {
		s = seasonEpRE.ReplaceAllString(s, "")
		s = seasonRangeRE.ReplaceAllString(s, "")
		// Remove the complete season-tagged episode first. Otherwise
		// "Season IV: 3" becomes "Season IV" + "3" and the
		// separator/episode can leak into the primary title.
		s = seasonTaggedBareEpisodeRE.ReplaceAllString(s, "")
		s = seasonWordRE.ReplaceAllString(s, "")
		s = seasonOrdinalRE.ReplaceAllString(s, "")
		s = seasonRomanWordRE.ReplaceAllString(s, "")
		s = seasonOnlyRE.ReplaceAllString(s, "")
	}

	// New: remove season-tagged dangling episodes:
	//
	//	Show S4: 13
	//	Show Season 4: 13
	//	Show 4th season: 13
	//	Show III: 13
	//
	// and old-style:
	//
	//	Show - 13
	if hasEpisodes(r) {
		s = seasonTaggedBareEpisodeRE.ReplaceAllString(s, "")
		s = bareTrailingEpisodeRE.ReplaceAllString(s, "")
	}

	s = yearRE.ReplaceAllString(s, "")

	s = regexp.MustCompile(`\s{2,}`).ReplaceAllString(s, " ")
	s = strings.Trim(s, " -|:")
	// The first plain segment is the title; a remaining plain segment after an
	// episode reference and before technical metadata is the episode title.
	if s == "" {
		return
	}
	if hasEpisodes(r) {
		ms := episodeRE.FindAllStringIndex(maskBracketContent(raw), -1)
		if len(ms) > 0 {
			tail := stripDecorative(raw[ms[len(ms)-1][1]:])
			tail = stripFilenameExtension(tail)
			tail = normalizeFilenameSeparators(tail)
			if tail != "" && !startsTech(tail) {
				r.EpisodeTitle = cleanEpisodeTitle(tail)
			}
		}
	}
	if r.EpisodeTitle != "" {
		if i := strings.LastIndex(strings.ToLower(s), strings.ToLower(r.EpisodeTitle)); i >= 0 {
			s = strings.TrimSpace(s[:i])
		}
	}
	if r.PrimaryTitle == "" {
		r.PrimaryTitle = s
	}
	// Fix title pollution from trailing technical words when title had no brackets.
	r.PrimaryTitle = cleanTitle(r.PrimaryTitle)
}

func stripFilenameExtension(s string) string {
	// Common media/torrent extensions. Only strip a terminal extension.
	return regexp.MustCompile(`(?i)\.(?:mkv|mp4|avi|mov|webm|m4v|ts|m2ts|wmv|flac|mp3|torrent)$`).
		ReplaceAllString(strings.TrimSpace(s), "")
}

func normalizeFilenameSeparators(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] != '.' {
			b.WriteByte(s[i])
			continue
		}
		// Keep decimal points such as E6.5 and codec/channel forms such as AAC2.0.
		if i > 0 && i+1 < len(s) && s[i-1] >= '0' && s[i-1] <= '9' && s[i+1] >= '0' &&
			s[i+1] <= '9' {
			b.WriteByte('.')
			continue
		}
		b.WriteByte(' ')
	}
	return b.String()
}

func cleanTitle(s string) string {
	s = strings.TrimSpace(s)
	s = regexp.MustCompile(`(?i)\s+Remastered$`).ReplaceAllString(s, "")
	s = regexp.MustCompile(`(?i)\s+(?:1080p|720p|480p|2160p)\b.*$`).ReplaceAllString(s, "")
	return strings.TrimSpace(s)
}
func stripDecorative(s string) string {
	for len(s) > 0 && strings.ContainsRune(" -|", rune(s[0])) {
		s = s[1:]
	}
	return strings.TrimSpace(s)
}
func startsTech(s string) bool {
	s = strings.TrimSpace(s)
	if regexp.MustCompile(`(?i)^(?:(?:2160|1440|1080|720|540|480|360|240)p\b|WEB(?:-DL|Rip)?\b|CR\b|AMZN\b|HIDI(?:VE)?\b|IQIYI\b|BILI\b|LIV\b|YTB\b|BD\b|BluRay\b|AV1\b|HEVC\b|AVC\b|H\.?26[45]\b|x26[45]\b|AAC(?:2\.0)?\b|Opus\b|DDP(?:2\.0)?\b|DTS\b|FLAC\b)`).
		MatchString(s) {
		return true
	}
	return regexp.MustCompile(`(?i)^(?:E\d|\.\d)`).MatchString(s)
}
func cleanEpisodeTitle(s string) string {
	for _, sep := range []string{"[", "(", "|"} {
		if i := strings.Index(s, sep); i >= 0 {
			s = s[:i]
		}
	}
	if m := regexp.MustCompile(`(?i)\s+\b(?:2160|1440|1080|720|540|480|360|240)p\b`).
		FindStringIndex(s); m != nil {
		s = s[:m[0]]
	}
	if m := regexp.MustCompile(`(?i)\s+(?:WEB-DL|WEBRip|BluRay|BD|AMZN|CR|HIDI|HIDIVE|IQIYI|BILI|LIV|YTB)\b`).
		FindStringIndex(s); m != nil {
		s = s[:m[0]]
	}
	return strings.TrimSpace(s)
}
func splitTopLevel(s string, sep rune) []string {
	var out []string
	start, depth := 0, 0
	for i, r := range s {
		switch r {
		case '[', '(', '{':
			depth++
		case ']', ')', '}':
			if depth > 0 {
				depth--
			}
		}
		if r == sep && depth == 0 {
			out = append(out, s[start:i])
			start = i + len(string(r))
		}
	}
	out = append(out, s[start:])
	return out
}

func matchingClose(s string, start int, open, close rune) int {
	depth := 0
	for i, r := range s[start:] {
		if r == open {
			depth++
		}
		if r == close {
			depth--
			if depth == 0 {
				return start + i
			}
		}
	}
	return -1
}

func parseTagBlock(x string, r *Metadata) {
	x = strings.TrimSpace(x)
	if strings.HasPrefix(strings.ToLower(x), "tags:") {
		x = x[5:]
	}
	for _, p := range strings.Split(x, ";") {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		kv := strings.SplitN(p, "=", 2)
		if len(kv) == 2 {
			k, v := strings.TrimSpace(kv[0]), strings.TrimSpace(kv[1])
			r.TagFields[k] = v
			if k == "A" {
				r.AudioLanguages = append(r.AudioLanguages, strings.Split(v, ",")...)
			}
			if k == "S" {
				r.SubtitleLanguages = append(r.SubtitleLanguages, strings.Split(v, ",")...)
			}
		} else if m := regexp.MustCompile(`^([A-Za-z]+)(.+)$`).FindStringSubmatch(p); len(m) == 3 {
			r.TagFields[m[1]] = m[2]
		}
	}
}

func containsCI(s, sub string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(sub))
}
func containsAnyCI(s string, subs ...string) bool {
	for _, x := range subs {
		if containsCI(s, x) {
			return true
		}
	}
	return false
}
func dedupe[T ~[]string](xs *T) {
	seen := map[string]struct{}{}
	out := (*xs)[:0]
	for _, x := range *xs {
		x = strings.TrimSpace(x)
		k := strings.ToLower(x)
		if k == "" {
			continue
		}
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, x)
	}
	*xs = out
}

func romanToInt(s string) (int, bool) {
	s = strings.ToUpper(strings.TrimSpace(s))
	if s == "" {
		return 0, false
	}

	values := map[byte]int{
		'I': 1,
		'V': 5,
		'X': 10,
		'L': 50,
		'C': 100,
		'D': 500,
		'M': 1000,
	}

	total := 0
	prev := 0

	for i := len(s) - 1; i >= 0; i-- {
		v, ok := values[s[i]]
		if !ok {
			return 0, false
		}

		if v < prev {
			total -= v
		} else {
			total += v
			prev = v
		}
	}

	// Reject non-canonical / implausibly large values. Season numbers
	// should be ordinary Roman numerals, not arbitrary Roman strings.
	if total <= 0 || total > 999 {
		return 0, false
	}

	return total, true
}

func parseSeasonNumber(s string) (int, bool) {
	s = strings.TrimSpace(s)

	if n, err := strconv.Atoi(s); err == nil {
		if n >= 0 && n <= 999 {
			return n, true
		}
		return 0, false
	}

	return romanToInt(s)
}

type Tags []Tag

func (s Tags) String() string {
	var parts []string

	for _, season := range s {
		for _, episode := range season.Episodes {
			if episode.End != 0 {
				parts = append(parts,
					fmt.Sprintf("S%dE%g-E%g", season.Number, episode.Start, episode.End),
				)
			} else {
				parts = append(parts,
					fmt.Sprintf("S%dE%g", season.Number, episode.Start),
				)
			}
		}

		// A season without explicit episodes is a season pack.
		if len(season.Episodes) == 0 {
			parts = append(parts, fmt.Sprintf("S%d", season.Number))
		}
	}

	return strings.Join(parts, " ")
}

func (t Tag) String() string {
	var parts []string

	for _, episode := range t.Episodes {
		if episode.End != 0 {
			parts = append(parts,
				fmt.Sprintf("S%dE%g-E%g", t.Number, episode.Start, episode.End),
			)
		} else {
			parts = append(parts,
				fmt.Sprintf("S%dE%g", t.Number, episode.Start),
			)
		}
	}

	// A season without explicit episodes is a season pack.
	if len(t.Episodes) == 0 {
		parts = append(parts, fmt.Sprintf("S%d", t.Number))
	}

	return strings.Join(parts, " ")
}

func (s Tags) IsZero() bool {
	return len(s) == 0
}

func (t Tag) IsZero() bool {
	return t.Number == 0
}

func (s Tags) LastEpisode() float32 {
	if len(s) == 0 {
		return -1
	}

	for _, v := range slices.Backward(s) {
		if len(v.Episodes) == 0 {
			continue
		}

		epRange := v.Episodes[len(v.Episodes)-1]
		if epRange.End != 0 {
			return epRange.End
		}
		return epRange.Start
	}

	return -1
}

func (s Tags) LastSeason() int {
	for _, tag := range slices.Backward(s) {
		if tag.Number != 0 {
			return tag.Number
		}
	}

	return 0
}

type Resolutions []string

func (r Resolutions) Highest() string {
	var highest string
	var highestHeight int

	for _, resolution := range r {
		height, err := strconv.Atoi(strings.TrimSuffix(strings.ToLower(resolution), "p"))
		if err != nil {
			continue
		}

		if height > highestHeight {
			highestHeight = height
			highest = resolution
		}
	}

	return highest
}

func (r Resolutions) Compare(other Resolutions) int {
	a := r.Highest()
	b := other.Highest()

	if a == b {
		return 0
	}

	av := resolutionHeight(a)
	bv := resolutionHeight(b)

	switch {
	case av < bv:
		return -1
	default:
		return 1
	}
}

func resolutionHeight(resolution string) int {
	resolution = strings.TrimSpace(strings.ToLower(resolution))
	resolution = strings.TrimSuffix(resolution, "p")

	height, err := strconv.Atoi(resolution)
	if err != nil {
		return 0
	}

	return height
}

var seasonRE = regexp.MustCompile(
	`(?i)\bS(\d{1,3})(?:E(\d{1,4}(?:\.\d+)?)(?:[-~]E?(\d{1,4}(?:\.\d+)?))?)?\b`,
)

func ParseTags(s string) Tags {
	var seasons Tags

	for _, match := range seasonRE.FindAllStringSubmatch(s, -1) {
		seasonNumber, _ := strconv.Atoi(match[1])

		season := Tag{
			Number: seasonNumber,
		}

		if match[2] != "" {
			start, _ := strconv.ParseFloat(match[2], 32)

			var end float32
			if match[3] != "" {
				parsedEnd, _ := strconv.ParseFloat(match[3], 32)
				end = float32(parsedEnd)
			}

			season.Episodes = append(season.Episodes, EpisodeRange{
				Start: float32(start),
				End:   end,
			})
		}

		seasons = append(seasons, season)
	}

	return seasons
}

func normalizeTags(tags any) []Tag {
	var result []Tag

	switch v := tags.(type) {
	case Tag:
		result = append(result, v)
	case Tags:
		result = append(result, v...)
	}

	slices.SortFunc(result, func(a, b Tag) int {
		return a.compare(b)
	})

	result = deduplicateTags(result)

	return result
}

func deduplicateTags(tags []Tag) []Tag {
	if len(tags) < 2 {
		return tags
	}

	result := make([]Tag, 0, len(tags))

	for _, tag := range tags {
		found := false

		for _, existing := range result {
			if tag.compare(existing) == 0 {
				found = true
				break
			}
		}

		if !found {
			result = append(result, tag)
		}
	}

	return result
}

func (m Tags) Compare(other any) int {
	return compareTags(normalizeTags(m), normalizeTags(other))
}

func (t Tag) Compare(other any) int {
	return compareTags(normalizeTags(t), normalizeTags(other))
}

func compareTags(a, b []Tag) int {
	switch {
	case len(a) == 0 && len(b) == 0:
		return 0
	case len(a) == 0:
		return -1
	case len(b) == 0:
		return 1
	}

	a = highestTags(a)
	b = highestTags(b)

	return a[0].compare(b[0])
}

func highestTags(tags []Tag) []Tag {
	if len(tags) == 0 {
		return nil
	}

	highest := tags[0]

	for _, tag := range tags[1:] {
		if tag.compare(highest) > 0 {
			highest = tag
		}
	}

	return []Tag{highest}
}
