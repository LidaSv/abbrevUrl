package app

import (
	"github.com/gorilla/mux"
	"io"
	"math/rand"
	"net/http"
	"strconv"
)

const URLPrefix = "http://localhost:8080/"

type AllURL struct {
	ID       string `json:"id"`
	LongURL  string `json:"longURL"`
	ShortURL string `json:"shortURL"`
}

var (
	urls []AllURL
)

func ShortenLinkHander(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	var url AllURL
	defer r.Body.Close()
	longURLByte, err := io.ReadAll(r.Body)
	if len(longURLByte) == 0 || err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Incorrect URL"))
	}
	longURL := string(longURLByte)

	w.WriteHeader(http.StatusCreated)
	url.LongURL = longURL
	url.ID = strconv.Itoa(rand.Intn(1000000))
	url.ShortURL = URLPrefix + url.ID
	urls = append(urls, url)
	w.Write([]byte(url.ShortURL))
}

func GetShortenHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	newID := mux.Vars(r)

	for _, val := range urls {
		if val.ID == newID["id"] {
			w.Header().Set("Location", val.LongURL)
			w.WriteHeader(http.StatusTemporaryRedirect)
		}
	}
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte("Short URL not in memory"))
}
