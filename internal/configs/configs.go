package configs

import (
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/sonalys/animeman/internal/utils"
	"gopkg.in/yaml.v3"
)

type AnimeListType string

const (
	AnimeListTypeMAL     AnimeListType = "myanimelist"
	AnimeListTypeAnilist AnimeListType = "anilist"
)

func (t AnimeListType) Validate() error {
	if t != AnimeListTypeAnilist && t != AnimeListTypeMAL {
		return fmt.Errorf("'%s' is invalid. should be [myanimelist,anilist]", t)
	}
	return nil
}

type AnimeListConfig struct {
	Type     AnimeListType `yaml:"type"`
	Username string        `yaml:"username"`
}

func (c AnimeListConfig) Validate() error {
	if err := c.Type.Validate(); err != nil {
		return fmt.Errorf("type: %w", err)
	}
	if c.Username == "" {
		return fmt.Errorf("username: is empty")
	}
	return nil
}

type RSSType string

const (
	RSSTypeNyaa RSSType = "nyaa"
)

func (t RSSType) Validate() error {
	if t != RSSTypeNyaa {
		return fmt.Errorf("'%s' is invalid. should be [nyaa]", t)
	}
	return nil
}

type RSSConfig struct {
	Type             RSSType           `yaml:"type"`
	Sources          []string          `yaml:"sources"`
	Qualities        []string          `yaml:"qualities"`
	PollFrequency    time.Duration     `yaml:"pollFrequency"`
	CustomParameters map[string]string `yaml:"customParameters"`
}

func (c RSSConfig) Validate() error {
	if err := c.Type.Validate(); err != nil {
		return fmt.Errorf("type: %w", err)
	}
	if c.PollFrequency < time.Minute {
		return fmt.Errorf("pollFrequency: should be at least 1 minute")
	}
	return nil
}

type TorrentClientType string

const (
	TorrentClientTypeQBittorrent TorrentClientType = "qbittorrent"
)

func (t TorrentClientType) Validate() error {
	if t != TorrentClientTypeQBittorrent {
		return fmt.Errorf("'%s' is invalid. should be [qbittorrent]", t)
	}
	return nil
}

type TorrentConfig struct {
	Type             TorrentClientType `yaml:"type"`
	Host             string            `yaml:"host"`
	Username         string            `yaml:"username"`
	Password         string            `yaml:"password"`
	Category         string            `yaml:"category"`
	DownloadPath     string            `yaml:"downloadPath"`
	CreateShowFolder bool              `yaml:"createShowFolder"`
	RenameTorrent    *bool             `yaml:"renameTorrent,omitempty"`
}

func (c TorrentConfig) Validate() error {
	if err := c.Type.Validate(); err != nil {
		return fmt.Errorf("type: %w", err)
	}
	if c.Host == "" {
		return fmt.Errorf("host: is empty")
	}
	return nil
}

type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelError LogLevel = "error"
)

type Config struct {
	AnimeListConfig `yaml:"animeList"`
	RSSConfig       `yaml:"rssConfig"`
	TorrentConfig   `yaml:"torrentConfig"`
	LogLevel        LogLevel `yaml:"logLevel"`
}

func (l LogLevel) Convert() zerolog.Level {
	switch l {
	case LogLevelDebug:
		return zerolog.DebugLevel
	case LogLevelError:
		return zerolog.ErrorLevel
	default:
		return zerolog.InfoLevel
	}
}

func (c Config) Validate() error {
	if err := c.AnimeListConfig.Validate(); err != nil {
		return fmt.Errorf("animeList.%w", err)
	}
	if err := c.RSSConfig.Validate(); err != nil {
		return fmt.Errorf("rssConfig.%w", err)
	}
	if err := c.TorrentConfig.Validate(); err != nil {
		return fmt.Errorf("torrentConfig.%w", err)
	}
	return nil
}

func GenerateBoilerplateConfig() {
	file, err := os.Create("config.yaml")
	if err != nil {
		log.Fatal().Msgf("failed to open a new config.yaml file: %s", err)
	}
	err = yaml.NewEncoder(file).Encode(Config{
		AnimeListConfig: AnimeListConfig{
			Type:     AnimeListTypeMAL,
			Username: "YOUR_USERNAME",
		},
		RSSConfig: RSSConfig{
			Sources:       []string{},
			Qualities:     []string{"1080 HEVC", "720"},
			PollFrequency: 15 * time.Minute,
		},
		TorrentConfig: TorrentConfig{
			Category:         "Animes",
			DownloadPath:     "/downloads/animes",
			Host:             "http://192.168.1.240:8088",
			Username:         "admin",
			Password:         "adminadmin",
			CreateShowFolder: true,
			RenameTorrent:    utils.Pointer(true),
			Type:             TorrentClientTypeQBittorrent,
		},
	})
	if err != nil {
		log.Fatal().Msgf("failed to save config.yaml file: %s", err)
	}
}

func ReadConfig(path string) (Config, error) {
	file, err := os.Open(path)
	if err != nil {
		GenerateBoilerplateConfig()
		log.Fatal().Msg("file config.yaml not detected, please open the created file and configure it correctly")
	}
	var config Config
	if err = yaml.NewDecoder(file).Decode(&config); err != nil {
		log.Fatal().Msgf("could not read config.yaml: %s", err)
	}
	return config, config.Validate()
}
