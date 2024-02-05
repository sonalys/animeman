package qbittorrent

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"strings"
)

var ErrUnauthorized = fmt.Errorf("unauthorized")

func NewErrConnection(err error) error {
	return fmt.Errorf("connection error: %w", err)
}

type TorrentURL []string

func (t TorrentURL) ApplyAddTorrent(w *multipart.Writer) {
	field, err := w.CreateFormField("urls")
	if err != nil {
		panic(err)
	}
	if _, err := io.WriteString(field, strings.Join(t, "\n")); err != nil {
		panic(err)
	}
}

type Tags []string

func (t Tags) ApplyAddTorrent(w *multipart.Writer) {
	field, err := w.CreateFormField("tags")
	if err != nil {
		panic(err)
	}
	if _, err := io.WriteString(field, strings.Join(t, ",")); err != nil {
		panic(err)
	}
}

type SavePath string

func (s SavePath) ApplyAddTorrent(w *multipart.Writer) {
	field, err := w.CreateFormField("savepath")
	if err != nil {
		panic(err)
	}
	io.WriteString(field, string(s))
}

type Category string

func (s Category) ApplyAddTorrent(w *multipart.Writer) {
	field, err := w.CreateFormField("category")
	if err != nil {
		panic(err)
	}
	io.WriteString(field, string(s))
}

type Torrent struct {
	Name string `json:"name"`
}

func (t Tags) ApplyListTorrent(v url.Values) {
	for _, tag := range t {
		v.Add("tag", url.QueryEscape(tag))
	}
}
