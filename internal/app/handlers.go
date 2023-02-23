package app

import (
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"

	"abbrevUrl/internal/storage"
)

const (
	URLPrefix       = "http://localhost:8080/"
	paramURL        = "id"
	typeLocation    = "Location"
	typeContentType = "Content-Type"
	bodyContentType = "text/plain"
)

var (
	MyInter storage.MyInter
	MyStruc *storage.MyStruc
)

func ShortenLinkHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(typeContentType, bodyContentType)
	MyInter = MyStruc

	defer r.Body.Close()
	longURLByte, err := io.ReadAll(r.Body)
	if err != nil || len(longURLByte) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Incorrect URL"))
		return
	}

	newID := MyInter.HaveLongURL(string(longURLByte))

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(URLPrefix + newID))
}

func GetShortenHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(typeContentType, bodyContentType)
	MyInter = MyStruc

	newID := chi.URLParam(r, paramURL)

	if newID == "" {
		http.Error(w, "ID param is missed", http.StatusBadRequest)
		return
	}

	longURL, status := MyInter.HaveShortURL(newID)

	w.Header().Set(typeLocation, longURL)
	w.WriteHeader(status)
	w.Write([]byte(longURL))
}
