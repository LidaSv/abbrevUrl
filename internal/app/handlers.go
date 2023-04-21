package app

import (
	"abbrevUrl/internal/compress"
	"abbrevUrl/internal/storage"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
)

const (
	paramURL            = "id"
	typeLocation        = "Location"
	typeContentType     = "Content-Type"
	bodyContentType     = "text/plain"
	bodyContentTypeJSON = "application/json"
)

type Storage interface {
	HaveLongURL(string, string) string
	HaveShortURL(string) string
	Inc(string, string, string)
	TakeAllURL(string) []storage.AllJSONGet
}

type Hand struct {
	url Storage
}

type JSONLink struct {
	LongURL  string `json:"url,omitempty"`
	ShortURL string `json:"result,omitempty"`
}

func HelpHandler(url Storage) *Hand {
	return &Hand{url: url}
}

func getCookies(r *http.Request) (string, error) {
	name := "clientCookie"
	z, err := r.Cookie(name)
	if err != nil {
		log.Println("Not cookie")
		return "", errors.New("Not cookie")
	}

	//log.Println(z.Value)
	if len(z.Value) == 5 {
		IP := z.Value
		return IP, nil
	}
	IP, err := compress.UnhashCookie(z.Value, name)
	if err != nil {
		log.Println("Not able to unhash Cookie")
		return "", errors.New("Not able to unhash Cookie")
	}
	//log.Println(IP)
	return IP, nil
}

func (s *Hand) AllJSONGetShortenHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(typeContentType, bodyContentTypeJSON)

	IP, err := getCookies(r)
	if err != nil {
		fmt.Fprint(w, err)
		return
	}

	l := s.url.TakeAllURL(IP)

	if l == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	txBz, err := json.MarshalIndent(l, "", "  ")
	if err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write(txBz)
}

func (s *Hand) ShortenJSONLinkHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(typeContentType, bodyContentTypeJSON)

	IP, err := getCookies(r)
	if err != nil {
		fmt.Fprint(w, err)
		return
	}

	longURLByte, err := compress.ReadBody(w, r)
	defer r.Body.Close()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(longURLByte)
		return
	}

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

	shortURL := s.url.HaveLongURL(value.LongURL, IP)

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

	IP, err := getCookies(r)
	if err != nil {
		fmt.Fprint(w, err)
		return
	}

	longURLByte, err := compress.ReadBody(w, r)
	defer r.Body.Close()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(longURLByte)
		return
	}

	longURL := string(longURLByte)

	shortURL := s.url.HaveLongURL(longURL, IP)

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
	http.Redirect(w, r, longURL, http.StatusTemporaryRedirect)
	w.Write([]byte(longURL))
}
