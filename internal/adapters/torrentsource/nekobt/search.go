package nekobt

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/rs/zerolog/log"
	"github.com/sonalys/animeman/internal/parser"
	"github.com/sonalys/animeman/internal/ports/animelist"
	"github.com/sonalys/animeman/internal/ports/torrentsource"
	"github.com/sonalys/animeman/internal/tags"
	"github.com/sonalys/animeman/internal/utils"
	"github.com/sonalys/animeman/internal/utils/nyaaquerier"
)

const (
	// pageSize is the torznab `limit` per page.
	pageSize = 100
	// maxPages caps pagination so a pathological feed can't loop forever.
	maxPages = 5
)

type (
	rss struct {
		XMLName xml.Name `xml:"rss"`
		Channel channel  `xml:"channel"`
	}

	channel struct {
		Title string `xml:"title"`
		Items []item `xml:"item"`
	}

	item struct {
		Title   string `xml:"title"`
		Link    string `xml:"link"`
		GUID    string `xml:"guid"`
		PubDate string `xml:"pubDate"`
		Size    int64  `xml:"size"`
		Attrs   []attr `xml:"attr"`
	}

	attr struct {
		Name  string `xml:"name,attr"`
		Value string `xml:"value,attr"`
	}
)

// Search implements torrentsource.Source.
// nekoBT matches the entry by AniList/MAL id directly, so no title filtering is needed.
// Results are sorted oldest-first by the source, so when the newest tag already
// downloaded in the torrent client is older than everything on the first page,
// we keep paginating until we find something newer (or run out of pages).
func (api *API) Search(
	ctx context.Context,
	entry animelist.Entry,
	opts torrentsource.SearchOptions,
) ([]torrentsource.Torrent, error) {
	base, err := api.buildQuery(entry, opts)
	if err != nil {
		return nil, fmt.Errorf("building query: %w", err)
	}

	var torrents []torrentsource.Torrent

	for offset := 0; offset < maxPages*pageSize; offset += pageSize {
		items, err := api.fetchPage(ctx, base, offset)
		if err != nil {
			return nil, err
		}

		if len(items) == 0 {
			break
		}

		filtered := utils.Filter(items,
			filterSeeders(1),
			filterMetadata(entry),
			filterSources(opts.Sources),
		)

		torrents = append(torrents, utils.Map(filtered, func(item item) torrentsource.Torrent {
			return torrentsource.Torrent{
				Title:   item.Title,
				Link:    item.Link,
				Seeders: item.seeders(),
				Hash:    item.attr("infohash"),
			}
		})...)

		// Pages are sorted oldest-first: if the smallest tag on this page is
		// still older than (or equal to) the latest downloaded tag, everything
		// on later pages can only be newer, so keep going.
		if !shouldPaginate(items, opts.LatestTag) {
			break
		}
	}

	torrents = parser.Prioritize(entry, torrents, opts)

	log.
		Ctx(ctx).
		Debug().
		Int("results", len(torrents)).
		Msg("search results")

	return torrents, nil
}

// shouldPaginate reports whether the source may have more results after this
// page. nekoBT returns newest results first, so paginate while even the
// smallest tag found remains newer than the latest downloaded tag.
func shouldPaginate(items []item, latestTag tags.Tag) bool {
	if latestTag.IsZero() {
		// Nothing downloaded yet: the first page already has everything.
		return false
	}

	if len(items) == 0 {
		return false
	}

	smallest := tags.Tag{}
	for _, it := range items {
		tag := parser.Parse(it.Title, 1, nil).Tag
		if smallest.IsZero() || tag.Compare(smallest) < 0 {
			smallest = tag
		}
	}

	return smallest.Compare(latestTag) > 0
}

// fetchPage fetches one torznab result page at the given offset.
func (api *API) fetchPage(ctx context.Context, baseQuery string, offset int) ([]item, error) {
	req := utils.Must(http.NewRequestWithContext(ctx, http.MethodGet, TORZNAB_URL, nil))

	values, err := url.ParseQuery(baseQuery)
	if err != nil {
		return nil, fmt.Errorf("parsing query: %w", err)
	}
	values.Set("offset", strconv.Itoa(offset))
	values.Set("limit", strconv.Itoa(pageSize))
	req.URL.RawQuery = values.Encode()

	resp, err := api.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching response: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("request failed: %s", string(utils.Must(io.ReadAll(resp.Body))))
	}

	var feed rss
	if err := xml.NewDecoder(resp.Body).Decode(&feed); err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	return feed.Channel.Items, nil
}

func (api *API) buildQuery(entry animelist.Entry, opt torrentsource.SearchOptions) (string, error) {
	q := url.Values{}

	q.Set("t", "search")

	mediaID, err := resolveMediaID(entry)
	if err != nil {
		return "", err
	}

	q.Set("media_id", mediaID)

	// Quality tokens that are video types or codecs are sent as dedicated
	// torznab params (video_type, video_codec) instead of `q` tokens.
	_, params := splitQualityFilters(opt.Qualities)
	maps.Copy(q, params)

	// media_id already narrows to the entry, so `q` only carries the
	// remaining quality filters and the user suffix, same format as nyaa.
	// Sources are NOT sent in `q`: they are release-group names, and nekoBT's
	// `|` OR operator is unreliable (any OR alternative matching nothing
	// returns zero results). They are filtered from the results instead.
	if query := buildQuery(opt); query != "" {
		q.Set("q", query)
	}

	if api.config.APIKey != "" {
		q.Set("apikey", api.config.APIKey)
	}

	// Apply user-configured torznab parameters last, so they can
	// override anything built above (sort, sub_lang, mtl, hardsub, ...).
	for name, value := range api.config.CustomParameters {
		q.Set(name, value)
	}

	return q.Encode(), nil
}

// buildQuery builds the `q` search query from the search options.
// Quality tokens that are video types or codecs are extracted into torznab
// params by splitQualityFilters, so `q` only ever carries resolutions and
// unknown tokens.
//
// nekoBT's `|` OR operator is unreliable: any OR expression containing an
// alternative that matches nothing returns zero results (verified against
// the live endpoint). To keep OR groups safe, quality tokens present in
// EVERY configured quality are hoisted into a plain AND (e.g. `1080` from
// ["1080", "1080", "1080"]), and the remaining tokens are
// clustered into OR dimensions of tokens that never co-occur (e.g.
// ["1080", "720"] becomes `(1080|720)`).
func buildQuery(opt torrentsource.SearchOptions) string {
	var parts []nyaaquerier.Node

	text, _ := splitQualityFilters(opt.Qualities)
	if common, dimensions := splitQualities(text); len(common) > 0 || len(dimensions) > 0 {
		for _, token := range common {
			parts = append(parts, nyaaquerier.PhraseOf(token))
		}
		for _, dimension := range dimensions {
			parts = append(parts, nyaaquerier.Or(utils.Map(dimension, nyaaquerier.PhraseOf)))
		}
	}

	if opt.SearchSuffix != "" {
		// The suffix is user-provided query syntax (e.g. `-"dub"`), pass it
		// through verbatim instead of sanitizing it into a phrase.
		parts = append(parts, nyaaquerier.Raw(opt.SearchSuffix))
	}

	return nyaaquerier.And(parts).String()
}

// splitQualityFilters extracts nekoBT torznab filters from the configured
// qualities. Tokens matching a video type or codec (see
// https://wiki.nekobt.to/info/metadata/) become video_type/video_codec
// params; the remaining tokens (resolutions, unknown words) are returned
// as text qualities for the `q` query. A quality whose tokens are all
// consumed contributes no text.
func splitQualityFilters(qualities []string) (text []string, params url.Values) {
	params = url.Values{}
	var codecs, types []string

	for _, quality := range qualities {
		var leftover []string
		for token := range strings.FieldsSeq(quality) {
			if codec, ok := videoCodecAliases[normalizeToken(token)]; ok {
				codecs = append(codecs, codec)
				continue
			}
			if videoType, ok := videoTypeAliases[normalizeToken(token)]; ok {
				types = append(types, videoType)
				continue
			}
			leftover = append(leftover, token)
		}
		if len(leftover) > 0 {
			text = append(text, strings.Join(leftover, " "))
		}
	}

	if len(codecs) > 0 {
		slices.Sort(codecs)
		params.Set("video_codec", strings.Join(slices.Compact(codecs), ","))
	}
	if len(types) > 0 {
		slices.Sort(types)
		params.Set("video_type", strings.Join(slices.Compact(types), ","))
	}

	return text, params
}

// normalizeToken lowercases a token and strips dots/dashes/spaces so
// aliases like `H.265`, `h265` and `H-265` all match.
func normalizeToken(token string) string {
	return strings.Map(func(r rune) rune {
		switch r {
		case '.', '-', ' ':
			return -1
		default:
			return unicode.ToLower(r)
		}
	}, token)
}

// videoCodecAliases maps normalized quality tokens to nekoBT video_codec
// filter values (numeric IDs from the metadata lookup table).
// https://wiki.nekobt.to/info/metadata/
var videoCodecAliases = map[string]string{
	"h264":  "1", // H264 (AVC, x264)
	"avc":   "1",
	"x264":  "1",
	"h265":  "2", // H265 (HEVC, x265)
	"hevc":  "2",
	"x265":  "2",
	"av1":   "3", // AV1
	"vp9":   "4", // VP9
	"mpeg2": "5", // MPEG-2
	"mpeg4": "6", // MPEG-4
	"wmv":   "7", // WMV
	"vc1":   "8", // VC1
}

// videoTypeAliases maps normalized quality tokens to nekoBT video_type
// filter values (numeric IDs from the metadata lookup table).
// https://wiki.nekobt.to/info/metadata/
var videoTypeAliases = map[string]string{
	"hybrid":    "15", // Hybrid
	"remux":     "14", // BD - Remux
	"bdremux":   "14",
	"bdencode":  "13", // BD - Encode
	"bdmini":    "12", // BD - Mini
	"bd":        "11", // BD - Disc
	"bluray":    "11",
	"web":       "9", // WEB-DL
	"webdl":     "9",
	"webencode": "8",  // WEB - Encode
	"webmini":   "7",  // WEB - Mini
	"dvdremux":  "5",  // DVD - Remux
	"dvdencode": "6",  // DVD - Encode
	"dvd":       "16", // DVD - Disc
	"tvraw":     "4",  // TV - Raw
	"tvencode":  "3",  // TV - Encode
	"laserdisc": "2",  // LaserDisc
	"vhs":       "1",  // VHS
}

// splitQualities splits the configured qualities into tokens shared by all
// of them (common) and OR dimensions for the rest. A dimension is a set of
// tokens that never co-occur in any quality, so exactly one of them matches
// a title and the OR group never contains an alternative that matches
// nothing. Example: ["1080", "720"]
// yields common=nil and dimensions=[[1080 720]].
func splitQualities(qualities []string) (common []string, dimensions [][]string) {
	if len(qualities) == 0 {
		return nil, nil
	}

	counts := make(map[string]int)
	for _, quality := range qualities {
		seen := make(map[string]struct{})
		for token := range strings.FieldsSeq(quality) {
			if _, ok := seen[token]; ok {
				continue
			}
			seen[token] = struct{}{}
			counts[token]++
		}
	}

	for token, count := range counts {
		if count == len(qualities) {
			common = append(common, token)
		}
	}
	slices.Sort(common)

	// Cluster the remaining tokens: two tokens belong to the same dimension
	// iff they never appear together in a quality. Greedy: assign each token
	// to the first dimension compatible with all its members.
	var dims [][]string
	for token := range counts {
		if counts[token] == len(qualities) {
			continue
		}
		placed := false
		for i, dim := range dims {
			compatible := true
			for _, member := range dim {
				if coOccurs(token, member, qualities) {
					compatible = false
					break
				}
			}
			if compatible {
				dims[i] = append(dim, token)
				placed = true
				break
			}
		}
		if !placed {
			dims = append(dims, []string{token})
		}
	}

	for _, dim := range dims {
		slices.Sort(dim)
		dimensions = append(dimensions, dim)
	}
	slices.SortFunc(dimensions, func(a, b []string) int {
		return slices.Compare(a, b)
	})

	return common, dimensions
}

// coOccurs reports whether two tokens appear together in any quality.
func coOccurs(a, b string, qualities []string) bool {
	for _, quality := range qualities {
		fields := strings.Fields(quality)
		if slices.Contains(fields, a) && slices.Contains(fields, b) {
			return true
		}
	}
	return false
}

// resolveMediaID returns the nekoBT external id for the entry, e.g. `anilist-20594`.
func resolveMediaID(entry animelist.Entry) (string, error) {
	switch {
	case entry.AnilistID != 0:
		return fmt.Sprintf("anilist-%d", entry.AnilistID), nil
	case entry.MALID != 0:
		return fmt.Sprintf("mal-%d", entry.MALID), nil
	default:
		return "", fmt.Errorf("entry %q has no anilist or mal id", strings.Join(entry.Titles, ", "))
	}
}

// filterSources keeps only torrents whose title contains one of the
// configured sources (release groups), mirroring how the parser extracts
// the release group.
func filterSources(sources []string) func(item) bool {
	return func(item item) bool {
		if len(sources) == 0 {
			return true
		}
		return parser.Parse(item.Title, 1, sources).ReleaseGroup != ""
	}
}

func filterSeeders(minSeeders int) func(item) bool {
	return func(item item) bool {
		return item.seeders() >= minSeeders
	}
}

func filterMetadata(entry animelist.Entry) func(item) bool {
	return func(item item) bool {
		// Compare the published date of the torrent with the entry's start and end dates.
		pubDate, err := time.Parse(time.RFC1123Z, item.PubDate)
		if err != nil {
			return false
		}

		// Compares publishing date with anime start date, 2 days offset to prevent wrong timezone and hour precision.
		if !entry.StartDate.IsZero() && pubDate.Before(entry.StartDate.AddDate(0, 0, -2)) {
			return false
		}

		return true
	}
}

func (i item) attr(name string) string {
	for _, a := range i.Attrs {
		if a.Name == name {
			return a.Value
		}
	}
	return ""
}

func (i item) seeders() int {
	v, err := strconv.Atoi(i.attr("seeders"))
	if err != nil {
		return 0
	}
	return v
}
