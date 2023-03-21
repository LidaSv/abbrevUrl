package server

import (
	"abbrevUrl/internal/app"
	"abbrevUrl/internal/storage"
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"net/http"
	"os"
	"os/signal"
	"time"
)

const (
	port = ":8080"
)

func AddServer() {
	r := chi.NewRouter()

	st := storage.Iter()
	s := app.HelpHandler(st)

	r.Route("/", func(r chi.Router) {
		r.Post("/", s.ShortenLinkHandler)
		r.Get("/{id}", s.GetShortenHandler)
	})

	server := http.Server{
		Addr:              "localhost" + port,
		Handler:           r,
		ReadHeaderTimeout: time.Second,
	}

	chErrors := make(chan error)

	go func() {
		err := server.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			chErrors <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	select {
	case <-stop:
		signal.Stop(stop)
		_ = server.Shutdown(context.Background())
	case <-chErrors:
		_ = server.Shutdown(context.Background())

	}

}
