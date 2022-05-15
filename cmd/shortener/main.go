package main

import (
	"context"
	"github.com/caarlos0/env/v6"
	"github.com/iryzzh/practicum-go-shortener/internal/app/handlers"
	"github.com/iryzzh/practicum-go-shortener/internal/app/server"
	"github.com/iryzzh/practicum-go-shortener/internal/app/store/memstore"
	"log"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"
)

func parseConfig() *server.Config {
	cfg := &server.Config{}

	if err := env.Parse(&cfg); err != nil {
		log.Fatal(err)
	}

	return cfg
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg := parseConfig()
	store := memstore.New()
	handler := handlers.New(cfg.URLLen, cfg.BaseURL, store)
	srv := server.New(cfg, handler)

	g, _ := errgroup.WithContext(ctx)
	g.Go(func() error {
		return srv.Serve(ctx)
	})

	if err := g.Wait(); err != nil {
		log.Fatal(err)
	}
}
