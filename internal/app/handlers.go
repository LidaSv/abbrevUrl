package app

import (
	"github.com/go-chi/chi/v5"
	"io"
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
	urls  []AllURL
	index int
)

func IndexID(index *int) int {
	*index++
	return *index
}

func ShortenLinkHander(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	var url AllURL
	defer r.Body.Close()
	longURLByte, err := io.ReadAll(r.Body)
	if err != nil || len(longURLByte) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Incorrect URL"))
		return
	}
	longURL := string(longURLByte)
	for _, val := range urls {
		if val.LongURL == longURL {
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(val.ShortURL))
			return
		}
	}

	w.WriteHeader(http.StatusCreated)
	url.ID = "new" + strconv.Itoa(IndexID(&index))
	url.LongURL = longURL
	url.ShortURL = URLPrefix + url.ID
	urls = append(urls, url)
	w.Write([]byte(url.ShortURL))
}

func GetShortenHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	newID := chi.URLParam(r, "id")
	if newID == "" {
		http.Error(w, "ID param is missed", http.StatusBadRequest)
		return
	}
	for _, val := range urls {
		if val.ID == newID {
			w.Header().Set("Location", val.LongURL)
			w.WriteHeader(http.StatusTemporaryRedirect)
		}
	}
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte("Short URL not in memory"))
}
