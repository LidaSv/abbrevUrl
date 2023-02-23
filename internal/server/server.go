package server

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"abbrevUrl/internal/app"
)

const (
	port = ":8080"
)

func AddServer() {
	r := chi.NewRouter()

	r.Post("/", app.ShortenLinkHandler)
	r.Get("/{id}", app.GetShortenHandler)
	log.Fatal(http.ListenAndServe(port, r))
}
