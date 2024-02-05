package qbittorrent

import "io"

func (api *API) Version() (string, error) {
	var path = api.host + "/app/version"
	resp, err := api.client.Get(path)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", ErrUnauthorized
	}
	version, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	return string(version), nil
}
