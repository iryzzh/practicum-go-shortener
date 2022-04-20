package main

import (
	"flag"
	"github.com/iryzzh/practicum-go-shortener/internal/app/shortener"
	"github.com/iryzzh/practicum-go-shortener/internal/app/store/memstore"
	"log"
)

func parseConfig() *shortener.Config {
	cfg := &shortener.Config{}

	flag.StringVar(&cfg.Network, "n", "tcp", "The network must be \"tcp\", \"tcp4\", \"tcp6\", \"unix\" or \"unixpacket\"")
	flag.StringVar(&cfg.BindAddress, "a", "localhost", "Address to bind")
	flag.StringVar(&cfg.BindPort, "p", "8080", "Port to bind")
	flag.IntVar(&cfg.UrlLen, "l", 8, "Generated link length")

	flag.Parse()

	return cfg
}

func main() {
	cfg := parseConfig()
	store := memstore.New()
	app := shortener.New(cfg, store)

	if err := app.Serve(); err != nil {
		log.Fatal(err)
	}
}
