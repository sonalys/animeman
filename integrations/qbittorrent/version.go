package qbittorrent

import (
	"io"
	"net/http"

	"github.com/sonalys/animeman/internal/utils"
)

func (api *API) Version() (string, error) {
	var path = api.host + "/app/version"
	resp, err := api.Do(utils.Must(http.NewRequest(http.MethodGet, path, nil)))
	if err != nil {
		return "", NewErrConnection(err)
	}
	defer resp.Body.Close()
	version, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	return string(version), nil
}
