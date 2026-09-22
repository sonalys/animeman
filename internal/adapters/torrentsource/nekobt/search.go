package nekobt

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
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
	"github.com/sonalys/animeman/internal/utils"
	"github.com/sonalys/animeman/internal/utils/nyaaquerier"
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
func (api *API) Search(
	ctx context.Context,
	entry animelist.Entry,
	opts torrentsource.SearchOptions,
) ([]torrentsource.Torrent, error) {
	req := utils.Must(http.NewRequestWithContext(ctx, http.MethodGet, TORZNAB_URL, nil))

	q, err := api.buildQuery(entry, opts)
	if err != nil {
		return nil, fmt.Errorf("building query: %w", err)
	}

	req.URL.RawQuery = q

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

	items := utils.Filter(feed.Channel.Items,
		filterSeeders(1),
		filterMetadata(entry),
	)

	torrents := utils.Map(items, func(item item) torrentsource.Torrent {
		return torrentsource.Torrent{
			Title:   item.Title,
			Link:    item.Link,
			Seeders: item.seeders(),
			Hash:    item.attr("infohash"),
		}
	})

	torrents = parser.Prioritize(entry, torrents, opts)

	log.
		Ctx(ctx).
		Debug().
		Int("results", len(torrents)).
		Msg("search results")

	return torrents, nil
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
	for name, values := range params {
		q[name] = values
	}

	// media_id already narrows to the entry, so `q` only carries the
	// remaining quality/source filters and the user suffix, same format as nyaa.
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
// ["1080", "720"] becomes `(1080|720)`). Sources stay an OR: they are
// release-group names that nekoBT indexes, and at least one always matches.
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

	if len(opt.Sources) > 0 {
		sourceNodes := utils.Map(opt.Sources, nyaaquerier.PhraseOf)
		parts = append(parts, nyaaquerier.Or(sourceNodes))
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
		for _, token := range strings.Fields(quality) {
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
// filter values. https://wiki.nekobt.to/info/metadata/
var videoCodecAliases = map[string]string{
	"h264":  "H264",
	"avc":   "H264",
	"x264":  "H264",
	"h265":  "H265",
	"hevc":  "H265",
	"x265":  "H265",
	"av1":   "AV1",
	"vp9":   "VP9",
	"mpeg2": "MPEG-2",
	"mpeg4": "MPEG-4",
	"wmv":   "WMV",
	"vc1":   "VC1",
}

// videoTypeAliases maps normalized quality tokens to nekoBT video_type
// filter values. https://wiki.nekobt.to/info/metadata/
var videoTypeAliases = map[string]string{
	"hybrid":    "Hybrid",
	"remux":     "BD - Remux",
	"bdremux":   "BD - Remux",
	"bdencode":  "BD - Encode",
	"bdmini":    "BD - Mini",
	"bd":        "BD - Disc",
	"bluray":    "BD - Disc",
	"web":       "WEB-DL",
	"webdl":     "WEB-DL",
	"webencode": "WEB - Encode",
	"webmini":   "WEB - Mini",
	"dvdremux":  "DVD - Remux",
	"dvdencode": "DVD - Encode",
	"dvd":       "DVD - Disc",
	"tvraw":     "TV - Raw",
	"tvencode":  "TV - Encode",
	"laserdisc": "LaserDisc",
	"vhs":       "VHS",
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
