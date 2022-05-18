package handlers

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/iryzzh/practicum-go-shortener/internal/app/model"
	"github.com/iryzzh/practicum-go-shortener/internal/app/store"
	"github.com/iryzzh/practicum-go-shortener/internal/pkg/utils"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

var (
	ErrIncorrectID  = errors.New("incorrect ID")
	ErrIncorrectURL = errors.New("url is incorrect")
	minURLLength    = 12
)

type Handler struct {
	*chi.Mux

	Store   store.Store
	LinkLen int
	BaseURL string
}

func New(linkLen int, baseURL string, store store.Store) *Handler {
	s := &Handler{
		Mux:     chi.NewMux(),
		LinkLen: linkLen,
		BaseURL: baseURL,
		Store:   store,
	}

	s.Use(middleware.RequestID)
	s.Use(middleware.RealIP)
	s.Use(middleware.Logger)
	s.Use(middleware.Recoverer)
	s.Use(middleware.Compress(5))
	s.Use(gzipMiddleware)

	// timeout
	s.Use(middleware.Timeout(5 * time.Second))

	s.Route("/", func(r chi.Router) {
		r.Route("/{id}", func(r chi.Router) {
			r.Use(s.ParseURL)
			r.Get("/", RedirectHandler)
		})
		r.Post("/", s.PostHandler)
		r.Route("/api", func(r chi.Router) {
			r.Post("/shorten", s.shorten)
		})
	})

	return s
}

func gzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(`Content-Encoding`) == `gzip` {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				log.Fatal(err)
			}
			r.Body = gz
			gz.Close()
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Handler) shorten(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	var url model.URL

	err := json.NewDecoder(r.Body).Decode(&url)
	if err != nil || len(url.URLOrigin) < minURLLength {
		s.fail(w, ErrIncorrectURL)
		return
	}

	url.URLShort = utils.RandStringBytesMaskImprSrcUnsafe(s.LinkLen)
	if err := s.Store.URL().Create(&url); err != nil {
		s.fail(w, err)
		return
	}

	resp := map[string]interface{}{
		"result": s.BaseURL + "/" + url.URLShort,
	}

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Panic(err)
	}
}

func (s *Handler) ParseURL(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		url, err := s.Store.URL().FindByID(id)
		if err != nil {
			s.fail(w, ErrIncorrectID)
			return
		}
		ctx := context.WithValue(r.Context(), "location", url.URLOrigin)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RedirectHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	http.Redirect(w, r, ctx.Value("location").(string), http.StatusTemporaryRedirect)
}

func (s *Handler) PostHandler(w http.ResponseWriter, r *http.Request) {
	p := strings.Replace(r.URL.Path, "/", "", -1)
	if len(p) > 0 {
		s.fail(w, nil)
		return
	}

	b, err := io.ReadAll(r.Body)
	if err != nil {
		s.fail(w, err)
		return
	}

	m := &model.URL{
		URLOrigin: string(b),
		URLShort:  utils.RandStringBytesMaskImprSrcUnsafe(s.LinkLen),
	}

	if err := s.Store.URL().Create(m); err != nil {
		s.fail(w, err)
		return
	}

	resp := s.BaseURL + "/" + m.URLShort

	w.Header().Set("content-type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(resp))
}

func gzipReader(r *http.Request) (io.Reader, error) {
	reader := r.Body

	if r.Header.Get(`Content-Encoding`) == `gzip` {
		gz, err := gzip.NewReader(reader)
		if err != nil {
			return nil, err
		}
		reader = gz
		gz.Close()
	}

	return reader, nil
}

func (s *Handler) fail(w http.ResponseWriter, e error) {
	w.WriteHeader(http.StatusBadRequest)

	type errorResponse struct {
		Error string `json:"error"`
	}

	if e != nil {
		err := json.NewEncoder(w).Encode(errorResponse{Error: e.Error()})
		if err != nil {
			log.Panic(err)
		}
	}
}
