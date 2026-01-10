package nyaa

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/sonalys/animeman/internal/utils"
)

type ListOptions struct {
	Titles              []string
	VerticalResolutions []string
	Sources             []string
}

func (opt ListOptions) Query() string {
	var b strings.Builder

	titles := utils.Map(opt.Titles, strconv.Quote)
	fmt.Fprintf(&b, "(%s)", strings.Join(titles, "|"))

	if resolutions := opt.VerticalResolutions; len(resolutions) > 0 {
		fmt.Fprintf(&b, " (%s)", strings.Join(resolutions, "|"))
	}

	if sources := opt.Sources; len(sources) > 0 {
		fmt.Fprintf(&b, " (%s)", strings.Join(sources, "|"))
	}

	return b.String()
}

func (api *API) List(ctx context.Context, options ListOptions) ([]Entry, error) {
	var path = API_URL

	req := utils.Must(http.NewRequestWithContext(ctx, http.MethodGet, path, nil))

	q := req.URL.Query()
	for name, value := range api.config.ListParameters {
		q.Set(name, value)
	}

	q.Add("q", options.Query())

	req.URL.RawQuery = q.Encode()

	t1 := time.Now()

	resp, err := api.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching response: %w", err)
	}
	defer resp.Body.Close()

	log.
		Trace().
		Str("url", req.URL.String()).
		Int("status_code", resp.StatusCode).
		Dur("duration", time.Since(t1)).
		Msg("rss search finished")

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("request failed: %s", string(utils.Must(io.ReadAll(resp.Body))))
	}

	var feed RSS
	if err := xml.NewDecoder(resp.Body).Decode(&feed); err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	return feed.Channel.Entries, nil
}
