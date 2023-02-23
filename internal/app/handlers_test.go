package app

import (
	"bytes"
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
		})

	}
}
