package qbittorrent

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
)

type (
	API struct {
		host               string
		username, password string
		client             *http.Client
	}
)

func New(ctx context.Context, host, username, password string) *API {
	jar, err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}
	api := &API{
		host:     fmt.Sprintf("%s/api/v2", host),
		username: username,
		password: password,
		client: &http.Client{
			Timeout: 3 * time.Second,
			Jar:     jar,
		},
	}
	var version string
	api.Wait(ctx)
	if version, err = api.Version(); err != nil {
		log.Fatal().Msgf("failed to connect to qBittorrent: %s", err)
	}
	log.Info().Msgf("connected to qBittorrent:%s", version)
	return api
}

func (api *API) Do(req *http.Request) (*http.Response, error) {
	ctx := context.WithoutCancel(req.Context())
	localReq := req.Clone(ctx)
	resp, err := api.client.Do(localReq)
	switch {
	case errors.Is(err, syscall.ECONNREFUSED) ||
		errors.Is(err, syscall.ECONNABORTED) ||
		errors.Is(err, syscall.ECONNRESET):
		log.Warn().Msgf("qBittorrent disconnected")
		api.Wait(ctx)
		return api.Do(req)
	case err == nil && resp.StatusCode >= 400:
		if loginErr := api.Login(api.username, api.password); loginErr != nil {
			return resp, loginErr
		}
		return api.Do(req)
	}
	return resp, err
}
