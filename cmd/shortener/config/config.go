package config

import (
	"flag"
	"github.com/caarlos0/env/v6"
	"sync"
)

type Config struct {
	Network         string `env:"NETWORK" envDefault:"tcp"`
	BindAddress     string `env:"SERVER_ADDRESS" envDefault:"localhost:8080"`
	BaseURL         string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	URLLen          int    `env:"LINK_LEN" envDefault:"8"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
	SessionKey      string `env:"SESSION_KEY" envDefault:"secret-key"`
}

var once sync.Once

func New() (*Config, error) {
	cfg := &Config{}

	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	once.Do(func() {
		flag.StringVar(&cfg.BindAddress, "a", cfg.BindAddress, "bind address")
		flag.StringVar(&cfg.BaseURL, "b", cfg.BaseURL, "base url")
		flag.StringVar(&cfg.FileStoragePath, "f", cfg.FileStoragePath, "file storage path")
		flag.StringVar(&cfg.DatabaseDSN, "d", cfg.DatabaseDSN, "database dsn")
		flag.StringVar(&cfg.SessionKey, "s", cfg.SessionKey, "session key")

		flag.Parse()
	})

	return cfg, nil
}
