package handlers_test

import (
	"bytes"
	"encoding/json"
	"github.com/iryzzh/practicum-go-shortener/internal/app/handlers"
	"github.com/iryzzh/practicum-go-shortener/internal/app/model"
	"github.com/iryzzh/practicum-go-shortener/internal/app/store/memstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"path"
	"strings"
	"testing"
)

var (
	linkLen       = 8
	serverAddress = "localhost:8080"
	baseURL       = "http://" + serverAddress
)

func newTestServer(handler http.Handler) *httptest.Server {
	l, err := net.Listen("tcp", serverAddress)
	if err != nil {
		log.Fatal(err)
	}

	ts := httptest.NewUnstartedServer(handler)
	ts.Listener.Close()
	ts.Listener = l

	ts.Start()
	return ts
}

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
			handler := handlers.New(linkLen, baseURL, st)
			ts := newTestServer(handler)
			defer ts.Close()

			resp, _ := testRequest(t, ts, "GET", "/"+tt.params.id, nil)
			defer resp.Body.Close() // CI go vet

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
			handler := handlers.New(linkLen, baseURL, st)
			ts := newTestServer(handler)
			defer ts.Close()

			body := strings.NewReader(tt.params.body)
			r, b := testRequest(t, ts, "POST", "/", body)
			defer r.Body.Close() // CI go vet

			assert.Equal(t, r.StatusCode, tt.want.code)
			assert.Condition(t, func() bool {
				if r.StatusCode == http.StatusBadRequest {
					return true
				}
				assert.True(t, strings.HasPrefix(b, baseURL))
				return assert.Equal(t, linkLen, len(path.Base(b)))
			})
		})
	}
}

func TestHandler_API_Post(t *testing.T) {
	endpoint := "/api/shorten"
	st := memstore.New()

	type Response struct {
		Result string `json:"result"`
	}

	tests := []struct {
		name           string
		body           map[string]interface{}
		wantStatusCode int
	}{
		{
			name:           "test correct link",
			body:           map[string]interface{}{"url": "http://example.com"},
			wantStatusCode: http.StatusCreated,
		},
		{
			name:           "test incorrect body",
			body:           map[string]interface{}{"url2": "http://example.com"},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "test incorrect link",
			body:           map[string]interface{}{"url": "http:\\wrong.com"},
			wantStatusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := handlers.New(linkLen, baseURL, st)
			ts := newTestServer(handler)
			defer ts.Close()

			body, _ := json.Marshal(tt.body)
			r, b := testRequest(t, ts, "POST", endpoint, bytes.NewReader(body))
			defer r.Body.Close()

			assert.Equal(t, tt.wantStatusCode, r.StatusCode)
			assert.Condition(t, func() bool {
				if r.StatusCode == http.StatusBadRequest {
					return true
				}

				var resp Response
				if err := json.Unmarshal([]byte(b), &resp); err != nil {
					t.Fatal(err)
				}

				assert.True(t, r.Header.Get("content-type") == "application/json")

				assert.True(t, strings.Contains(resp.Result, baseURL))
				return assert.Equal(t, linkLen, len(path.Base(resp.Result)))
			})
		})
	}
}
