package qbittorrent

import (
	"context"
	"errors"
	"net/http"
	"net/http/cookiejar"
	"time"

	"github.com/sonalys/animeman/integrations/qbittorrent"
	"github.com/sonalys/animeman/internal/utils"
	"github.com/sonalys/animeman/pkg/v1/torrentclient"
)

type (
	Wrapper struct {
		client *qbittorrent.API
	}
)

func New(ctx context.Context, host, username, password string) *Wrapper {
	client := &http.Client{
		Timeout: 3 * time.Second,
		Jar:     utils.Must(cookiejar.New(nil)),
	}
	return &Wrapper{
		client: qbittorrent.New(ctx, client, host, username, password),
	}
}

func convertTorrent(in []qbittorrent.Torrent) []torrentclient.Torrent {
	out := make([]torrentclient.Torrent, 0, len(in))
	for i := range in {
		out = append(out, torrentclient.Torrent{
			Name:     in[i].Name,
			Category: in[i].Category,
			Hash:     in[i].Hash,
			Tags:     in[i].GetTags(),
		})
	}
	return out
}

func convertError(err error) error {
	switch {
	case errors.Is(err, qbittorrent.ErrUnauthorized):
		return torrentclient.ErrUnauthorized
	default:
		return err
	}
}

func (w *Wrapper) List(ctx context.Context, args ...torrentclient.ArgListTorrent) ([]torrentclient.Torrent, error) {
	resp, err := w.client.List(ctx, utils.ConvertInterfaceList[torrentclient.ArgListTorrent, qbittorrent.ArgListTorrent](args)...)
	if err != nil {
		return nil, convertError(err)
	}
	return convertTorrent(resp), nil
}

func (w *Wrapper) AddTorrent(ctx context.Context, args ...torrentclient.ArgAddTorrent) error {
	return convertError(w.client.AddTorrent(ctx, utils.ConvertInterfaceList[torrentclient.ArgAddTorrent, qbittorrent.ArgAddTorrent](args)...))
}

func (w *Wrapper) AddTorrentTags(ctx context.Context, ids []string, args ...torrentclient.AddTorrentTagsArg) error {
	return convertError(w.client.AddTorrentTags(ctx, ids, utils.ConvertInterfaceList[torrentclient.AddTorrentTagsArg, qbittorrent.AddTorrentTagsArg](args)...))
}
