package app

import (
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"abbrevUrl/internal/storage"
)

const (
	URLPrefix       = "http://localhost:8080/"
	paramURL        = "id"
	typeLocation    = "Location"
	typeContentType = "Content-Type"
	bodyContentType = "text/plain"
	firstID         = "1"
)

func ShortenLinkHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(typeContentType, bodyContentType)

	longURLByte, err := io.ReadAll(r.Body)
	if err != nil || len(longURLByte) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Incorrect URL"))
		return
	}
	defer r.Body.Close()
	longURL := string(longURLByte)

	if value, ok := storage.CacheLongURL[longURL]; ok {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(URLPrefix + value))
		return
	}

	longURLCut := strings.Replace(strings.Replace(longURL, "www.", "", -1), "https://", "", -1)

	domenCut := strings.Split(longURLCut, ".")[0]
	var newID string
	if val, ok := storage.CacheDomen[domenCut]; ok {
		newID = domenCut + strconv.Itoa(val+1)
		storage.CacheDomen[domenCut] = val + 1
	} else {
		newID = domenCut + firstID
		storage.CacheDomen[domenCut] = 1
	}

	storage.CacheNewID[newID] = longURL
	storage.CacheLongURL[longURL] = newID

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(URLPrefix + newID))
}

func GetShortenHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(typeContentType, bodyContentType)

	newID := chi.URLParam(r, paramURL)

	if newID == "" {
		http.Error(w, "ID param is missed", http.StatusBadRequest)
		return
	}

	if value, ok := storage.CacheNewID[newID]; ok {
		w.Header().Set(typeLocation, value)
		w.WriteHeader(http.StatusTemporaryRedirect)
		return
	}

	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte("Short URL not in memory"))
}
