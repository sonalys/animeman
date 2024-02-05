package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/sonalys/animeman/controller"
	"github.com/sonalys/animeman/integrations/myanimelist"
	"github.com/sonalys/animeman/integrations/nyaa"
	"github.com/sonalys/animeman/integrations/qbittorrent"
	"github.com/sonalys/animeman/internal/utils"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Sources             []string      `yaml:"sources"`
	Qualities           []string      `yaml:"qualities"`
	Category            string        `yaml:"category"`
	DownloadPath        string        `yaml:"downloadPath"`
	CreateShowFolder    bool          `yaml:"createShowFolder"`
	MALUser             string        `yaml:"malUser"`
	QBitTorrentHost     string        `yaml:"qBitTorrentHost"`
	QBitTorrentUsername string        `yaml:"qBitTorrentUsername"`
	QBitTorrentPassword string        `yaml:"qBitTorrentPassword"`
	PollFrequency       time.Duration `yaml:"pollFrequency"`
}

func createMockConfig() {
	file, err := os.Create("config.yaml")
	if err != nil {
		panic(fmt.Errorf("failed to open a new config.yaml file: %w", err))
	}
	err = yaml.NewEncoder(file).Encode(Config{
		Sources:             []string{},
		Qualities:           []string{"1080", "720"},
		Category:            "Animes",
		DownloadPath:        "/downloads/animes",
		MALUser:             "raicon",
		QBitTorrentHost:     "http://192.168.1.240:8088",
		QBitTorrentUsername: "admin",
		QBitTorrentPassword: "adminadmin",
		CreateShowFolder:    true,
		PollFrequency:       15 * time.Minute,
	})
	if err != nil {
		panic(fmt.Errorf("failed to save config.yaml file: %w", err))
	}
}

func readConfig() Config {
	file, err := os.ReadFile(utils.Coalesce(os.Getenv("CONFIG_PATH"), "config.yaml"))
	if err != nil {
		createMockConfig()
		panic("file config.yaml not detected, please open the created file and configure it correctly")
	}
	var config Config
	err = yaml.Unmarshal(file, &config)
	if err != nil {
		panic(fmt.Errorf("could not read config.yaml: %w", err))
	}
	return config
}

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out: os.Stderr,
	})
}

func main() {
	log.Info().Msg("starting Animeman")
	config := readConfig()
	malAPI := myanimelist.New(config.MALUser)
	nyaaAPI := nyaa.New()
	qb := qbittorrent.New(config.QBitTorrentHost, config.QBitTorrentUsername, config.QBitTorrentPassword)
	c := controller.New(controller.Dependencies{
		MAL:  malAPI,
		NYAA: nyaaAPI,
		QB:   qb,
		Config: controller.Config{
			Sources:          config.Sources,
			Qualitites:       config.Qualities,
			Category:         config.Category,
			DownloadPath:     config.DownloadPath,
			CreateShowFolder: config.CreateShowFolder,
			PollFrequency:    config.PollFrequency,
		},
	})
	sig := make(chan os.Signal, 1)
	ctx, cancel := context.WithCancel(context.Background())
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		log.Info().Msg("context cancelled")
		cancel()
	}()
	c.Start(ctx)
}
