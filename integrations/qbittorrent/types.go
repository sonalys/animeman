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

func (t Tags) ApplyAddTorrentTagsArg(v url.Values) {
	v.Set("tags", strings.Join(t, ","))
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

func (c Category) ApplyAddTorrent(w *multipart.Writer) {
	field, err := w.CreateFormField("category")
	if err != nil {
		panic(err)
	}
	io.WriteString(field, string(c))
}

func (c Category) ApplyListTorrent(v url.Values) {
	v.Add("category", url.QueryEscape(string(c)))
}

type Torrent struct {
	Name     string `json:"name"`
	Category string `json:"category"`
	Hash     string `json:"hash"`
}

func (t Tags) ApplyListTorrent(v url.Values) {
	for _, tag := range t {
		v.Add("tag", url.QueryEscape(tag))
	}
}

type Paused bool

func (p Paused) ApplyAddTorrent(w *multipart.Writer) {
	field, err := w.CreateFormField("paused")
	if err != nil {
		panic(err)
	}
	io.WriteString(field, fmt.Sprint(p))
}
