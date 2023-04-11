package app

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
)

const (
	paramURL            = "id"
	typeLocation        = "Location"
	typeContentType     = "Content-Type"
	bodyContentType     = "text/plain"
	bodyContentTypeJSON = "application/json"
)

type Inter interface {
	HaveLongURL(string) (string, map[string]string)
	HaveShortURL(string) string
	Inc(string, string)
}

type Hand struct {
	url Inter
}

type JSONLink struct {
	LongURL  string `json:"url,omitempty"`
	ShortURL string `json:"result,omitempty"`
}

func HelpHandler(url Inter) *Hand {
	return &Hand{url: url}
}

func (s *Hand) ShortenJSONLinkHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(typeContentType, bodyContentTypeJSON)

	longURLByte, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Incorrect URL"))
		return
	}
	defer r.Body.Close()

	value := JSONLink{}
	if err := json.Unmarshal(longURLByte, &value); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Incorrect type URL"))
		return
	}

	if value.LongURL == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Incorrect URL"))
		return
	}

	shortURL, _ := s.url.HaveLongURL(value.LongURL)

	tx := JSONLink{
		ShortURL: shortURL,
	}
	txBz, err := json.Marshal(tx)
	if err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(txBz)
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

	shortURL, cache := s.url.HaveLongURL(longURL)

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintln(w, cache)
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
	http.Redirect(w, r, longURL, http.StatusTemporaryRedirect)
	//w.WriteHeader(http.StatusTemporaryRedirect)
	w.Write([]byte(longURL))
}
