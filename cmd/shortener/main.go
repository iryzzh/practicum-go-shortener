package main

import (
	"context"
	"github.com/caarlos0/env/v6"
	"github.com/iryzzh/practicum-go-shortener/internal/app/handlers"
	"github.com/iryzzh/practicum-go-shortener/internal/app/server"
	"github.com/iryzzh/practicum-go-shortener/internal/app/store"
	"github.com/iryzzh/practicum-go-shortener/internal/app/store/filestore"
	"github.com/iryzzh/practicum-go-shortener/internal/app/store/memstore"
	"log"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"
)

type Config struct {
	Network         string `env:"NETWORK" envDefault:"tcp"`
	BindAddress     string `env:"SERVER_ADDRESS" envDefault:"localhost:8080"`
	BaseURL         string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	URLLen          int    `env:"LINK_LEN" envDefault:"8"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
}

func parseConfig() *Config {
	cfg := &Config{}

	if err := env.Parse(cfg); err != nil {
		log.Fatal(err)
	}

	return cfg
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg := parseConfig()

	var s store.Store

	switch {
	case cfg.FileStoragePath != "":
		file, err := os.OpenFile(cfg.FileStoragePath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0777)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		s = filestore.New(file)
	default:
		s = memstore.New()
	}

	handler := handlers.New(cfg.URLLen, cfg.BaseURL, s)
	srv := server.New(cfg.Network, cfg.BindAddress, handler)

	g, _ := errgroup.WithContext(ctx)
	g.Go(func() error {
		return srv.Serve(ctx)
	})

	if err := g.Wait(); err != nil {
		log.Fatal(err)
	}
}
