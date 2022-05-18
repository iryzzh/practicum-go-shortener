package config

import (
	"flag"
	"github.com/caarlos0/env/v6"
)

type Config struct {
	Network         string `env:"NETWORK" envDefault:"tcp"`
	BindAddress     string `env:"SERVER_ADDRESS" envDefault:"localhost:8080"`
	BaseURL         string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	URLLen          int    `env:"LINK_LEN" envDefault:"8"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
}

func New() (*Config, error) {
	cfg := &Config{}
	err := env.Parse(cfg)

	regStrVar(&cfg.BindAddress, "a", cfg.BindAddress, "bind address")
	regStrVar(&cfg.BaseURL, "b", cfg.BaseURL, "base url")
	regStrVar(&cfg.FileStoragePath, "f", cfg.FileStoragePath, "file storage path")

	flag.Parse()

	return cfg, err
}

func regStrVar(s *string, name string, value string, usage string) {
	if flag.Lookup(name) == nil {
		flag.StringVar(s, name, value, usage)
	}
}
