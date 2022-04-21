package handlers_test

import (
	"github.com/iryzzh/practicum-go-shortener/internal/app/handlers"
	"github.com/iryzzh/practicum-go-shortener/internal/app/model"
	"github.com/iryzzh/practicum-go-shortener/internal/app/store/memstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var (
	linkLen = 8
)

func testRequest(t *testing.T, ts *httptest.Server, method, path string, body io.Reader) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, body)
	require.NoError(t, err)

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(req)
	require.NoError(t, err)

	respBody, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	defer resp.Body.Close()

	return resp, string(respBody)
}

func TestHandler_Get(t *testing.T) {
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
			handler := handlers.New(linkLen, st)
			ts := httptest.NewServer(handler)
			defer ts.Close()

			resp, _ := testRequest(t, ts, "GET", "/"+tt.params.id, nil)
			assert.Equal(t, resp.StatusCode, tt.want.code)
			assert.Equal(t, resp.Header.Get("Location"), tt.want.location)
		})
	}
}

func TestHandler_Post(t *testing.T) {
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
			handler := handlers.New(linkLen, st)
			ts := httptest.NewServer(handler)
			defer ts.Close()

			body := strings.NewReader(tt.params.body)
			r, b := testRequest(t, ts, "POST", "/", body)
			assert.Equal(t, r.StatusCode, tt.want.code)

			assert.Condition(t, func() bool {
				if r.StatusCode == http.StatusBadRequest {
					return true
				}
				url := r.Request.URL.String()
				assert.True(t, strings.HasPrefix(b, url))
				l := strings.TrimPrefix(b, url)
				return assert.Equal(t, linkLen, len(l))
			})
		})
	}
}
