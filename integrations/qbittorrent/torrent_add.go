package qbittorrent

import (
	"bytes"
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
)

type ArgAddTorrent interface {
	ApplyAddTorrent(*multipart.Writer)
}

func (api *API) AddTorrent(ctx context.Context, args ...ArgAddTorrent) error {
	var path = api.host + "/torrents/add"
	var b bytes.Buffer
	formdata := multipart.NewWriter(&b)
	for _, f := range args {
		f.ApplyAddTorrent(formdata)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, path, &b)
	if err != nil {
		return fmt.Errorf("creating request failed: %w", err)
	}
	req.Header.Set("Content-Type", formdata.FormDataContentType())
	resp, err := api.Do(req)
	if err != nil {
		return fmt.Errorf("post torrents/add failed: %w", err)
	}
	resp.Body.Close()
	return nil
}
