package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"abbrevUrl/internal/app"
)

const (
	port = ":8080"
)

func AddServer() error {
	r := chi.NewRouter()
	s := app.Server{}

	r.Route("/", func(r chi.Router) {
		r.Post("/", s.ShortenLinkHandler)
		r.Get("/{id}", s.GetShortenHandler)
	})

	err := http.ListenAndServe(port, r)
	return err
}
