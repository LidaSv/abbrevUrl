package storage

import (
	"github.com/caarlos0/env/v6"
	"github.com/spf13/pflag"
	"log"
	"math/rand"
	"os"
	"strings"
	"sync"
)

type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS" envDefault:"localhost:8080"`
	FileStoragePath string `env:"FILE_STORAGE_PATH" envDefault:"internal/storage/cache.log"`
	BaseURL         string `env:"BASE_URL"`
}

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

	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	flag := pflag.FlagSet{}
	FlagBaseURL := flag.String("b", "http://"+cfg.ServerAddress, "a string")
	flag.Parse(os.Args[1:])

	path, exists := os.LookupEnv("BASE_URL")

	var BaseURLNew string
	if exists {
		BaseURLNew = path
	} else {
		BaseURLNew = *FlagBaseURL
	}

	if BaseURLNew[len(BaseURLNew)-1:] == "/" {
		BaseURLNew = BaseURLNew[:len(BaseURLNew)-1]
	}

	if val != "" {
		shortURL := BaseURLNew + "/" + val
		return shortURL
	}

	//Сокращение URL
	replacer := strings.NewReplacer("https://", "", "/", "", "http://", "", "www.", "", ".", "")
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
