package app

import (
	"abbrevUrl/internal/middleware"
	"abbrevUrl/internal/storage"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	paramURL            = "id"
	typeLocation        = "Location"
	typeContentType     = "Content-Type"
	bodyContentType     = "text/plain"
	bodyContentTypeJSON = "application/json"
)

type Storage interface {
	HaveLongURL(string, string) (string, error)
	HaveShortURL(string) (string, error)
	Inc(string, string, string)
	TakeAllURL(string) []storage.AllJSONGet
	ShortenDBLink(string) (string, error)
	DatabaseDsns(string) *pgx.Conn
}

type Hand struct {
	url Storage
}

type JSONLink struct {
	LongURL  string `json:"url,omitempty"`
	ShortURL string `json:"result,omitempty"`
}

type OriginLinks struct {
	ID          string `json:"correlation_id,omitempty"`
	OriginalURL string `json:"original_url,omitempty"`
}

type OriginLinksShort struct {
	ID       string `json:"correlation_id,omitempty"`
	ShortURL string `json:"short_url,omitempty"`
}

func HelpHandler(url Storage) *Hand {
	return &Hand{url: url}
}

type ShortURL []string

func (s *Hand) DeleteShortLink(w http.ResponseWriter, r *http.Request) {

	//workersCount := 10
	//done := make(chan bool)

	_, err := getCookies(r)
	if err != nil {
		fmt.Fprint(w, err)
		return
	}

	shortURLByte, err := middleware.ReadBody(w, r)
	defer r.Body.Close()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(shortURLByte)
		return
	}

	var t ShortURL
	err = json.Unmarshal(shortURLByte, &t)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Fatal("Unmarshal: ", err)
	}
	param := "{" + strings.Join(t, ",") + "}"
	db := s.url.DatabaseDsns(param)

	if db != nil {
		_, err := db.Exec(context.Background(),
			`update long_short_urls
				 set flg_delete = 1
				 where short_url = any($1)
				 ;`, param)
		if err != nil {
			log.Fatal("update: ", err)
		}
	}

	if db != nil {
		time.AfterFunc(10*time.Second, func() {
			_, err := db.Exec(context.Background(),
				`delete from long_short_urls
					 where flg_delete = 1;`)
			if err != nil {
				log.Fatal("delete: ", err)
			}
		})
	}

	w.WriteHeader(http.StatusAccepted)
}

func (s *Hand) ShortenDBLinkHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(typeContentType, bodyContentTypeJSON)

	longURLByte, err := middleware.ReadBody(w, r)
	defer r.Body.Close()
	if err != nil {
		log.Println("Read body: ", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(longURLByte)
		return
	}

	var value []OriginLinks
	err = json.Unmarshal(longURLByte, &value)
	if err != nil {
		log.Println("Unmarshal: ", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Incorrect type URL"))
		return
	}

	var short []OriginLinksShort
	for _, t := range value {
		shortURL, _ := s.url.ShortenDBLink(t.OriginalURL)
		z := OriginLinksShort{
			ID:       t.ID,
			ShortURL: shortURL,
		}
		short = append(short, z)
	}

	if short == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	txBz, err := json.MarshalIndent(short, "", "  ")
	if err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(txBz)
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

	longURLByte, err := middleware.ReadBody(w, r)
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

	shortURL, errLongURL := s.url.HaveLongURL(value.LongURL, IP)

	tx := JSONLink{
		ShortURL: shortURL,
	}
	txBz, err := json.Marshal(tx)
	if err != nil {
		panic(err)
	}

	if errLongURL != nil {
		w.WriteHeader(http.StatusConflict)
		w.Write(txBz)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write(txBz)
}

func (s *Hand) ShortenLinkHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(typeContentType, bodyContentType)

	IP, err := getCookies(r)
	if err != nil {
		fmt.Fprint(w, err)
		log.Print("getCookies: ", err)
		return
	}

	longURLByte, err := middleware.ReadBody(w, r)
	defer r.Body.Close()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err)
		return
	}

	longURL := string(longURLByte)

	shortURL, errLongURL := s.url.HaveLongURL(longURL, IP)

	if errLongURL != nil {
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(shortURL))
		return
	}

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

	longURL, err := s.url.HaveShortURL(newID)
	if err != nil {
		w.Header().Set(typeLocation, longURL)
		http.Redirect(w, r, longURL, http.StatusGone)
		w.Write([]byte(longURL))
		return
	}

	if longURL == "Short URL not in memory" {
		w.Header().Set(typeLocation, longURL)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(longURL))
		return
	}

	w.Header().Set(typeLocation, longURL)
	http.Redirect(w, r, longURL, http.StatusTemporaryRedirect)
	w.Write([]byte(longURL))
}

func getCookies(r *http.Request) (string, error) {
	name := "clientCookie"
	z, err := r.Cookie(name)
	if err != nil {
		log.Println("Not cookie")
		return "", errors.New("not cookie")
	}

	if len(z.Value) == 5 {
		IP := z.Value
		return IP, nil
	}
	IP, err := middleware.UnhashCookie(z.Value, name)
	if err != nil {
		log.Println("Not able to unhash Cookie")
		return "", errors.New("not able to unhash Cookie")
	}
	return IP, nil
}

func (s *Hand) PingPSQL(w http.ResponseWriter, r *http.Request) {

	db := s.url.DatabaseDsns("")

	if db != nil {
		err := db.Ping(context.Background())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Unable to connect to database: %v\n", err)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("DB connection"))
		return
	}

	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprint(w, "Unable to connect to database: db doesn't have url")
}
