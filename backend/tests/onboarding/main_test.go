package main

import (
	"testing"

	"github.com/sonalys/animeman/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Onboarding(t *testing.T) {
	user, err := domain.NewUser("username", []byte("password"))
	require.NoError(t, err)
	assert.NotNil(t, user)

	prowlarrConfig := user.NewProwlarrConfiguration("http://192.168.1.219:9696", "api_key")
	require.NotNil(t, prowlarrConfig)

	torrentConfig := user.NewTorrentClientConfiguration(
		domain.TorrentSourceQBitTorrent, "http://192.168.1.219:8088", "username", nil)
	require.NotNil(t, torrentConfig)

	// animeList := user.NewAnimeList("username", domain.AnimeListSourceAnilist)
	// require.NotNil(t, animeList)
}
