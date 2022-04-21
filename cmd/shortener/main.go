package main

import (
	"context"
	"flag"
	"github.com/iryzzh/practicum-go-shortener/internal/app/shortener"
	"github.com/iryzzh/practicum-go-shortener/internal/app/store/memstore"
	"log"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"
)

func parseConfig() *shortener.Config {
	cfg := &shortener.Config{}

	flag.StringVar(&cfg.Network, "n", "tcp", "The network must be \"tcp\", \"tcp4\", \"tcp6\", \"unix\" or \"unixpacket\"")
	flag.StringVar(&cfg.BindAddress, "a", "localhost", "Address to bind")
	flag.StringVar(&cfg.BindPort, "p", "8080", "Port to bind")
	flag.IntVar(&cfg.URLLen, "l", 8, "Generated link length")

	flag.Parse()

	return cfg
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg := parseConfig()
	store := memstore.New()
	app := shortener.New(cfg, store)

	g, _ := errgroup.WithContext(ctx)
	g.Go(func() error {
		return app.Serve(ctx)
	})

	if err := g.Wait(); err != nil {
		log.Fatal(err)
	}
}
