package app

import (
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"

	"abbrevUrl/internal/storage"
)

const (
	paramURL        = "id"
	typeLocation    = "Location"
	typeContentType = "Content-Type"
	bodyContentType = "text/plain"
)

type Server storage.CacheURL

func (s *Server) ShortenLinkHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(typeContentType, bodyContentType)

	longURLByte, err := io.ReadAll(r.Body)
	if err != nil || len(longURLByte) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Incorrect URL"))
		return
	}
	defer r.Body.Close()
	longURL := string(longURLByte)

	CacheURL := storage.CacheURL{LongURL: longURL}
	shortURL := CacheURL.HaveLongURL()

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
}

func (s *Server) GetShortenHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(typeContentType, bodyContentType)

	newID := chi.URLParam(r, paramURL)

	if newID == "" {
		http.Error(w, "ID param is missed", http.StatusBadRequest)
		return
	}
	CacheURL := storage.CacheURL{ID: newID}
	longURL := CacheURL.HaveShortURL()

	if longURL == "Short URL not in memory" {
		w.Header().Set(typeLocation, longURL)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(longURL))
	}

	w.Header().Set(typeLocation, longURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
	w.Write([]byte(longURL))
}
