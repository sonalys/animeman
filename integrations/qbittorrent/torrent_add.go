package qbittorrent

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/sonalys/animeman/internal/utils"
	"github.com/sonalys/animeman/pkg/v1/torrentclient"
)

func digestArg(arg *torrentclient.AddTorrentConfig) (io.Reader, string) {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	field := utils.Must(w.CreateFormField("urls"))
	utils.Must(io.WriteString(field, strings.Join(arg.URLs, "\n")))
	field = utils.Must(w.CreateFormField("tags"))
	utils.Must(io.WriteString(field, strings.Join(arg.Tags, ",")))
	field = utils.Must(w.CreateFormField("category"))
	utils.Must(io.WriteString(field, fmt.Sprint(arg.Category)))
	field = utils.Must(w.CreateFormField("paused"))
	utils.Must(io.WriteString(field, fmt.Sprint(arg.Paused)))
	field = utils.Must(w.CreateFormField("savepath"))
	utils.Must(io.WriteString(field, fmt.Sprint(arg.SavePath)))
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
	resp.Body.Close()
	return nil
}
