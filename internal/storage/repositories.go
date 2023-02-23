package storage

import (
	"math/rand"
	"net/http"
	"time"
)

var (
	CacheLongURL map[string]string // cache map[longURL]newID
)

func init() {
	CacheLongURL = make(map[string]string)
}

var (
	charset = "abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

type MyInter interface {
	HaveLongURL(string) string
	HaveShortURL(string) (string, int)
}

type MyStruc struct{}

func randSeq() string {
	rand.Seed(time.Now().UnixNano())
	newID := make([]byte, 7)
	for i := range newID {
		newID[i] = charset[rand.Intn(len(charset))]
	}
	return string(newID)
}

func (t *MyStruc) HaveLongURL(longURL string) string {

	if val, ok := CacheLongURL[longURL]; ok {
		return val
	}
	newID := randSeq()
	CacheLongURL[longURL] = newID
	return newID
}

func (t *MyStruc) HaveShortURL(u string) (string, int) {
	for key, val := range CacheLongURL {
		if u == val {
			return key, http.StatusTemporaryRedirect
		}
	}
	return "Short URL not in memory", http.StatusBadRequest
}
