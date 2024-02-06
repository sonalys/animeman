package config

import (
	"os"
	"time"

	"github.com/rs/zerolog/log"
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

func GenerateBoilerplateConfig() {
	file, err := os.Create("config.yaml")
	if err != nil {
		log.Fatal().Msgf("failed to open a new config.yaml file: %s", err)
	}
	err = yaml.NewEncoder(file).Encode(Config{
		Sources:             []string{},
		Qualities:           []string{"1080 HEVC", "720"},
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
		log.Fatal().Msgf("failed to save config.yaml file: %s", err)
	}
}

func ReadConfig(path string) Config {
	file, err := os.ReadFile(path)
	if err != nil {
		GenerateBoilerplateConfig()
		log.Fatal().Msg("file config.yaml not detected, please open the created file and configure it correctly")
	}
	var config Config
	err = yaml.Unmarshal(file, &config)
	if err != nil {
		log.Fatal().Msgf("could not read config.yaml: %s", err)
	}
	return config
}
