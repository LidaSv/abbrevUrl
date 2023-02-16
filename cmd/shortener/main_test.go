package main

import (
	"bytes"
	"io"
	"log"
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
			if resp != nil {
				defer resp.Body.Close()
			}
			if err != nil {
				log.Fatal(err)
			}
			if resp.StatusCode != tt.args.wantStatus {
				t.Errorf("want status %d, have %d", tt.args.wantStatus, resp.StatusCode)
			}
			resData, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("cannot read response body")
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
	resp1, err := http.Post(URLPrefix, "text/plain", body)
	if resp1 != nil {
		defer resp1.Body.Close()
	}
	if err != nil {
		log.Fatal(err)
	}
	if resp1.StatusCode != http.StatusCreated {
		t.Errorf("want status 201, have %d", resp1.StatusCode)
	}

	// Save URL we want ro retrieve
	utlToShot := "https://habr.com/ru/article/713190/"
	body = bytes.NewBuffer([]byte(utlToShot))
	resp2, err := http.Post(URLPrefix, "text/plain", body)
	if resp2 != nil {
		defer resp2.Body.Close()
	}
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp2.StatusCode != http.StatusCreated {
		t.Errorf("want status 201, have %d", resp2.StatusCode)
	}
	resData, err := io.ReadAll(resp2.Body)
	if err != nil {
		t.Fatalf("cannot read response body")
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

	resp3, err := client.Get(reqURL)
	if resp3 != nil {
		defer resp3.Body.Close()
	}
	if err != nil {
		t.Fatalf("cannot make GET with shorted ID %s: %v", resID, err)
	}
	if resp3.StatusCode != http.StatusTemporaryRedirect {
		t.Errorf("want status 307 for invalid ID, have %d", resp3.StatusCode)
	}
	locHeader := resp3.Header.Get("Location")
	if locHeader != utlToShot {
		t.Errorf("invalid shorted URL received %s, want %s", locHeader, utlToShot)
	}
}
