package storage

import (
	"math/rand"
	"strings"
	"sync"
)

const (
	URLPrefix = "http://localhost:8080/"
)

type CacheURL struct {
	ID       string
	LongURL  string
	ShortURL string
}

type UrlStorage struct {
	mutex sync.RWMutex
	Urls  map[string]string
}

func Iter() *UrlStorage {
	return &UrlStorage{Urls: make(map[string]string)}
}

func (u *UrlStorage) randSeq(longURL string) string {

	newURL := longURL

	newID := make([]byte, 7)
	for i := range newID {
		newID[i] = newURL[rand.Intn(len(newURL))]
	}

	u.mutex.RLock()
	_, ok := u.Urls[longURL]
	defer u.mutex.RUnlock()

	if ok {
		u.randSeq(longURL)
	}
	return string(newID)
}

func (u *UrlStorage) getShortURL(longURL string) string {
	u.mutex.RLock()
	val, ok := u.Urls[longURL]
	defer u.mutex.RUnlock()
	if ok {
		return val
	}
	return ""
}

func (u *UrlStorage) Inc(longURL, newID string) {
	u.mutex.Lock()
	u.Urls[longURL] = newID
	u.Urls[newID] = longURL
	u.mutex.Unlock()
}

func (u *UrlStorage) HaveLongURL(longURL string) string {
	var appURL CacheURL

	val := u.getShortURL(longURL)

	if val != "" {
		s := URLPrefix + val
		return s
	}

	//Сокращение URL
	replacer := strings.NewReplacer("https://", "", "/", "", "http://", "", "www.", "", ".", "")
	repl := replacer.Replace(longURL)
	newID := u.randSeq(repl)

	appURL.ID = newID
	appURL.LongURL = longURL
	appURL.ShortURL = URLPrefix + newID

	u.Inc(longURL, newID)

	return appURL.ShortURL

}

func (u *UrlStorage) HaveShortURL(ID string) string {
	u.mutex.RLock()
	val, ok := u.Urls[ID]
	u.mutex.RUnlock()

	if ok {
		return val
	}
	return "Short URL not in memory"
}
