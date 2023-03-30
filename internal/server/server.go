package server

import (
	"abbrevUrl/internal/app"
	"abbrevUrl/internal/storage"
	"context"
	"errors"
	"github.com/caarlos0/env/v6"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"
)

type Config struct {
	ServerAddress int    `env:"SERVER_ADDRESS"`
	BaseURL       string `env:"BASE_URL"`
}

func AddServer() {
	r := chi.NewRouter()

	st := storage.Iter()
	s := app.HelpHandler(st)

	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	r.Route("/", func(r chi.Router) {
		r.Post("/api/shorten", s.ShortenJSONLinkHandler)
		r.Post("/", s.ShortenLinkHandler)
		r.Get(cfg.BaseURL, s.GetShortenHandler)
	})

	server := http.Server{
		Addr:              "localhost:" + strconv.Itoa(cfg.ServerAddress),
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
