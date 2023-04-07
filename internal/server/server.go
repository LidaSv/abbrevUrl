package server

import (
	"abbrevUrl/internal/app"
	"abbrevUrl/internal/storage"
	"context"
	"errors"
	"flag"
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
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
}

func AddServer() {
	r := chi.NewRouter()

	st := storage.Iter()
	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	FlagServerAddress := flag.String("a", "localhost:8080", "a string")
	FlagFileStoragePath := flag.String("f", "", "a string")
	flag.Parse()

	var fileName string
	if cfg.FileStoragePath != "" || *FlagFileStoragePath != "" {
		if cfg.FileStoragePath != "" {
			fileName = cfg.FileStoragePath
		} else {
			fileName = *FlagFileStoragePath
		}
		storage.ReadCache(fileName, st)
	}

	s := app.HelpHandler(st)

	r.Route("/", func(r chi.Router) {
		r.Post("/api/shorten", s.ShortenJSONLinkHandler)
		r.Post("/", s.ShortenLinkHandler)
		r.Get("/{id:[0-9a-z]+}", s.GetShortenHandler)
	})

	path, exists := os.LookupEnv("SERVER_ADDRESS")

	var serv string
	if exists {
		serv = path
	} else if *FlagServerAddress != "" {
		serv = *FlagServerAddress
	} else {
		serv = cfg.ServerAddress
	}

	replacer := strings.NewReplacer("https://", "", "http://", "")
	ServerAdd := replacer.Replace(serv)

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
