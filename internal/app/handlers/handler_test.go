package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/iryzzh/practicum-go-shortener/cmd/shortener/config"
	"github.com/iryzzh/practicum-go-shortener/internal/app/handlers"
	"github.com/iryzzh/practicum-go-shortener/internal/app/model"
	"github.com/iryzzh/practicum-go-shortener/internal/app/store"
	"github.com/iryzzh/practicum-go-shortener/internal/app/store/memstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func newTestServer(st store.Store) (*httptest.Server, error) {
	cfg, err := config.New()
	if err != nil {
		return nil, err
	}

	l, err := net.Listen("tcp", cfg.BindAddress)
	if err != nil {
		log.Fatal(err)
	}

	handler := handlers.New(cfg.URLLen, cfg.BaseURL, st, []byte(cfg.SessionKey))

	ts := httptest.NewUnstartedServer(handler)
	ts.Listener.Close()
	ts.Listener = l

	ts.Start()

	return ts, nil
}

func testRequest(t *testing.T, method, path string, body io.Reader, jar *cookiejar.Jar) (*http.Response, string) {
	req, err := http.NewRequest(method, path, body)
	require.NoError(t, err)

	if jar == nil {
		jar, err = cookiejar.New(nil)
		if err != nil {
			t.Fatal(err)
		}
	}

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Jar: jar,
	}

	resp, err := client.Do(req)
	require.NoError(t, err)

	respBody, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(respBody)
}

func TestHandler_Get(t *testing.T) {
	st := memstore.New()
	url := model.TestURL(t)
	if err := st.URL().Create(url); err != nil {
		t.Fatal(err)
	}

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
			ts, err := newTestServer(st)
			if err != nil {
				t.Fatal(err)
			}
			defer ts.Close()

			resp, _ := testRequest(t, "GET", ts.URL+"/"+tt.params.id, nil, nil)
			defer resp.Body.Close() // CI go vet

			assert.Equal(t, tt.want.code, resp.StatusCode)
			assert.Equal(t, resp.Header.Get("Location"), tt.want.location)
		})
	}
}

func TestHandler_Post(t *testing.T) {
	st := memstore.New()

	type want struct {
		code int
	}
	tests := []struct {
		name string
		want want
		body string
	}{
		{
			name: "Post valid url",
			want: want{http.StatusCreated},
			body: "https://example.com",
		},
		{
			name: "Post invalid url #1",
			want: want{http.StatusBadRequest},
			body: "httsp://ya.ru",
		},
		{
			name: "Post invalid url #2",
			want: want{http.StatusBadRequest},
			body: "httsp://ya.ru",
		},
		{
			name: "Post invalid url #3",
			want: want{http.StatusBadRequest},
			body: "httsp://ya.ru",
		},
		{
			name: "Post conflict",
			want: want{http.StatusConflict},
			body: model.TestURL(t).URLOrigin,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.want.code == http.StatusConflict {
				if err := st.URL().Create(model.TestURL(t)); err != nil {
					t.Fatal(err)
				}
			}

			ts, err := newTestServer(st)
			if err != nil {
				t.Fatal(err)
			}
			defer ts.Close()

			body := strings.NewReader(tt.body)
			r, b := testRequest(t, "POST", ts.URL, body, nil)
			defer r.Body.Close()

			assert.Equal(t, tt.want.code, r.StatusCode)

			if r.StatusCode != http.StatusBadRequest {
				result, _ := testRequest(t, "GET", b, nil, nil)
				defer result.Body.Close()
				assert.Equal(t, tt.body, result.Header.Get("location"))
			}
		})
	}
}

func TestHandler_API_Shorten_Post(t *testing.T) {
	st := memstore.New()
	endpoint := "/api/shorten"

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
		{
			name:           "test conflict",
			body:           map[string]interface{}{"url": model.TestURL(t).URLOrigin},
			wantStatusCode: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantStatusCode == http.StatusConflict {
				if err := st.URL().Create(model.TestURL(t)); err != nil {
					t.Fatal(err)
				}
			}
			ts, err := newTestServer(st)
			if err != nil {
				t.Fatal(err)
			}
			defer ts.Close()

			body, _ := json.Marshal(tt.body)
			r, b := testRequest(t, "POST", ts.URL+endpoint, bytes.NewReader(body), nil)
			defer r.Body.Close()

			assert.Equal(t, tt.wantStatusCode, r.StatusCode)

			if r.StatusCode == http.StatusBadRequest {
				return
			}

			assert.True(t, r.Header.Get("content-type") == "application/json")

			var resp Response
			if err := json.Unmarshal([]byte(b), &resp); err != nil {
				t.Fatal(err)
			}

			result, _ := testRequest(t, "GET", resp.Result, nil, nil)
			defer result.Body.Close()

			assert.Equal(t, tt.body["url"].(string), result.Header.Get("location"))
		})
	}
}

func TestHandler_API_Shorten_Batch(t *testing.T) {
	st := memstore.New()
	endpoint := "/api/shorten/batch"

	type Request []struct {
		CorrelationID string `json:"correlation_id"`
		OriginalURL   string `json:"original_url"`
	}

	type Response []struct {
		CorrelationID string `json:"correlation_id"`
		ShortURL      string `json:"short_url"`
	}

	tests := []struct {
		name           string
		body           Request
		wantStatusCode int
	}{
		{
			name: "test correct batch",
			body: Request{
				{
					CorrelationID: "a57024fe-3ffe-494b-a85f-9496cdc09ae6",
					OriginalURL:   "http://wixbzuqq.yandex/whlbtt0uq0/ytutnpencn839",
				},
				{
					CorrelationID: "b24dc7d3-2a22-4741-bc55-b29bce02696d",
					OriginalURL:   "http://vbeexc.ru/eitl9fupmfn/ciucbp8kc4iuf",
				},
			},
			wantStatusCode: http.StatusCreated,
		},
		{
			name: "test incorrect batch",
			body: Request{
				{
					CorrelationID: "a57024fe-3ffe-494b-a85f-9496cdc09ae6",
					OriginalURL:   "http://wrong",
				},
				{
					CorrelationID: "b24dc7d3-2a22-4741-bc55-b29bce02696d",
					OriginalURL:   "http://vbeexc.ru/eitl9fupmfn/ciucbp8kc4iuf",
				},
			},
			wantStatusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, err := newTestServer(st)
			if err != nil {
				t.Fatal(err)
			}
			defer ts.Close()
			body, _ := json.Marshal(tt.body)
			r, b := testRequest(t, "POST", ts.URL+endpoint, bytes.NewReader(body), nil)
			defer r.Body.Close()

			assert.Equal(t, tt.wantStatusCode, r.StatusCode)

			if r.StatusCode == http.StatusBadRequest {
				return
			}

			assert.True(t, r.Header.Get("content-type") == "application/json")

			var resp Response
			if err := json.Unmarshal([]byte(b), &resp); err != nil {
				t.Fatal(err)
			}

			for i, v := range resp {
				assert.Equal(t, tt.body[i].CorrelationID, v.CorrelationID)

				result, _ := testRequest(t, "GET", v.ShortURL, nil, nil)
				defer result.Body.Close()
				assert.Equal(t, result.Header.Get("location"), tt.body[i].OriginalURL)
			}
		})
	}
}

func TestHandler_API_User_Urls_Delete(t *testing.T) {
	st := memstore.New()
	endpoint := "/api/user/urls"

	tests := []struct {
		name           string
		sameCookie     bool
		wantStatusCode int
	}{
		{
			name:           "deleting belonging records",
			sameCookie:     true,
			wantStatusCode: http.StatusGone,
		},
		{
			name:           "deleting not belonging records",
			sameCookie:     false,
			wantStatusCode: http.StatusTemporaryRedirect,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, err := newTestServer(st)
			if err != nil {
				t.Fatal(err)
			}
			defer ts.Close()

			var jar *cookiejar.Jar

			if tt.sameCookie {
				jar, err = cookiejar.New(nil)
				if err != nil {
					t.Fatal(err)
				}
			}

			var values []string
			var wg sync.WaitGroup

			valuesCount := 10

			once := sync.Once{}
			done := make(chan struct{})
			mut := sync.Mutex{}

			for i := 0; i < valuesCount; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()

					url := model.TestURLGenerated(t)

					res, body := testRequest(t, "POST", ts.URL, strings.NewReader(url.URLOrigin), jar)
					res.Body.Close()

					assert.Equal(t, http.StatusCreated, res.StatusCode)

					mut.Lock()
					values = append(values, filepath.Base(body))
					mut.Unlock()

					once.Do(func() {
						close(done)
					})
				}()
				<-done
			}
			wg.Wait()

			valuesJSON, err := json.Marshal(values)
			if err != nil {
				t.Fatal(err)
			}
			response, _ := testRequest(t, "DELETE", ts.URL+endpoint, bytes.NewReader(valuesJSON), jar)
			defer response.Body.Close()

			assert.Equal(t, http.StatusAccepted, response.StatusCode)

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			_ = ctx
			defer cancel()

			for _, v := range values {
				wg.Add(1)

				go func(v string) {
					defer wg.Done()

					for {
						response, _ := testRequest(t, "GET", ts.URL+"/"+v, nil, nil)
						response.Body.Close()

						select {
						case <-ctx.Done():
							return
						default:
							if tt.wantStatusCode == response.StatusCode {
								return
							}
						}
					}
				}(v)
			}
			wg.Wait()
		})
	}
}
