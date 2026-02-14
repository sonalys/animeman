package main

import (
	"net/url"
	"testing"
	"time"

	"github.com/sonalys/animeman/internal/domain/authentication"
	"github.com/sonalys/animeman/internal/domain/collections"
	"github.com/sonalys/animeman/internal/domain/indexing"
	"github.com/sonalys/animeman/internal/domain/stream"
	"github.com/sonalys/animeman/internal/domain/transfer"
	"github.com/sonalys/animeman/internal/domain/users"
	"github.com/sonalys/animeman/internal/domain/watchlists"
	"github.com/sonalys/animeman/internal/utils/errutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Onboarding(t *testing.T) {
	user, err := users.NewUser("username", []byte("password"))
	require.NoError(t, err)
	assert.NotNil(t, user)

	prowlarrConfig := user.NewIndexerClient(
		indexing.IndexerTypeProwlarr,
		*errutils.Must(url.Parse("http://192.168.1.219:9696")),
		authentication.NewAPIKeyAuthentication("apiKey"),
	)
	require.NotNil(t, prowlarrConfig)

	torrentConfig := user.NewTransferClient(
		transfer.ClientTypeQBittorrent,
		*errutils.Must(url.Parse("http://192.168.1.219:8088")),
		authentication.NewUserPasswordAuthentication("username", []byte{}),
	)
	require.NotNil(t, torrentConfig)

	collection := user.NewCollection(
		"My Collection",
		"/volume1/media/anime",
		[]string{},
		true,
	)
	require.NotNil(t, collection)

	qualityProfile := collections.NewQualityProfile(
		"FullHD only",
		stream.Resolution1080p,
		stream.Resolution1080p,
		[]stream.VideoCodec{},
		[]string{},
	)

	media := collection.NewMedia(
		[]collections.Title{
			collections.NewTitle(collections.TitleTypeNative, "en-us", "Media title"),
		},
		collections.MonitoringStatusAll,
		collections.NewMediaMetadata([]string{}, time.Time{}, time.Time{}),
		qualityProfile.ID,
	)
	require.NotNil(t, media)

	season := media.NewSeason(
		1,
		collections.AiringStatusAiring,
		collections.SeasonMetadata{},
	)
	require.NotNil(t, season)

	episode := season.NewEpisode(
		collections.MediaTypeTV,
		"1",
		[]collections.Title{},
		new(time.Now()),
	)
	require.NotNil(t, episode)

	watchlist := user.NewExternalWatchList(
		watchlists.WatchlistSourceAniList,
		"username",
		time.Hour,
	)
	require.NotNil(t, watchlist)

	entry := watchlist.NewEntry(
		media.ID,
		season.ID,
		watchlists.WatchlistStatusWatching,
	)

	entry.SetLastWatchedEpisode(episode.ID)
}
