package qbittorrent

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func (api *API) Login(username, password string) error {
	var path = api.host + "/auth/login"
	data := url.Values{
		"username": []string{username},
		"password": []string{password},
	}.Encode()
	req, err := http.NewRequest(http.MethodPost, path, strings.NewReader(data))
	if err != nil {
		return fmt.Errorf("login request creation failed: %w", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := api.Do(req)
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}
	resp.Body.Close()
	return nil
}
