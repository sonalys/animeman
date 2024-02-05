package qbittorrent

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"time"

	"github.com/rs/zerolog/log"
)

type (
	API struct {
		host   string
		client *http.Client
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
	var version string
	if version, err = api.Version(); err != nil {
		if !errors.Is(err, ErrUnauthorized) {
			log.Fatal().Msgf("failed to initialize: %s", err)
		}
		if err := api.Login(username, password); err != nil {
			log.Fatal().Msgf("could not initialize qBittorrent: %s", err)
		}
		version, err = api.Version()
		if err == ErrUnauthorized {
			panic(err)
		}
	}
	log.Info().Msgf("connected to qBitTorrent:%s", version)
	return api
}

func (api *API) Do(req *http.Request) (*http.Response, error) {
	resp, err := api.client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == 400 {
		return nil, ErrUnauthorized
	}
	return resp, nil
}
