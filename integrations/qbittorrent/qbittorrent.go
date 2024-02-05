package qbittorrent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

type (
	API struct {
		host   string
		client *http.Client
	}

	ArgAddTorrent interface {
		ApplyAddTorrent(*multipart.Writer)
	}

	ArgListTorrent interface {
		ApplyListTorrent(url.Values)
	}
)

func New(host, username, password string) *API {
	jar, err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}
	api := &API{
		host: fmt.Sprintf("%s/api/v2", host),
		client: &http.Client{
			Timeout: 3 * time.Second,
			Jar:     jar,
		},
	}
	log.Info().Msg("connecting to qBitTorrent")
	if _, err := api.Version(); err != nil {
		if err := api.Login(username, password); err != nil {
			log.Fatal().Msgf("could not initialize qBittorrent: %s", err)
		}
	}
	return api
}

func (api *API) Version() (string, error) {
	var path = api.host + "/app/version"
	resp, err := api.client.Get(path)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("unauthorized")
	}
	version, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	return string(version), nil
}

func (api *API) Login(username, password string) error {
	var path = api.host + "/auth/login"
	req, err := http.NewRequest(http.MethodPost, path, nil)
	if err != nil {
		return fmt.Errorf("login request creation failed: %w", err)
	}
	req.URL.RawQuery = url.Values{
		"username": []string{username},
		"password": []string{password},
	}.Encode()
	resp, err := api.client.Do(req)
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("login failed: %w", err)
	}
	return nil
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
	resp, err := api.client.Do(req)
	if err != nil {
		return fmt.Errorf("post torrents/add failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("unauthorized")
	}
	return nil
}

type Torrent struct {
	Name string `json:"name"`
}

func (t Tags) ApplyListTorrent(v url.Values) {
	for _, tag := range t {
		v.Add("tag", url.QueryEscape(tag))
	}
}

func (api *API) List(ctx context.Context, args ...ArgListTorrent) ([]Torrent, error) {
	var path = api.host + "/torrents/info"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("list request failed: %w", err)
	}
	values := url.Values{}
	for _, f := range args {
		f.ApplyListTorrent(values)
	}
	req.URL.RawQuery = values.Encode()
	resp, err := api.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not list torrents: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unauthorized")
	}
	var respBody []Torrent
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return nil, fmt.Errorf("could not read response: %w", err)
	}
	return respBody, nil
}
