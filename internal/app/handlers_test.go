package app

import (
	"abbrevUrl/internal/storage"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUrlShort(t *testing.T) {
	type args struct {
		url        string
		wantStatus int
		wantURL    string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Normal link generate #1",
			args: args{
				url:        "https://github.com/go-resty/resty",
				wantStatus: 201,
			},
		},
		{
			name: "Normal link generate #2",
			args: args{
				url:        "https://vk.com",
				wantStatus: 201,
			},
		},
		{
			name: "Not normal link generate #1",
			args: args{
				url:        "",
				wantStatus: 400,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := storage.Iter()
			s := HelpHandler(st)
			body := bytes.NewBuffer([]byte(tt.args.url))
			request := httptest.NewRequest(http.MethodPost, "/", body)
			w := httptest.NewRecorder()
			h := http.HandlerFunc(s.ShortenLinkHandler)
			h.ServeHTTP(w, request)
			res := w.Result()
			defer res.Body.Close()

			if len(tt.args.url) == 0 {
				if res.StatusCode != http.StatusBadRequest {
					t.Errorf("want status 400, have %d", w.Code)
				}
			} else {
				if res.StatusCode != http.StatusCreated {
					t.Errorf("want status 201, have %d", w.Code)
				}
			}
		})

	}
}

func TestShortenJSONLinkHandler(t *testing.T) {
	type args struct {
		url        string
		wantStatus int
		incorrect  string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Normal link generate #1",
			args: args{
				url:        `{"url": "https://github.com/go-resty/resty"}`,
				wantStatus: 201,
				incorrect:  "",
			},
		},
		{
			name: "Not normal link generate #1",
			args: args{
				url:        `{"URL": "https://vk.com"}`,
				wantStatus: 400,
				incorrect:  "Incorrect URL",
			},
		},
		{
			name: "Not normal link generate #2",
			args: args{
				url:        `{"url": ""}`,
				wantStatus: 400,
				incorrect:  "Incorrect URL",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := storage.Iter()
			s := HelpHandler(st)
			body := bytes.NewBuffer([]byte(tt.args.url))
			request := httptest.NewRequest(http.MethodPost, "/api/shorten", body)
			w := httptest.NewRecorder()
			h := http.HandlerFunc(s.ShortenJSONLinkHandler)
			h.ServeHTTP(w, request)
			res := w.Result()
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}

			value := JSONLink{}
			if err := json.Unmarshal([]byte(tt.args.url), &value); err != nil {
				t.Error(err)
			}

			if value.LongURL == "" {
				if res.StatusCode != http.StatusBadRequest {
					t.Errorf("want status 400, have %d", w.Code)
				}
				if string(resBody) != tt.args.incorrect {
					t.Errorf("expected body %s, got %s", tt.args.incorrect, string(resBody))
				}
			} else {
				if res.StatusCode != http.StatusCreated {
					t.Errorf("want status 201, have %d", w.Code)
					t.Error(string(resBody))
				}
			}
		})

	}
}
