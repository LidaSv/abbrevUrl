package main

import (
	"abbrevUrl/internal/app"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
)

func main() {
	r := chi.NewRouter()

	r.Post("/", app.ShortenLinkHander)
	r.Get("/{id}", app.GetShortenHandler)
	log.Fatal(http.ListenAndServe(":8080", r))
}
