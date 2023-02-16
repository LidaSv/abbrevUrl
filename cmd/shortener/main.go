package main

import (
	"abbrevUrl/internal/app"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/", app.ShortenLinkHander).Methods("POST")
	router.HandleFunc("/{id:[0-9a-z]+}", app.GetShortenHandler).Methods("GET")
	log.Fatal(http.ListenAndServe(":8080", router))
}
