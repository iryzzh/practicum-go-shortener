package handlers_test

import (
	"github.com/iryzzh/practicum-go-shortener/internal/app/handlers"
	"github.com/iryzzh/practicum-go-shortener/internal/app/model"
	"github.com/iryzzh/practicum-go-shortener/internal/app/store/memstore"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var (
	linkLen = 8
)

func Test_handler_GetHandler(t *testing.T) {
	st := memstore.New()
	url := model.TestURL(t)
	st.URL().Create(url)

	type params struct {
		id string
	}
	type want struct {
		code     int
		location string
	}
	tests := []struct {
		name   string
		params params
		want   want
	}{
		{
			name: "Get with valid id",
			want: want{
				code:     http.StatusTemporaryRedirect,
				location: url.URLOrigin,
			},
			params: params{
				id: url.URLShort,
			},
		},
		{
			name: "Get with invalid id",
			want: want{code: http.StatusBadRequest},
			params: params{
				"invalid-id",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &handlers.Handler{
				ServeMux: http.NewServeMux(),
				Store:    st,
				LinkLen:  0,
			}
			r := httptest.NewRequest(http.MethodGet, "/"+tt.params.id, nil)
			w := httptest.NewRecorder()
			h := http.HandlerFunc(s.GetHandler)
			h.ServeHTTP(w, r)

			result := w.Result()
			defer result.Body.Close()

			if result.StatusCode != tt.want.code {
				t.Errorf("Expected status code %d, got %d", tt.want.code, w.Code)
			}
			if result.Header.Get("Location") != tt.want.location {
				t.Errorf("Expected location %v, got %v", tt.want.location, result.Header.Get("Location"))
			}
		})
	}
}

func Test_handler_PostHandler(t *testing.T) {
	st := memstore.New()

	type params struct {
		body   string
		target string
	}
	type want struct {
		code int
	}
	tests := []struct {
		name   string
		params params
		want   want
	}{
		{
			name: "Post valid url",
			want: want{http.StatusCreated},
			params: params{
				body:   "https://example.com",
				target: "/",
			},
		},
		{
			name: "Post invalid url #1",
			want: want{http.StatusBadRequest},
			params: params{
				body:   "httsp://ya.ru",
				target: "/",
			},
		},
		{
			name: "Post invalid url #2",
			want: want{http.StatusBadRequest},
			params: params{
				body:   "y\\a.ru",
				target: "/",
			},
		},
		{
			name: "Post invalid url #3",
			want: want{http.StatusBadRequest},
			params: params{
				body:   "y\\a.ru",
				target: "/invalid/path",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &handlers.Handler{
				ServeMux: http.NewServeMux(),
				Store:    st,
				LinkLen:  linkLen,
			}
			reader := strings.NewReader(tt.params.body)
			r := httptest.NewRequest(http.MethodPost, tt.params.target, reader)
			r.Host = "localhost:8080"
			w := httptest.NewRecorder()
			h := http.HandlerFunc(s.PostHandler)
			h.ServeHTTP(w, r)

			result := w.Result()
			defer result.Body.Close()

			if result.StatusCode != tt.want.code {
				t.Errorf("Expected status code %d, got %d", tt.want.code, w.Code)
			}
		})
	}
}
