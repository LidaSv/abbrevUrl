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
	"strings"
	"time"
)

type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS" envDefault:"localhost:8080"`
	BaseURL         string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	FileStoragePath string `env:"FILE_STORAGE_PATH" envDefault:"/Users/ldsviyazova/Desktop/GitHub/abbrevUrl/internal/storage/cache.log"`
}

func AddServer() {
	r := chi.NewRouter()

	st := storage.Iter()
	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	if cfg.FileStoragePath != "" {
		fileName := cfg.FileStoragePath
		storage.ReadCache(fileName, st)
	}
	s := app.HelpHandler(st)

	r.Route("/", func(r chi.Router) {
		r.Post("/api/shorten", s.ShortenJSONLinkHandler)
		r.Post("/", s.ShortenLinkHandler)
		r.Get("/{id:[0-9a-z]+}", s.GetShortenHandler)
	})

	replacer := strings.NewReplacer("https://", "", "http://", "")
	ServerAdd := replacer.Replace(cfg.ServerAddress)

	if ServerAdd[len(ServerAdd)-1:] == "/" {
		ServerAdd = ServerAdd[:len(ServerAdd)-1]
	}

	server := http.Server{
		Addr:              ServerAdd,
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
		if cfg.FileStoragePath != "" {
			fileName := cfg.FileStoragePath
			storage.WriterCache(fileName, st)
		}
	case <-chErrors:
		_ = server.Shutdown(context.Background())
		if cfg.FileStoragePath != "" {
			fileName := cfg.FileStoragePath
			storage.WriterCache(fileName, st)
		}
	}
}
