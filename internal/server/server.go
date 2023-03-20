package server

import (
	"abbrevUrl/internal/app"
	"abbrevUrl/internal/storage"
	"context"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"os"
	"os/signal"
)

const (
	port = ":8080"
)

func AddServer() error {
	r := chi.NewRouter()

	st := storage.Iter()
	s := app.HelpHandler(st)

	r.Route("/", func(r chi.Router) {
		r.Post("/", s.ShortenLinkHandler)
		r.Get("/{id}", s.GetShortenHandler)
	})

	var srv http.Server

	chErrors := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		if err := srv.Shutdown(context.Background()); err != nil {
			log.Printf("HTTP server Shutdown: %v", err)
		}
		close(chErrors)
	}()

	err := http.ListenAndServe(port, r)
	if err != http.ErrServerClosed {
		return err
	}

	<-chErrors
	return err
}
