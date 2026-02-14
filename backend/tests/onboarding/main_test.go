package main

import (
	"testing"
	"time"

	"github.com/sonalys/animeman/internal/domain"
	"github.com/sonalys/animeman/internal/domain/stream"
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

	collection := user.NewCollection(
		"My Collection",
		"/volume1/media/anime",
		[]string{},
		true,
	)
	require.NotNil(t, collection)

	qualityProfile := domain.NewQualityProfile(
		"fullhd only",
		stream.Resolution1080p,
		stream.Resolution1080p,
		[]stream.VideoCodec{},
		[]string{},
	)

	media := collection.NewMedia(
		[]domain.Title{
			domain.NewTitle(domain.TitleTypeNative, "en-us", "Media title"),
		},
		domain.MonitoringStatusAll,
		domain.NewMediaMetadata([]string{}, time.Time{}, time.Time{}),
		qualityProfile.ID,
	)
	require.NotNil(t, media)

	season := media.NewSeason(
		1,
		domain.AiringStatusAiring,
		domain.SeasonMetadata{},
	)
	require.NotNil(t, season)

	episode := season.NewEpisode(
		domain.MediaTypeTV,
		"1",
		[]domain.Title{},
		new(time.Now()),
	)
	require.NotNil(t, episode)

	watchlist := user.NewExternalWatchList(
		domain.WatchlistSourceAniList,
		"username",
		time.Hour,
	)
	require.NotNil(t, watchlist)

	entry := watchlist.NewEntry(
		media.ID,
		season.ID,
		domain.WatchlistStatusWatching,
	)

	entry.SetLastWatchedEpisode(episode.ID)
}
