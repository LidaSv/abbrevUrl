package app

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
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
				wantURL:    "Incorrect URL",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := bytes.NewBuffer([]byte(tt.args.url))
			request := httptest.NewRequest(http.MethodPost, "/", body)
			w := httptest.NewRecorder()
			h := http.HandlerFunc(ShortenLinkHandler)
			h.ServeHTTP(w, request)
			res := w.Result()

			if len(tt.args.url) == 0 {
				if res.StatusCode != http.StatusBadRequest {
					t.Errorf("want status 400, have %d", w.Code)
				}
			} else {
				if res.StatusCode != http.StatusCreated {
					t.Errorf("want status 201, have %d", w.Code)
				}
			}
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}

			now := time.Now()
			newID := strconv.FormatInt(now.Unix(), 16)
			if len(tt.args.url) == 0 {
				if string(resBody) != tt.args.wantURL {
					t.Errorf("Expected body %s, got %s", tt.args.wantURL, w.Body.String())
				}
			} else {
				if string(resBody) != URLPrefix+newID {
					t.Errorf("Expected body %s, got %s", URLPrefix+newID, w.Body.String())
				}
			}
		})

	}
}
