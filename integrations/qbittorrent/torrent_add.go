package qbittorrent

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
)

type ArgAddTorrent interface {
	ApplyAddTorrent(*multipart.Writer)
}

func (t TorrentURL) ApplyAddTorrent(w *multipart.Writer) {
	field, err := w.CreateFormField("urls")
	if err != nil {
		panic(err)
	}
	if _, err := io.WriteString(field, strings.Join(t, "\n")); err != nil {
		panic(err)
	}
}

func (t Tags) ApplyAddTorrent(w *multipart.Writer) {
	field, err := w.CreateFormField("tags")
	if err != nil {
		panic(err)
	}
	if _, err := io.WriteString(field, strings.Join(t, ",")); err != nil {
		panic(err)
	}
}

func (c Category) ApplyAddTorrent(w *multipart.Writer) {
	field, err := w.CreateFormField("category")
	if err != nil {
		panic(err)
	}
	io.WriteString(field, string(c))
}

func (p Paused) ApplyAddTorrent(w *multipart.Writer) {
	field, err := w.CreateFormField("paused")
	if err != nil {
		panic(err)
	}
	io.WriteString(field, fmt.Sprint(p))
}

func (s SavePath) ApplyAddTorrent(w *multipart.Writer) {
	field, err := w.CreateFormField("savepath")
	if err != nil {
		panic(err)
	}
	io.WriteString(field, string(s))
}

func (api *API) AddTorrent(ctx context.Context, args ...ArgAddTorrent) error {
	var path = api.host + "/torrents/add"
	var b bytes.Buffer
	formdata := multipart.NewWriter(&b)
	for _, f := range args {
		f.ApplyAddTorrent(formdata)
	}
	req, err := http.NewRequest(http.MethodPost, path, &b)
	if err != nil {
		return fmt.Errorf("creating request failed: %w", err)
	}
	req.Header.Set("Content-Type", formdata.FormDataContentType())
	resp, err := api.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("post torrents/add failed: %w", err)
	}
	resp.Body.Close()
	return nil
}
