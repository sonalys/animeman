package qbittorrent

import (
	"fmt"
	"strings"

	"github.com/sonalys/animeman/internal/utils/sliceutils"
)

type (
	Torrent struct {
		Name     string `json:"name"`
		Category string `json:"category"`
		Hash     string `json:"hash"`
		Tags     string `json:"tags"`
	}
)

func NewErrConnection(err error) error {
	return fmt.Errorf("connection error: %w", err)
}

func (t Torrent) GetTags() []string {
	return sliceutils.Map[string, string](strings.Split(t.Tags, ","), func(s string) string { return strings.TrimSpace(s) })
}
