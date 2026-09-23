package qbittorrent

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/sonalys/animeman/internal/ports/torrentclient"
	"github.com/sonalys/animeman/internal/utils/must"
)

func digestArg(arg *torrentclient.AddTorrentConfig) (io.Reader, string) {
	var b bytes.Buffer

	w := multipart.NewWriter(&b)

	field := must.Must(w.CreateFormField("urls"))
	must.Must(io.WriteString(field, strings.Join(arg.URLs, "\n")))
	field = must.Must(w.CreateFormField("tags"))
	must.Must(io.WriteString(field, strings.Join(arg.Tags, ",")))
	field = must.Must(w.CreateFormField("category"))
	must.Must(io.WriteString(field, fmt.Sprint(arg.Category)))
	field = must.Must(w.CreateFormField("paused"))
	must.Must(io.WriteString(field, fmt.Sprint(arg.Paused)))
	field = must.Must(w.CreateFormField("savepath"))
	must.Must(io.WriteString(field, fmt.Sprint(arg.SavePath)))

	if arg.Name != nil {
		field = must.Must(w.CreateFormField("rename"))
		must.Must(io.WriteString(field, fmt.Sprint(*arg.Name)))
	}

	return &b, w.FormDataContentType()
}

func (api *API) AddTorrent(ctx context.Context, arg *torrentclient.AddTorrentConfig) error {
	var path = api.host + "/torrents/add"
	r, contentType := digestArg(arg)

	req, err := http.NewRequest(http.MethodPost, path, r)
	if err != nil {
		return fmt.Errorf("creating request failed: %w", err)
	}

	req.Header.Set("Content-Type", contentType)

	resp, err := api.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("post torrents/add failed: %w", err)
	}

	return resp.Body.Close()
}
