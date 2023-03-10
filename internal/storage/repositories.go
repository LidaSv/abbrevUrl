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
	HaveShortURL(string) string
}

type MyURL struct {
	ID       string `json:"id"`
	LongUrl  string `json:"longUrl"`
	ShortUrl string `json:"shortUrl"`
}

var urls = map[string]MyURL{}

func (l *MyURL) randSeq(longURL string) string {

	newURL := longURL

	newID := make([]byte, 7)
	for i := range newID {
		newID[i] = newURL[rand.Intn(len(newURL))]
	}

	if _, ok := urls[l.LongUrl]; ok {
		l.randSeq(longURL)
	}
	return string(newID)
}

func (l *MyURL) HaveLongURL() string {
	var appURL MyURL

	if val, ok := urls[l.LongUrl]; ok {
		return val.ShortUrl
	}

	//Сокращение URL
	replacer := strings.NewReplacer("https://", "", "/", "", "http://", "", "www.", "", ".", "")
	repl := replacer.Replace(l.LongUrl)
	newID := l.randSeq(repl)

	appURL.ID = newID
	appURL.LongUrl = l.LongUrl
	appURL.ShortUrl = URLPrefix + newID

	urls[l.LongUrl] = appURL
	urls[appURL.ID] = appURL

	return appURL.ShortUrl

}

func (l *MyURL) HaveShortURL() string {
	if val, ok := urls[l.ID]; ok {
		return val.LongUrl
	}
	return "Short URL not in memory"
}
