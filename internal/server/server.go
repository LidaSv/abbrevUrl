package server

import (
	"abbrevUrl/internal/app"
	"abbrevUrl/internal/middleware"
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
	FileStoragePath string `env:"FILE_STORAGE_PATH" envDefault:"./tmp/cache"`
	DatabaseDsn     string `env:"DATABASE_DSN" envDefault:"host=localhost port=6422 user=postgres password=123 dbname=postgres"`
	//envDefault:"host=localhost port=6422 user=postgres password=123 dbname=postgres"
}

func AddServer() {
	r := chi.NewRouter()

	r.Use(middleware.CookieHandle, middleware.GzipHandle)

	st := storage.Iter()
	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	FlagServerAddress := flag.String("a", cfg.ServerAddress, "a string")
	FlagBaseURL := flag.String("b", "http://"+cfg.ServerAddress, "a string")
	FlagFileStoragePath := flag.String("f", cfg.FileStoragePath, "a string")
	FlagDatabaseDsn := flag.String("d", cfg.DatabaseDsn, "a string")
	flag.Parse()

	basePath, baseExists := os.LookupEnv("BASE_URL")

	if baseExists {
		st.BaseURL = basePath
	} else {
		st.BaseURL = *FlagBaseURL
	}

	dbPath, dbExists := os.LookupEnv("DATABASE_DSN")

	if dbExists {
		st.DatabaseDsn = dbPath
	} else {
		st.DatabaseDsn = *FlagDatabaseDsn
	}

	filePath, fileExist := os.LookupEnv("FILE_STORAGE_PATH")

	var fileName string
	if fileExist {
		fileName = filePath
	} else {
		fileName = *FlagFileStoragePath
	}

	if st.DatabaseDsn != "" {
		storage.ReadDBCashe(st.DatabaseDsn, st)
	} else {
		storage.ReadCache(fileName, st)
	}

	s := app.HelpHandler(st)

	r.Route("/", func(r chi.Router) {
		r.Post("/api/shorten", s.ShortenJSONLinkHandler)
		r.Post("/", s.ShortenLinkHandler)
		r.Get("/{id:[0-9a-z]+}", s.GetShortenHandler)
		r.Get("/api/user/urls", s.AllJSONGetShortenHandler)
		r.Get("/ping", s.PingPSQL)
	})

	serverPath, serverExists := os.LookupEnv("SERVER_ADDRESS")

	var serv string
	if serverExists {
		serv = serverPath
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
		if st.DatabaseDsn != "" {
			storage.WriteDBCashe(st.DatabaseDsn, st)
		} else {
			storage.WriterCache(fileName, st)
		}
	case <-chErrors:
		_ = server.Shutdown(context.Background())
		if st.DatabaseDsn != "" {
			storage.WriteDBCashe(st.DatabaseDsn, st)
		} else {
			storage.WriterCache(fileName, st)
		}
	}
}
