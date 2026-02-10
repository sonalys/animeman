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
	"github.com/sonalys/animeman/internal/utils"
)

func digestArg(arg *domain.AddTorrentConfig) (io.Reader, string) {
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

	if arg.Name != nil {
		field = utils.Must(w.CreateFormField("rename"))
		utils.Must(io.WriteString(field, fmt.Sprint(*arg.Name)))
	}

	return &b, w.FormDataContentType()
}

func (api *API) AddTorrent(ctx context.Context, arg *domain.AddTorrentConfig) error {
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
