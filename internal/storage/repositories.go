package storage

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
	"log"
	"math/rand"
	"strings"
	"sync"
)

type URLStorage struct {
	mutex      sync.RWMutex
	Urls       map[string]string
	IPUrls     map[string][]string
	BaseURL    string
	LocalDB    *pgxpool.Pool
	DeleteURLs string
}

type AllJSONGet struct {
	IP          string `json:"-"`
	ShortURL    string `json:"short_url,omitempty"`
	OriginalURL string `json:"original_url,omitempty"`
}

func Iter() *URLStorage {
	return &URLStorage{
		Urls:   make(map[string]string),
		IPUrls: map[string][]string{},
	}
}

func (u *URLStorage) DatabaseDsns(p string) *pgxpool.Pool {
	u.DeleteURLs = p
	return u.LocalDB
}

func (u *URLStorage) ShortenDBLink(longURL string) (string, error) {

	val, err := u.getShortURL(longURL)

	BaseURLNew := u.BaseURL

	if BaseURLNew[len(BaseURLNew)-1:] == "/" {
		BaseURLNew = BaseURLNew[:len(BaseURLNew)-1]
	}

	if val != "" {
		shortURL := BaseURLNew + "/" + val
		return shortURL, err
	}

	//Сокращение URL
	replacer := strings.NewReplacer("https://", "", "/", "", "http://", "", "www.", "", ".", "", "-", "")
	repl := replacer.Replace(longURL)
	newID := u.randSeq(repl)

	shortURL := BaseURLNew + "/" + newID

	u.Inc(longURL, newID, "")

	return shortURL, err

}

func (u *URLStorage) TakeAllURL(IP string) []AllJSONGet {

	var l []AllJSONGet

	if len(u.IPUrls) == 0 {
		return nil
	}

	i := 0

	for key, value := range u.IPUrls {
		//log.Println(key, IP)
		if key == IP {
			z := AllJSONGet{
				IP:          key,
				ShortURL:    value[i],
				OriginalURL: value[i+1],
			}
			i += 2
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

func (u *URLStorage) getShortURL(longURL string) (string, error) {
	u.mutex.RLock()
	val, ok := u.Urls[longURL]
	defer u.mutex.RUnlock()
	if ok {
		return val, errors.New(`DB has short url`)
	}
	return "", nil
}

func (u *URLStorage) Inc(longURL, newID, IP string) {
	u.mutex.Lock()
	u.Urls[longURL] = newID
	u.Urls[newID] = longURL
	u.IPUrls[IP] = append(u.IPUrls[IP], u.BaseURL+"/"+newID, longURL)
	u.mutex.Unlock()

	if u.LocalDB != nil {
		_, err := u.LocalDB.Exec(context.Background(),
			`insert into long_short_urls (long_url, short_url, id_short_url)
					values ($1, $2, $3)
				;`, longURL, u.BaseURL+"/"+newID, newID)
		if err != nil {
			log.Fatal("insert inc: ", err)
		}
	}
}

func (u *URLStorage) HaveLongURL(longURL, IP string) (string, error) {

	val, err := u.getShortURL(longURL)

	BaseURLNew := u.BaseURL

	if BaseURLNew[len(BaseURLNew)-1:] == "/" {
		BaseURLNew = BaseURLNew[:len(BaseURLNew)-1]
	}

	// Проверка ошибки
	if val != "" {
		shortURL := BaseURLNew + "/" + val
		return shortURL, err
	}

	//Сокращение URL
	replacer := strings.NewReplacer("https://", "", "/", "", "http://", "", "www.", "", ".", "", "-", "")
	repl := replacer.Replace(longURL)
	newID := u.randSeq(repl)

	shortURL := BaseURLNew + "/" + newID
	u.Inc(longURL, newID, IP)

	return shortURL, err

}

func (u *URLStorage) HaveShortURL(ID string) (string, error) {

	u.mutex.RLock()
	longURL, ok := u.Urls[ID]
	u.mutex.RUnlock()
	if ok {
		if strings.Contains(u.DeleteURLs, ID) {
			return longURL, errors.New(`this URL delete`)
		}
		return longURL, nil
	}
	return "Short URL not in memory", nil
}
