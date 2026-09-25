package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/sonalys/animeman/cmd/service/controller"
	"github.com/sonalys/animeman/internal/pkg/coalesce"
	"github.com/sonalys/animeman/internal/ports/animelist"
	"github.com/sonalys/animeman/internal/ports/torrentsource"
)

func newHealthcheck(deps controller.Dependencies) *http.Server {
	healthMux := http.NewServeMux()
	healthMux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		var failed []string
		if _, err := deps.TorrentClient.List(r.Context(), nil); err != nil {
			failed = append(failed, "torrentClient")
		}
		if _, err := deps.TorrentSource.Search(
			r.Context(),
			animelist.Entry{},
			torrentsource.SearchOptions{},
		); err != nil {
			failed = append(failed, "torrentSource")
		}

		if len(failed) > 0 {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintf(w, "unhealthy: %s", strings.Join(failed, ", "))
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	healthMux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		if _, err := deps.AnimeListSource.GetCurrentlyWatching(r.Context()); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	healthServer := &http.Server{
		Addr:    coalesce.OrDefault(os.Getenv("HEALTH_ADDR"), ":8080"),
		Handler: healthMux,
	}

	return healthServer
}
