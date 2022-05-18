package main

import (
	"context"
	"github.com/iryzzh/practicum-go-shortener/cmd/shortener/config"
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

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.New()
	if err != nil {
		log.Fatal(err)
	}

	var s store.Store

	switch {
	case cfg.FileStoragePath != "":
		var file *os.File
		s, file = filestore.New(cfg.FileStoragePath)
		defer file.Close()
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
