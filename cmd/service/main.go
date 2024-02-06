package main

import (
	"context"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/sonalys/animeman/integrations/myanimelist"
	"github.com/sonalys/animeman/integrations/nyaa"
	"github.com/sonalys/animeman/integrations/qbittorrent"
	"github.com/sonalys/animeman/internal/config"
	"github.com/sonalys/animeman/internal/discovery"
	"github.com/sonalys/animeman/internal/utils"
)

func isLaunchedByDebugger() bool {
	// gops executable must be in the path. See https://github.com/google/gops
	gopsOut, err := exec.Command("gops", strconv.Itoa(os.Getppid())).Output()
	return err == nil && strings.Contains(string(gopsOut), "dlv")
}

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out: os.Stderr,
	})
	if !isLaunchedByDebugger() {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}

func main() {
	log.Info().Msg("starting Animeman")
	config := config.ReadConfig(utils.Coalesce(os.Getenv("CONFIG_PATH"), "config.yaml"))
	c := discovery.New(discovery.Dependencies{
		MAL:  myanimelist.New(config.MALUser),
		NYAA: nyaa.New(),
		QB:   qbittorrent.New(config.QBitTorrentHost, config.QBitTorrentUsername, config.QBitTorrentPassword),
		Config: discovery.Config{
			Sources:          config.Sources,
			Qualitites:       config.Qualities,
			Category:         config.Category,
			DownloadPath:     config.DownloadPath,
			CreateShowFolder: config.CreateShowFolder,
			PollFrequency:    config.PollFrequency,
		},
	})
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	if err := c.Start(ctx); err != nil {
		log.Error().Msgf("failed to finish discover: %s", err)
	} else {
		log.Info().Msg("finished successfully")
	}
	done()
}
