package main

import (
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

type AnimeListType string

const (
	AnimeListTypeMAL     AnimeListType = "myanimelist"
	AnimeListTypeAnilist AnimeListType = "anilist"

	DefaultRenamingScript = `
		// Joins parts with a single space.
		join(
			// Filters only non-empty parts.
			filter(
				[
					// Put release group inside square brackets if any.
					format("[%s]", releaseGroup),
					title,
					// Format tag as a string. e.g. S1E4.
					tag.String(), 
					// Put vertical resolution inside square brackets + p, if any.
					format("[%s]", verticalResolution),
				], 
				# != "",
			), 
			" ",
		)`
)

// Validate is broadened to accept short aliases for each adapter type.
func (t AnimeListType) Validate() (AnimeListType, error) {
	switch t {
	case "", AnimeListTypeMAL, "mal":
		return AnimeListTypeMAL, nil
	case AnimeListTypeAnilist:
		return AnimeListTypeAnilist, nil
	default:
		return "", fmt.Errorf("'%s' is invalid. should be [myanimelist,anilist]", t)
	}
}

type AnimeListConfig struct {
	Type     AnimeListType `yaml:"type"`
	Username string        `yaml:"username"`
	CacheTTL time.Duration `yaml:"cacheTTL"`
}

func (c *AnimeListConfig) Validate() error {
	typ, err := c.Type.Validate()
	if err != nil {
		return fmt.Errorf("type: %w", err)
	}
	c.Type = typ
	if c.Username == "" {
		return fmt.Errorf("username: is empty")
	}
	if c.CacheTTL == 0 {
		c.CacheTTL = 30 * time.Minute
	}
	if c.CacheTTL < 30*time.Minute {
		return fmt.Errorf("cacheTTL: must be at least 30 minutes")
	}
	return nil
}

type TorrentSourceType string

const (
	TorrentSourceTypeNyaa   TorrentSourceType = "nyaa"
	TorrentSourceTypeNekoBT TorrentSourceType = "nekobt"
)

func (t TorrentSourceType) Validate() (TorrentSourceType, error) {
	switch t {
	case "", TorrentSourceTypeNyaa:
		return TorrentSourceTypeNyaa, nil
	case TorrentSourceTypeNekoBT:
		return TorrentSourceTypeNekoBT, nil
	default:
		return "", fmt.Errorf("'%s' is invalid. should be [nyaa,nekobt]", t)
	}
}

// NyaaConfig holds nyaa-specific settings.
type NyaaConfig struct {
	CustomParameters map[string]string `yaml:"customParameters"`
}

// NekobtConfig holds nekoBT-specific settings.
type NekobtConfig struct {
	APIKey string `yaml:"apiKey"`
	// CustomParameters sets extra torznab query parameters, e.g.
	// `sort: seeders`, `sub_lang: en`, `mtl: "false"`, `hardsub: "false"`.
	// They override the defaults built by the adapter.
	CustomParameters map[string]string `yaml:"customParameters"`
}

type TorrentSourceConfig struct {
	Type   TorrentSourceType `yaml:"type"`
	Nyaa   *NyaaConfig       `yaml:"nyaa,omitempty"`
	Nekobt *NekobtConfig     `yaml:"nekobt,omitempty"`
}

func (c *TorrentSourceConfig) Validate() error {
	typ, err := c.Type.Validate()
	if err != nil {
		return fmt.Errorf("type: %w", err)
	}
	c.Type = typ
	switch c.Type {
	case TorrentSourceTypeNyaa:
		if c.Nyaa == nil {
			c.Nyaa = new(NyaaConfig)
		}
	case TorrentSourceTypeNekoBT:
		if c.Nekobt == nil {
			c.Nekobt = new(NekobtConfig)
		}
		if c.Nekobt.APIKey == "" {
			return fmt.Errorf("nekobt.apiKey: is empty")
		}
	}
	return nil
}

type TorrentClientType string

const (
	TorrentClientTypeQBittorrent TorrentClientType = "qbittorrent"
)

func (t TorrentClientType) Validate() (TorrentClientType, error) {
	switch t {
	case "", TorrentClientTypeQBittorrent, "qbit":
		return TorrentClientTypeQBittorrent, nil
	default:
		return "", fmt.Errorf("'%s' is invalid. should be [qbittorrent]", t)
	}
}

// QBittorrentConfig holds qbittorrent-specific settings.
type QBittorrentConfig struct {
	Host     string `yaml:"host"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type TorrentConfig struct {
	Type        TorrentClientType  `yaml:"type"`
	QBittorrent *QBittorrentConfig `yaml:"qbittorrent,omitempty"`
}

func (c *TorrentConfig) Validate() error {
	typ, err := c.Type.Validate()
	if err != nil {
		return fmt.Errorf("type: %w", err)
	}
	c.Type = typ
	if c.Type == TorrentClientTypeQBittorrent {
		if c.QBittorrent == nil {
			c.QBittorrent = new(QBittorrentConfig)
		}
		if c.QBittorrent.Host == "" {
			return fmt.Errorf("qbittorrent.host: is empty")
		}
	}
	return nil
}

// DiscoveryConfig holds behavior settings shared by all adapters.
type DiscoveryConfig struct {
	SearchSuffix     string        `yaml:"searchSuffix"`
	Sources          []string      `yaml:"sources"`
	Qualities        []string      `yaml:"qualities"`
	PollFrequency    time.Duration `yaml:"pollFrequency"`
	Category         string        `yaml:"category"`
	DownloadPath     string        `yaml:"downloadPath"`
	CreateShowFolder bool          `yaml:"createShowFolder"`
	RenameTorrent    *bool         `yaml:"renameTorrent,omitempty"`
	RenameScript     string        `yaml:"renameScript,omitempty"`
}

func (c *DiscoveryConfig) Validate() error {
	if c.PollFrequency == 0 {
		c.PollFrequency = 15 * time.Minute
	}
	if c.PollFrequency < time.Minute {
		return fmt.Errorf("pollFrequency: should be at least 1 minute")
	}
	if c.RenameScript == "" {
		// Default renaming logic to keep backwards compatibility.
		// Example output: [ReleaseGroup] Show Title S01E02 [1080p]
		c.RenameScript = DefaultRenamingScript
	}
	return nil
}

// ShokoConfig holds shoko server settings.
type ShokoConfig struct {
	Host   string `yaml:"host"`
	APIKey string `yaml:"apiKey"`
}

func (c *ShokoConfig) Validate() error {
	if c.Host == "" {
		return fmt.Errorf("host: is empty")
	}
	if c.APIKey == "" {
		return fmt.Errorf("apiKey: is empty")
	}
	return nil
}

type LogLevel string

const (
	LogLevelTrace LogLevel = "trace"
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelError LogLevel = "error"
)

type Config struct {
	AnimeListConfig     `         yaml:"animeList"`
	TorrentSourceConfig `         yaml:"torrentSource"`
	TorrentConfig       `         yaml:"torrentClient"`
	DiscoveryConfig     `         yaml:"discovery"`
	ShokoConfig         `         yaml:"shoko,omitempty"`
	LogLevel            LogLevel `yaml:"logLevel"`
}

func (l LogLevel) Convert() zerolog.Level {
	switch l {
	case LogLevelTrace:
		return zerolog.TraceLevel
	case LogLevelDebug:
		return zerolog.DebugLevel
	case LogLevelError:
		return zerolog.ErrorLevel
	default:
		return zerolog.InfoLevel
	}
}

func (c *Config) Validate() error {
	if err := c.AnimeListConfig.Validate(); err != nil {
		return fmt.Errorf("animeList.%w", err)
	}
	if err := c.TorrentSourceConfig.Validate(); err != nil {
		return fmt.Errorf("torrentSource.%w", err)
	}
	if err := c.TorrentConfig.Validate(); err != nil {
		return fmt.Errorf("torrentClient.%w", err)
	}
	if err := c.DiscoveryConfig.Validate(); err != nil {
		return fmt.Errorf("discovery.%w", err)
	}
	if c.ShokoConfig.Host != "" {
		if err := c.ShokoConfig.Validate(); err != nil {
			return fmt.Errorf("shoko.%w", err)
		}
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
			CacheTTL: 30 * time.Minute,
		},
		TorrentSourceConfig: TorrentSourceConfig{
			Type: TorrentSourceTypeNyaa,
			Nyaa: &NyaaConfig{
				CustomParameters: map[string]string{
					"c": "1_2", // Anime - English-translated.
				},
			},
			Nekobt: &NekobtConfig{
				APIKey: "YOUR_API_KEY",
				CustomParameters: map[string]string{
					"sub_lang":   "en",    // Only English subtitles.
					"audio_lang": "ja",    // Only Japanese audio.
					"mtl":        "false", // No machine-translated subtitles.
					"hardsub":    "false", // No hardsubbed releases.
				},
			},
		},
		TorrentConfig: TorrentConfig{
			Type: TorrentClientTypeQBittorrent,
			QBittorrent: &QBittorrentConfig{
				Host:     "http://192.168.1.240:8088",
				Username: "admin",
				Password: "adminadmin",
			},
		},
		SearchSuffix:     `-"dub"`,
		Sources:          []string{},
		Qualities:        []string{"1080 HEVC", "720"},
		PollFrequency:    15 * time.Minute,
		Category:         "Animes",
		DownloadPath:     "/downloads/animes",
		CreateShowFolder: true,
		RenameTorrent:    new(true),
		Host:             "http://192.168.1.240:8111",
		APIKey:           "YOUR_API_KEY",
		LogLevel:         LogLevelInfo,
	})
	if err != nil {
		log.Fatal().Msgf("failed to save config.yaml file: %s", err)
	}
}

func ReadConfig(path string) (Config, error) {
	file, err := os.Open(path)
	if err != nil {
		GenerateBoilerplateConfig()
		log.Fatal().
			Msg("file config.yaml not detected, please open the created file and configure it correctly")
	}
	var config Config
	if err = yaml.NewDecoder(file).Decode(&config); err != nil {
		log.Fatal().Msgf("could not read config.yaml: %s", err)
	}
	return config, config.Validate()
}
