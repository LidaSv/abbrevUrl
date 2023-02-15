package main

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"
)

const URLPrefix = "http://localhost:8080/"

func TestUrlShort(t *testing.T) {
	type args struct {
		url        string
		wantStatus int
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Normal link generate",
			args: args{
				url:        "https://github.com/go-resty/resty",
				wantStatus: 201,
			},
		},
		{
			name: "Normal link generate",
			args: args{
				url:        "",
				wantStatus: 400,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := bytes.NewBuffer([]byte(tt.args.url))
			resp, err := http.Post(URLPrefix, "text/plain", body)
			defer resp.Body.Close()
			if err != nil {
				t.Errorf("request failed: %v", err)
			}
			if resp.StatusCode != tt.args.wantStatus {
				t.Errorf("want status %d, have %d", tt.args.wantStatus, resp.StatusCode)
			}
			resData, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Errorf("cannot read response body")
			}
			if resp.StatusCode != http.StatusBadRequest {
				resURL := string(resData)
				if !strings.HasPrefix(resURL, URLPrefix) {
					t.Errorf("not fount http://localhost:8080/ prefix in response: %s", resURL)
				}
			}
		})

	}
}

func TestUrlLongReceive(t *testing.T) {
	//Save first url
	tmpURL := "https://vk.com"
	body := bytes.NewBuffer([]byte(tmpURL))
	resp, err := http.Post(URLPrefix, "text/plain", body)
	defer resp.Body.Close()
	if err != nil {
		t.Errorf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("want status 201, have %d", resp.StatusCode)
	}
	// Save URL we want ro retrieve
	utlToShot := "https://habr.com/ru/article/713190/"
	body = bytes.NewBuffer([]byte(utlToShot))
	resp, err = http.Post(URLPrefix, "text/plain", body)
	defer resp.Body.Close()
	if err != nil {
		t.Errorf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("want status 201, have %d", resp.StatusCode)
	}
	resData, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("cannot read response body")
	}
	resURL := string(resData)
	if !strings.HasPrefix(resURL, URLPrefix) {
		t.Errorf("nit fount http prefix in response: %s", resURL)
	}
	resID := strings.Replace(resURL, URLPrefix, "", 1)
	reqURL := URLPrefix + resID
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp2, err := client.Get(reqURL)
	defer resp2.Body.Close()
	if err != nil {
		t.Fatalf("cannot make GET with shorted ID %s: %v", resID, err)
	}
	if resp2.StatusCode != http.StatusTemporaryRedirect {
		t.Errorf("want status 307 for invalid ID, have %d", resp.StatusCode)
	}
	locHeader := resp2.Header.Get("Location")
	if locHeader != utlToShot {
		t.Errorf("invalid shorted URL received %s, want %s", locHeader, utlToShot)
	}
}
