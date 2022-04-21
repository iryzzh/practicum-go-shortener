package shortener

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/iryzzh/practicum-go-shortener/internal/app/model"
	"github.com/iryzzh/practicum-go-shortener/internal/app/store"
	"github.com/iryzzh/practicum-go-shortener/internal/pkg/utils"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type Config struct {
	Network     string
	BindAddress string
	BindPort    string
	URLLen      int
}

type Shortener struct {
	cfg   *Config
	store store.Store
	mux   *http.ServeMux
}

func New(cfg *Config, store store.Store) *Shortener {
	return &Shortener{
		cfg:   cfg,
		store: store,
	}
}

func (s *Shortener) Serve() error {
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	ctx := context.Background()

	listener, err := s.getListener()
	if err != nil {
		return err
	}

	if listener == nil {
		return nil
	}

	defer listener.Close()

	mux := http.NewServeMux()

	mux.HandleFunc("/", s.shortHandler)

	srv := &http.Server{
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Println("Starting HTTP server on", listener.Addr())

	serveError := make(chan error, 1)
	go func() {
		select {
		case serveError <- srv.Serve(listener):
		case <-ctx.Done():
		}
	}()

	<-done
	log.Println("Stopping HTTP Server")
	timeout, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := srv.Shutdown(timeout); err == timeout.Err() {
		srv.Close()
	}

	return nil
}

func (s *Shortener) shortHandler(w http.ResponseWriter, r *http.Request) {
	type errorResponse struct {
		Error string `json:"error"`
	}

	enc := json.NewEncoder(w)

	switch r.Method {
	case http.MethodGet:
		resp, err := s.decodeURL(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			enc.Encode(errorResponse{Error: err.Error()})
			return
		}
		http.Redirect(w, r, resp, http.StatusTemporaryRedirect)
	case http.MethodPost:
		resp, err := s.createURL(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			enc.Encode(errorResponse{Error: err.Error()})
			return
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(resp))
	default:
		w.WriteHeader(http.StatusBadRequest)
		enc.Encode(errorResponse{Error: "only GET and POST methods are supported"})
	}
}

func (s *Shortener) createURL(r *http.Request) (string, error) {
	if r.URL.Path != "/" {
		return "", fmt.Errorf("path not found")
	}
	url, err := io.ReadAll(r.Body)
	if err != nil {
		return "", err
	}

	m := &model.URL{
		URLOrigin: string(url),
		URLShort:  utils.RandStringBytesMaskImprSrcUnsafe(s.cfg.URLLen),
	}

	if err := s.store.URL().Create(m); err != nil {
		return "", err
	}

	ret := "http://" + r.Host + "/" + m.URLShort
	return ret, nil
}

func (s *Shortener) decodeURL(r *http.Request) (string, error) {
	split := strings.Split(r.URL.Path, "/")
	id := split[1]
	if id == "" {
		return "", fmt.Errorf("must specify id")
	}
	if len(split) > 2 || len(id) != s.cfg.URLLen {
		return "", fmt.Errorf("incorrect id length")
	}

	url, err := s.store.URL().FindByID(id)
	if err != nil {
		return "", err
	}

	return url.URLOrigin, nil
}

func (s *Shortener) getListener() (net.Listener, error) {
	l, err := net.Listen(s.cfg.Network, s.cfg.BindAddress+":"+s.cfg.BindPort)
	if err != nil {
		return nil, err
	}

	return l, nil
}
