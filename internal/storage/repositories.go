package storage

import (
	"math/rand"
	"strings"
	"sync"
)

const (
	URLPrefix = "http://localhost:8080/"
)

//type CacheURL struct {
//	ID       string
//	LongURL  string
//	ShortURL string
//}

type URLStorage struct {
	mutex sync.RWMutex
	Urls  map[string]string
}

func Iter() *URLStorage {
	return &URLStorage{Urls: make(map[string]string)}
}

func (u *URLStorage) randSeq(longURL string) string {

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

func (u *URLStorage) getShortURL(longURL string) string {
	u.mutex.RLock()
	val, ok := u.Urls[longURL]
	defer u.mutex.RUnlock()
	if ok {
		return val
	}
	return ""
}

func (u *URLStorage) Inc(longURL, newID string) {
	u.mutex.Lock()
	u.Urls[longURL] = newID
	u.Urls[newID] = longURL
	u.mutex.Unlock()
}

func (u *URLStorage) HaveLongURL(longURL string) string {

	val := u.getShortURL(longURL)

	if val != "" {
		shortURL := URLPrefix + val
		return shortURL
	}

	//Сокращение URL
	replacer := strings.NewReplacer("https://", "", "/", "", "http://", "", "www.", "", ".", "")
	repl := replacer.Replace(longURL)
	newID := u.randSeq(repl)

	shortURL := URLPrefix + newID
	u.Inc(longURL, newID)

	return shortURL

}

func (u *URLStorage) HaveShortURL(ID string) string {
	u.mutex.RLock()
	longURL, ok := u.Urls[ID]
	u.mutex.RUnlock()

	if ok {
		return longURL
	}
	return "Short URL not in memory"
}
