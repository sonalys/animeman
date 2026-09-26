package metadata

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/sonalys/animeman/internal/pkg/sliceutils"
)

type Metadata struct {
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
}

const (
	seasonNumberPattern  = `\d{1,3}`
	episodeNumberPattern = `\d{1,4}(?:\.\d+)?`
	seasonValuePattern   = `(?:` + seasonNumberPattern + `|[IVXLCDM]+)`
	resolutionPattern    = `(?:2160|1440|1080|720|540|480|360|240)`
	videoCodecPattern    = `(?:AV1|HEVC|H\.265|H265|AVC|H\.264|H264|x265|x264)`
	techAudioPattern     = `(?:AAC(?:2\.0)?|Opus|DDP(?:2\.0)?|DTS|FLAC)`
	sourcePattern        = `(?:WEB-DL|WEBRip|BluRay|BD|AMZN|CR|HIDI|HIDIVE|IQIYI|BILI|LIV|YTB)`
)

var (
	// seasonEpRE matches explicit season + episode notation.
	// Examples: S1E1, S01E12, S1E1-E24, S2E3~E12, S1E1E2E3
	seasonEpRE = regexp.MustCompile(
		`(?i)\bS(` + seasonNumberPattern + `)\s*E(` + episodeNumberPattern + `)(?:\s*[-~]\s*(?:E)?(` + episodeNumberPattern + `))?(?:E` + episodeNumberPattern + `)*\b`,
	)

	// seasonRangeRE matches a season range.
	// Examples: S1-S2, S1~S2, S1-2
	seasonRangeRE = regexp.MustCompile(
		`(?i)\bS(` + seasonNumberPattern + `)\s*(?:-|~)\s*S?(` + seasonNumberPattern + `)\b`,
	)

	// seasonOnlyRE matches a standalone season tag.
	// Examples: S1, S02, S12
	seasonOnlyRE = regexp.MustCompile(`(?i)\bS(` + seasonNumberPattern + `)\b`)

	// episodeRE matches explicit episode notation.
	// Examples: E1, E12, E12-E24, E12~E24
	episodeRE = regexp.MustCompile(
		`(?i)E(` + episodeNumberPattern + `)(?:\s*(?:-|~)\s*E?(` + episodeNumberPattern + `))?`,
	)

	// numericSeasonEpisodeRE matches numeric season-episode notation.
	// Examples: 1-12, 01-12, 2-5, 1-12.5
	numericSeasonEpisodeRE = regexp.MustCompile(
		`(?i)\b([1-9]\d?)\s*-\s*(` + episodeNumberPattern + `)\b`,
	)

	// seasonXEpisodeRE matches season x episode notation, including ranges.
	// Examples: 1x12, 01x12, 1×12, 2x5-8, 2x5~8, 2x5.5
	seasonXEpisodeRE = regexp.MustCompile(
		`(?i)\b([1-9]\d?)\s*[x×]\s*(` + episodeNumberPattern + `)(?:\s*[-~]\s*(` + episodeNumberPattern + `))?\b`,
	)

	// seasonWordRE matches "season N" notation, including Roman numerals.
	// Examples: Season 1, season 12, Season IV
	seasonWordRE = regexp.MustCompile(
		`(?i)\bseason\s+(` + seasonValuePattern + `)\b`,
	)

	// seasonOrdinalRE matches ordinal season notation.
	// Examples: 1st season, 2nd season, 3rd season, 12th season
	seasonOrdinalRE = regexp.MustCompile(
		`(?i)\b(` + seasonNumberPattern + `)(?:st|nd|rd|th)\s+season\b`,
	)

	// seasonRomanWordRE matches a Roman numeral followed by "season".
	// Examples: I season, II season, IV season
	seasonRomanWordRE = regexp.MustCompile(`(?i)\b([IVXLCDM]+)\s+season\b`)

	// seasonTaggedBareEpisodeRE matches season-qualified bare episode numbers.
	// Examples: S1:12, S2-5, Season 1:12, 2nd season:12, IV:12
	seasonTaggedBareEpisodeRE = regexp.MustCompile(
		`(?i)\b(?:S(` + seasonNumberPattern + `)|season\s+(` + seasonValuePattern + `)|(` + seasonNumberPattern + `)(?:st|nd|rd|th)\s+season|([IVXLCDM]+))\s*[:\-]\s*(` + episodeNumberPattern + `)(?:\.[A-Za-z0-9]+)?(?:\s|$)`,
	)

	// bareTrailingEpisodeRE matches a bare episode number after a dash.
	// Examples: "Title - 12", "Title - 1", "Title - 12.5"
	bareTrailingEpisodeRE = regexp.MustCompile(
		`(?i)(?:^|\s)-\s*(` + episodeNumberPattern + `)(?:\s|$)`,
	)

	// yearRE matches a four-digit year beginning with 19 or 20.
	// Examples: 1999, 2006, 2024
	yearRE = regexp.MustCompile(`\b(?:19|20)\d{2}\b`)

	// epRangeRE matches a bare episode range.
	// Examples: 01-24, 1-12, 0501~600, 01 ~ 74, 1-12.5
	epRangeRE = regexp.MustCompile(
		`(?i)\b(` + episodeNumberPattern + `)\s*(?:-|~)\s*(` + episodeNumberPattern + `)\b`,
	)

	// resRE matches a supported video resolution.
	// Examples: 480p, 720p, 1080p, 1440p, 2160p
	resRE = regexp.MustCompile(`(?i)\b` + resolutionPattern + `p\b`)

	// dimRE matches video dimensions.
	// Examples: 1920x1080, 1280x720, 3840x2160
	dimRE = regexp.MustCompile(`\b\d{3,5}x\d{3,5}\b`)

	// checksumRE matches an eight-character hexadecimal checksum.
	// Examples: deadbeef, A1B2C3D4, 0123abcd
	checksumRE = regexp.MustCompile(`(?i)\b[0-9a-f]{8}\b`)

	// codecRE matches supported video codec names.
	// Examples: AVC, H264, H.264, HEVC, H265, H.265, x264, x265, AV1
	codecRE = regexp.MustCompile(`(?i)\b` + videoCodecPattern + `\b`)

	// audioCodecRE matches supported audio codec names.
	// Examples: AAC, AAC2.0, Opus, E-AC-3, EAC3, DDP, DDP2.0, DTS, DTS-HD MA, FLAC, AC3
	audioCodecRE = regexp.MustCompile(
		`(?i)\b(?:AAC(?:2\.0)?|Opus|E-?AC-?3|DDP(?:2\.0)?|DD2\.0|DTS-HD(?:\s+MA)?|DTS|FLAC|AC3)\b`,
	)
)

var (
	filenameGroupRE = regexp.MustCompile(`(?i)-([A-Za-z][A-Za-z0-9_-]*)$`)

	seasonEpisodeRangeExceptionRE = regexp.MustCompile(
		`(?i)^S` + seasonNumberPattern + `\s+-\s+` + episodeNumberPattern + `$`,
	)

	bitDepthRE = regexp.MustCompile(`(?i)\b(?:8|10|12)-?bit\b`)

	encoderRE = regexp.MustCompile(`(?i)\b(?:NVENC|Veryslow)\b`)

	tagResidueRE = regexp.MustCompile(`\s+\{Tags:.*$`)

	releaseSuffixRE = regexp.MustCompile(`(?i)\s+-[A-Za-z0-9]+\s*$`)

	leadingEpisodeNumberRE = regexp.MustCompile(`^` + episodeNumberPattern + `\b`)

	multipleSpaceRE = regexp.MustCompile(`\s{2,}`)

	filenameExtensionRE = regexp.MustCompile(
		`(?i)\.(?:mkv|mp4|avi|mov|webm|m4v|ts|m2ts|wmv|flac|mp3|torrent)$`,
	)

	remasteredSuffixRE = regexp.MustCompile(`(?i)\s+Remastered$`)

	resolutionSuffixRE = regexp.MustCompile(`(?i)\s+` + resolutionPattern + `p\b.*$`)

	tagFieldRE = regexp.MustCompile(`^([A-Za-z]+)(.+)$`)

	startsTechRE = regexp.MustCompile(
		`(?i)^(?:` +
			resolutionPattern + `p\b|` +
			`WEB(?:-DL|Rip)?\b|` +
			`CR\b|` +
			`AMZN\b|` +
			`HIDI(?:VE)?\b|` +
			`IQIYI\b|` +
			`BILI\b|` +
			`LIV\b|` +
			`YTB\b|` +
			`BD\b|` +
			`BluRay\b|` +
			videoCodecPattern + `\b|` +
			techAudioPattern + `\b)`,
	)

	startsEpisodeRE = regexp.MustCompile(`(?i)^(?:E\d|\.\d)`)

	episodeTitleResolutionRE = regexp.MustCompile(
		`(?i)\s+\b` + resolutionPattern + `p\b`,
	)

	episodeTitleSourceRE = regexp.MustCompile(
		`(?i)\s+` + sourcePattern + `\b`,
	)
)

// Parse extracts metadata from a release name.
func Parse(raw string, fallbackSeason int, sources []string) Metadata {
	metadata := Metadata{
		Tags:      parseTags(raw, fallbackSeason),
		TagFields: map[string]string{},
	}

	tokens := Tokenize(raw)
	metadata.Group = extractGroup(tokens)
	if metadata.Group == "" {
		metadata.Group = extractFilenameGroup(raw)
	}

	parts := splitTopLevel(raw, '|')
	if len(parts) > 1 {
		for _, p := range parts[1:] {
			x := strings.TrimSpace(p)
			if i := strings.Index(x, "("); i >= 0 {
				x = strings.TrimSpace(x[:i])
			}
			if x != "" {
				appendUnique(&metadata.AlternateTitles, x)
			}
		}
	}

	metadata.Checksum = extractChecksum(tokens)
	metadata.IsBatch = containsAnyCI(raw, "batch", "mini-batch")
	metadata.IsComplete = containsAnyCI(raw, "complete series", "complete", "01-", "01 ~")
	metadata.IsRemastered = containsAnyCI(raw, "remastered")
	metadata.IsRepack = containsAnyCI(raw, "repack")

	if containsAnyCI(raw, "weekly") {
		appendUnique(&metadata.Labels, "weekly")
	}

	for _, t := range tokens {
		if t.Kind == TokenTagBlock {
			parseTagBlock(t.Text, &metadata)
		}
	}

	parseTech(raw, &metadata)
	parseYear(raw, &metadata)
	parseSemanticGroups(tokens, &metadata)
	parseMainText(raw, &metadata)

	return metadata
}

func ParsePrimaryTitle(raw string) string {
	metadata := Metadata{
		Tags: ParseTags(raw),
	}

	parseMainText(raw, &metadata)

	return metadata.PrimaryTitle
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
	if m := filenameGroupRE.FindStringSubmatch(s); m != nil {
		g := m[1]
		if !containsAnyCI(g, "DL", "WEBRip", "WEB-DL", "BluRay", "BD", "AMZN", "CR") {
			return g
		}
	}
	return ""
}

func extractChecksum(ts []Token) string {
	for _, t := range ts {
		if t.Kind != TokenBracket {
			continue
		}

		x := strings.TrimSpace(t.Text)
		if checksumRE.MatchString(x) && strings.TrimSpace(checksumRE.FindString(x)) == x {
			return strings.ToUpper(x)
		}
	}
	return ""
}

func maskBracketContent(s string) string {
	b := []byte(s)
	depth := 0

	for i, ch := range b {
		switch ch {
		case '[':
			depth++
			b[i] = ' '
		case ']':
			if depth > 0 {
				depth--
			}
			b[i] = ' '
		default:
			if depth > 0 {
				b[i] = ' '
			}
		}
	}

	return string(b)
}

func hasEpisodes(tags Tags) bool {
	for _, season := range tags {
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

func dedupeSeasonEpisodes(tags Tags) {
	for si := range tags {
		seen := make(map[EpisodeRange]struct{}, len(tags[si].Episodes))
		out := tags[si].Episodes[:0]

		for _, ep := range tags[si].Episodes {
			key := EpisodeRange{Start: ep.Start, End: ep.End}
			if _, ok := seen[key]; ok {
				continue
			}

			seen[key] = struct{}{}
			out = append(out, ep)
		}

		tags[si].Episodes = out
	}
}

func ParseTags(s string) Tags {
	return parseTags(s, 1)
}

func parseTags(s string, fallbackSeason int) Tags {
	var tags Tags
	parseS := maskBracketContent(s)

	seasonTaggedMatches := seasonTaggedBareEpisodeRE.FindAllStringSubmatchIndex(parseS, -1)
	seasonRanges := seasonRangeRE.FindAllStringSubmatchIndex(parseS, -1)

	for _, m := range seasonRanges {
		raw := parseS[m[0]:m[1]]
		if seasonEpisodeRangeExceptionRE.MatchString(raw) {
			continue
		}

		start, _ := strconv.Atoi(parseS[m[2]:m[3]])
		end, _ := strconv.Atoi(parseS[m[4]:m[5]])
		if start >= end {
			continue
		}

		for season := start; season <= end; season++ {
			tags.AppendSeason(season)
		}
	}

	for _, m := range seasonOnlyRE.FindAllStringSubmatchIndex(parseS, -1) {
		if m[1] < len(s) && s[m[1]] >= '0' && s[m[1]] <= '9' {
			continue
		}
		if inSeasonRange(m[0], seasonRanges, parseS) {
			continue
		}

		n, _ := strconv.Atoi(s[m[2]:m[3]])
		tags.AppendSeason(n)
	}

	for _, m := range seasonWordRE.FindAllStringSubmatchIndex(parseS, -1) {
		rawSeason := s[m[2]:m[3]]
		if isRomanSeasonTag(rawSeason) && overlapsMatch(m[0], m[1], seasonTaggedMatches) {
			continue
		}

		if n, ok := parseSeasonNumber(rawSeason); ok {
			tags.AppendSeason(n)
		}
	}

	for _, m := range seasonOrdinalRE.FindAllStringSubmatchIndex(parseS, -1) {
		n, err := strconv.Atoi(s[m[2]:m[3]])
		if err == nil {
			tags.AppendSeason(n)
		}
	}

	for _, m := range seasonRomanWordRE.FindAllStringSubmatchIndex(parseS, -1) {
		if overlapsMatch(m[0], m[1], seasonTaggedMatches) {
			continue
		}

		rawSeason := s[m[2]:m[3]]
		if n, ok := romanToInt(rawSeason); ok {
			tags.AppendSeason(n)
		}
	}

	seasonEpMatches := seasonEpRE.FindAllStringSubmatchIndex(parseS, -1)
	for _, m := range seasonEpMatches {
		n, err := strconv.Atoi(s[m[2]:m[3]])
		if err == nil {
			tags.AppendSeason(n)
		}
	}

	for _, m := range episodeRE.FindAllStringSubmatchIndex(parseS, -1) {
		start, end := s[m[2]:m[3]], ""
		if m[4] >= 0 {
			end = s[m[4]:m[5]]
		}

		season := seasonBefore(m[0], seasonEpMatches, parseS)
		appendEpisode(&tags, season, fallbackSeason, start, end)
	}

	for _, m := range seasonTaggedMatches {
		isSeasonRange := false
		for _, rr := range seasonRanges {
			if m[0] >= rr[0] && m[0] < rr[1] {
				raw := parseS[rr[0]:rr[1]]
				if !seasonEpisodeRangeExceptionRE.MatchString(raw) {
					isSeasonRange = true
				}
				break
			}
		}
		if isSeasonRange {
			continue
		}

		var seasonRaw string
		switch {
		case m[2] >= 0:
			seasonRaw = strings.TrimSpace(s[m[2]:m[3]])
		case m[4] >= 0:
			seasonRaw = strings.TrimSpace(s[m[4]:m[5]])
		case m[6] >= 0:
			seasonRaw = strings.TrimSpace(s[m[6]:m[7]])
		case m[8] >= 0:
			seasonRaw = strings.TrimSpace(s[m[8]:m[9]])
		}

		season, ok := parseSeasonNumber(seasonRaw)
		if !ok {
			continue
		}
		tags.AppendSeason(season)

		epStart, epEnd := m[10], m[11]
		if epStart < 0 || epEnd < 0 {
			continue
		}

		appendEpisode(&tags, season, fallbackSeason, s[epStart:epEnd], "")
	}

	numericSeasonMatches := numericSeasonEpisodeRE.FindAllStringSubmatchIndex(parseS, -1)
	for _, m := range numericSeasonMatches {
		season, _ := strconv.Atoi(parseS[m[2]:m[3]])
		episode := parseS[m[4]:m[5]]
		tags.AppendSeason(season)
		appendEpisode(&tags, season, fallbackSeason, episode, "")
	}

	seasonXMatches := seasonXEpisodeRE.FindAllStringSubmatchIndex(parseS, -1)
	for _, m := range seasonXMatches {
		season, _ := strconv.Atoi(parseS[m[2]:m[3]])
		start := s[m[4]:m[5]]
		end := ""

		if m[6] >= 0 {
			end = s[m[6]:m[7]]
		}

		tags.AppendSeason(season)
		appendEpisode(&tags, season, fallbackSeason, start, end)
	}

	for _, m := range epRangeRE.FindAllStringSubmatchIndex(parseS, -1) {
		if overlapsMatch(m[0], m[1], numericSeasonMatches) ||
			overlapsMatch(m[0], m[1], seasonXMatches) {
			continue
		}

		a, _ := strconv.Atoi(parseS[m[2]:m[3]])
		b, _ := strconv.Atoi(parseS[m[4]:m[5]])

		if a >= 1900 && a <= 2099 && b >= 1900 && b <= 2099 {
			continue
		}
		if inSeasonRange(m[0], seasonRanges, parseS) {
			continue
		}

		season := seasonBefore(
			m[0],
			seasonOnlyRE.FindAllStringSubmatchIndex(parseS, -1),
			parseS,
		)
		appendEpisode(&tags, season, fallbackSeason, s[m[2]:m[3]], s[m[4]:m[5]])
	}

	for _, m := range bareTrailingEpisodeRE.FindAllStringSubmatchIndex(parseS, -1) {
		if overlapsMatch(m[0], m[1], numericSeasonMatches) ||
			overlapsMatch(m[0], m[1], seasonTaggedMatches) {
			continue
		}

		start := s[m[2]:m[3]]
		n, _ := strconv.Atoi(start)
		if n >= 1900 && n <= 2099 {
			continue
		}

		season := seasonBefore(
			m[0],
			seasonOnlyRE.FindAllStringSubmatchIndex(parseS, -1),
			parseS,
		)
		appendEpisode(&tags, season, fallbackSeason, start, "")
	}

	dedupeSeasonEpisodes(tags)

	return tags
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

func inSeasonRange(start int, ranges [][]int, s string) bool {
	for _, r := range ranges {
		if start >= r[0] && start < r[1] {
			return !seasonEpisodeRangeExceptionRE.MatchString(s[r[0]:r[1]])
		}
	}
	return false
}

func parseYear(s string, r *Metadata) {
	if m := yearRE.FindStringSubmatch(s); m != nil {
		r.Year, _ = strconv.Atoi(m[0])
	}
}

func parseTech(s string, r *Metadata) {
	appendUnique(&r.Resolutions, sliceutils.Map(resRE.FindAllString(s, -1), strings.ToLower)...)
	appendUnique(&r.Dimensions, sliceutils.Map(dimRE.FindAllString(s, -1), strings.ToLower)...)

	for _, x := range bitDepthRE.FindAllString(s, -1) {
		appendUnique(&r.BitDepths, x)
	}
	for _, x := range codecRE.FindAllString(s, -1) {
		appendUnique(&r.Codecs, normalizeCodec(x))
	}

	appendUnique(&r.AudioCodecs, audioCodecRE.FindAllString(s, -1)...)

	for _, x := range []string{
		"WEB-DL", "WEBRip", "CTHP", "WebRip", "BD", "BluRay",
		"AMZN", "CR", "HIDI", "HIDIVE", "IQIYI", "BILI", "LIV",
		"OV", "YTB", "VHS",
	} {
		if containsCI(s, x) {
			appendUnique(&r.Sources, x)
		}
	}

	for _, x := range []string{
		"Dual Audio", "Dual-Audio", "DUAL", "MULTi", "Multi-Audio",
		"Multi-Subs", "MultiSub", "English Dub", "English-Sub",
		"Korean Audio", "D-SUB", "M-SUB",
	} {
		if containsCI(s, x) {
			if strings.Contains(strings.ToLower(x), "sub") {
				appendUnique(&r.SubtitleFlags, x)
			} else {
				appendUnique(&r.AudioFlags, x)
			}
		}
	}

	for _, x := range encoderRE.FindAllString(s, -1) {
		appendUnique(&r.Encoders, x)
	}
	if containsCI(s, "END") {
		appendUnique(&r.ReleaseFlags, "END")
	}
}

func normalizeCodec(s string) string {
	switch strings.ToUpper(s) {
	case "H265":
		return "H.265"
	case "H264":
		return "H.264"
	default:
		return s
	}
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
			classifyParenthesis(t.Text, r)
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
		appendUnique(&r.ReleaseFlags, x)
		appendUnique(&r.Labels, x)
		return
	}

	if resRE.MatchString(x) || codecRE.MatchString(x) || audioCodecRE.MatchString(x) ||
		containsCI(x, "WEB") || containsCI(x, "BD") ||
		containsCI(x, "Multi") || containsCI(x, "Dub") {
		return
	}

	if containsAnyCI(x, "Complete", "Uncensored", "weekly") {
		appendUnique(&r.ReleaseFlags, x)
		appendUnique(&r.Labels, x)
		return
	}

	if !containsAnyCI(x, "Mini-Batch") {
		r.Unknown = append(r.Unknown, x)
	}
}

func classifyParenthesis(x string, r *Metadata) {
	x = strings.TrimSpace(x)
	if x == "" {
		return
	}

	if resRE.MatchString(x) || codecRE.MatchString(x) || audioCodecRE.MatchString(x) ||
		containsAnyCI(x, "WEB", "BD") {
		return
	}

	if hasEpisodes(r.Tags) && epRangeRE.MatchString(x) {
		return
	}

	if containsAnyCI(x,
		"Multi-Subs",
		"Multi-Audio",
		"Dual-Audio",
		"English-Sub",
		"Korean Audio",
	) {
		parts := strings.Split(x, ",")
		var aliases []string

		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}

			if containsAnyCI(
				p,
				"Multi-Subs",
				"Multi-Audio",
				"Dual-Audio",
				"English-Sub",
				"Korean Audio",
			) {
				appendUnique(&r.ReleaseFlags, p)
				continue
			}

			aliases = append(aliases, p)
		}

		appendUnique(&r.AlternateTitles, aliases...)
		return
	}

	if strings.Contains(x, "|") || strings.Contains(x, ";") {
		appendUnique(&r.AlternateTitles, splitAlias(x)...)
		return
	}

	if yearRE.MatchString(x) {
		return
	}

	if strings.Contains(x, ",") {
		appendUnique(&r.AlternateTitles, splitAlias(x)...)
		return
	}

	if !containsAnyCI(x,
		"weekly",
		"batch",
		"dual",
		"multi-sub",
		"multi-audio",
	) {
		appendUnique(&r.AlternateTitles, x)
	}
}

func splitAlias(x string) []string {
	var out []string

	for _, p := range strings.FieldsFunc(x, func(r rune) bool {
		return r == '|' || r == ';'
	}) {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}

	return out
}

// ExtractPrimaryTitle extracts the primary title without performing the full
// metadata parse.
func ExtractPrimaryTitle(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	r := Metadata{
		Tags: ParseTags(raw),
	}
	if strings.HasPrefix(raw, "[") && matchingClose(raw, 0, '[', ']') >= 0 {
		r.Group = "[group]"
	}

	parseMainText(raw, &r)
	return r.PrimaryTitle
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
	s = stripFilenameExtension(s)

	hasEp := hasEpisodes(r.Tags)

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
	s = tagResidueRE.ReplaceAllString(s, "")
	s = normalizeFilenameSeparators(s)
	s = releaseSuffixRE.ReplaceAllString(s, "")
	s = strings.Join(strings.Fields(s), " ")

	if i := strings.Index(s, " - "); i >= 0 {
		right := strings.TrimSpace(s[i+3:])
		if leadingEpisodeNumberRE.MatchString(right) {
			s = strings.TrimSpace(s[:i])
		}
	}

	s = seasonEpRE.ReplaceAllString(s, "")
	if hasEp {
		s = episodeRE.ReplaceAllString(s, "")
		s = epRangeRE.ReplaceAllString(s, "")
	}

	if len(r.Tags) > 0 {
		s = seasonEpRE.ReplaceAllString(s, "")
		s = seasonRangeRE.ReplaceAllString(s, "")
		s = seasonTaggedBareEpisodeRE.ReplaceAllString(s, "")
		s = seasonWordRE.ReplaceAllString(s, "")
		s = seasonOrdinalRE.ReplaceAllString(s, "")
		s = seasonRomanWordRE.ReplaceAllString(s, "")
		s = seasonOnlyRE.ReplaceAllString(s, "")
	}

	if hasEp {
		s = seasonTaggedBareEpisodeRE.ReplaceAllString(s, "")
		s = bareTrailingEpisodeRE.ReplaceAllString(s, "")
	}

	s = yearRE.ReplaceAllString(s, "")
	s = multipleSpaceRE.ReplaceAllString(s, " ")
	s = strings.Trim(s, " -|:")

	if s == "" {
		return
	}

	if hasEp {
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
	r.PrimaryTitle = cleanTitle(r.PrimaryTitle)
}

func stripFilenameExtension(s string) string {
	return filenameExtensionRE.ReplaceAllString(strings.TrimSpace(s), "")
}

func normalizeFilenameSeparators(s string) string {
	var b strings.Builder

	for i := 0; i < len(s); i++ {
		if s[i] != '.' {
			b.WriteByte(s[i])
			continue
		}

		if i > 0 && i+1 < len(s) &&
			s[i-1] >= '0' && s[i-1] <= '9' &&
			s[i+1] >= '0' && s[i+1] <= '9' {
			b.WriteByte('.')
			continue
		}

		b.WriteByte(' ')
	}

	return b.String()
}

func cleanTitle(s string) string {
	s = strings.TrimSpace(s)
	s = remasteredSuffixRE.ReplaceAllString(s, "")
	s = resolutionSuffixRE.ReplaceAllString(s, "")
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
	return startsTechRE.MatchString(s) || startsEpisodeRE.MatchString(s)
}

func cleanEpisodeTitle(s string) string {
	for _, sep := range []string{"[", "(", "|"} {
		if i := strings.Index(s, sep); i >= 0 {
			s = s[:i]
		}
	}

	if m := episodeTitleResolutionRE.FindStringIndex(s); m != nil {
		s = s[:m[0]]
	}
	if m := episodeTitleSourceRE.FindStringIndex(s); m != nil {
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

	return append(out, s[start:])
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

	for p := range strings.SplitSeq(x, ";") {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}

		kv := strings.SplitN(p, "=", 2)
		if len(kv) == 2 {
			k, v := strings.TrimSpace(kv[0]), strings.TrimSpace(kv[1])
			r.TagFields[k] = v

			switch k {
			case "A":
				appendUnique(&r.AudioLanguages, strings.Split(v, ",")...)
			case "S":
				appendUnique(&r.SubtitleLanguages, strings.Split(v, ",")...)
			}
			continue
		}

		if m := tagFieldRE.FindStringSubmatch(p); len(m) == 3 {
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

func appendUnique[T ~[]string](xs *T, values ...string) {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}

		duplicate := false
		for _, existing := range *xs {
			if strings.EqualFold(strings.TrimSpace(existing), value) {
				duplicate = true
				break
			}
		}

		if !duplicate {
			*xs = append(*xs, value)
		}
	}
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
