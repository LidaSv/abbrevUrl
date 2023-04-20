package storage

import (
	"math/rand"
	"strings"
	"sync"
)

type URLStorage struct {
	mutex   sync.RWMutex
	Urls    map[string]string
	BaseURL string
}

type AllJSONGet struct {
	ShortURL    string `json:"short_url,omitempty"`
	OriginalURL string `json:"original_url,omitempty"`
}

func Iter() *URLStorage {
	return &URLStorage{Urls: make(map[string]string)}
}

func (u *URLStorage) TakeAllURL() []AllJSONGet {

	var l []AllJSONGet

	if len(u.Urls) == 0 {
		return nil
	}

	for key, value := range u.Urls {
		if strings.HasPrefix(value, "https://") {
			z := AllJSONGet{
				ShortURL:    u.BaseURL + "/" + key,
				OriginalURL: value,
			}
			l = append(l, z)
		}
	}
	return l
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

	BaseURLNew := u.BaseURL

	if BaseURLNew[len(BaseURLNew)-1:] == "/" {
		BaseURLNew = BaseURLNew[:len(BaseURLNew)-1]
	}

	if val != "" {
		shortURL := BaseURLNew + "/" + val
		return shortURL
	}

	//Сокращение URL
	replacer := strings.NewReplacer("https://", "", "/", "", "http://", "", "www.", "", ".", "", "-", "")
	repl := replacer.Replace(longURL)
	newID := u.randSeq(repl)

	shortURL := BaseURLNew + "/" + newID
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
