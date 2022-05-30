package handlers

import (
	"compress/gzip"
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/iryzzh/practicum-go-shortener/internal/app/model"
	"github.com/iryzzh/practicum-go-shortener/internal/app/store"
	"github.com/iryzzh/practicum-go-shortener/internal/pkg/utils"
	"github.com/json-iterator/go"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

var (
	ErrIncorrectID  = errors.New("incorrect ID")
	ErrIncorrectURL = errors.New("url is incorrect")
	minURLLength    = 12

	json = jsoniter.ConfigCompatibleWithStandardLibrary
)

type Handler struct {
	*chi.Mux

	Store         store.Store
	LinkLen       int
	BaseURL       string
	sessionsStore *sessions.CookieStore
	cookieName    string
}

func New(linkLen int, baseURL string, store store.Store) *Handler {
	s := &Handler{
		Mux:           chi.NewMux(),
		LinkLen:       linkLen,
		BaseURL:       baseURL,
		Store:         store,
		sessionsStore: sessions.NewCookieStore(utils.GenerateRandomKey(32)),
		cookieName:    "_session_",
	}

	s.Use(middleware.RequestID)
	s.Use(middleware.RealIP)
	s.Use(middleware.Logger)
	s.Use(middleware.Recoverer)
	s.Use(middleware.Compress(5))
	s.Use(gzipMiddleware)

	// timeout
	s.Use(middleware.Timeout(5 * time.Second))

	s.Use(s.SessionsHandler)

	s.Route("/", func(r chi.Router) {
		r.Route("/{id}", func(r chi.Router) {
			r.With(s.ParseURL).Get("/", RedirectHandler)
		})
		r.Post("/", s.PostHandler)
	})

	s.Route("/api", func(r chi.Router) {
		r.Post("/shorten", s.shorten)
		r.Post("/shorten/batch", s.batch)
		r.Get("/user/urls", s.userUrls)
	})

	s.Get("/ping", s.Status)

	return s
}

func (s *Handler) batch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		s.fail(w, err)
		return
	}

	type response struct {
		CorrelationID string `json:"correlation_id"`
		OriginalURL   string `json:"original_url,omitempty"`
		ShortURL      string `json:"short_url,omitempty"`
	}

	var r1 []response
	var r2 []response

	err = json.Unmarshal(b, &r1)
	if err != nil {
		s.fail(w, err)
		return
	}

	for _, v := range r1 {
		shortURL := utils.RandStringBytesMaskImprSrcUnsafe(s.LinkLen)

		if err := s.Store.URL().Create(&model.URL{
			URLOrigin: v.OriginalURL,
			URLShort:  shortURL,
		}); err != nil {
			s.fail(w, err)
			return
		}
		r2 = append(r2, response{
			CorrelationID: v.CorrelationID,
			ShortURL:      s.BaseURL + "/" + shortURL,
		})
	}

	if len(r2) > 0 {
		w.Header().Set("content-type", "application/json")

		w.WriteHeader(http.StatusCreated)

		if err := json.NewEncoder(w).Encode(r2); err != nil {
			s.fail(w, err)
			return
		}
	}

	w.WriteHeader(http.StatusInternalServerError)
}

func (s *Handler) Status(w http.ResponseWriter, r *http.Request) {
	if err := s.Store.Ping(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Handler) userUrls(w http.ResponseWriter, r *http.Request) {
	session, _ := s.sessionsStore.Get(r, s.cookieName)

	if session.Values["uuid"] != nil {
		user, err := s.Store.User().FindByUUID(session.Values["uuid"].(string))
		if err != nil {
			if err == store.ErrUserNotFound {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			s.fail(w, err)
			return
		}

		type response struct {
			ShortURL    string `json:"short_url"`
			OriginalURL string `json:"original_url"`
		}

		var resp []response

		urls, err := s.Store.URL().FindByUserID(user.ID)
		if err != nil {
			s.fail(w, err)
			return
		}

		for _, v := range urls {
			resp = append(resp, response{
				ShortURL:    s.BaseURL + "/" + v.URLShort,
				OriginalURL: v.URLOrigin,
			})
		}

		if len(resp) > 0 {
			w.Header().Set("content-type", "application/json")
			if err := json.NewEncoder(w).Encode(resp); err != nil {
				s.fail(w, err)
				return
			}
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Handler) SessionsHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := s.sessionsStore.Get(r, s.cookieName)
		if session.Values["uuid"] == nil {
			user := &model.User{}
			user.UUID = uuid.New().String()
			if err := s.Store.User().Create(user); err == nil {
				session.Values["uuid"] = user.UUID
				if err := session.Save(r, w); err != nil {
					log.Println("session save error:", err)
				}
			}
		}
		next.ServeHTTP(w, r)
	})
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
	url := &model.URL{}

	err := json.NewDecoder(r.Body).Decode(&url)
	if err != nil || len(url.URLOrigin) < minURLLength {
		s.fail(w, ErrIncorrectURL)
		return
	}

	url.URLShort = utils.RandStringBytesMaskImprSrcUnsafe(s.LinkLen)
	if err := s.Store.URL().Create(url); err != nil {
		s.fail(w, err)
		return
	}

	if err := s.SaveURL(r, url); err != nil {
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

func (s *Handler) SaveURL(r *http.Request, url *model.URL) error {
	session, _ := s.sessionsStore.Get(r, s.cookieName)
	if session.Values["uuid"] != nil {
		user, err := s.Store.User().FindByUUID(session.Values["uuid"].(string))
		if err != nil {
			return err
		}
		if err := s.Store.URL().UpdateUserID(url, user.ID); err != nil {
			return err
		}
	}

	return nil
}

func (s *Handler) ParseURL(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		url, err := s.Store.URL().FindByUUID(id)
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
	if location, ok := ctx.Value("location").(string); ok {
		http.Redirect(w, r, location, http.StatusTemporaryRedirect)
		return
	}
	w.WriteHeader(http.StatusInternalServerError)
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

	url := &model.URL{
		URLOrigin: string(b),
		URLShort:  utils.RandStringBytesMaskImprSrcUnsafe(s.LinkLen),
	}

	if err := s.Store.URL().Create(url); err != nil {
		s.fail(w, err)
		return
	}

	if err := s.SaveURL(r, url); err != nil {
		s.fail(w, err)
		return
	}

	resp := s.BaseURL + "/" + url.URLShort

	w.Header().Set("content-type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(resp))
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
