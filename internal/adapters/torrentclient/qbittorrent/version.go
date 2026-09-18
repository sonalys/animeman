package qbittorrent

import (
	"context"
	"io"
	"net/http"

	"github.com/sonalys/animeman/internal/utils"
)

func (api *API) Version(ctx context.Context) (string, error) {
	var path = api.host + "/app/version"
	resp, err := api.Do(ctx, utils.Must(http.NewRequest(http.MethodGet, path, nil)))
	if err != nil {
		return "", NewErrConnection(err)
	}
	defer resp.Body.Close()
	return string(utils.Must(io.ReadAll(resp.Body))), nil
}
