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

func New(linkLen int, baseURL string, store store.Store, sessionKey []byte) *Handler {
	s := &Handler{
		Mux:           chi.NewMux(),
		LinkLen:       linkLen,
		BaseURL:       baseURL,
		Store:         store,
		sessionsStore: sessions.NewCookieStore(sessionKey),
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
		r.Delete("/user/urls", s.DeleteUrlsHandler)
	})

	s.Get("/ping", s.Status)

	return s
}

func (s *Handler) DeleteUrlsHandler(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.fail(w, err)
		return
	}
	defer r.Body.Close()

	var values []string
	err = json.Unmarshal(b, &values)
	if err != nil {
		s.fail(w, err)
		return
	}

	go s.deleteUrls(r, values)

	w.WriteHeader(http.StatusAccepted)
}

func (s *Handler) deleteUrls(r *http.Request, values []string) {
	defer func() {
		if x := recover(); x != nil {
			log.Println("runtime panic:", x)
			s.deleteUrls(r, values)
		}
	}()

	session, _ := s.sessionsStore.Get(r, s.cookieName)
	if session.Values["uuid"] == nil {
		return
	}
	user, err := s.Store.User().FindByUUID(session.Values["uuid"].(string))
	if err != nil {
		log.Println("ERR:", err)
		return
	}
	urls, err := s.Store.URL().FindByUserID(user.ID)
	if err != nil {
		log.Println("user URL err:", err)
		return
	}

	var ids []int
	for _, url := range urls {
		for _, v := range values {
			if url.URLShort == v {
				if !url.IsDeleted {
					ids = append(ids, url.ID)
				}
			}
		}
	}

	if len(ids) > 0 {
		err := s.Store.URL().BatchDelete(ids)
		if err != nil {
			log.Println("delete error:", err)
		}
	}
}

func (s *Handler) batch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.fail(w, err)
		return
	}
	defer r.Body.Close()

	type data struct {
		CorrelationID string  `json:"correlation_id"`
		OriginalURL   *string `json:"original_url,omitempty"`
		ShortURL      *string `json:"short_url,omitempty"`
	}

	var result []data

	err = json.Unmarshal(b, &result)
	if err != nil {
		s.fail(w, err)
	}

	for i, v := range result {
		shortURL := utils.RandString(s.LinkLen)

		if err := s.Store.URL().Create(&model.URL{
			URLOrigin: *v.OriginalURL,
			URLShort:  shortURL,
		}); err != nil {
			s.fail(w, err)
			return
		}

		str := s.BaseURL + "/" + shortURL
		result[i].OriginalURL = nil
		result[i].ShortURL = &str
	}

	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(result); err != nil {
		s.fail(w, err)
		return
	}
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
		if errors.Is(err, store.ErrUserNotFound) {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if err != nil {
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

	url.URLShort = utils.RandString(s.LinkLen)

	err = s.Store.URL().Create(url)
	if errors.Is(err, store.ErrURLExist) {
		encodeJSON(w, http.StatusConflict, map[string]interface{}{
			"result": s.BaseURL + "/" + url.URLShort,
		})
		return
	}
	if err != nil {
		s.fail(w, err)
		return
	}

	if err := s.SaveURL(r, url); err != nil {
		s.fail(w, err)
		return
	}

	encodeJSON(w, http.StatusCreated, map[string]interface{}{
		"result": s.BaseURL + "/" + url.URLShort,
	})
}

func encodeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
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

// https://pkg.go.dev/context#WithValue
type ctxKeyLocation struct{}

func (s *Handler) ParseURL(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		url, err := s.Store.URL().FindByUUID(id)
		if err != nil {
			s.fail(w, ErrIncorrectID)
			return
		}

		if url.IsDeleted {
			w.WriteHeader(http.StatusGone)
			return
		}

		ctx := context.WithValue(r.Context(), ctxKeyLocation{}, url.URLOrigin)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RedirectHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if location, ok := ctx.Value(ctxKeyLocation{}).(string); ok {
		http.Redirect(w, r, location, http.StatusTemporaryRedirect)
		return
	}
	w.WriteHeader(http.StatusInternalServerError)
}

func (s *Handler) PostHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "text/plain")

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
		URLShort:  utils.RandString(s.LinkLen),
	}

	err = s.Store.URL().Create(url)
	if errors.Is(err, store.ErrURLExist) {
		basicResponse(w, http.StatusConflict, []byte(s.BaseURL+"/"+url.URLShort))
		return
	}
	if err != nil {
		s.fail(w, err)
		return
	}

	if err := s.SaveURL(r, url); err != nil {
		s.fail(w, err)
		return
	}

	basicResponse(w, http.StatusCreated, []byte(s.BaseURL+"/"+url.URLShort))
}

func basicResponse(w http.ResponseWriter, statusCode int, body []byte) {
	w.WriteHeader(statusCode)
	w.Write(body)
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
