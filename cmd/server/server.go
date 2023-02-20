package server

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"abbrevUrl/internal/app"
)

const (
	port     = ":8080"
	mainURL  = "/"
	paramURL = "/{id}"
)

func AddServer() {
	r := chi.NewRouter()

	r.Post(mainURL, app.ShortenLinkHandler)
	r.Get(paramURL, app.GetShortenHandler)
	log.Fatal(http.ListenAndServe(port, r))
}
