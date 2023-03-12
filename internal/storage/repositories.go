package storage

import (
	"math/rand"
	"strings"
)

const (
	URLPrefix = "http://localhost:8080/"
)

type MyInter interface {
	HaveLongURL() string
	HaveShortURL() string
}

type CacheURL struct {
	ID       string
	LongURL  string
	ShortURL string
}

var urls = map[string]CacheURL{}

func (l *CacheURL) randSeq(longURL string) string {

	newURL := longURL

	newID := make([]byte, 7)
	for i := range newID {
		newID[i] = newURL[rand.Intn(len(newURL))]
	}

	if _, ok := urls[l.LongURL]; ok {
		l.randSeq(longURL)
	}
	return string(newID)
}

func (l *CacheURL) HaveLongURL() string {
	var appURL CacheURL

	if val, ok := urls[l.LongURL]; ok {
		return val.ShortURL
	}

	//Сокращение URL
	replacer := strings.NewReplacer("https://", "", "/", "", "http://", "", "www.", "", ".", "")
	repl := replacer.Replace(l.LongURL)
	newID := l.randSeq(repl)

	appURL.ID = newID
	appURL.LongURL = l.LongURL
	appURL.ShortURL = URLPrefix + newID

	urls[l.LongURL] = appURL
	urls[appURL.ID] = appURL

	return appURL.ShortURL

}

func (l *CacheURL) HaveShortURL() string {
	if val, ok := urls[l.ID]; ok {
		return val.LongURL
	}
	return "Short URL not in memory"
}
