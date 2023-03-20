package app

import (
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
)

const (
	paramURL        = "id"
	typeLocation    = "Location"
	typeContentType = "Content-Type"
	bodyContentType = "text/plain"
)

type Inter interface {
	HaveLongURL(string) string
	HaveShortURL(string) string
}

type Hand struct {
	url Inter
}

func HelpHandler(url Inter) *Hand {
	return &Hand{url: url}
}

func (s *Hand) ShortenLinkHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(typeContentType, bodyContentType)

	longURLByte, err := io.ReadAll(r.Body)
	if err != nil || len(longURLByte) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Incorrect URL"))
		return
	}
	defer r.Body.Close()
	longURL := string(longURLByte)

	shortURL := s.url.HaveLongURL(longURL)

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
}

func (s *Hand) GetShortenHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(typeContentType, bodyContentType)

	newID := chi.URLParam(r, paramURL)

	if newID == "" {
		http.Error(w, "ID param is missed", http.StatusBadRequest)
		return
	}

	longURL := s.url.HaveShortURL(newID)

	if longURL == "Short URL not in memory" {
		w.Header().Set(typeLocation, longURL)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(longURL))
	}

	w.Header().Set(typeLocation, longURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
	w.Write([]byte(longURL))
}
