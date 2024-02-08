package torrentclient

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"strings"
)

var ErrUnauthorized = fmt.Errorf("unauthorized")

type (
	TorrentURL []string
	Tag        string
	Tags       []string
	SavePath   string
	Category   string
	Paused     bool

	Torrent struct {
		Name     string
		Category string
		Hash     string
		Tags     []string
	}
)

type ArgListTorrent interface {
	ApplyListTorrent(url.Values)
}

type ArgAddTorrent interface {
	ApplyAddTorrent(*multipart.Writer)
}

type AddTorrentTagsArg interface {
	ApplyAddTorrentTags(url.Values)
}

func (t Tag) ApplyListTorrent(v url.Values) {
	v.Add("tag", string(t))
}

func (c Category) ApplyListTorrent(v url.Values) {
	v.Add("category", string(c))
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
