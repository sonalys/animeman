package qbittorrent

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/sonalys/animeman/internal/domain"
	"github.com/sonalys/animeman/internal/utils/errutils"
)

func digestArg(arg *domain.AddTorrentConfig) (io.Reader, string) {
	var b bytes.Buffer

	w := multipart.NewWriter(&b)

	field := errutils.Must(w.CreateFormField("urls"))
	errutils.Must(io.WriteString(field, strings.Join(arg.URLs, "\n")))
	field = errutils.Must(w.CreateFormField("tags"))
	errutils.Must(io.WriteString(field, strings.Join(arg.Tags, ",")))
	field = errutils.Must(w.CreateFormField("category"))
	errutils.Must(io.WriteString(field, fmt.Sprint(arg.Category)))
	field = errutils.Must(w.CreateFormField("paused"))
	errutils.Must(io.WriteString(field, fmt.Sprint(arg.Paused)))
	field = errutils.Must(w.CreateFormField("savepath"))
	errutils.Must(io.WriteString(field, fmt.Sprint(arg.SavePath)))

	if arg.Name != nil {
		field = errutils.Must(w.CreateFormField("rename"))
		errutils.Must(io.WriteString(field, fmt.Sprint(*arg.Name)))
	}

	return &b, w.FormDataContentType()
}

func (api *Client) AddTorrent(ctx context.Context, arg *domain.AddTorrentConfig) error {
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
