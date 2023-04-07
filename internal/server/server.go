package server

import (
	"abbrevUrl/internal/app"
	"abbrevUrl/internal/storage"
	"context"
	"errors"
	"github.com/caarlos0/env/v6"
	"github.com/go-chi/chi/v5"
	"github.com/spf13/pflag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"
)

type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS"`
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

	flag := pflag.FlagSet{}
	FlagServerAddress := flag.String("a", "localhost:8080", "a string")
	FlagFileStoragePath := flag.String("f", "/Users/ldsviyazova/Desktop/GitHub/abbrevUrl/internal/storage/cache.log", "a string")
	flag.Parse(os.Args[1:])

	filepath, exist := os.LookupEnv("FILE_STORAGE_PATH")

	var fileName string
	if exist {
		fileName = filepath
	} else {
		fileName = *FlagFileStoragePath
	}

	storage.ReadCache(fileName, st)

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
	} else {
		serv = *FlagServerAddress
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
		storage.WriterCache(fileName, st)
	case <-chErrors:
		_ = server.Shutdown(context.Background())
		storage.WriterCache(fileName, st)
	}
}
