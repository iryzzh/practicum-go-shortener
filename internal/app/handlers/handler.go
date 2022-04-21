package handlers

import (
	"encoding/json"
	"errors"
	"github.com/iryzzh/practicum-go-shortener/internal/app/model"
	"github.com/iryzzh/practicum-go-shortener/internal/app/store"
	"github.com/iryzzh/practicum-go-shortener/internal/pkg/utils"
	"io"
	"log"
	"net/http"
	"path"
	"regexp"
	"strings"
)

var (
	ErrIncorrectID = errors.New("incorrect ID")
)

type Handler struct {
	*http.ServeMux

	Store   store.Store
	LinkLen int
}

func New(linkLen int, store store.Store) http.Handler {
	s := &Handler{
		ServeMux: http.NewServeMux(),
		LinkLen:  linkLen,
		Store:    store,
	}

	s.HandleFunc("/", s.multiplexer)

	handler := http.Handler(s)
	handler = wrapLogger(handler)

	return handler
}

func wrapLogger(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		handler.ServeHTTP(w, r)
	})
}

func (s *Handler) multiplexer(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.GetHandler(w, r)
	case http.MethodPost:
		s.PostHandler(w, r)
	}
}

func (s *Handler) GetHandler(w http.ResponseWriter, r *http.Request) {
	id := s.id(r)
	if id == "" {
		s.fail(w, ErrIncorrectID)
		return
	}
	url, err := s.Store.URL().FindByID(id)
	if err != nil {
		s.fail(w, err)
		return
	}

	http.Redirect(w, r, url.URLOrigin, http.StatusTemporaryRedirect)
}

func (s *Handler) PostHandler(w http.ResponseWriter, r *http.Request) {
	p := strings.Replace(r.URL.Path, "/", "", -1)
	if len(p) > 0 {
		s.fail(w, nil)
		return
	}

	header := r.Header.Get("Content-Type")
	switch {
	case strings.EqualFold(header, "text/plain"), header == "":
	default:
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

	resp := "http://" + r.Host + "/" + m.URLShort

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(resp))
}

func (s *Handler) id(r *http.Request) string {
	base := path.Base(r.URL.Path)
	id := strings.Trim(base, "/")
	re := regexp.MustCompile(`^[a-zA-Z\d]*$`)
	if !re.MatchString(id) {
		return ""
	}

	return id
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

func (s *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet, http.MethodPost:
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	handler, _ := s.Handler(r)
	handler.ServeHTTP(w, r)
}
